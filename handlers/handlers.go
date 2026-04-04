package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"mirahub/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"github.com/resend/resend-go/v3"
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

type ContactForm struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type QuoteForm struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Company string `json:"company"`
	Parts   string `json:"parts"`
	Notes   string `json:"notes"`
}

func QuoteHandler(c *gin.Context) {
	var form QuoteForm

	// 1. Parse JSON
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 2. Setup Resend Client
	apiKey := os.Getenv("RESEND_API_KEY")
	emailUser := os.Getenv("EMAIL_USER") // Your receiving email

	if apiKey == "" || emailUser == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Email configuration missing"})
		return
	}
	client := resend.NewClient(apiKey)

	// 3. Prepare the HTML Body
	htmlContent := fmt.Sprintf(`
        <h2>New Quote Request</h2>
        <p><strong>Name:</strong> %s</p>
        <p><strong>Email:</strong> %s</p>
        <p><strong>Phone:</strong> %s</p>
        <p><strong>Company:</strong> %s</p>
        <p><strong>Parts:</strong><br/>%s</p>
        <p><strong>Notes:</strong><br/>%s</p>
    `, form.Name, form.Email, form.Phone, form.Company, form.Parts, form.Notes)

	// 4. Create the Request
	params := &resend.SendEmailRequest{
		From:    form.Email, // Replace with your verified domain in production
		To:      []string{emailUser},
		Cc:      []string{"watichirambo@gmail.com"},
		ReplyTo: form.Email,
		Subject: "New Quote Request from " + form.Name,
		Html:    htmlContent,
	}

	// 5. Send
	sent, err := client.Emails.Send(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Quote request sent",
		"email_id": sent.Id,
	})
}

func ContactHandler(c *gin.Context) {
	var form ContactForm

	// 1. Parse JSON
	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 2. Initialize Resend Client
	// Ideally, move the client initialization outside this handler for better performance
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Resend API key not set"})
		return
	}
	client := resend.NewClient(apiKey)

	// 3. Prepare the Email
	params := &resend.SendEmailRequest{
		From:    form.Email, // Replace with your verified domain in production
		To:      []string{os.Getenv("EMAIL_USER")},
		Cc:      []string{"watichirambo@gmail.com"},
		Subject: form.Subject,
		ReplyTo: form.Email,
		Text: fmt.Sprintf(
			"Name: %s\nEmail: %s\n\nMessage:\n%s",
			form.Name, form.Email, form.Message,
		),
	}

	// 4. Send Email
	sent, err := client.Emails.Send(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Email sent successfully",
		"email_id": sent.Id,
	})
}

// Register a new user
func Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	db := c.MustGet("db").(*sqlx.DB)

	if req.Username == "" || req.Password == "" || req.Email == "" || req.Phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "All fields required"})
		return
	}

	var exists bool
	err := db.Get(&exists, `
		SELECT EXISTS(
			SELECT 1 FROM users 
			WHERE username=$1 OR email=$2 OR phone=$3
		)
	`, req.Username, req.Email, req.Phone)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	if exists {
		c.JSON(400, gin.H{"error": "User already exists"})
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	_, err = db.Exec(`
		INSERT INTO users(username,email,phone,password_hash,role)
		VALUES($1,$2,$3,$4,$5)
	`, req.Username, req.Email, req.Phone, string(hash), "user")

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{"message": "User registered"})
}

