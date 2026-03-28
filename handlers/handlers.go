package handlers

import (
	"database/sql"
	"log"
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
    INSERT INTO users (id, username, email, phone, password_hash, role)
    VALUES (
        1,
        'admin',
        'admin@mirahub.com',
        '0990000000',
        '$2a$12$uMl7jYQZ.A4dHqK5bMEwEu6k3Gak8z0N5L8lYEBeo4Qg.UL1rJ9fy',
        'admin'
    )
    ON CONFLICT (id) DO NOTHING;
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
        ('Z80',         'OF23',  'A1-1B',      'Oil Filter', 1, 1, 1, NULL,             0, 0.00, 1),
        ('Z131',        'OF7',   'A1-4C',      'Oil Filter', 1, 1, 1, 'Toyota Old Hi',  0, 0.00, 1),
        ('Z347',        'OF18',  'A1-5B',      'Oil Filter', 1, 1, 1, NULL,             0, 0.00, 1),
        ('Z95',         'OF10',  'A1-A+',      'Oil Filter', 1, 1, 1, 'Ford Ranger 2',  0, 0.00, 1),

        ('7801-21040',  'AF11',  'A9-3AB',     'Air Filter', 1, 1, 1, 'Toyota Prius',   0, 0.00, 1),
        ('16546-41B00', 'AF2',   'A9-1C',      'Air Filter', 1, 1, 1, 'Nissan March',   0, 0.00, 1),
        ('17801-11130', 'AF9',   'A9-4AB',     'Air Filter', 1, 1, 1, 'Toyota Land C',  0, 0.00, 1),
        ('17801-38050', 'AF25',  'A9-2AB',     'Air Filter', 1, 1, 1, 'Toyota Land C',  0, 0.00, 1),
        ('17801-33040', 'AF22',  'A5-1B',      'Air Filter', 1, 1, 1, 'Toyota Probox',  0, 0.00, 1),
        ('17801-30040', 'AF17',  'A9-3AB',     'Air Filter', 1, 1, 1, 'Toyota Land C',  0, 0.00, 1),
        ('17801-28030', 'AF16',  'A5-6B',      'Air Filter', 1, 1, 1, 'Toyota Camry',   0, 0.00, 1),
        ('17801-31090', 'AF20',  'A9-1AB',     'Air Filter', 1, 1, 1, 'Toyota FJ Cru',  0, 0.00, 1),
        ('17801-31120', 'AF21',  '',           'Air Filter', 1, 1, 1, 'Toyota RAV 4',   0, 0.00, 1),
        ('17801-50040', 'AF27',  'A9-4AB',     'Air Filter', 1, 1, 1, 'Toyota Land C',  0, 0.00, 1),
        ('17801-38011', 'AF24',  'A9-2C',      'Air Filter', 1, 1, 1, 'Toyota Camry',   0, 0.00, 1),
        ('17801-30060', 'AF18',  'A6-3AB',     'Air Filter', 1, 1, 1, 'Toyota Hiace',   0, 0.00, 1),
        ('17801-37020', 'AF23',  'A9-2AB',     'Air Filter', 1, 1, 1, 'Toyota RAV 4',   0, 0.00, 1),
        ('17801-30070', 'AF19',  'A9-1AB',     'Air Filter', 1, 1, 1, 'Toyota Hiace',   0, 0.00, 1),
        ('17801-77050', 'AF34',  'A5-5B',      'Air Filter', 1, 1, 1, 'Toyota RAV 4',   0, 0.00, 1),
        ('17801-97402', 'AF37',  'A5-5C',      'Air Filter', 1, 1, 1, 'Toyota Passo',   0, 0.00, 1),
        ('17801-21060', 'AF13',  '',           'Air Filter', 1, 1, 1, 'Toyota Passo 1', 0, 0.00, 1),
        ('80292-T5R-E01','AF46', 'A5-1B',      'Air Filter', 1, 1, 1, NULL,             0, 0.00, 1),
        ('17801-B1010', 'AF38',  'A6-6AB',     'Air Filter', 1, 1, 1, 'Daihatsu Terio', 0, 0.00, 1),
        ('17801-28010', 'AF15',  'A6-5AB',     'Air Filter', 1, 1, 1, 'Toyota RAV 4',   0, 0.00, 1),
        ('17801-54110', 'AF28',  'A6-4AB',     'Air Filter', 1, 1, 1, 'Toyota Hiace/1', 0, 0.00, 1)
    ON CONFLICT (code) DO NOTHING;
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

	log.Println("🔍 LOGIN HANDLER RUNNING WITH IDENTIFIER:", req.Identifier)

	var user models.User
	row := db.QueryRow(`
        SELECT id, username, password_hash, role
        FROM users
        WHERE username=$1 OR email=$1 OR phone=$1
    `, req.Identifier)

	if err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Role); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// compare password with bcrypt hash
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

	rows, err := db.Query(`
    SELECT id, code, name, category_id, supplier_id, warehouse_id, stock, price, created_by
    FROM products
`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	products := []models.Product{}
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Code, &p.Name, &p.CategoryID, &p.SupplierID, &p.WarehouseID, &p.Stock, &p.Price, &p.CreatedBy); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		products = append(products, p)
	}

	c.JSON(http.StatusOK, gin.H{"products": products})
}

