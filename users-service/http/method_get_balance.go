package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	"time"
	"user/models"
)

type BalanceResponse struct {
	Email   string `json:"email"`
	Balance string `json:"balance"`
}

type BalanceRequest struct {
	Email string `json:"email"`
}

type NatsBalanceRequest struct {
	ID        string `json:"id"`
	RequestID string `json:"request_id"`
}

type NatsBalanceResponse struct {
	RequestID string  `json:"request_id"`
	Balance   float64 `json:"balance"`
}

func (srv *Server) getBalance(w http.ResponseWriter, r *http.Request) {
	requestId := uuid.New().String()
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	balanceRequest := BalanceRequest{}
	err := json.NewDecoder(r.Body).Decode(&balanceRequest)
	if err != nil || balanceRequest.Email == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	log.Println("get balance request for email", balanceRequest.Email)

	var user models.DBUser
	if err := srv.db.Where("email = ?", balanceRequest.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("error user not found")
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("error fetching user by email: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	natsRequest, err := json.Marshal(NatsBalanceRequest{
		RequestID: requestId,
		ID:        user.ID,
	})
	if err != nil {
		log.Printf("error marshall nats balance request: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}

	msg, err := srv.natsConn.Request("balance-request", natsRequest, 10*time.Second)
	if err != nil {
		log.Printf("error sending nats request: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	natsResponse := &NatsBalanceResponse{}
	err = json.Unmarshal(msg.Data, natsResponse)
	if err != nil {
		log.Printf("error unmarshalling nats balance response: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(BalanceResponse{
		Email:   user.Email,
		Balance: fmt.Sprintf("%.3f", natsResponse.Balance),
	})
	if err != nil {
		log.Printf("error encoding balance response: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}