func SeedVehiclesAndProducts(db *sqlx.DB) error {
	tx := db.MustBegin()

	// Step 1: Insert vehicles
	vehicles := []string{
		"Toyota Prius",
		"Toyota Aqua",
		"Ford Ranger",
		"Toyota Hiace",
	}

	var vehicleIDs = make(map[string]int)

	for _, v := range vehicles {
		var id int
		err := tx.QueryRow(`
			INSERT INTO vehicles (name)
			VALUES ($1)
			ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, v).Scan(&id)

		if err != nil {
			tx.Rollback()
			return err
		}

		vehicleIDs[v] = id
	}

	// Step 2: Insert product
	var productID int
	err := tx.QueryRow(`
		INSERT INTO products (code, item_code, hold, name, category_id, supplier_id, warehouse_id, stock, price, created_by, image_url)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (code) DO NOTHING
		RETURNING id
	`,
		"Z80", "OF23", "A1-1B", "Oil Filter", 1, 1, 1, 0, 0.00, 1, "/uploads/products/oil.jpg",
	).Scan(&productID)

	if err != nil {
		tx.Rollback()
		return err
	}

	// Step 3: Link product to vehicles
	productVehicles := []string{
		"Toyota Prius",
		"Toyota Aqua",
	}

	for _, v := range productVehicles {
		_, err := tx.Exec(`
			INSERT INTO product_vehicles (product_id, vehicle_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, productID, vehicleIDs[v])

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func AttachVehicle(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("id")

		var req struct {
			VehicleID int `json:"vehicle_id"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Exec(`
			INSERT INTO product_vehicles (product_id, vehicle_id)
			VALUES ($1, $2)
			ON CONFLICT DO NOTHING
		`, productID, req.VehicleID)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "vehicle attached"})
	}
}

func DetachVehicle(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		productID := c.Param("id")

		var req struct {
			VehicleID int `json:"vehicle_id"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Exec(`
			DELETE FROM product_vehicles
			WHERE product_id = $1 AND vehicle_id = $2
		`, productID, req.VehicleID)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "vehicle detached"})
	}
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

	db := c.MustGet("db").(*sqlx.DB)

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT secret not configured"})
		return
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

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			c.Abort()
			return
		}

		// ✅ Safe Bearer parsing
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT secret not configured"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Optional: enforce signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// ✅ Safe claims extraction
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// ✅ Safely extract fields
		uidRaw, exists := claims["user_id"]
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user_id missing in token"})
			c.Abort()
			return
		}

		uidFloat, ok := uidRaw.(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id type"})
			c.Abort()
			return
		}

		c.Set("user_id", int(uidFloat))

		if username, ok := claims["username"].(string); ok {
			c.Set("username", username)
		}

		if role, ok := claims["role"].(string); ok {
			c.Set("role", role)
		}

		c.Next()
	}
}

func GetProducts(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	rows, err := db.Query(`
		SELECT 
    p.id,
    p.code,
    p.name,
    p.category_id,
    p.supplier_id,
    p.warehouse_id,
    p.stock,
    p.price,
    p.hold,
    p.item_code,
    p.image_url,
    p.created_by,
    v.id AS vehicle_id,
    v.name AS vehicle_name
FROM products p
LEFT JOIN product_vehicles pv ON pv.product_id = p.id
LEFT JOIN vehicles v ON v.id = pv.vehicle_id
ORDER BY p.id
	`)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	productMap := make(map[int]*models.Product)

	for rows.Next() {
		var (
			p           models.Product
			vehicleID   sql.NullInt64
			vehicleName sql.NullString
			imageURL    sql.NullString
		)

		err := rows.Scan(
			&p.ID,
			&p.Code,
			&p.Name,
			&p.CategoryID,
			&p.SupplierID,
			&p.WarehouseID,
			&p.Stock,
			&p.Price,
			&p.Hold,
			&p.ItemCode,
			&imageURL,
			&p.CreatedBy,
			&vehicleID,
			&vehicleName,
		)

		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		if imageURL.Valid {
			p.ImageURL = imageURL.String
		}

		if _, exists := productMap[p.ID]; !exists {
			p.Vehicles = []models.Vehicle{}
			productMap[p.ID] = &p
		}

		if vehicleID.Valid {
			productMap[p.ID].Vehicles = append(productMap[p.ID].Vehicles, models.Vehicle{
				ID:   int(vehicleID.Int64),
				Name: vehicleName.String,
			})
		}
	}

	var products []models.Product
	for _, p := range productMap {
		products = append(products, *p)
	}

	c.JSON(200, gin.H{"products": products})
}