func CreateSale(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	// ✅ Extract user ID from JWT
	claims := c.MustGet("claims").(jwt.MapClaims)
	userID := int(claims["user_id"].(float64))

	type Body struct {
		ProductID int     `json:"product_id"`
		Quantity  int     `json:"quantity"`
		Price     float64 `json:"price"`
	}

	var b Body
	if err := c.ShouldBindJSON(&b); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	// ✅ Begin atomic transaction
	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction"})
		return
	}

	// Rollback helper
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// ✅ Lock product row (FOR UPDATE prevents race conditions)
	var currentStock int
	err = tx.QueryRow(`
        SELECT stock 
        FROM products 
        WHERE id = ? 
        FOR UPDATE
    `, b.ProductID).Scan(&currentStock)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found"})
		return
	}

	// ✅ Prevent negative stock
	if b.Quantity > currentStock {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity exceeds available stock"})
		return
	}

	// ✅ Insert sale + return sale ID + timestamp
	var saleID int
	var saleDate time.Time

	err = tx.QueryRow(`
        INSERT INTO sales (product_id, user_id, quantity, price)
        VALUES (?, ?, ?, ?)
        RETURNING id, sale_date
    `, b.ProductID, userID, b.Quantity, b.Price).Scan(&saleID, &saleDate)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save sale"})
		return
	}

	// ✅ Deduct stock
	_, err = tx.Exec(`
        UPDATE products
        SET stock = stock - ?
        WHERE id = ?
    `, b.Quantity, b.ProductID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stock"})
		return
	}

	// ✅ Commit atomic transaction
	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Transaction failed"})
		return
	}

	// ✅ Return full sale record
	c.JSON(http.StatusCreated, gin.H{
		"message":    "Sale created successfully",
		"sale_id":    saleID,
		"product_id": b.ProductID,
		"user_id":    userID,
		"quantity":   b.Quantity,
		"price":      b.Price,
		"sale_date":  saleDate,
	})
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

func GetSales(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	rows, err := db.Query(`
        SELECT 
            s.id, s.product_id, s.user_id, s.quantity, s.price, s.sale_date,
            p.name AS product_name,
            u.username AS username
        FROM sales s
        JOIN products p ON p.id = s.product_id
        JOIN users u ON u.id = s.user_id
        ORDER BY s.id DESC
    `)
	if err != nil {
		c.JSON(500, gin.H{"error": "Database query failed"})
		return
	}
	defer rows.Close()

	type SaleResponse struct {
		ID          int       `json:"id"`
		ProductID   int       `json:"product_id"`
		UserID      int       `json:"user_id"`
		Quantity    int       `json:"quantity"`
		Price       float64   `json:"price"`
		SaleDate    time.Time `json:"sale_date"`
		ProductName string    `json:"product_name"`
		Username    string    `json:"username"`
	}

	var sales []SaleResponse

	for rows.Next() {
		var s SaleResponse
		err := rows.Scan(
			&s.ID, &s.ProductID, &s.UserID, &s.Quantity, &s.Price, &s.SaleDate,
			&s.ProductName, &s.Username,
		)
		if err != nil {
			c.JSON(500, gin.H{"error": "Scan failed"})
			return
		}
		sales = append(sales, s)
	}

	c.JSON(200, sales)
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
