package handlers

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"mirahub/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

// -------------------- Auth --------------------

// LoginRequest represents user login input
type LoginRequest struct {
	Identifier string `json:"identifier"` // username, email, or phone
	Password   string `json:"password"`
}

// RegisterRequest represents new user input
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Password string `json:"password"`
}

// Register a new user
func Register(c *gin.Context) {
	var req RegisterRequest

	// Validate JSON input
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Basic validation
	if req.Username == "" || req.Password == "" || req.Email == "" || req.Phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username, email, phone, and password are required"})
		return
	}

	db := c.MustGet("db").(*sql.DB)

	// Check if username, email, or phone already exists
	var exists bool
	err := db.QueryRow(
		`SELECT EXISTS(
			SELECT 1 FROM users WHERE username=$1 OR email=$2 OR phone=$3
		)`,
		req.Username, req.Email, req.Phone,
	).Scan(&exists)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username, email, or phone already exists"})
		return
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}

	// Insert new user
	_, err = db.Exec(
		`INSERT INTO users(username, email, phone, password_hash, role)
		 VALUES ($1,$2,$3,$4,$5)`,
		req.Username,
		req.Email,
		req.Phone,
		string(hashed),
		"user",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

func SeedAll(c *gin.Context, db *sqlx.DB) {
	tx := db.MustBegin()

	// ✅ Seed Users
	_, err := tx.Exec(`
		INSERT INTO users (id, username, email, phone, password, role)
		VALUES (
			1,
			'admin',
			'admin@mirahub.com',
			'0990000000',
			'$2a$12$uMl7jYQZ.A4dHqK5bMEwEu6k3Gak8z0N5L8lYEBeo4Qg.UL1rJ9fy',
			'admin'
		)
		ON CONFLICT (id) DO NOTHING
	`)
	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// ✅ Seed Categories
	_, err = tx.Exec(`
        INSERT INTO categories (id, name)
        VALUES 
            (1, 'Electronics'),
            (2, 'Accessories'),
            (3, 'Automotive')
        ON CONFLICT (id) DO NOTHING
    `)
	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// ✅ Seed Suppliers
	_, err = tx.Exec(`
        INSERT INTO suppliers (id, name, contact_info)
        VALUES
            (1, 'MegaTech', '0123456789'),
            (2, 'AutoSuppliers Ltd', '013339991')
        ON CONFLICT (id) DO NOTHING
    `)
	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// ✅ Seed Warehouses
	_, err = tx.Exec(`
        INSERT INTO warehouses (id, name, location)
        VALUES
            (1, 'Main Warehouse', 'Blantyre'),
            (2, 'Secondary Warehouse', 'Lilongwe')
        ON CONFLICT (id) DO NOTHING
    `)
	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// ✅ Seed Products
	_, err = tx.Exec(`
        INSERT INTO products
            (code, item_code, hold, name, category_id, supplier_id, warehouse_id, vehicle, stock, price, created_by)
        VALUES
            ('P001', 'ITM001', false, 'Laptop', 1, 1, 1, '', 10, 450000, 1),
            ('P002', 'ITM002', false, 'Keyboard', 1, 1, 1, '', 15, 15000, 1),
            ('P003', 'ITM003', false, 'Mouse', 1, 1, 1, '', 20, 8000, 1),
            ('P004', 'ITM004', false, 'Monitor', 1, 1, 1, '', 5, 120000, 1),
            ('P005', 'ITM005', false, 'Printer', 1, 1, 1, '', 8, 95000, 1)
        ON CONFLICT (code) DO NOTHING
    `)
	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"message": "✅ Database seeded successfully!",
		"data": gin.H{
			"users":      1,
			"categories": 3,
			"suppliers":  2,
			"warehouses": 2,
			"products":   5,
		},
	})
}

// SeedProducts creates 5 dummy products
func SeedProducts(c *gin.Context, db *sqlx.DB) {
	products := []models.Product{
		{Code: "P001", Name: "Laptop", CategoryID: 1, SupplierID: 1, WarehouseID: 1, Stock: 10, CreatedBy: 1},
		{Code: "P002", Name: "Keyboard", CategoryID: 1, SupplierID: 1, WarehouseID: 1, Stock: 15, CreatedBy: 1},
		{Code: "P003", Name: "Mouse", CategoryID: 1, SupplierID: 1, WarehouseID: 1, Stock: 20, CreatedBy: 1},
		{Code: "P004", Name: "Monitor", CategoryID: 1, SupplierID: 1, WarehouseID: 1, Stock: 5, CreatedBy: 1},
		{Code: "P005", Name: "Printer", CategoryID: 1, SupplierID: 1, WarehouseID: 1, Stock: 8, CreatedBy: 1},
	}

	tx := db.MustBegin()
	for _, p := range products {
		_, err := tx.Exec(`INSERT INTO products (code, name, category_id, supplier_id, warehouse_id, stock, created_by) 
            VALUES ($1,$2,$3,$4,$5,$6,$7)
            ON CONFLICT (code) DO NOTHING`,
			p.Code, p.Name, p.CategoryID, p.SupplierID, p.WarehouseID, p.Stock, p.CreatedBy)
		if err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "5 products seeded successfully!"})
}

