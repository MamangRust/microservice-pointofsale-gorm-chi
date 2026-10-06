package models

import "time"

type Role struct {
	RoleID    int32      `gorm:"column:role_id;primaryKey" json:"role_id"`
	RoleName  string     `gorm:"column:role_name" json:"role_name"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `gorm:"column:deleted_at;softDelete" json:"deleted_at"`
}

func (Role) TableName() string { return "roles" }

type User struct {
	UserID           int32      `gorm:"column:user_id;primaryKey" json:"user_id"`
	Firstname        string     `gorm:"column:firstname" json:"firstname"`
	Lastname         string     `gorm:"column:lastname" json:"lastname"`
	Email            string     `gorm:"column:email" json:"email"`
	Password         string     `gorm:"column:password" json:"password"`
	VerificationCode string     `gorm:"column:verification_code" json:"verification_code"`
	IsVerified       *bool      `gorm:"column:is_verified" json:"is_verified"`
	CreatedAt        *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt        *time.Time `gorm:"column:deleted_at;softDelete" json:"deleted_at"`
}

func (User) TableName() string { return "users" }

type UserRole struct {
	UserRoleID int32      `gorm:"column:user_role_id;primaryKey" json:"user_role_id"`
	UserID     int32      `gorm:"column:user_id" json:"user_id"`
	RoleID     int32      `gorm:"column:role_id" json:"role_id"`
	CreatedAt  *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt  *time.Time `gorm:"column:deleted_at;softDelete" json:"deleted_at"`
}

func (UserRole) TableName() string { return "user_roles" }

type RefreshToken struct {
	RefreshTokenID int32      `gorm:"column:refresh_token_id;primaryKey" json:"refresh_token_id"`
	UserID         int32      `gorm:"column:user_id" json:"user_id"`
	Token          string     `gorm:"column:token" json:"token"`
	Expiration     time.Time  `gorm:"column:expiration" json:"expiration"`
	CreatedAt      *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt      *time.Time `gorm:"column:deleted_at;softDelete" json:"deleted_at"`
}

func (RefreshToken) TableName() string { return "refresh_tokens" }

type ResetToken struct {
	ResetTokenID int32     `gorm:"column:id;primaryKey" json:"id"`
	UserID       int64     `gorm:"column:user_id" json:"user_id"`
	Token        string    `gorm:"column:token" json:"token"`
	ExpiryDate   time.Time `gorm:"column:expiry_date" json:"expiry_date"`
}

func (ResetToken) TableName() string { return "reset_tokens" }

type Cashier struct {
	CashierID  int32      `gorm:"column:cashier_id;primaryKey" json:"cashier_id"`
	MerchantID int32      `gorm:"column:merchant_id" json:"merchant_id"`
	UserID     int32      `gorm:"column:user_id" json:"user_id"`
	Name       string     `gorm:"column:name" json:"name"`
	CreatedAt  *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt  *time.Time `gorm:"column:deleted_at;softDelete" json:"deleted_at"`
}

func (Cashier) TableName() string { return "cashiers" }

type Merchant struct {
	MerchantID   int32      `gorm:"column:merchant_id;primaryKey" json:"merchant_id"`
	UserID       int32      `gorm:"column:user_id" json:"user_id"`
	Name         string     `gorm:"column:name" json:"name"`
	Description  *string    `gorm:"column:description" json:"description"`
	Address      *string    `gorm:"column:address" json:"address"`
	ContactEmail *string    `gorm:"column:contact_email" json:"contact_email"`
	ContactPhone *string    `gorm:"column:contact_phone" json:"contact_phone"`
	Status       string     `gorm:"column:status" json:"status"`
	CreatedAt    *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt    *time.Time `gorm:"column:deleted_at;softDelete" json:"deleted_at"`
}

func (Merchant) TableName() string { return "merchants" }

type Category struct {
	CategoryID   int32      `gorm:"column:category_id;primaryKey" json:"category_id"`
	Name         string     `gorm:"column:name" json:"name"`
	Description  *string    `gorm:"column:description" json:"description"`
	SlugCategory *string    `gorm:"column:slug_category" json:"slug_category"`
	CreatedAt    *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt    *time.Time `gorm:"column:deleted_at;softDelete" json:"deleted_at"`
}

func (Category) TableName() string { return "categories" }

type Order struct {
	OrderID    int32      `gorm:"column:order_id;primaryKey" json:"order_id"`
	MerchantID int32      `gorm:"column:merchant_id" json:"merchant_id"`
	CashierID  int32      `gorm:"column:cashier_id" json:"cashier_id"`
	TotalPrice int64      `gorm:"column:total_price" json:"total_price"`
	CreatedAt  *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt  *time.Time `gorm:"column:deleted_at;softDelete" json:"deleted_at"`
}

func (Order) TableName() string { return "orders" }

type OrderItem struct {
	OrderItemID int32      `gorm:"column:order_item_id;primaryKey" json:"order_item_id"`
	OrderID     int32      `gorm:"column:order_id" json:"order_id"`
	ProductID   int32      `gorm:"column:product_id" json:"product_id"`
	Quantity    int32      `gorm:"column:quantity" json:"quantity"`
	Price       int64      `gorm:"column:price" json:"price"`
	CreatedAt   *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;softDelete" json:"deleted_at"`
}