func nullInt(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func nullFloat(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func CreateProduct(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	code := c.PostForm("code")
	itemCode := c.PostForm("item_code")
	name := c.PostForm("name")

	categoryID := c.PostForm("category_id")
	supplierID := c.PostForm("supplier_id")
	warehouseID := c.PostForm("warehouse_id")
	stock := c.PostForm("stock")
	price := c.PostForm("price")
	hold := c.PostForm("hold")
	createdBy := c.PostForm("created_by")

	// ✅ Image upload
	file, _ := c.FormFile("image")
	var imageURL string

	if file != nil {
		path := "./uploads/products/" + file.Filename
		if err := c.SaveUploadedFile(file, path); err != nil {
			c.JSON(500, gin.H{"error": "image upload failed"})
			return
		}
		imageURL = "/uploads/products/" + file.Filename
	}

	vehicleIDs := c.PostFormArray("vehicle_ids")

	tx, err := db.Beginx()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	var productID int
	err = tx.QueryRow(`
		INSERT INTO products 
		(code,item_code,name,category_id,supplier_id,warehouse_id,stock,price,hold,created_by,image_url)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id
	`,
		code,
		itemCode,
		name,
		nullInt(categoryID),
		nullInt(supplierID),
		nullInt(warehouseID),
		nullInt(stock),
		nullFloat(price),
		hold,
		nullInt(createdBy),
		imageURL,
	).Scan(&productID)

	if err != nil {
		tx.Rollback()
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// ✅ Link vehicles
	for _, vid := range vehicleIDs {
		_, err := tx.Exec(`
			INSERT INTO product_vehicles(product_id, vehicle_id)
			VALUES ($1,$2)
			ON CONFLICT DO NOTHING
		`, productID, vid)

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

	c.JSON(201, gin.H{
		"message": "product created",
		"id":      productID,
	})
}

func GetCustomerByID(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	id := c.Param("id")

	var cust models.Customer

	err := db.QueryRow(`
        SELECT id, name, email, phone, created_by, created_at
        FROM customers
        WHERE id = $1
    `, id).Scan(
		&cust.ID,
		&cust.Name,
		&cust.Email,
		&cust.Phone,
		&cust.CreatedBy,
		&cust.CreatedAt,
	)

	if err == sql.ErrNoRows {
		c.JSON(404, gin.H{"error": "Customer not found"})
		return
	}

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch customer: " + err.Error()})
		return
	}

	c.JSON(200, cust)
}

func UpdateCustomer(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	id := c.Param("id")

	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	res, err := db.Exec(`
        UPDATE customers
        SET name = $1, email = $2, phone = $3
        WHERE id = $4
    `, input.Name, input.Email, input.Phone, id)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update customer: " + err.Error()})
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Customer not found"})
		return
	}

	c.JSON(200, gin.H{
		"message": "Customer updated",
		"id":      id,
	})
}

func DeleteCustomer(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	id := c.Param("id")

	res, err := db.Exec(`
        DELETE FROM customers
        WHERE id = $1
    `, id)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete customer: " + err.Error()})
		return
	}

	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(404, gin.H{"error": "Customer not found"})
		return
	}

	c.JSON(200, gin.H{
		"message": "Customer deleted",
		"id":      id,
	})
}