// Login authenticates a user and returns a JWT
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	db := c.MustGet("db").(*sql.DB)

	var user models.User
	row := db.QueryRow("SELECT id, username, password_hash, role FROM users WHERE username=$1", req.Identifier)
	if err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Role); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "supersecretkey"
	}

	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

// Logout clears the JWT cookie
func Logout(c *gin.Context) {
	// Set cookie with empty value and expired time
	c.SetCookie(
		"token", // name
		"",      // value
		-1,      // maxAge (negative to delete)
		"/",     // path
		"",      // domain (empty = current)
		false,   // secure
		true,    // httpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// -------------------- Middleware --------------------

// AuthMiddleware checks for JWT token
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			c.Abort()
			return
		}

		tokenString := authHeader[len("Bearer "):]
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "supersecretkey"
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		c.Set("user_id", int(claims["user_id"].(float64)))
		c.Set("username", claims["username"])
		c.Set("role", claims["role"])
		c.Next()
	}
}

// -------------------- Products --------------------

func GetProducts(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	rows, err := db.Query("SELECT id, code, name, category_id, supplier_id, warehouse_id, stock, created_by FROM products")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	products := []models.Product{}
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.CategoryID, &p.SupplierID, &p.WarehouseID, &p.Stock, &p.CreatedBy); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		products = append(products, p)
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}

func GetSales(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	rows, _ := db.Query("SELECT id, product_id, user_id, quantity, sale_date FROM sales")
	defer rows.Close()

	var list []models.Sale
	for rows.Next() {
		var m models.Sale
		rows.Scan(&m.ID, &m.ProductID, &m.UserID, &m.Quantity, &m.SaleDate)
		list = append(list, m)
	}
	c.JSON(http.StatusOK, gin.H{"sales": list})
}

func CreateInvoice(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	userID := c.GetInt("user_id")

	var m models.Invoice
	c.ShouldBindJSON(&m)

	db.Exec("INSERT INTO invoices(sale_id,user_id,total) VALUES($1,$2,$3)",
		m.SaleID, userID, m.Total)

	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}

func CreateQuotation(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	userID := c.GetInt("user_id")

	var m models.Quotation
	c.ShouldBindJSON(&m)

	db.Exec("INSERT INTO quotations(product_id,user_id,price) VALUES($1,$2,$3)",
		m.ProductID, userID, m.Price)

	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}

func CreateReceipt(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	userID := c.GetInt("user_id")

	var m models.Receipt
	c.ShouldBindJSON(&m)

	db.Exec("INSERT INTO receipts(invoice_id,user_id,amount) VALUES($1,$2,$3)",
		m.InvoiceID, userID, m.Amount)

	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}

// -------------------- Sales --------------------

type CreateSaleRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

func GetCategories(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	rows, _ := db.Query("SELECT id, name FROM categories")
	defer rows.Close()

	var list []models.Category
	for rows.Next() {
		var m models.Category
		rows.Scan(&m.ID, &m.Name)
		list = append(list, m)
	}
	c.JSON(http.StatusOK, gin.H{"categories": list})
}

func CreateCategory(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	var m models.Category
	c.ShouldBindJSON(&m)
	db.Exec("INSERT INTO categories(name) VALUES($1)", m.Name)
	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}

