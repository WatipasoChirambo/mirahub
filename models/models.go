package models

import "time"

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

// Products
type Product struct {
	ID          int     `db:"id" json:"id"`
	Code        string  `db:"code" json:"code"`
	Name        string  `db:"name" json:"name"`
	CategoryID  int     `db:"category_id" json:"category_id"`
	SupplierID  int     `db:"supplier_id" json:"supplier_id"`
	WarehouseID int     `db:"warehouse_id" json:"warehouse_id"`
	Stock       int     `db:"stock" json:"stock"`
	Price       float64 `db:"price" json:"price"` // Added price per unit
	CreatedBy   int     `db:"created_by" json:"created_by"`
}

// Sales
type Sale struct {
	ID        int       `db:"id" json:"id"`
	ProductID int       `db:"product_id" json:"product_id"`
	UserID    int       `db:"user_id" json:"user_id"` // who made the sale
	Quantity  int       `db:"quantity" json:"quantity"`
	Price     float64   `db:"price" json:"price"` // price at the time of sale
	SaleDate  time.Time `db:"sale_date" json:"sale_date"`
}

// Invoices
type Invoice struct {
	ID          int       `db:"id" json:"id"`
	SaleID      int       `db:"sale_id" json:"sale_id"`
	UserID      int       `db:"user_id" json:"user_id"` // user who issued the invoice
	InvoiceDate time.Time `db:"invoice_date" json:"invoice_date"`
	Total       float64   `db:"total" json:"total"`
}

// Quotations
type Quotation struct {
	ID        int       `db:"id" json:"id"`
	ProductID int       `db:"product_id" json:"product_id"`
	UserID    int       `db:"user_id" json:"user_id"` // user who created the quote
	QuoteDate time.Time `db:"quote_date" json:"quote_date"`
	Price     float64   `db:"price" json:"price"`
}

// Receipts
type Receipt struct {
	ID          int       `db:"id" json:"id"`
	InvoiceID   int       `db:"invoice_id" json:"invoice_id"`
	UserID      int       `db:"user_id" json:"user_id"` // user who received payment
	ReceiptDate time.Time `db:"receipt_date" json:"receipt_date"`
	Amount      float64   `db:"amount" json:"amount"`
}
