package models

import "time"

type KafkaUser struct {
	ID        string `gorm:"primary_key"`
	Email     string
	Balance   string
	CreatedAt time.Time
}

type DBUser struct {
	UserID    string `gorm:"primary_key"`
	Balance   float64
	CreatedAt time.Time
}

func (DBUser) TableName() string {
	return "users"
}
