package models

import "time"

// Voucher struct holds all data for vouchers, voucher used only by one user.
type Voucher struct {
	ID        int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Code      string    `json:"code" gorm:"unique; not null"`
	Value     float64   `json:"value" gorm:"not null"`
	Redeemed  bool      `json:"redeemed" gorm:"default:false"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}