func CreateCustomer(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	var input struct {
		Name     string  `json:"name"`
		Email    *string `json:"email"`
		Phone    *string `json:"phone"`
		Whatsapp *string `json:"whatsapp"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	createdBy := c.GetInt("user_id")

	var id int
	err := db.QueryRow(`
		INSERT INTO customers (name, email, phone, whatsapp, created_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`,
		input.Name,
		input.Email,
		input.Phone,
		input.Whatsapp,
		createdBy,
	).Scan(&id)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create customer: " + err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"message": "Customer created",
		"id":      id,
	})
}

func GetCustomers(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	rows, err := db.Query(`
        SELECT 
            id,
            name,
            email,
            phone,
            created_by,
            created_at
        FROM customers
        ORDER BY id ASC
    `)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database query failed: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	customers := []models.Customer{}

	for rows.Next() {
		var cust models.Customer

		err := rows.Scan(
			&cust.ID,
			&cust.Name,
			&cust.Email,
			&cust.Phone,
			&cust.CreatedBy,
			&cust.CreatedAt,
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Scan error: " + err.Error(),
			})
			return
		}

		customers = append(customers, cust)
	}

	// ✅ Check for iteration errors
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Row iteration failed: " + err.Error(),
		})
		return
	}

	// ✅ Same format as GetProducts
	c.JSON(http.StatusOK, gin.H{
		"customers": customers,
	})
}

func CreateOrder(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	userID := c.GetInt("user_id")
	if userID == 0 {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var input struct {
		CustomerID int `json:"customer_id"`
		Items      []struct {
			ProductID int     `json:"product_id"`
			Quantity  int     `json:"quantity"`
			Price     float64 `json:"price"`
		} `json:"items"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	// 1. Create order
	var orderID int
	err = tx.QueryRow(`
		INSERT INTO orders (customer_id, user_id)
		VALUES ($1, $2)
		RETURNING id
	`, input.CustomerID, userID).Scan(&orderID)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create order"})
		return
	}

	// 2. Insert order items + update stock
	for _, item := range input.Items {

		// Lock product
		var stock int
		err := tx.QueryRow(`
			SELECT stock FROM products WHERE id = $1 FOR UPDATE
		`, item.ProductID).Scan(&stock)

		if err != nil {
			c.JSON(500, gin.H{"error": "Product not found"})
			return
		}

		if item.Quantity > stock {
			c.JSON(400, gin.H{"error": "Insufficient stock"})
			return
		}

		// Deduct stock
		_, err = tx.Exec(`
			UPDATE products SET stock = stock - $1 WHERE id = $2
		`, item.Quantity, item.ProductID)

		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to update stock"})
			return
		}

		// Insert order item
		_, err = tx.Exec(`
			INSERT INTO order_items (order_id, product_id, quantity, price)
			VALUES ($1, $2, $3, $4)
		`, orderID, item.ProductID, item.Quantity, item.Price)

		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to insert order items"})
			return
		}
	}

	// Commit
	if err := tx.Commit(); err != nil {
		c.JSON(500, gin.H{"error": "Transaction commit failed"})
		return
	}

	c.JSON(201, gin.H{
		"order_id": orderID,
	})
}

func GetOrders(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	rows, err := db.Query(`
		SELECT id, customer_id, user_id, created_at
		FROM orders
		ORDER BY created_at DESC
	`)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch orders"})
		return
	}
	defer rows.Close()

	var orders []map[string]interface{}

	for rows.Next() {
		var id, customerID, userID int
		var createdAt string

		rows.Scan(&id, &customerID, &userID, &createdAt)

		orders = append(orders, gin.H{
			"id":          id,
			"customer_id": customerID,
			"user_id":     userID,
			"created_at":  createdAt,
		})
	}

	c.JSON(200, gin.H{
		"orders": orders,
	})
}

func CreateCustomers(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	var input struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	// Authenticated user
	createdBy := c.GetInt("user_id")

	var id int
	err := db.QueryRow(`
        INSERT INTO customers (name, email, phone, created_by)
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `, input.Name, input.Email, input.Phone, createdBy).Scan(&id)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create customer: " + err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"message": "Customer created",
		"id":      id,
	})
}

