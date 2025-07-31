package models

// DB interface for databases
type DB interface {
	RegisterUser(user *User) error
	GetUserByEmail(email string) (User, error)
	GetUserByID(userID int) (User, error)
	UpdateUserByID(user *User) error
	UpdatePassword(email string, hashedPassword []byte) error
	UpdateUserVerification(userID int, verified bool) error
	ListAllUsers() ([]User, error)
	ListAdmins() ([]User, error)
	DeleteUserByID(userID int) error
	CreateVoucher(voucher *Voucher) error
	ListAllVouchers() ([]Voucher, error)
	GetVoucherByCode(code string) (Voucher, error)
	RedeemVoucher(code string) error
	CreateTransaction(transaction *Transaction) error
	CreditUserBalance(userID int, amount uint64) error
	CreateInvoice(invoice *Invoice) error
	GetInvoice(id int) (Invoice, error)
	ListUserInvoices(userID int) ([]Invoice, error)
	ListInvoices() ([]Invoice, error)
	UpdateInvoicePDF(id int, data []byte) error
	CreateUserNode(userNode *UserNodes) error
	ListUserNodes(userID int) ([]UserNodes, error)
	// SSH Key methods
	CreateSSHKey(sshKey *SSHKey) error
	ListUserSSHKeys(userID int) ([]SSHKey, error)
	DeleteSSHKey(sshKeyID int, userID int) error
	GetSSHKeyByID(sshKeyID int, userID int) (SSHKey, error)
	// Notification methods
	CreateNotification(notification *Notification) error
	GetUserNotifications(userID string, limit, offset int) ([]Notification, error)
	MarkNotificationAsRead(notificationID uint, userID string) error
	MarkAllNotificationsAsRead(userID string) error
	GetUnreadNotificationCount(userID string) (int64, error)
	DeleteNotification(notificationID uint, userID string) error
	// Cluster methods
	CreateCluster(userID string, cluster *Cluster) error
	ListUserClusters(userID string) ([]Cluster, error)
	GetClusterByName(userID string, projectName string) (Cluster, error)
	UpdateCluster(cluster *Cluster) error
	DeleteCluster(userID string, projectName string) error
	// pending records methods
	CreatePendingRecord(record *PendingRecord) error
	ListAllPendingRecords() ([]PendingRecord, error)
	ListOnlyPendingRecords() ([]PendingRecord, error)
	ListUserPendingRecords(userID int) ([]PendingRecord, error)
	UpdatePendingRecordTransferredAmount(id int, amount uint64) error
}
