package models

import (
	"database/sql"
	"time"
)

// Categories
type Category struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

// Suppliers
type Supplier struct {
	ID          int    `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	ContactInfo string `db:"contact_info" json:"contact_info"`
}

// Warehouses
type Warehouse struct {
	ID       int    `db:"id" json:"id"`
	Name     string `db:"name" json:"name"`
	Location string `db:"location" json:"location"`
	Capacity int    `db:"capacity" json:"capacity"`
	Manager  string `db:"manager" json:"manager"`
	Status   string `db:"status" json:"status"`
}

// Users
type User struct {
	ID        int       `db:"id" json:"id"`
	Username  string    `db:"username" json:"username"`
	Email     string    `db:"email" json:"email"`
	Phone     string    `db:"phone" json:"phone"`
	Password  string    `db:"password_hash" json:"-"` // never expose
	Role      string    `db:"role" json:"role"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Vehicle struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

type Product struct {
	ID          int            `db:"id" json:"id"`
	Code        string         `db:"code" json:"code"`
	ItemCode    string         `db:"item_code" json:"item_code"`
	Hold        sql.NullString `json:"hold"`
	Name        string         `db:"name" json:"name"`
	CategoryID  sql.NullInt64  `db:"category_id" json:"category_id"`   // Changed to sql.NullInt64
	SupplierID  sql.NullInt64  `db:"supplier_id" json:"supplier_id"`   // Changed to sql.NullInt64
	WarehouseID sql.NullInt64  `db:"warehouse_id" json:"warehouse_id"` // Changed to sql.NullInt64
	Stock       int            `db:"stock" json:"stock"`
	Price       float64        `db:"price" json:"price"`

	CreatedBy *int    `db:"created_by" json:"created_by"`
	ImageURL  *string `db:"image_url" json:"image_url"`

	// Many-to-many
	Vehicles []Vehicle `json:"vehicles"`
}

// Sales
type Sale struct {
	ID         int       `db:"id" json:"id"`
	ProductID  int       `db:"product_id" json:"product_id"`
	UserID     int       `db:"user_id" json:"user_id"` // who made the sale
	CustomerID int       `db:"customer_id" json:"customer_id"`
	Quantity   int       `db:"quantity" json:"quantity"`
	Price      float64   `db:"price" json:"price"` // price at the time of sale
	SaleDate   time.Time `db:"sale_date" json:"sale_date"`
	Total      float64   `db:"total" json:"total"`
}

type SaleResponse struct {
	ID        int       `json:"id"`
	ProductID int       `json:"product_id"`
	UserID    int       `json:"user_id"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	SaleDate  time.Time `json:"sale_date"`

	ProductName string `json:"product_name"`

	CreatedByID       *int    `json:"created_by_id"`
	CreatedByUsername *string `json:"created_by_username"`
	Total             float64 `db:"total" json:"total"`
}

// FileAttachment represents a file attached to a document
type FileAttachment struct {
	ID           int       `db:"id" json:"id"`
	DocumentType string    `db:"document_type" json:"document_type"` // invoice, quotation, receipt, order
	DocumentID   int       `db:"document_id" json:"document_id"`     // ID of the related document
	FileName     string    `db:"file_name" json:"file_name"`
	FilePath     string    `db:"file_path" json:"file_path"`
	FileSize     int64     `db:"file_size" json:"file_size"`
	MimeType     string    `db:"mime_type" json:"mime_type"`
	UploadedBy   int       `db:"uploaded_by" json:"uploaded_by"` // user ID
	UploadedAt   time.Time `db:"uploaded_at" json:"uploaded_at"`
	Description  *string   `db:"description" json:"description"`
}

// Invoices - Updated with file support
type Invoice struct {
	ID          int        `db:"id" json:"id"`
	SaleID      int        `db:"sale_id" json:"sale_id"`
	UserID      int        `db:"user_id" json:"user_id"` // user who issued the invoice
	InvoiceDate time.Time  `db:"invoice_date" json:"invoice_date"`
	Total       float64    `db:"total" json:"total"`
	Status      string     `db:"status" json:"status"` // draft, sent, paid, overdue, cancelled
	DueDate     *time.Time `db:"due_date" json:"due_date"`

	// File attachments
	Attachments []FileAttachment `json:"attachments,omitempty"`

	// PDF generation fields
	PDFPath        *string    `db:"pdf_path" json:"pdf_path"` // path to generated PDF
	PDFGeneratedAt *time.Time `db:"pdf_generated_at" json:"pdf_generated_at"`
}

// InvoiceRequest for creating/updating invoices with file
type InvoiceRequest struct {
	SaleID      int        `json:"sale_id" binding:"required"`
	UserID      int        `json:"user_id" binding:"required"`
	InvoiceDate time.Time  `json:"invoice_date"`
	Total       float64    `json:"total" binding:"required"`
	Status      string     `json:"status"`
	DueDate     *time.Time `json:"due_date"`
}

// Quotations - Updated with file support
type Quotation struct {
	ID         int        `db:"id" json:"id"`
	ProductID  int        `db:"product_id" json:"product_id"`
	UserID     int        `db:"user_id" json:"user_id"` // user who created the quote
	QuoteDate  time.Time  `db:"quote_date" json:"quote_date"`
	Price      float64    `db:"price" json:"price"`
	Status     string     `db:"status" json:"status"` // draft, sent, accepted, rejected, expired
	ValidUntil *time.Time `db:"valid_until" json:"valid_until"`
	Notes      *string    `db:"notes" json:"notes"`

	// File attachments
	Attachments []FileAttachment `json:"attachments,omitempty"`

	// PDF generation fields
	PDFPath        *string    `db:"pdf_path" json:"pdf_path"`
	PDFGeneratedAt *time.Time `db:"pdf_generated_at" json:"pdf_generated_at"`
}

// Receipts - Updated with file support
type Receipt struct {
	ID            int       `db:"id" json:"id"`
	InvoiceID     int       `db:"invoice_id" json:"invoice_id"`
	UserID        int       `db:"user_id" json:"user_id"` // user who received payment
	ReceiptDate   time.Time `db:"receipt_date" json:"receipt_date"`
	Amount        float64   `db:"amount" json:"amount"`
	PaymentMethod string    `db:"payment_method" json:"payment_method"` // cash, card, bank transfer, etc.
	ReferenceNo   *string   `db:"reference_no" json:"reference_no"`
	Notes         *string   `db:"notes" json:"notes"`

	// File attachments
	Attachments []FileAttachment `json:"attachments,omitempty"`

	// PDF generation fields
	PDFPath        *string    `db:"pdf_path" json:"pdf_path"`
	PDFGeneratedAt *time.Time `db:"pdf_generated_at" json:"pdf_generated_at"`
}

type Customer struct {
	ID          int                    `db:"id" json:"id"`
	Name        string                 `db:"name" json:"name"`
	Email       *string                `db:"email" json:"email"`
	Phone       *string                `db:"phone" json:"phone"`
	Whatsapp    *string                `db:"whatsapp" json:"whatsapp"`
	Preferences map[string]interface{} `db:"preferences" json:"preferences"`
	Segment     string                 `db:"segment" json:"segment"`
	CreatedBy   *int                   `db:"created_by" json:"created_by"`
	CreatedAt   time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time              `db:"updated_at" json:"updated_at"`
}

type Order struct {
	ID         int       `db:"id" json:"id"`
	CustomerID int       `db:"customer_id" json:"customer_id"`
	UserID     int       `db:"user_id" json:"user_id"` // who created the order
	OrderDate  time.Time `db:"order_date" json:"order_date"`
	Status     string    `db:"status" json:"status"` // pending, confirmed, shipped, delivered, cancelled

	// File attachments
	Attachments []FileAttachment `json:"attachments,omitempty"`

	// PDF generation fields
	PDFPath        *string    `db:"pdf_path" json:"pdf_path"`
	PDFGeneratedAt *time.Time `db:"pdf_generated_at" json:"pdf_generated_at"`
}

type OrderItem struct {
	ID        int     `db:"id" json:"id"`
	OrderID   int     `db:"order_id" json:"order_id"`
	ProductID int     `db:"product_id" json:"product_id"`
	Quantity  int     `db:"quantity" json:"quantity"`
	Price     float64 `db:"price" json:"price"` // price at time of order
}

// Helper struct for file upload response
type FileUploadResponse struct {
	FileID      int    `json:"file_id"`
	FileName    string `json:"file_name"`
	FilePath    string `json:"file_path"`
	FileSize    int64  `json:"file_size"`
	MimeType    string `json:"mime_type"`
	DownloadURL string `json:"download_url"`
}