func CreateSale(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	// ✅ Get authenticated user
	userID := c.GetInt("user_id")
	if userID == 0 {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// ✅ Input struct
	var input struct {
		ProductID int     `json:"product_id"`
		Quantity  int     `json:"quantity"`
		Price     float64 `json:"price"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	// ✅ Basic validation
	if input.ProductID <= 0 || input.Quantity <= 0 || input.Price <= 0 {
		c.JSON(400, gin.H{"error": "Invalid product_id, quantity, or price"})
		return
	}

	// ✅ Start transaction
	tx, err := db.Begin()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Ensure rollback on failure
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// ✅ Lock product row
	var stock int
	err = tx.QueryRow(`
		SELECT stock 
		FROM products 
		WHERE id = $1 
		FOR UPDATE
	`, input.ProductID).Scan(&stock)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{"error": "Product not found"})
		} else {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}

	// ✅ Check stock
	if input.Quantity > stock {
		c.JSON(400, gin.H{"error": "Insufficient stock"})
		return
	}

	// ✅ Deduct stock
	_, err = tx.Exec(`
		UPDATE products 
		SET stock = stock - $1 
		WHERE id = $2
	`, input.Quantity, input.ProductID)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// ✅ Calculate total
	total := input.Price * float64(input.Quantity)

	// ✅ Insert sale
	var saleID int
	var saleDate time.Time

	err = tx.QueryRow(`
		INSERT INTO sales (product_id, user_id, quantity, price, total)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, sale_date
	`, input.ProductID, userID, input.Quantity, input.Price, total).Scan(&saleID, &saleDate)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   "Insert failed",
			"details": err.Error(),
		})
		return
	}

	// ✅ Commit transaction
	if err = tx.Commit(); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	// ✅ Response
	c.JSON(201, gin.H{
		"id":         saleID,
		"product_id": input.ProductID,
		"user_id":    userID,
		"quantity":   input.Quantity,
		"price":      input.Price,
		"sale_date":  saleDate,
		"total":      total,
	})
}

func CreateInvoice(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	userID := c.GetInt("user_id")

	var m models.Invoice
	c.ShouldBindJSON(&m)

	db.Exec("INSERT INTO invoices(sale_id,user_id,total) VALUES($1,$2,$3)",
		m.SaleID, userID, m.Total)

	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}

func CreateQuotation(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	userID := c.GetInt("user_id")

	var m models.Quotation
	c.ShouldBindJSON(&m)

	db.Exec("INSERT INTO quotations(product_id,user_id,price) VALUES($1,$2,$3)",
		m.ProductID, userID, m.Price)

	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}

func CreateReceipt(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
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
	db := c.MustGet("db").(*sqlx.DB)
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
	db := c.MustGet("db").(*sqlx.DB)
	var m models.Category
	c.ShouldBindJSON(&m)
	db.Exec("INSERT INTO categories(name) VALUES($1)", m.Name)
	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}

func UpdateCategory(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	id := c.Param("id")
	var m models.Category
	c.ShouldBindJSON(&m)
	db.Exec("UPDATE categories SET name=$1 WHERE id=$2", m.Name, id)
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func DeleteCategory(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	id := c.Param("id")
	db.Exec("DELETE FROM categories WHERE id=$1", id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func GetSuppliers(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
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
	db := c.MustGet("db").(*sqlx.DB)
	var m models.Supplier
	c.ShouldBindJSON(&m)
	db.Exec("INSERT INTO suppliers(name, contact_info) VALUES($1,$2)", m.Name, m.ContactInfo)
	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}

func UpdateSupplier(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	id := c.Param("id")
	var m models.Supplier
	c.ShouldBindJSON(&m)
	db.Exec("UPDATE suppliers SET name=$1, contact_info=$2 WHERE id=$3", m.Name, m.ContactInfo, id)
	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func DeleteSupplier(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	id := c.Param("id")
	db.Exec("DELETE FROM suppliers WHERE id=$1", id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func UpdateProduct(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	id := c.Param("id")

	var p models.Product
	c.ShouldBindJSON(&p)

	db.Exec(`UPDATE products SET code=$1,name=$2,category_id=$3,supplier_id=$4,warehouse_id=$5,stock=$6 WHERE id=$7`,
		p.Code, p.Name, p.CategoryID, p.SupplierID, p.WarehouseID, p.Stock, id)

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func DeleteProduct(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	id := c.Param("id")
	db.Exec("DELETE FROM products WHERE id=$1", id)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func CreateWarehouse(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	userID := c.GetInt("user_id")
	if userID == 0 {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var input struct {
		Name     string `json:"name"`
		Location string `json:"location"`
		Capacity int    `json:"capacity"`
		Manager  string `json:"manager"`
		Status   string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	if input.Status == "" {
		input.Status = "active"
	}

	var id int
	err := db.QueryRow(`
		INSERT INTO warehouses (name, location, capacity, manager, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, input.Name, input.Location, input.Capacity, input.Manager, input.Status).Scan(&id)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create warehouse"})
		return
	}

	c.JSON(201, gin.H{
		"id": id,
	})
}

func GetWarehouseStock(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	rows, err := db.Query(`
		SELECT 
			w.id,
			w.name,
			COALESCE(SUM(p.stock), 0) as total_stock
		FROM warehouses w
		LEFT JOIN products p ON p.warehouse_id = w.id
		GROUP BY w.id, w.name
		ORDER BY w.id
	`)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch stock"})
		return
	}
	defer rows.Close()

	var result []gin.H

	for rows.Next() {
		var id int
		var name string
		var totalStock int

		rows.Scan(&id, &name, &totalStock)

		result = append(result, gin.H{
			"id":          id,
			"name":        name,
			"total_stock": totalStock,
		})
	}

	c.JSON(200, gin.H{
		"warehouses": result,
	})
}

