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

// Products
type Product struct {
	ID          int     `db:"id" json:"id"`
	Code        string  `db:"code" json:"code"`
	Name        string  `db:"name" json:"name"`
	CategoryID  int     `db:"category_id" json:"category_id"`
	SupplierID  int     `db:"supplier_id" json:"supplier_id"`
	WarehouseID int     `db:"warehouse_id" json:"warehouse_id"`
	Stock       int     `db:"stock" json:"stock"`
	Price       float64 `db:"price" json:"price"`         // ✅ Per unit
	Hold        string  `db:"hold" json:"hold"`           // ✅ NEW
	Vehicle     string  `db:"vehicle" json:"vehicle"`     // ✅ NEW
	ItemCode    string  `db:"item_code" json:"item_code"` // ✅ NEW
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
	Total     float64   `db:"total" json:"total"`
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

// Customers
type Customer struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Email     string    `db:"email" json:"email"`
	Phone     string    `db:"phone" json:"phone"`
	CreatedBy int       `db:"created_by" json:"created_by"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Order struct {
	ID         int       `db:"id" json:"id"`
	CustomerID int       `db:"customer_id" json:"customer_id"`
	UserID     int       `db:"user_id" json:"user_id"` // who created the order
	OrderDate  time.Time `db:"order_date" json:"order_date"`
}

type OrderItem struct {
	ID        int     `db:"id" json:"id"`
	OrderID   int     `db:"order_id" json:"order_id"`
	ProductID int     `db:"product_id" json:"product_id"`
	Quantity  int     `db:"quantity" json:"quantity"`
	Price     float64 `db:"price" json:"price"` // price at time of order
}
