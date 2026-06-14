package models

import (
	"time"

	"gorm.io/gorm"
)

// ==========================================
// ORGANIZATION
// ==========================================

type SamitiSetting struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	Name            string     `gorm:"size:255;not null" json:"name"`
	RegistrationNo  *string    `gorm:"size:100" json:"registration_no"`
	Address         string     `gorm:"size:500;not null" json:"address"`
	WardNo          int        `gorm:"not null" json:"ward_no"`
	Municipality    string     `gorm:"size:255;not null" json:"municipality"`
	District        string     `gorm:"size:255;not null" json:"district"`
	Province        string     `gorm:"size:255;not null" json:"province"`
	ContactPhone    *string    `gorm:"size:20" json:"contact_phone"`
	ContactEmail    *string    `gorm:"size:255" json:"contact_email"`
	Description     *string    `gorm:"type:text" json:"description"`
	Logo            *string    `gorm:"size:500" json:"logo"`
	MapImage        *string    `gorm:"size:500" json:"map_image"`
	Latitude        *float64   `json:"latitude"`
	Longitude       *float64   `json:"longitude"`
	EstablishedDate *time.Time `json:"established_date"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type SamitiHead struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      *uint      `gorm:"uniqueIndex" json:"user_id"`
	Name        string     `gorm:"size:255;not null" json:"name"`
	Post        string     `gorm:"size:50;not null" json:"post"` // chairperson, secretary, treasurer, member
	Phone       *string    `gorm:"size:20" json:"phone"`
	Email       *string    `gorm:"size:255" json:"email"`
	Address     *string    `gorm:"size:500" json:"address"`
	Photo       *string    `gorm:"size:500" json:"photo"`
	TenureStart *time.Time `json:"tenure_start"`
	TenureEnd   *time.Time `json:"tenure_end"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	Remarks     *string    `gorm:"type:text" json:"remarks"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// Optional login account for committee members who work as admin/staff.
	User                      *User `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"user,omitempty"`
	BootstrapAdminDeactivated bool  `gorm:"-" json:"bootstrap_admin_deactivated,omitempty"`
	ExistingAccountPromoted   bool  `gorm:"-" json:"existing_account_promoted,omitempty"`
}

// ==========================================
// USERS & AUTH
// ==========================================

type User struct {
	ID                  uint       `gorm:"primaryKey" json:"id"`
	Name                string     `gorm:"size:255;not null" json:"name"`
	Email               *string    `gorm:"size:255;uniqueIndex" json:"email"`
	Phone               string     `gorm:"size:20;uniqueIndex;not null" json:"phone"`
	Password            string     `gorm:"size:255;not null" json:"-"`           // json:"-" hides password from API responses
	Role                string     `gorm:"size:20;default:member" json:"role"`   // admin, staff, member
	Status              string     `gorm:"size:20;default:active" json:"status"` // active, inactive
	IsBootstrapAdmin    bool       `gorm:"default:false;index" json:"is_bootstrap_admin"`
	MustChangePassword  bool       `gorm:"default:false" json:"must_change_password"`
	FailedLoginAttempts int        `gorm:"default:0" json:"-"`
	LastFailedLoginAt   *time.Time `json:"-"`
	LockedUntil         *time.Time `gorm:"index" json:"locked_until,omitempty"`
	LastLoginAt         *time.Time `json:"last_login_at,omitempty"`
	PasswordChangedAt   *time.Time `json:"password_changed_at,omitempty"`
	MFAEnabled          bool       `gorm:"default:false" json:"mfa_enabled"`
	MFASecretEncrypted  *string    `gorm:"type:text" json:"-"`
	MFABackupCodes      *string    `gorm:"type:jsonb" json:"-"`
	MFALastUsedStep     int64      `gorm:"default:0" json:"-"`
	EmailVerifiedAt     *time.Time `json:"email_verified_at"`
	PhoneVerifiedAt     *time.Time `json:"phone_verified_at"`
	RememberToken       *string    `gorm:"size:255" json:"-"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`

	// Relations
	Member *Member `gorm:"foreignKey:UserID" json:"member,omitempty"`
}

// BeforeCreate hook — hash password before saving
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// Password will be hashed in the service layer before calling Create
	return nil
}

// UserSession stores only SHA-256 hashes of opaque browser credentials. The
// raw session and CSRF tokens exist only in secure cookies on the client.
type UserSession struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     uint       `gorm:"not null;index" json:"user_id"`
	TokenHash  string     `gorm:"size:64;not null;uniqueIndex" json:"-"`
	CSRFHash   string     `gorm:"size:64;not null" json:"-"`
	IPAddress  *string    `gorm:"size:64" json:"ip_address,omitempty"`
	UserAgent  *string    `gorm:"size:500" json:"user_agent,omitempty"`
	ExpiresAt  time.Time  `gorm:"not null;index" json:"expires_at"`
	LastSeenAt time.Time  `gorm:"not null;index" json:"last_seen_at"`
	RevokedAt  *time.Time `gorm:"index" json:"revoked_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	User *User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// ==========================================
