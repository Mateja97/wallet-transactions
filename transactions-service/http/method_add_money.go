package http

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"log"
	"net/http"
	"transactions/models"
	"transactions/shared"
)

type AddMoneyRequest struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}

type AddMoneyResponse struct {
	Balance float64 `json:"updated_balance"`
}

func (srv *Server) addMoney(w http.ResponseWriter, r *http.Request) {
	log.Println("add money request")
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var request AddMoneyRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if _, err = uuid.Parse(request.UserID); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	tx := srv.db.Begin()
	if tx.Error != nil {
		log.Printf("error starting database transaction: %v", tx.Error)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	userLock := shared.Resources.GetUserLock(request.UserID)
	userLock.Lock()
	defer userLock.Unlock()

	var user models.DBUser
	if err := tx.Where("user_id = ?", request.UserID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("error user not found")
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("error fetching user by id: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	oldBalance := user.Balance
	user.Balance += request.Amount
	if err := tx.Save(&user).Error; err != nil {
		log.Printf("error updating user balance: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	transaction := models.DBTransaction{
		ID:            uuid.New().String(),
		UserID:        request.UserID,
		BalanceChange: request.Amount,
		OldBalance:    oldBalance,
		NewBalance:    user.Balance,
	}
	if err := tx.Create(&transaction).Error; err != nil {
		log.Printf("error recording transaction: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("error committing transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := AddMoneyResponse{
		Balance: user.Balance,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("error encoding credit response: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}

}
