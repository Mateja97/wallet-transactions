package models

import "time"

type DBUser struct {
	ID        string `gorm:"PRIMARY_KEY"`
	Email     string `gorm:"UNIQUE_INDEX"`
	CreatedAt time.Time
}

func (DBUser) TableName() string {
	return "users"
}