// MEMBERS
// ==========================================

type Member struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	UserID         *uint          `gorm:"uniqueIndex" json:"user_id"` // FK → users.id
	MembershipNo   string         `gorm:"size:50;uniqueIndex;not null" json:"membership_no"`
	Name           string         `gorm:"size:255;not null" json:"name"`
	AssistantName  string         `gorm:"size:255;not null" json:"assistant_name"`
	FatherName     string         `gorm:"size:255;not null" json:"father_name"`
	WardNo         int            `gorm:"not null" json:"ward_no"`
	Tole           string         `gorm:"size:255;not null" json:"tole"`
	Phone          *string        `gorm:"size:20" json:"phone"`
	Photo          *string        `gorm:"size:500" json:"photo"`
	AssistantPhoto *string        `gorm:"size:500" json:"assistant_photo"`
	JoinedDate     *time.Time     `json:"joined_date"`
	Status         string         `gorm:"size:20;default:active" json:"status"`
	Remarks        *string        `gorm:"type:text" json:"remarks"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	User          *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	FamilyMembers []FamilyMember `gorm:"foreignKey:MemberID" json:"family_members,omitempty"`
	Requests      []Request      `gorm:"foreignKey:MemberID" json:"requests,omitempty"`
	Payments      []Payment      `gorm:"foreignKey:MemberID" json:"payments,omitempty"`
	Transactions  []Transaction  `gorm:"foreignKey:MemberID" json:"transactions,omitempty"`
	Fines         []Fine         `gorm:"foreignKey:MemberID" json:"fines,omitempty"`
}

type FamilyMember struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	MemberID      uint      `gorm:"not null;index" json:"member_id"`
	Name          string    `gorm:"size:255;not null" json:"name"`
	Relation      string    `gorm:"size:100;not null" json:"relation"` // father, mother, spouse, son, daughter, etc.
	Age           *int      `json:"age"`
	Gender        *string   `gorm:"size:10" json:"gender"`
	CitizenshipNo *string   `gorm:"size:50" json:"citizenship_no"`
	Remarks       *string   `gorm:"type:text" json:"remarks"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ==========================================
// FISCAL YEAR & FEES
// ==========================================

type FiscalYear struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:20;uniqueIndex;not null" json:"name"` // e.g. "2080/81"
	StartDate time.Time `gorm:"not null" json:"start_date"`
	EndDate   time.Time `gorm:"not null" json:"end_date"`
	IsActive  bool      `gorm:"default:false" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`

	// Populated only when a fiscal year is activated.
	CarriedStockItems  int `gorm:"-" json:"carried_stock_items,omitempty"`
	CarriedRateItems   int `gorm:"-" json:"carried_rate_items,omitempty"`
	CarriedFeeItems    int `gorm:"-" json:"carried_fee_items,omitempty"`
	AssignedMemberFees int `gorm:"-" json:"assigned_member_fees,omitempty"`

	// Relations
	FeeSettings   []FeeSetting   `gorm:"foreignKey:FiscalYearID" json:"fee_settings,omitempty"`
	Stocks        []Stock        `gorm:"foreignKey:FiscalYearID" json:"stocks,omitempty"`
	ResourceRates []ResourceRate `gorm:"foreignKey:FiscalYearID" json:"resource_rates,omitempty"`
}

