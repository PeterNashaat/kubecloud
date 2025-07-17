package models

import "time"

type PendingRecord struct {
	ID                   int       `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID               int       `json:"user_id" gorm:"not null"`
	TFTAmount            uint64    `json:"amount" gorm:"not null"`
	TransferredTFTAmount uint64    `json:"transferred_amount" gorm:"not null"`
	CreatedAt            time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt            time.Time `json:"updated_at" gorm:"not null"`
}