func GetWarehouses(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	rows, err := db.Query(`
		SELECT id, name, location, capacity, manager, status, created_at
		FROM warehouses
		ORDER BY id DESC
	`)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch warehouses"})
		return
	}
	defer rows.Close()

	var warehouses []map[string]interface{}

	for rows.Next() {
		var id, capacity int
		var name, location, manager, status string
		var createdAt string

		rows.Scan(&id, &name, &location, &capacity, &manager, &status, &createdAt)

		warehouses = append(warehouses, gin.H{
			"id":         id,
			"name":       name,
			"location":   location,
			"capacity":   capacity,
			"manager":    manager,
			"status":     status,
			"created_at": createdAt,
		})
	}

	c.JSON(200, gin.H{
		"warehouses": warehouses,
	})
}

func UpdateWarehouse(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	id := c.Param("id")

	var input struct {
		Name     string `json:"name"`
		Location string `json:"location"`
		Capacity int    `json:"capacity"`
		Manager  string `json:"manager"`
		Status   string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	_, err := db.Exec(`
		UPDATE warehouses
		SET name=$1, location=$2, capacity=$3, manager=$4, status=$5
		WHERE id=$6
	`, input.Name, input.Location, input.Capacity, input.Manager, input.Status, id)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update warehouse"})
		return
	}

	c.JSON(200, gin.H{"message": "Warehouse updated"})
}

func DeleteWarehouse(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	id := c.Param("id")

	_, err := db.Exec(`DELETE FROM warehouses WHERE id=$1`, id)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete warehouse"})
		return
	}

	c.JSON(200, gin.H{"message": "Warehouse deleted"})
}

func GetSales(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	rows, err := db.Query(`
		SELECT 
			s.id,
			s.product_id,
			s.user_id,
			s.quantity,
			s.price,
			s.sale_date,
			p.name AS product_name,
			u.id AS created_by_id,
			u.username AS created_by_username
		FROM sales s
		JOIN products p ON p.id = s.product_id
		LEFT JOIN users u ON u.id = p.created_by
		ORDER BY s.id DESC
	`)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   "Database query failed",
			"details": err.Error(),
		})
		return
	}
	defer rows.Close()

	var sales []models.SaleResponse

	for rows.Next() {
		var s models.SaleResponse

		err := rows.Scan(
			&s.ID,
			&s.ProductID,
			&s.UserID,
			&s.Quantity,
			&s.Price,
			&s.SaleDate,
			&s.ProductName,
			&s.CreatedByID,
			&s.CreatedByUsername,
		)

		if err != nil {
			c.JSON(500, gin.H{
				"error":   "Scan failed",
				"details": err.Error(),
			})
			return
		}

		sales = append(sales, s)
	}

	if err := rows.Err(); err != nil {
		c.JSON(500, gin.H{
			"error":   "Row iteration error",
			"details": err.Error(),
		})
		return
	}

	// ✅ Always return empty array instead of null
	if sales == nil {
		sales = []models.SaleResponse{}
	}

	c.JSON(200, sales)
}

// -------------------- Invoices --------------------

func GetInvoices(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
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
	db := c.MustGet("db").(*sqlx.DB)
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
	db := c.MustGet("db").(*sqlx.DB)
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