func (OrderItem) TableName() string { return "order_items" }

type Product struct {
	ProductID    int32      `gorm:"column:product_id;primaryKey" json:"product_id"`
	MerchantID   int32      `gorm:"column:merchant_id" json:"merchant_id"`
	CategoryID   int32      `gorm:"column:category_id" json:"category_id"`
	Name         string     `gorm:"column:name" json:"name"`
	Description  *string    `gorm:"column:description" json:"description"`
	Price        int32      `gorm:"column:price" json:"price"`
	CountInStock int32      `gorm:"column:count_in_stock" json:"count_in_stock"`
	Brand        *string    `gorm:"column:brand" json:"brand"`
	Weight       *int32     `gorm:"column:weight" json:"weight"`
	SlugProduct  *string    `gorm:"column:slug_product" json:"slug_product"`
	ImageProduct *string    `gorm:"column:image_product" json:"image_product"`
	Barcode      *string    `gorm:"column:barcode" json:"barcode"`
	CreatedAt    *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt    *time.Time `gorm:"column:deleted_at;softDelete" json:"deleted_at"`
}

func (Product) TableName() string { return "products" }

type Transaction struct {
	TransactionID int32      `gorm:"column:transaction_id;primaryKey" json:"transaction_id"`
	OrderID       int32      `gorm:"column:order_id" json:"order_id"`
	MerchantID    int32      `gorm:"column:merchant_id" json:"merchant_id"`
	PaymentMethod string     `gorm:"column:payment_method" json:"payment_method"`
	Amount        int32      `gorm:"column:amount" json:"amount"`
	ChangeAmount  *int32     `gorm:"column:change_amount" json:"change_amount"`
	PaymentStatus *string    `gorm:"column:payment_status" json:"payment_status"`
	CreatedAt     *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt     *time.Time `gorm:"column:deleted_at;softDelete" json:"deleted_at"`
}

func (Transaction) TableName() string { return "transactions" }

type MerchantDocument struct {
	DocumentID   int32      `gorm:"column:document_id;primaryKey" json:"document_id"`
	MerchantID   int32      `gorm:"column:merchant_id" json:"merchant_id"`
	DocumentType string     `gorm:"column:document_type" json:"document_type"`
	DocumentUrl  string     `gorm:"column:document_url" json:"document_url"`
	Status       string     `gorm:"column:status" json:"status"`
	Note         *string    `gorm:"column:note" json:"note"`
	UploadedAt   *time.Time `gorm:"column:uploaded_at" json:"uploaded_at"`
	CreatedAt    *time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    *time.Time `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt    *time.Time `gorm:"column:deleted_at;softDelete" json:"deleted_at"`
}

func (MerchantDocument) TableName() string { return "merchant_documents" }
