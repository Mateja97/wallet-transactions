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

type TransferMoneyRequest struct {
	FromUserID string  `json:"from_user_id"`
	ToUserID   string  `json:"to_user_id"`
	Amount     float64 `json:"amount_to_transfer"`
}

func (srv *Server) transferMoney(w http.ResponseWriter, r *http.Request) {
	log.Println("transfer money request")
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var request TransferMoneyRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || request.FromUserID == "" || request.ToUserID == "" || request.Amount <= 0 || request.ToUserID == request.FromUserID {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	tx := srv.db.Begin()
	if tx.Error != nil {
		log.Printf("error starting database transaction: %v", tx.Error)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	userToLock := shared.Resources.GetUserLock(request.ToUserID)
	userFromLock := shared.Resources.GetUserLock(request.FromUserID)

	userToLock.Lock()
	userFromLock.Lock()
	defer userToLock.Unlock()
	defer userFromLock.Unlock()

	var fromUser, toUser models.DBUser
	if err := tx.Where("user_id = ?", request.FromUserID).First(&fromUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("error user not found")
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("error fetching user by id: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if err := tx.Where("user_id = ?", request.ToUserID).First(&toUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "To user not found", http.StatusNotFound)
			return
		}
		log.Printf("error fetching to user: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if fromUser.Balance < request.Amount {
		http.Error(w, "Insufficient balance", http.StatusBadRequest)
		return
	}

	oldFromUserBalance := fromUser.Balance
	fromUser.Balance -= request.Amount
	oldToUserBalance := toUser.Balance
	toUser.Balance += request.Amount

	if err := tx.Save(fromUser).Error; err != nil {
		log.Printf("error updating from user balance: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	if err := tx.Save(toUser).Error; err != nil {
		log.Printf("error updating to user balance: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	fromUserTransaction := models.DBTransaction{
		ID:            uuid.New().String(),
		UserID:        request.FromUserID,
		BalanceChange: -request.Amount,
		OldBalance:    oldFromUserBalance,
		NewBalance:    fromUser.Balance,
	}

	if err := tx.Create(&fromUserTransaction).Error; err != nil {
		log.Printf("error recording from user transaction: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	toUserTransaction := models.DBTransaction{
		ID:            uuid.New().String(),
		UserID:        request.ToUserID,
		BalanceChange: request.Amount,
		OldBalance:    oldToUserBalance,
		NewBalance:    toUser.Balance,
	}

	if err := tx.Create(&toUserTransaction).Error; err != nil {
		log.Printf("error recording transaction: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("error committing transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
