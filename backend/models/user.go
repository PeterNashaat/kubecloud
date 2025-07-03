package models

import "time"

// User represents a user in the system
type User struct {
	ID                int       `gorm:"primaryKey;autoIncrement;column:id"`
	StripeCustomerID  string    `json:"stripe_customer_id"`
	Username          string    `json:"username" binding:"required"`
	Email             string    `json:"email" gorm:"unique" binding:"required"`
	Password          []byte    `json:"password" binding:"required"`
	UpdatedAt         time.Time `json:"updated_at"`
	Verified          bool      `json:"verified"`
	Code              int       `json:"code"`
	Admin             bool      `json:"admin"`
	CreditCardBalance float64   `json:"credit_card_balance" gorm:"default:0"` // money from credit card
	CreditedBalance   float64   `json:"credited_balance" gorm:"default:0"`    // manually added by admin or from vouchers
	Mnemonic          string    `json:"-" gorm:"column:mnemonic"`
	SSHKey            string    `json:"ssh_key"`
	Debt              float64   `json:"debt"`
}

// SSHKey represents an SSH key for a user
type SSHKey struct {
	ID        int       `gorm:"primaryKey;autoIncrement;column:id"`
	UserID    int       `gorm:"user_id" binding:"required"`
	Name      string    `json:"name" binding:"required"`
	PublicKey string    `json:"public_key" binding:"required"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserNodes model holds info of reserved nodes of user
type UserNodes struct {
	ID         int       `gorm:"primaryKey;autoIncrement;column:id"`
	UserID     int       `gorm:"user_id" binding:"required"`
	ContractID uint64    `gorm:"contract_id" binding:"required"`
	NodeID     uint32    `gorm:"node_id" binding:"required"`
	CreatedAt  time.Time `json:"created_at"`
}
