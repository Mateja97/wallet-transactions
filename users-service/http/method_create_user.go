package http

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"log"
	"net/http"
	"time"
	"user/models"
)

type CreateUserRequest struct {
	Email string `json:"email"`
}

type CreateUserResponse struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (srv *Server) createUser(w http.ResponseWriter, r *http.Request) {
	log.Println("create user post request")
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var createUserRequest CreateUserRequest

	err := json.NewDecoder(r.Body).Decode(&createUserRequest)
	if err != nil || createUserRequest.Email == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	user := models.DBUser{
		ID:        uuid.NewString(),
		Email:     createUserRequest.Email,
		CreatedAt: time.Now(),
	}

	tx := srv.db.Begin()
	if tx.Error != nil {
		log.Printf("error starting database transaction: %v", tx.Error)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered in createUser handler:", r)
			tx.Rollback()
		}
	}()

	if err := tx.Create(&user).Error; err != nil {
		log.Printf("error db create: %v", err)
		tx.Rollback()
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	msg, err := json.Marshal(user)
	if err != nil {
		log.Printf("error json marshall: %v", err)
		tx.Rollback()
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = srv.kafkaWriter.WriteMessages(r.Context(), kafka.Message{
		Key:   []byte(user.ID),
		Value: msg,
	})
	if err != nil {
		log.Printf("error kafka write: %v", err.Error())
		tx.Rollback()
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("error committing transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(CreateUserResponse{
		UserID:    user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
	if err != nil {
		log.Printf("error encoding balance response: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}
