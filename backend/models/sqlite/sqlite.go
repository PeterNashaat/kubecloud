package sqlite

import (
	"fmt"
	"kubecloud/models"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Sqlite struct implements db interface with sqlite
type Sqlite struct {
	db *gorm.DB
}

// NewSqliteStorage connects to the database file
func NewSqliteStorage(file string) (*Sqlite, error) {
	db, err := gorm.Open(sqlite.Open(file), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate models
	err = db.AutoMigrate(&models.User{}, &models.Voucher{}, models.Transaction{}, models.Invoice{}, models.NodeItem{})
	if err != nil {
		return nil, err
	}

	return &Sqlite{db: db}, nil
}

// Close closes the database connection
func (s *Sqlite) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// RegisterUser registers a new user to the system
func (s *Sqlite) RegisterUser(user *models.User) error {
	return s.db.Create(user).Error
}

// GetUserByEmail returns user by its email if found
func (s *Sqlite) GetUserByEmail(email string) (models.User, error) {
	var user models.User
	query := s.db.First(&user, "email = ?", email)
	return user, query.Error
}

// GetUserByEmail returns user by its email if found
func (s *Sqlite) GetUserByID(userID int) (models.User, error) {
	var user models.User
	query := s.db.First(&user, "id = ?", userID)
	return user, query.Error
}

// UpdateUserByID updates user data by its ID
func (s *Sqlite) UpdateUserByID(user *models.User) error {
	user.UpdatedAt = time.Now()
	return s.db.Model(&models.User{}).
		Where("id = ?", user.ID).
		Updates(user).Error
}

// UpdatePassword updates password of user by its email
func (s *Sqlite) UpdatePassword(email string, hashedPassword []byte) error {
	result := s.db.Model(&models.User{}).
		Where("email = ?", email).
		Updates(map[string]interface{}{
			"password":   hashedPassword,
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no user found with email %s", email)
	}

	return nil
}

// UpdateUserVerification updates verified status of user by its ID
func (s *Sqlite) UpdateUserVerification(userID int, verified bool) error {
	result := s.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("verified", verified)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no user found with ID %d", userID)
	}

	return nil
}

// ListAllUsers lists all users in system
func (s *Sqlite) ListAllUsers() ([]models.User, error) {
	var users []models.User

	err := s.db.Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil

}

// DeleteUserByID deletes user by its ID
func (s *Sqlite) DeleteUserByID(userID int) error {
	return s.db.Where("id = ?", userID).Delete(&models.User{}).Error
}

// CreateVoucher creates new voucher in system
func (s *Sqlite) CreateVoucher(voucher *models.Voucher) error {
	return s.db.Create(voucher).Error
}

// ListAllVouchers gets all vouchers in system
func (s *Sqlite) ListAllVouchers() ([]models.Voucher, error) {
	var vouchers []models.Voucher

	err := s.db.Find(&vouchers).Error
	if err != nil {
		return nil, err
	}
	return vouchers, nil
}

// GetVoucherByCode returns voucher by its code
func (s *Sqlite) GetVoucherByCode(code string) (models.Voucher, error) {
	var voucher models.Voucher
	query := s.db.First(&voucher, "code = ?", code)
	return voucher, query.Error
}

// RedeemVoucher updates status if voucher
func (s *Sqlite) RedeemVoucher(code string) error {
	result := s.db.Model(&models.Voucher{}).
		Where("code = ?", code).
		Update("redeemed", true)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no voucher found with Code %d", code)
	}

	return nil
}

// CreateTransaction creates a payment transaction
func (s *Sqlite) CreateTransaction(transaction *models.Transaction) error {
	return s.db.Create(transaction).Error
}

// CreditUserBalance add credited balance to user by its ID
func (s *Sqlite) CreditUserBalance(userID int, amount uint64) error {
	return s.db.Model(&models.User{}).
		Where("id = ?", userID).
		UpdateColumn("credited_balance", gorm.Expr("credited_balance + ?", amount)).
		Error
}

// CreateInvoice creates new invoice
func (s *Sqlite) CreateInvoice(invoice *models.Invoice) error {
	return s.db.Create(&invoice).Error
}

// GetInvoice returns an invoice by ID
func (s *Sqlite) GetInvoice(id int) (models.Invoice, error) {
	var invoice models.Invoice
	return invoice, s.db.First(&invoice, id).Error
}

// ListUserInvoices returns all invoices of user
func (s *Sqlite) ListUserInvoices(userID int) ([]models.Invoice, error) {
	var invoices []models.Invoice
	return invoices, s.db.Where("user_id = ?", userID).Find(&invoices).Error
}

// ListInvoices returns all invoices (admin)
func (s *Sqlite) ListInvoices() ([]models.Invoice, error) {
	var invoices []models.Invoice
	return invoices, s.db.Find(&invoices).Error
}

func (s *Sqlite) UpdateInvoicePDF(id int, data []byte) error {
	return s.db.Model(&models.Invoice{}).Where("id = ?", id).Updates(map[string]interface{}{"file_data": data}).Error
}

// CreateUserNode creates new node record for user
func (s *Sqlite) CreateUserNode(userNode *models.UserNodes) error {
	return s.db.Create(&userNode).Error
}

// ListUserNodes returns all nodes records for user by its ID
func (s *Sqlite) ListUserNodes(userID int) ([]models.UserNodes, error) {
	var userNodes []models.UserNodes
	return userNodes, s.db.Where("user_id = ?", userID).Find(&userNodes).Error
}