type FeeSetting struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	FiscalYearID  uint      `gorm:"not null;uniqueIndex:idx_fee_fiscal_year" json:"fiscal_year_id"`
	MembershipFee float64   `gorm:"type:numeric(14,2);not null" json:"membership_fee"`
	CreatedAt     time.Time `json:"created_at"`

	FiscalYear *FiscalYear `gorm:"foreignKey:FiscalYearID" json:"fiscal_year,omitempty"`
}

// ==========================================
// RESOURCES
// ==========================================

type ResourceType struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `gorm:"size:100;uniqueIndex;not null" json:"name"` // timber, firewood, grass
	Unit string `gorm:"size:20;not null" json:"unit"`              // cft, bundle, kg

	// Relations
	Items []ResourceItem `gorm:"foreignKey:ResourceTypeID" json:"items,omitempty"`
}

type ResourceItem struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	ResourceTypeID uint   `gorm:"not null;index" json:"resource_type_id"`
	Name           string `gorm:"size:255;not null" json:"name"` // e.g. "Sal Wood", "Pine Timber"

	// Relations
	Type   *ResourceType  `gorm:"foreignKey:ResourceTypeID" json:"type,omitempty"`
	Rates  []ResourceRate `gorm:"foreignKey:ResourceItemID" json:"rates,omitempty"`
	Stocks []Stock        `gorm:"foreignKey:ResourceItemID" json:"stocks,omitempty"`
}

type ResourceRate struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ResourceItemID uint      `gorm:"not null;index;uniqueIndex:idx_rate_item_fy" json:"resource_item_id"`
	FiscalYearID   uint      `gorm:"not null;index;uniqueIndex:idx_rate_item_fy" json:"fiscal_year_id"`
	RatePerUnit    float64   `gorm:"type:numeric(14,2);not null" json:"rate_per_unit"`
	CreatedAt      time.Time `json:"created_at"`

	// Relations
	Item       *ResourceItem `gorm:"foreignKey:ResourceItemID" json:"item,omitempty"`
	FiscalYear *FiscalYear   `gorm:"foreignKey:FiscalYearID" json:"fiscal_year,omitempty"`
}

type Stock struct {
	ID                uint    `gorm:"primaryKey" json:"id"`
	ResourceItemID    uint    `gorm:"not null;index;uniqueIndex:idx_stock_item_fy" json:"resource_item_id"`
	FiscalYearID      uint    `gorm:"not null;index;uniqueIndex:idx_stock_item_fy" json:"fiscal_year_id"`
	TotalQuantity     float64 `gorm:"not null" json:"total_quantity"`
	RemainingQuantity float64 `gorm:"not null" json:"remaining_quantity"`
	ReservedQuantity  float64 `gorm:"not null;default:0" json:"reserved_quantity"`

	// Relations
	Item       *ResourceItem `gorm:"foreignKey:ResourceItemID" json:"item,omitempty"`
	FiscalYear *FiscalYear   `gorm:"foreignKey:FiscalYearID" json:"fiscal_year,omitempty"`
}

// ==========================================
// REQUESTS (WORKFLOW)
// ==========================================

