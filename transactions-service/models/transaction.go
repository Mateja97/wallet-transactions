package models

import "time"

type DBTransaction struct {
	ID            string `gorm:"primary_key"`
	UserID        string
	BalanceChange float64
	OldBalance    float64
	NewBalance    float64

	CreatedAt time.Time
}

func (DBTransaction) TableName() string {
	return "transactions"
}