func UpdateCategory(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	id := c.Param("id")
	var m models.Category
	c.ShouldBindJSON(&m)
	db.Exec("UPDATE categories SET name=$1 WHERE id=$2", m.Name, id)
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func DeleteCategory(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	id := c.Param("id")
	db.Exec("DELETE FROM categories WHERE id=$1", id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func GetSuppliers(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	rows, _ := db.Query("SELECT id, name, contact_info FROM suppliers")
	defer rows.Close()

	var list []models.Supplier
	for rows.Next() {
		var m models.Supplier
		rows.Scan(&m.ID, &m.Name, &m.ContactInfo)
		list = append(list, m)
	}
	c.JSON(http.StatusOK, gin.H{"suppliers": list})
}

func CreateSupplier(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	var m models.Supplier
	c.ShouldBindJSON(&m)
	db.Exec("INSERT INTO suppliers(name, contact_info) VALUES($1,$2)", m.Name, m.ContactInfo)
	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}

func UpdateSupplier(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	id := c.Param("id")
	var m models.Supplier
	c.ShouldBindJSON(&m)
	db.Exec("UPDATE suppliers SET name=$1, contact_info=$2 WHERE id=$3", m.Name, m.ContactInfo, id)
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func DeleteSupplier(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	id := c.Param("id")
	db.Exec("DELETE FROM suppliers WHERE id=$1", id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func CreateProduct(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	userID := c.GetInt("user_id")

	var p models.Product
	c.ShouldBindJSON(&p)

	db.Exec(`INSERT INTO products(code,name,category_id,supplier_id,warehouse_id,stock,created_by)
	         VALUES($1,$2,$3,$4,$5,$6,$7)`,
		p.Code, p.Name, p.CategoryID, p.SupplierID, p.WarehouseID, p.Stock, userID)

	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}

func UpdateProduct(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	id := c.Param("id")

	var p models.Product
	c.ShouldBindJSON(&p)

	db.Exec(`UPDATE products SET code=$1,name=$2,category_id=$3,supplier_id=$4,warehouse_id=$5,stock=$6 WHERE id=$7`,
		p.Code, p.Name, p.CategoryID, p.SupplierID, p.WarehouseID, p.Stock, id)

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func DeleteProduct(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	id := c.Param("id")
	db.Exec("DELETE FROM products WHERE id=$1", id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func GetWarehouses(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	rows, _ := db.Query("SELECT id, name, location FROM warehouses")
	defer rows.Close()

	var list []models.Warehouse
	for rows.Next() {
		var m models.Warehouse
		rows.Scan(&m.ID, &m.Name, &m.Location)
		list = append(list, m)
	}
	c.JSON(http.StatusOK, gin.H{"warehouses": list})
}

func CreateWarehouse(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	var m models.Warehouse
	c.ShouldBindJSON(&m)
	db.Exec("INSERT INTO warehouses(name, location) VALUES($1,$2)", m.Name, m.Location)
	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}

func UpdateWarehouse(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	id := c.Param("id")
	var m models.Warehouse
	c.ShouldBindJSON(&m)
	db.Exec("UPDATE warehouses SET name=$1, location=$2 WHERE id=$3", m.Name, m.Location, id)
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func DeleteWarehouse(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	id := c.Param("id")
	db.Exec("DELETE FROM warehouses WHERE id=$1", id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func CreateSale(c *gin.Context) {
	var req CreateSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	db := c.MustGet("db").(*sql.DB)
	userID := c.GetInt("user_id")

	// Insert sale
	_, err := db.Exec("INSERT INTO sales(product_id, user_id, quantity) VALUES($1,$2,$3)", req.ProductID, userID, req.Quantity)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Decrement stock
	_, err = db.Exec("UPDATE products SET stock = stock - $1 WHERE id=$2", req.Quantity, req.ProductID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Sale created"})
}

// -------------------- Invoices --------------------

func GetInvoices(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	rows, err := db.Query("SELECT id, sale_id, user_id, invoice_date, total FROM invoices")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	invoices := []models.Invoice{}
	for rows.Next() {
		var inv models.Invoice
		if err := rows.Scan(&inv.ID, &inv.SaleID, &inv.UserID, &inv.InvoiceDate, &inv.Total); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		invoices = append(invoices, inv)
	}

	c.JSON(http.StatusOK, gin.H{"invoices": invoices})
}

// -------------------- Quotations --------------------

func GetQuotations(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	rows, err := db.Query("SELECT id, product_id, user_id, quote_date, price FROM quotations")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	quotes := []models.Quotation{}
	for rows.Next() {
		var q models.Quotation
		if err := rows.Scan(&q.ID, &q.ProductID, &q.UserID, &q.QuoteDate, &q.Price); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		quotes = append(quotes, q)
	}

	c.JSON(http.StatusOK, gin.H{"quotations": quotes})
}

// -------------------- Receipts --------------------

func GetReceipts(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)
	rows, err := db.Query("SELECT id, invoice_id, user_id, receipt_date, amount FROM receipts")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	receipts := []models.Receipt{}
	for rows.Next() {
		var r models.Receipt
		if err := rows.Scan(&r.ID, &r.InvoiceID, &r.UserID, &r.ReceiptDate, &r.Amount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		receipts = append(receipts, r)
	}

	c.JSON(http.StatusOK, gin.H{"receipts": receipts})
}