type Request struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	MemberID          uint       `gorm:"not null;index" json:"member_id"`
	FiscalYearID      uint       `gorm:"not null;index" json:"fiscal_year_id"`
	ResourceItemID    uint       `gorm:"not null;index" json:"resource_item_id"`
	QuantityRequested float64    `gorm:"not null" json:"quantity_requested"`
	QuantityApproved  *float64   `json:"quantity_approved"`
	RatePerUnit       *float64   `json:"rate_per_unit"`
	TotalAmount       *float64   `json:"total_amount"`
	Status            string     `gorm:"size:20;default:pending" json:"status"` // pending, approved, rejected, completed
	RequestedAt       time.Time  `gorm:"autoCreateTime" json:"requested_at"`
	ApprovedBy        *uint      `json:"approved_by"`
	ApprovedAt        *time.Time `json:"approved_at"`
	Remarks           *string    `gorm:"type:text" json:"remarks"`
	CreatedAt         time.Time  `json:"created_at"`

	// Relations
	Member       *Member       `gorm:"foreignKey:MemberID" json:"member,omitempty"`
	FiscalYear   *FiscalYear   `gorm:"foreignKey:FiscalYearID" json:"fiscal_year,omitempty"`
	ResourceItem *ResourceItem `gorm:"foreignKey:ResourceItemID" json:"resource_item,omitempty"`
	Approver     *User         `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
	Payments     []Payment     `gorm:"foreignKey:RequestID" json:"payments,omitempty"`
}

// ==========================================
// PAYMENTS
// ==========================================

type Payment struct {
	ID                     uint       `gorm:"primaryKey" json:"id"`
	MemberID               uint       `gorm:"not null;index" json:"member_id"`
	RequestID              *uint      `gorm:"index" json:"request_id"`
	LedgerTransactionID    *uint      `gorm:"index" json:"ledger_transaction_id"`
	Amount                 float64    `gorm:"type:numeric(14,2);not null" json:"amount"`
	PaymentMethod          string     `gorm:"size:20;not null;index" json:"payment_method"` // cash, esewa
	TransactionID          *string    `gorm:"size:255" json:"transaction_id"`               // legacy/external reference
	GatewayTransactionUUID *string    `gorm:"size:100;uniqueIndex" json:"gateway_transaction_uuid"`
	GatewayReferenceID     *string    `gorm:"size:100;index" json:"gateway_reference_id"`
	GatewayStatus          *string    `gorm:"size:30" json:"gateway_status"`
	GatewayResponse        *string    `gorm:"type:text" json:"-"`
	ReceiptNo              *string    `gorm:"size:100;uniqueIndex" json:"receipt_no"`
	Remarks                *string    `gorm:"type:text" json:"remarks"`
	Status                 string     `gorm:"size:20;default:pending;index" json:"status"` // pending, paid, failed, canceled
	CreatedBy              *uint      `gorm:"index" json:"created_by"`
	PaidAt                 *time.Time `json:"paid_at"`
	VerifiedAt             *time.Time `json:"verified_at"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`

	// Relations
	Member            *Member      `gorm:"foreignKey:MemberID" json:"member,omitempty"`
	Request           *Request     `gorm:"foreignKey:RequestID" json:"request,omitempty"`
	LedgerTransaction *Transaction `gorm:"foreignKey:LedgerTransactionID" json:"ledger_transaction,omitempty"`
	Creator           *User        `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// ==========================================
// TRANSACTIONS (LEDGER — CORE SYSTEM)
// ==========================================

type Transaction struct {
	ID                uint         `gorm:"primaryKey" json:"id"`
	MemberID          uint         `gorm:"not null;index" json:"member_id"`
	FiscalYearID      uint         `gorm:"not null;index" json:"fiscal_year_id"`
	ResourceItemID    *uint        `gorm:"index" json:"resource_item_id"`
	RequestID         *uint        `gorm:"index" json:"request_id"`
	Type              string       `gorm:"size:50;not null;index" json:"type"`
	Source            string       `gorm:"size:30;default:system;index" json:"source"`
	RecordStatus      string       `gorm:"size:20;default:verified;index" json:"record_status"` // draft, verified, reversed
	Quantity          *float64     `json:"quantity"`
	RatePerUnit       *float64     `json:"rate_per_unit"`
	TotalAmount       float64      `gorm:"type:numeric(14,2);not null" json:"total_amount"`
	AmountPaid        float64      `gorm:"type:numeric(14,2);default:0" json:"amount_paid"`
	AmountRemaining   float64      `gorm:"type:numeric(14,2);not null" json:"amount_remaining"`
	ReceiptNo         string       `gorm:"size:100;uniqueIndex;not null" json:"receipt_no"`
	PhysicalReference *string      `gorm:"size:255" json:"physical_reference"`
	Date              time.Time    `gorm:"not null" json:"date"`
	Remarks           *string      `gorm:"type:text" json:"remarks"`
	EnteredBy         *uint        `gorm:"index" json:"entered_by"`
	VerifiedBy        *uint        `gorm:"index" json:"verified_by"`
	VerifiedAt        *time.Time   `json:"verified_at"`
	ReversedBy        *uint        `gorm:"index" json:"reversed_by"`
	ReversedAt        *time.Time   `json:"reversed_at"`
	ReversalReason    *string      `gorm:"type:text" json:"reversal_reason"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
	Documents         []FileUpload `gorm:"-" json:"documents,omitempty"`

	// Relations
	Member         *Member       `gorm:"foreignKey:MemberID" json:"member,omitempty"`
	FiscalYear     *FiscalYear   `gorm:"foreignKey:FiscalYearID" json:"fiscal_year,omitempty"`
	ResourceItem   *ResourceItem `gorm:"foreignKey:ResourceItemID" json:"resource_item,omitempty"`
	Request        *Request      `gorm:"foreignKey:RequestID" json:"request,omitempty"`
	EnteredByUser  *User         `gorm:"foreignKey:EnteredBy" json:"entered_by_user,omitempty"`
	VerifiedByUser *User         `gorm:"foreignKey:VerifiedBy" json:"verified_by_user,omitempty"`
}

// ==========================================
// EXPENSES
// ==========================================

type ExpenseCategory struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:255;not null;uniqueIndex" json:"name"`
	Description *string   `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`

	// Relations
	Expenses []Expense `gorm:"foreignKey:CategoryID" json:"expenses,omitempty"`
}

type Expense struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	FiscalYearID  uint           `gorm:"not null;index" json:"fiscal_year_id"`
	CategoryID    uint           `gorm:"not null;index" json:"category_id"`
	Title         string         `gorm:"size:255;not null" json:"title"`
	Amount        float64        `gorm:"type:numeric(14,2);not null" json:"amount"`
	ExpenseDate   time.Time      `gorm:"not null" json:"expense_date"`
	PaymentMethod string         `gorm:"size:20;not null" json:"payment_method"` // cash, bank, online
	PaidTo        string         `gorm:"size:255;not null" json:"paid_to"`
	ReceiptNo     *string        `gorm:"size:100" json:"receipt_no"`
	BillPhoto     *string        `gorm:"size:500" json:"bill_photo"`
	Remarks       *string        `gorm:"type:text" json:"remarks"`
	CreatedBy     uint           `gorm:"not null" json:"created_by"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	FiscalYear *FiscalYear      `gorm:"foreignKey:FiscalYearID" json:"fiscal_year,omitempty"`
	Category   *ExpenseCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Creator    *User            `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// ==========================================
// FINES
// ==========================================

type Fine struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	FiscalYearID     uint           `gorm:"not null;index" json:"fiscal_year_id"`
	MemberID         *uint          `json:"member_id"`
	Name             string         `gorm:"size:255" json:"name"` // for non-member violators
	ViolationType    string         `gorm:"size:255;not null" json:"violation_type"`
	Description      *string        `gorm:"type:text" json:"description"`
	FineAmount       float64        `gorm:"type:numeric(14,2);not null" json:"fine_amount"`
	IncidentDate     time.Time      `gorm:"not null" json:"incident_date"`
	Status           string         `gorm:"size:20;default:pending" json:"status"` // pending, paid, waived
	PaymentReference *string        `gorm:"size:255" json:"payment_reference"`
	Photo            *string        `gorm:"size:500" json:"photo"`
	Remarks          *string        `gorm:"type:text" json:"remarks"`
	CreatedBy        uint           `gorm:"not null" json:"created_by"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	FiscalYear *FiscalYear `gorm:"foreignKey:FiscalYearID" json:"fiscal_year,omitempty"`
	Member     *Member     `gorm:"foreignKey:MemberID" json:"member,omitempty"`
	Creator    *User       `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// ==========================================
// LETTERS
// ==========================================

type Letter struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Type         string     `gorm:"size:20;not null" json:"type"` // incoming, outgoing
	ReferenceNo  *string    `gorm:"size:100" json:"reference_no"`
	Title        string     `gorm:"size:255;not null" json:"title"`
	Subject      string     `gorm:"size:500;not null" json:"subject"`
	FromParty    *string    `gorm:"size:255" json:"from_party"`
	ToParty      *string    `gorm:"size:255" json:"to_party"`
	LetterDate   time.Time  `gorm:"not null" json:"letter_date"`
	ReceivedDate *time.Time `json:"received_date"`
	SentDate     *time.Time `json:"sent_date"`
	DocumentFile *string    `gorm:"size:500" json:"document_file"`
	Remarks      *string    `gorm:"type:text" json:"remarks"`
	Status       string     `gorm:"size:20;not null;default:active;index" json:"status"`
	CreatedBy    uint       `gorm:"not null" json:"created_by"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	// Relations
	Creator *User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
}

// ==========================================
// AUDIT LOG
// ==========================================

type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    *uint     `gorm:"index" json:"user_id"`
	Action    string    `gorm:"size:50;not null;index" json:"action"`  // create, update, delete, login, approve, reject
	Entity    string    `gorm:"size:100;not null;index" json:"entity"` // member, request, payment, expense, fine, etc.
	EntityID  *uint     `gorm:"index" json:"entity_id"`
	OldValues *string   `gorm:"type:jsonb" json:"old_values"` // JSON snapshot before change
	NewValues *string   `gorm:"type:jsonb" json:"new_values"` // JSON snapshot after change
	IPAddress *string   `gorm:"size:50" json:"ip_address"`
	UserAgent *string   `gorm:"size:500" json:"user_agent"`
	Remarks   *string   `gorm:"type:text" json:"remarks"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ==========================================
// NOTIFICATIONS
// ==========================================

type Notification struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	UserID     *uint      `gorm:"index" json:"user_id"`       // null = broadcast to all
	TargetRole *string    `gorm:"size:20" json:"target_role"` // admin, staff, member — for role-based
	Title      string     `gorm:"size:255;not null" json:"title"`
	Message    string     `gorm:"type:text;not null" json:"message"`
	Type       string     `gorm:"size:50;not null;index" json:"type"` // info, warning, success, payment, request, system
	Entity     *string    `gorm:"size:100" json:"entity"`             // member, request, payment, etc.
	EntityID   *uint      `json:"entity_id"`
	IsRead     bool       `gorm:"default:false;index" json:"is_read"`
	ReadAt     *time.Time `json:"read_at"`
	CreatedAt  time.Time  `gorm:"index" json:"created_at"`

	// Relations
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// NotificationReceipt isolates read state per user for broadcast and
// role-targeted notifications.
type NotificationReceipt struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	NotificationID uint       `gorm:"not null;uniqueIndex:idx_notification_user" json:"notification_id"`
	UserID         uint       `gorm:"not null;uniqueIndex:idx_notification_user;index" json:"user_id"`
	IsRead         bool       `gorm:"default:false;index" json:"is_read"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	Notification *Notification `gorm:"foreignKey:NotificationID;constraint:OnDelete:CASCADE" json:"notification,omitempty"`
	User         *User         `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// ==========================================
// FILE UPLOADS
// ==========================================

type FileUpload struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	OriginalName string    `gorm:"size:255;not null" json:"original_name"`
	StoredName   string    `gorm:"size:255;not null" json:"stored_name"`
	FilePath     string    `gorm:"size:500;not null" json:"-"`
	FileURL      string    `gorm:"size:500;not null" json:"file_url"`
	MimeType     string    `gorm:"size:100;not null" json:"mime_type"`
	FileSize     int64     `gorm:"not null" json:"file_size"`             // bytes
	Folder       string    `gorm:"size:100;not null;index" json:"folder"` // photos, documents, bills, receipts
	Entity       *string   `gorm:"size:100;index" json:"entity"`          // member, expense, letter
	EntityID     *uint     `gorm:"index" json:"entity_id"`
	UploadedBy   uint      `gorm:"not null;index" json:"uploaded_by"`
	Visibility   string    `gorm:"size:20;not null;default:private;index" json:"visibility"`
	SHA256       *string   `gorm:"size:64;index" json:"sha256,omitempty"`
	CreatedAt    time.Time `json:"created_at"`

	// Relations
	Uploader *User `gorm:"foreignKey:UploadedBy" json:"uploader,omitempty"`
}
