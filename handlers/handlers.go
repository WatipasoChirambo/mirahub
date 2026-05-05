package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

// Update your ReceiptWithItemsRequest struct to include UserID
type ReceiptWithItemsRequest struct {
	CustomerName  string `json:"customer_name"`
	CustomerEmail string `json:"customer_email"`
	CustomerPhone string `json:"customer_phone"`
	UserID        int    `json:"user_id"`
	Items         []struct {
		ProductID   int     `json:"product_id"`
		Name        string  `json:"name"`
		Code        string  `json:"code"`
		Quantity    int     `json:"quantity"`
		Price       float64 `json:"price"`
		Description string  `json:"description"`
	} `json:"items"`
	Subtotal      float64 `json:"subtotal"`
	TaxRate       float64 `json:"tax_rate"`
	TaxAmount     float64 `json:"tax_amount"`
	Discount      float64 `json:"discount"`
	Total         float64 `json:"total"`
	Notes         string  `json:"notes"`
	ReceiptNumber string  `json:"receipt_number"`
	PaymentMethod string  `json:"payment_method"`
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
		From:    "quote@mirahubautoparts.com",
		To:      []string{emailUser},
		Cc:      []string{"elizabeth.chabaluka@gmail.com"},
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

	if err := c.ShouldBindJSON(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Resend API key not set"})
		return
	}

	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    "contact@mirahubautoparts.com",
		To:      []string{os.Getenv("EMAIL_USER")},
		Cc:      []string{"elizabeth.chabaluka@gmail.com"},
		Subject: form.Subject,
		ReplyTo: form.Email,
		Text: fmt.Sprintf(
			"Name: %s\nEmail: %s\n\nMessage:\n%s",
			form.Name, form.Email, form.Message,
		),
	}

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

func SeedProducts(c *gin.Context, db *sqlx.DB) {
	createdBy := 1

	// Use sql.NullInt64 for nullable fields
	products := []struct {
		Code        string
		Name        string
		CategoryID  sql.NullInt64
		SupplierID  sql.NullInt64
		WarehouseID sql.NullInt64
		Stock       int
		CreatedBy   *int
	}{
		{
			Code:        "P001",
			Name:        "Laptop",
			CategoryID:  sql.NullInt64{Int64: 1, Valid: true},
			SupplierID:  sql.NullInt64{Int64: 1, Valid: true},
			WarehouseID: sql.NullInt64{Int64: 1, Valid: true},
			Stock:       10,
			CreatedBy:   &createdBy,
		},
		{
			Code:        "P002",
			Name:        "Keyboard",
			CategoryID:  sql.NullInt64{Int64: 1, Valid: true},
			SupplierID:  sql.NullInt64{Int64: 1, Valid: true},
			WarehouseID: sql.NullInt64{Int64: 1, Valid: true},
			Stock:       15,
			CreatedBy:   &createdBy,
		},
		{
			Code:        "P003",
			Name:        "Mouse",
			CategoryID:  sql.NullInt64{Int64: 1, Valid: true},
			SupplierID:  sql.NullInt64{Int64: 1, Valid: true},
			WarehouseID: sql.NullInt64{Int64: 1, Valid: true},
			Stock:       20,
			CreatedBy:   &createdBy,
		},
		{
			Code:        "P004",
			Name:        "Monitor",
			CategoryID:  sql.NullInt64{Int64: 1, Valid: true},
			SupplierID:  sql.NullInt64{Int64: 1, Valid: true},
			WarehouseID: sql.NullInt64{Int64: 1, Valid: true},
			Stock:       5,
			CreatedBy:   &createdBy,
		},
		{
			Code:        "P005",
			Name:        "Printer",
			CategoryID:  sql.NullInt64{Int64: 1, Valid: true},
			SupplierID:  sql.NullInt64{Int64: 1, Valid: true},
			WarehouseID: sql.NullInt64{Int64: 1, Valid: true},
			Stock:       8,
			CreatedBy:   &createdBy,
		},
	}

	tx, err := db.Beginx()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to start transaction: " + err.Error()})
		return
	}
	defer tx.Rollback()

	for _, p := range products {
		_, err := tx.Exec(`
			INSERT INTO products (code, name, category_id, supplier_id, warehouse_id, stock, created_by) 
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (code) DO NOTHING`,
			p.Code,
			p.Name,
			p.CategoryID,
			p.SupplierID,
			p.WarehouseID,
			p.Stock,
			p.CreatedBy,
		)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to insert product " + p.Code + ": " + err.Error()})
			return
		}
	}

	if err := tx.Commit(); err != nil {
		c.JSON(500, gin.H{"error": "Failed to commit transaction: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "5 products seeded successfully!"})
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	db := c.MustGet("db").(*sqlx.DB)

	var user models.User

	err := db.QueryRow(`
        SELECT id, username, email, phone, password_hash, role, created_at
        FROM users
        WHERE username=$1 OR email=$1 OR phone=$1
    `, req.Identifier).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Phone,
		&user.Password,
		&user.Role,
		&user.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// generate JWT
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

	// cookies
	c.SetCookie("token", tokenString, 86400, "/", "", false, true)
	c.SetCookie("user_id", fmt.Sprintf("%d", user.ID), 86400, "/", "", false, true)

	// IMPORTANT: never return password
	user.Password = ""

	// response
	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"user":  user,
	})
}

func Logout(c *gin.Context) {
	// Clear the token cookie
	c.SetCookie(
		"token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)

	// Clear the user_id cookie
	c.SetCookie(
		"user_id",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// -------------------- Middleware --------------------

// CORSMiddleware handles CORS settings
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // Your frontend URL
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// Try to get token from Authorization header first
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString = parts[1]
			} else {
				tokenString = authHeader
			}
		}

		// If not in header, try to get from cookie
		if tokenString == "" {
			cookie, err := c.Cookie("token")
			if err == nil && cookie != "" {
				tokenString = cookie
			}
		}

		// If still no token, return unauthorized
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authentication token"})
			c.Abort()
			return
		}

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "JWT secret not configured"})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
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

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

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

// DebugAuth endpoint for testing authentication
func DebugAuth(c *gin.Context) {
	userID := c.GetInt("user_id")
	username := c.GetString("username")
	role := c.GetString("role")

	// Get cookies for debugging
	tokenCookie, _ := c.Cookie("token")
	userIDCookie, _ := c.Cookie("user_id")

	c.JSON(200, gin.H{
		"authenticated":      userID != 0,
		"user_id":            userID,
		"username":           username,
		"role":               role,
		"token_cookie":       tokenCookie != "",
		"token_cookie_value": tokenCookie[:min(20, len(tokenCookie))] + "...",
		"user_id_cookie":     userIDCookie,
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestProducts - Debug endpoint
func TestProducts(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	// Get raw products
	rows, err := db.Query("SELECT id, code, name, hold, item_code FROM products LIMIT 5")
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var products []map[string]interface{}
	for rows.Next() {
		var id int
		var code, name, hold, itemCode string
		rows.Scan(&id, &code, &name, &hold, &itemCode)
		products = append(products, map[string]interface{}{
			"id":        id,
			"code":      code,
			"name":      name,
			"hold":      hold,
			"item_code": itemCode,
		})
	}

	c.JSON(200, gin.H{
		"products": products,
		"count":    len(products),
	})
}

func GetProducts(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	query := `
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
	`

	rows, err := db.Query(query)
	if err != nil {
		log.Printf("Query error: %v", err)
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	// Create a response struct to properly format the output
	type ProductResponse struct {
		ID          int      `json:"id"`
		Code        string   `json:"code"`
		Name        string   `json:"item"`
		ItemCode    string   `json:"item_code"`
		Hold        string   `json:"hold"`
		Stock       int      `json:"quantity"`
		Price       float64  `json:"price"`
		ImageURL    string   `json:"image"`
		CategoryID  *int     `json:"category_id,omitempty"`
		SupplierID  *int     `json:"supplier_id,omitempty"`
		WarehouseID *int     `json:"warehouse_id,omitempty"`
		Vehicles    []string `json:"vehicles"`
	}

	productMap := make(map[int]*ProductResponse)

	for rows.Next() {
		var (
			id          int
			code        string
			name        string
			categoryID  sql.NullInt64
			supplierID  sql.NullInt64
			warehouseID sql.NullInt64
			stock       int
			price       float64
			hold        sql.NullString
			itemCode    sql.NullString
			createdBy   sql.NullInt64
			imageURL    sql.NullString
			vehicleID   sql.NullInt64
			vehicleName sql.NullString
		)

		err := rows.Scan(
			&id,
			&code,
			&name,
			&categoryID,
			&supplierID,
			&warehouseID,
			&stock,
			&price,
			&hold,
			&itemCode,
			&imageURL,
			&createdBy,
			&vehicleID,
			&vehicleName,
		)
		if err != nil {
			log.Printf("Scan error: %v", err)
			c.JSON(500, gin.H{"error": "Scan failed: " + err.Error()})
			return
		}

		p, exists := productMap[id]
		if !exists {
			// Convert NULL values to empty strings or nil pointers
			holdStr := ""
			if hold.Valid {
				holdStr = hold.String
			}

			itemCodeStr := ""
			if itemCode.Valid {
				itemCodeStr = itemCode.String
			}

			imageURLStr := ""
			if imageURL.Valid {
				imageURLStr = imageURL.String
			}

			// Handle nullable IDs
			var catID *int
			if categoryID.Valid {
				idVal := int(categoryID.Int64)
				catID = &idVal
			}

			var supID *int
			if supplierID.Valid {
				idVal := int(supplierID.Int64)
				supID = &idVal
			}

			var whID *int
			if warehouseID.Valid {
				idVal := int(warehouseID.Int64)
				whID = &idVal
			}

			p = &ProductResponse{
				ID:          id,
				Code:        code,
				Name:        name,
				ItemCode:    itemCodeStr,
				Hold:        holdStr,
				Stock:       stock,
				Price:       price,
				ImageURL:    imageURLStr,
				CategoryID:  catID,
				SupplierID:  supID,
				WarehouseID: whID,
				Vehicles:    []string{},
			}
			productMap[id] = p
		}

		if vehicleID.Valid && vehicleID.Int64 > 0 && vehicleName.Valid {
			p.Vehicles = append(p.Vehicles, vehicleName.String)
		}
	}

	if err = rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		c.JSON(500, gin.H{"error": "Rows error: " + err.Error()})
		return
	}

	products := make([]ProductResponse, 0, len(productMap))
	for _, p := range productMap {
		products = append(products, *p)
	}

	log.Printf("Successfully fetched %d products", len(products))
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

// GetInvoiceByID retrieves a single invoice with its attachments
func GetInvoiceByID(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	id := c.Param("id")

	var invoice models.Invoice
	var pdfPath sql.NullString
	var pdfGeneratedAt sql.NullTime
	var dueDate sql.NullTime

	err := db.QueryRow(`
		SELECT id, sale_id, user_id, invoice_date, total, status, due_date, pdf_path, pdf_generated_at
		FROM invoices WHERE id = $1
	`, id).Scan(&invoice.ID, &invoice.SaleID, &invoice.UserID, &invoice.InvoiceDate,
		&invoice.Total, &invoice.Status, &dueDate, &pdfPath, &pdfGeneratedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{"error": "Invoice not found"})
		} else {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}

	if dueDate.Valid {
		invoice.DueDate = &dueDate.Time
	}
	if pdfPath.Valid {
		invoice.PDFPath = &pdfPath.String
	}
	if pdfGeneratedAt.Valid {
		invoice.PDFGeneratedAt = &pdfGeneratedAt.Time
	}

	// Fetch attachments
	var attachments []models.FileAttachment
	db.Select(&attachments, `
		SELECT id, document_type, document_id, file_name, file_path, file_size, mime_type, uploaded_by, uploaded_at, description
		FROM file_attachments
		WHERE document_type = 'invoice' AND document_id = $1
	`, id)
	invoice.Attachments = attachments

	c.JSON(200, gin.H{"invoice": invoice})
}

// GetQuotationByID retrieves a single quotation with its attachments
func GetQuotationByID(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	id := c.Param("id")

	var quotation models.Quotation
	var validUntil sql.NullTime
	var notes sql.NullString
	var pdfPath sql.NullString
	var pdfGeneratedAt sql.NullTime

	err := db.QueryRow(`
		SELECT id, product_id, user_id, quote_date, price, status, valid_until, notes, pdf_path, pdf_generated_at
		FROM quotations WHERE id = $1
	`, id).Scan(&quotation.ID, &quotation.ProductID, &quotation.UserID, &quotation.QuoteDate,
		&quotation.Price, &quotation.Status, &validUntil, &notes, &pdfPath, &pdfGeneratedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{"error": "Quotation not found"})
		} else {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}

	if validUntil.Valid {
		quotation.ValidUntil = &validUntil.Time
	}
	if notes.Valid {
		quotation.Notes = &notes.String
	}
	if pdfPath.Valid {
		quotation.PDFPath = &pdfPath.String
	}
	if pdfGeneratedAt.Valid {
		quotation.PDFGeneratedAt = &pdfGeneratedAt.Time
	}

	// Fetch attachments
	var attachments []models.FileAttachment
	db.Select(&attachments, `
		SELECT id, document_type, document_id, file_name, file_path, file_size, mime_type, uploaded_by, uploaded_at, description
		FROM file_attachments
		WHERE document_type = 'quotation' AND document_id = $1
	`, id)
	quotation.Attachments = attachments

	c.JSON(200, gin.H{"quotation": quotation})
}

// GetReceiptByID retrieves a single receipt with its attachments
func GetReceiptByID(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	id := c.Param("id")

	var receipt models.Receipt
	var referenceNo sql.NullString
	var notes sql.NullString
	var pdfPath sql.NullString
	var pdfGeneratedAt sql.NullTime

	err := db.QueryRow(`
		SELECT id, invoice_id, user_id, receipt_date, amount, payment_method, reference_no, notes, pdf_path, pdf_generated_at
		FROM receipts WHERE id = $1
	`, id).Scan(&receipt.ID, &receipt.InvoiceID, &receipt.UserID, &receipt.ReceiptDate,
		&receipt.Amount, &receipt.PaymentMethod, &referenceNo, &notes, &pdfPath, &pdfGeneratedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{"error": "Receipt not found"})
		} else {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}

	if referenceNo.Valid {
		receipt.ReferenceNo = &referenceNo.String
	}
	if notes.Valid {
		receipt.Notes = &notes.String
	}
	if pdfPath.Valid {
		receipt.PDFPath = &pdfPath.String
	}
	if pdfGeneratedAt.Valid {
		receipt.PDFGeneratedAt = &pdfGeneratedAt.Time
	}

	// Fetch attachments
	var attachments []models.FileAttachment
	db.Select(&attachments, `
		SELECT id, document_type, document_id, file_name, file_path, file_size, mime_type, uploaded_by, uploaded_at, description
		FROM file_attachments
		WHERE document_type = 'receipt' AND document_id = $1
	`, id)
	receipt.Attachments = attachments

	c.JSON(200, gin.H{"receipt": receipt})
}

// GenerateInvoicePDF generates a PDF for an invoice
func GenerateInvoicePDF(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	id := c.Param("id")

	// Fetch invoice data
	var invoice models.Invoice
	err := db.Get(&invoice, "SELECT * FROM invoices WHERE id = $1", id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Invoice not found"})
		return
	}

	// Fetch sale details
	var sale models.Sale
	err = db.Get(&sale, "SELECT * FROM sales WHERE id = $1", invoice.SaleID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Sale not found"})
		return
	}

	// Fetch product details
	var product models.Product
	err = db.Get(&product, "SELECT * FROM products WHERE id = $1", sale.ProductID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Product not found"})
		return
	}

	// Generate PDF logic here
	// This is a placeholder - you'll need to implement actual PDF generation
	pdfPath := fmt.Sprintf("./uploads/invoices/invoice_%s.pdf", id)

	// Update invoice with PDF path
	_, err = db.Exec(`
		UPDATE invoices 
		SET pdf_path = $1, pdf_generated_at = $2 
		WHERE id = $3
	`, pdfPath, time.Now(), id)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update invoice with PDF path"})
		return
	}

	c.JSON(200, gin.H{
		"message":  "PDF generated successfully",
		"pdf_path": pdfPath,
		"invoice":  invoice,
		"sale":     sale,
		"product":  product,
	})
}

// GenerateQuotationPDF generates a PDF for a quotation
func GenerateQuotationPDF(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	id := c.Param("id")

	// Fetch quotation data
	var quotation models.Quotation
	err := db.Get(&quotation, "SELECT * FROM quotations WHERE id = $1", id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Quotation not found"})
		return
	}

	// Fetch product details
	var product models.Product
	err = db.Get(&product, "SELECT * FROM products WHERE id = $1", quotation.ProductID)
	if err != nil {
		c.JSON(404, gin.H{"error": "Product not found"})
		return
	}

	// Generate PDF logic here
	pdfPath := fmt.Sprintf("./uploads/quotations/quotation_%s.pdf", id)

	// Update quotation with PDF path
	_, err = db.Exec(`
		UPDATE quotations 
		SET pdf_path = $1, pdf_generated_at = $2 
		WHERE id = $3
	`, pdfPath, time.Now(), id)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update quotation with PDF path"})
		return
	}

	c.JSON(200, gin.H{
		"message":   "PDF generated successfully",
		"pdf_path":  pdfPath,
		"quotation": quotation,
		"product":   product,
	})
}

// SendInvoiceEmail sends an invoice via email
func SendInvoiceEmail(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	id := c.Param("id")

	var input struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Email is required"})
		return
	}

	// Fetch invoice with PDF
	var invoice models.Invoice
	err := db.Get(&invoice, "SELECT * FROM invoices WHERE id = $1", id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Invoice not found"})
		return
	}

	if invoice.PDFPath == nil {
		c.JSON(400, gin.H{"error": "PDF not generated yet. Please generate PDF first."})
		return
	}

	// Email sending logic here
	// You can use the resend client similar to ContactHandler

	c.JSON(200, gin.H{
		"message": fmt.Sprintf("Invoice sent to %s", input.Email),
	})
}

// SendQuotationEmail sends a quotation via email
func SendQuotationEmail(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	id := c.Param("id")

	var input struct {
		Email string `json:"email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Email is required"})
		return
	}

	// Fetch quotation with PDF
	var quotation models.Quotation
	err := db.Get(&quotation, "SELECT * FROM quotations WHERE id = $1", id)
	if err != nil {
		c.JSON(404, gin.H{"error": "Quotation not found"})
		return
	}

	if quotation.PDFPath == nil {
		c.JSON(400, gin.H{"error": "PDF not generated yet. Please generate PDF first."})
		return
	}

	// Email sending logic here

	c.JSON(200, gin.H{
		"message": fmt.Sprintf("Quotation sent to %s", input.Email),
	})
}

func GetVehicles(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	var vehicles []models.Vehicle

	err := db.Select(&vehicles, `
        SELECT id, name 
        FROM vehicles 
        ORDER BY name
    `)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"vehicles": vehicles})
}

func GetVehicle(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	id := c.Param("id")

	var vehicle models.Vehicle

	err := db.Get(&vehicle, `
        SELECT id, name 
        FROM vehicles 
        WHERE id = $1
    `, id)

	if err != nil {
		c.JSON(404, gin.H{"error": "Vehicle not found"})
		return
	}

	c.JSON(200, gin.H{"vehicle": vehicle})
}

func CreateVehicle(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	var input struct {
		Name string `json:"name"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON"})
		return
	}

	if input.Name == "" {
		c.JSON(400, gin.H{"error": "Name is required"})
		return
	}

	var id int
	err := db.QueryRow(`
        INSERT INTO vehicles (name)
        VALUES ($1)
        RETURNING id
    `, input.Name).Scan(&id)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"message": "Vehicle created",
		"vehicle": gin.H{
			"id":   id,
			"name": input.Name,
		},
	})
}

func CreateProduct(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	code := c.PostForm("code")
	itemCode := c.PostForm("item_code")
	name := c.PostForm("name")

	userID := c.GetInt("user_id")

	if userID == 0 {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	categoryID := c.PostForm("category_id")
	supplierID := c.PostForm("supplier_id")
	warehouseID := c.PostForm("warehouse_id")
	stock := c.PostForm("stock")
	price := c.PostForm("price")
	hold := c.PostForm("hold")
	createdBy := c.PostForm("created_by")

	file, _ := c.FormFile("image")
	var imageURL string

	if file != nil {
		// Create directory if not exists
		if err := os.MkdirAll("./uploads/products", 0755); err != nil {
			c.JSON(500, gin.H{"error": "Failed to create upload directory"})
			return
		}

		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), file.Filename)
		path := "./uploads/products/" + filename
		if err := c.SaveUploadedFile(file, path); err != nil {
			c.JSON(500, gin.H{"error": "image upload failed"})
			return
		}
		imageURL = "/uploads/products/" + filename
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

	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Row iteration failed: " + err.Error(),
		})
		return
	}

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

	for _, item := range input.Items {
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

		_, err = tx.Exec(`
			UPDATE products SET stock = stock - $1 WHERE id = $2
		`, item.Quantity, item.ProductID)

		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to update stock"})
			return
		}

		_, err = tx.Exec(`
			INSERT INTO order_items (order_id, product_id, quantity, price)
			VALUES ($1, $2, $3, $4)
		`, orderID, item.ProductID, item.Quantity, item.Price)

		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to insert order items"})
			return
		}
	}

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

	var err error

	userID := c.GetInt("user_id")
	if userID == 0 {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	var input struct {
		ProductID  int `json:"product_id"`
		CustomerID int `json:"customer_id"`
		Quantity   int `json:"quantity"`
	}

	if bindErr := c.ShouldBindJSON(&input); bindErr != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	if input.ProductID <= 0 || input.Quantity <= 0 {
		c.JSON(400, gin.H{"error": "Invalid product_id or quantity"})
		return
	}

	if input.CustomerID <= 0 {
		c.JSON(400, gin.H{"error": "Invalid customer_id"})
		return
	}

	tx, err := db.Begin()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to start transaction"})
		return
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var stock int
	var productPrice float64

	err = tx.QueryRow(`
        SELECT stock, price
        FROM products 
        WHERE id = $1 
        FOR UPDATE
    `, input.ProductID).Scan(&stock, &productPrice)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(404, gin.H{"error": "Product not found"})
		} else {
			c.JSON(500, gin.H{"error": err.Error()})
		}
		return
	}

	if input.Quantity > stock {
		c.JSON(400, gin.H{"error": "Insufficient stock"})
		return
	}

	_, err = tx.Exec(`
        UPDATE products 
        SET stock = stock - $1 
        WHERE id = $2
    `, input.Quantity, input.ProductID)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	total := productPrice * float64(input.Quantity)

	var saleID int
	var saleDate time.Time

	err = tx.QueryRow(`
        INSERT INTO sales (product_id, customer_id, user_id, quantity, price, total)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, sale_date
    `, input.ProductID, input.CustomerID, userID, input.Quantity, productPrice, total).
		Scan(&saleID, &saleDate)

	if err != nil {
		c.JSON(500, gin.H{
			"error":   "Insert failed",
			"details": err.Error(),
		})
		return
	}

	err = tx.Commit()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"id":          saleID,
		"product_id":  input.ProductID,
		"customer_id": input.CustomerID,
		"user_id":     userID,
		"quantity":    input.Quantity,
		"price":       productPrice,
		"total":       total,
		"sale_date":   saleDate,
	})
}

// CreateInvoice with file attachment support
func CreateInvoice(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	userID := c.GetInt("user_id")

	var input models.InvoiceRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	var invoiceID int
	var invoiceDate time.Time

	err := db.QueryRow(`
		INSERT INTO invoices(sale_id, user_id, total, status, due_date)
		VALUES($1, $2, $3, $4, $5)
		RETURNING id, invoice_date
	`, input.SaleID, userID, input.Total, input.Status, input.DueDate).Scan(&invoiceID, &invoiceDate)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create invoice: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Invoice created",
		"id":           invoiceID,
		"invoice_date": invoiceDate,
	})
}

// CreateQuotation with file attachment support
func CreateQuotation(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	userID := c.GetInt("user_id")

	var input struct {
		ProductID  int        `json:"product_id"`
		Price      float64    `json:"price"`
		ValidUntil *time.Time `json:"valid_until"`
		Notes      string     `json:"notes"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	var quoteID int
	var quoteDate time.Time

	err := db.QueryRow(`
		INSERT INTO quotations(product_id, user_id, price, valid_until, notes, status)
		VALUES($1, $2, $3, $4, $5, 'draft')
		RETURNING id, quote_date
	`, input.ProductID, userID, input.Price, input.ValidUntil, input.Notes).Scan(&quoteID, &quoteDate)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create quotation: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Quotation created",
		"id":         quoteID,
		"quote_date": quoteDate,
	})
}

func CreateReceipt(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	// Get user_id from context (set by AuthMiddleware)
	userID := c.GetInt("user_id")
	if userID == 0 {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse request body
	var req struct {
		ReceiptNumber string  `json:"receipt_number"` // reference_no in DB
		Date          string  `json:"date"`           // receipt_date in DB
		CustomerID    int     `json:"customer_id"`
		Subtotal      float64 `json:"subtotal"`
		TaxRate       float64 `json:"tax_rate"`
		TaxAmount     float64 `json:"tax_amount"`
		Discount      float64 `json:"discount"`
		Total         float64 `json:"total"`
		PaymentMethod string  `json:"payment_method"`
		Notes         string  `json:"notes"`
		Status        string  `json:"status"`
		InvoiceID     *int    `json:"invoice_id"` // optional invoice reference
		Items         []struct {
			ProductID   int     `json:"product_id"`
			ProductName string  `json:"product_name"`
			ProductCode string  `json:"product_code"`
			Quantity    int     `json:"quantity"`
			Price       float64 `json:"price"`
			TotalPrice  float64 `json:"total_price"`
			Description string  `json:"description"`
		} `json:"items"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Validate required fields
	if len(req.Items) == 0 {
		c.JSON(400, gin.H{"error": "At least one item is required"})
		return
	}

	if req.CustomerID <= 0 {
		c.JSON(400, gin.H{"error": "Valid customer_id is required"})
		return
	}

	// Set default values if not provided
	if req.Status == "" {
		req.Status = "completed"
	}
	if req.PaymentMethod == "" {
		req.PaymentMethod = "cash"
	}
	if req.Date == "" {
		req.Date = time.Now().Format(time.RFC3339)
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to start transaction"})
		return
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Check stock availability for all items first
	for _, item := range req.Items {
		var stock int
		err = tx.QueryRow(`
			SELECT stock
			FROM products 
			WHERE id = $1 
			FOR UPDATE
		`, item.ProductID).Scan(&stock)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(404, gin.H{"error": fmt.Sprintf("Product not found: ID %d", item.ProductID)})
			} else {
				c.JSON(500, gin.H{"error": err.Error()})
			}
			return
		}

		if item.Quantity > stock {
			c.JSON(400, gin.H{
				"error":           "Insufficient stock",
				"product_id":      item.ProductID,
				"product_name":    item.ProductName,
				"available_stock": stock,
				"requested":       item.Quantity,
			})
			return
		}
	}

	// Insert receipt - matching your exact schema
	var receiptID int
	var receiptDate time.Time

	err = tx.QueryRow(`
		INSERT INTO receipts (
			invoice_id,
			user_id,
			receipt_date,
			amount,
			payment_method,
			reference_no,
			notes,
			customer_id,
			subtotal,
			tax_rate,
			tax_amount,
			discount,
			total,
			status,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, NOW())
		RETURNING id, receipt_date
	`,
		nullIntForIDPtr(req.InvoiceID), // invoice_id (can be NULL)
		userID,                         // user_id
		req.Date,                       // receipt_date
		req.Total,                      // amount (using total as amount)
		req.PaymentMethod,              // payment_method
		req.ReceiptNumber,              // reference_no
		req.Notes,                      // notes
		req.CustomerID,                 // customer_id
		req.Subtotal,                   // subtotal
		req.TaxRate,                    // tax_rate
		req.TaxAmount,                  // tax_amount
		req.Discount,                   // discount
		req.Total,                      // total
		req.Status,                     // status
	).Scan(&receiptID, &receiptDate)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create receipt: " + err.Error()})
		return
	}

	// Insert receipt items and update stock
	for _, item := range req.Items {
		// Get product details if not provided
		productName := item.ProductName
		productCode := item.ProductCode

		if productName == "" || productCode == "" {
			var dbProduct struct {
				Name string
				Code string
			}
			err = tx.QueryRow(`
				SELECT name, COALESCE(code, '') as code 
				FROM products WHERE id = $1
			`, item.ProductID).Scan(&dbProduct.Name, &dbProduct.Code)

			if err == nil {
				if productName == "" {
					productName = dbProduct.Name
				}
				if productCode == "" {
					productCode = dbProduct.Code
				}
			}
		}

		// Insert receipt item
		_, err = tx.Exec(`
			INSERT INTO receipt_items 
			(receipt_id, product_id, product_name, product_code, quantity, price, total, description)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`,
			receiptID,
			item.ProductID,
			productName,
			productCode,
			item.Quantity,
			item.Price,
			item.TotalPrice,
			item.Description,
		)

		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to create receipt items: " + err.Error()})
			return
		}

		// Update product stock
		result, err := tx.Exec(`
			UPDATE products 
			SET stock = stock - $1 
			WHERE id = $2 AND stock >= $1
		`,
			item.Quantity,
			item.ProductID,
		)

		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to update product stock: " + err.Error()})
			return
		}

		// Check if stock was actually updated
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(400, gin.H{
				"error":      "Stock update failed - insufficient stock",
				"product_id": item.ProductID,
			})
			return
		}
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to commit receipt: " + err.Error()})
		return
	}

	// Return success response
	c.JSON(201, gin.H{
		"message":        "Receipt created successfully",
		"id":             receiptID,
		"receipt_number": req.ReceiptNumber,
		"receipt_date":   receiptDate,
		"customer_id":    req.CustomerID,
		"subtotal":       req.Subtotal,
		"tax_rate":       req.TaxRate,
		"tax_amount":     req.TaxAmount,
		"discount":       req.Discount,
		"total":          req.Total,
		"payment_method": req.PaymentMethod,
		"status":         req.Status,
		"created_by":     userID,
		"items_count":    len(req.Items),
	})
}

// Helper function for nullable invoice_id
func nullIntForIDPtr(i *int) interface{} {
	if i == nil || *i == 0 {
		return nil
	}
	return *i
}

// Helper function for null strings
func nullString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// UploadFileAttachment for documents
func UploadFileAttachment(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	userID := c.GetInt("user_id")

	documentType := c.Param("type") // invoice, quotation, receipt, order
	documentID := c.Param("id")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "File is required"})
		return
	}

	// Validate file type
	allowedTypes := []string{"application/pdf", "image/jpeg", "image/png", "application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"}
	contentType := file.Header.Get("Content-Type")
	allowed := false
	for _, t := range allowedTypes {
		if t == contentType {
			allowed = true
			break
		}
	}
	if !allowed {
		c.JSON(400, gin.H{"error": "Invalid file type. Allowed: PDF, JPEG, PNG, DOC, DOCX"})
		return
	}

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		c.JSON(400, gin.H{"error": "File too large. Max size 10MB"})
		return
	}

	// Create directory if not exists
	dir := fmt.Sprintf("./uploads/%s/%s", documentType, documentID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(500, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%s%s", time.Now().Unix(), documentID, ext)
	filePath := filepath.Join(dir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(500, gin.H{"error": "Failed to save file"})
		return
	}

	// Save file record to database
	description := c.PostForm("description")

	var attachmentID int
	err = db.QueryRow(`
		INSERT INTO file_attachments (document_type, document_id, file_name, file_path, file_size, mime_type, uploaded_by, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, documentType, documentID, file.Filename, filePath, file.Size, contentType, userID, description).Scan(&attachmentID)

	if err != nil {
		// Delete the file if database insert fails
		os.Remove(filePath)
		c.JSON(500, gin.H{"error": "Failed to save file record: " + err.Error()})
		return
	}

	// Return file URL
	fileURL := fmt.Sprintf("/uploads/%s/%s/%s", documentType, documentID, filename)

	c.JSON(200, gin.H{
		"message":   "File uploaded successfully",
		"id":        attachmentID,
		"file_name": file.Filename,
		"file_url":  fileURL,
		"file_size": file.Size,
		"mime_type": contentType,
	})
}

// GetDocumentAttachments retrieves attachments for a document
func GetDocumentAttachments(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	documentType := c.Param("type")
	documentID := c.Param("id")

	var attachments []models.FileAttachment

	err := db.Select(&attachments, `
		SELECT id, document_type, document_id, file_name, file_path, file_size, mime_type, uploaded_by, uploaded_at, description
		FROM file_attachments
		WHERE document_type = $1 AND document_id = $2
		ORDER BY uploaded_at DESC
	`, documentType, documentID)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to fetch attachments: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"attachments": attachments,
	})
}

// DeleteFileAttachment removes a file attachment
func DeleteFileAttachment(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	attachmentID := c.Param("attachment_id")

	// Get file path first
	var filePath string
	err := db.QueryRow(`
		SELECT file_path FROM file_attachments WHERE id = $1
	`, attachmentID).Scan(&filePath)

	if err != nil {
		c.JSON(404, gin.H{"error": "Attachment not found"})
		return
	}

	// Delete from database
	_, err = db.Exec(`
		DELETE FROM file_attachments WHERE id = $1
	`, attachmentID)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete attachment record"})
		return
	}

	// Delete physical file
	if err := os.Remove(filePath); err != nil {
		// Log error but don't fail the request
		log.Printf("Warning: Failed to delete file %s: %v", filePath, err)
	}

	c.JSON(200, gin.H{
		"message": "Attachment deleted successfully",
	})
}

// -------------------- Sales --------------------

type CreateSaleRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

func GetCategories(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	rows, _ := db.Query("SELECT id, name, description FROM categories")
	defer rows.Close()

	var list []models.Category
	for rows.Next() {
		var m models.Category
		rows.Scan(&m.ID, &m.Name, &m.Description)
		list = append(list, m)
	}
	c.JSON(http.StatusOK, gin.H{"categories": list})
}

func CreateCategory(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	var m models.Category
	c.ShouldBindJSON(&m)
	db.Exec("INSERT INTO categories(name, description) VALUES($1, $2)", m.Name, m.Description)
	c.JSON(http.StatusCreated, gin.H{"message": "created"})
}

func UpdateCategory(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	id := c.Param("id")
	var m models.Category
	c.ShouldBindJSON(&m)
	db.Exec("UPDATE categories SET name=$1, description=$2 WHERE id=$3", m.Name, m.Description, id)
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
			s.total,  
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
			&s.Total,
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

	if sales == nil {
		sales = []models.SaleResponse{}
	}

	c.JSON(200, sales)
}

// GetInvoices with attachments
func GetInvoices(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	rows, err := db.Query(`
		SELECT id, sale_id, user_id, invoice_date, total, status, due_date, pdf_path, pdf_generated_at
		FROM invoices
		ORDER BY id DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	invoices := []models.Invoice{}
	for rows.Next() {
		var inv models.Invoice
		var pdfPath sql.NullString
		var pdfGeneratedAt sql.NullTime
		var dueDate sql.NullTime

		if err := rows.Scan(&inv.ID, &inv.SaleID, &inv.UserID, &inv.InvoiceDate, &inv.Total,
			&inv.Status, &dueDate, &pdfPath, &pdfGeneratedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if dueDate.Valid {
			inv.DueDate = &dueDate.Time
		}
		if pdfPath.Valid {
			inv.PDFPath = &pdfPath.String
		}
		if pdfGeneratedAt.Valid {
			inv.PDFGeneratedAt = &pdfGeneratedAt.Time
		}

		// Fetch attachments for this invoice
		var attachments []models.FileAttachment
		db.Select(&attachments, `
			SELECT id, document_type, document_id, file_name, file_path, file_size, mime_type, uploaded_by, uploaded_at, description
			FROM file_attachments
			WHERE document_type = 'invoice' AND document_id = $1
		`, inv.ID)
		inv.Attachments = attachments

		invoices = append(invoices, inv)
	}

	c.JSON(http.StatusOK, gin.H{"invoices": invoices})
}

// GetQuotations with attachments
func GetQuotations(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	rows, err := db.Query(`
		SELECT id, product_id, user_id, quote_date, price, status, valid_until, notes, pdf_path, pdf_generated_at
		FROM quotations
		ORDER BY id DESC
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	quotes := []models.Quotation{}
	for rows.Next() {
		var q models.Quotation
		var validUntil sql.NullTime
		var notes sql.NullString
		var pdfPath sql.NullString
		var pdfGeneratedAt sql.NullTime

		if err := rows.Scan(&q.ID, &q.ProductID, &q.UserID, &q.QuoteDate, &q.Price,
			&q.Status, &validUntil, &notes, &pdfPath, &pdfGeneratedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if validUntil.Valid {
			q.ValidUntil = &validUntil.Time
		}
		if notes.Valid {
			q.Notes = &notes.String
		}
		if pdfPath.Valid {
			q.PDFPath = &pdfPath.String
		}
		if pdfGeneratedAt.Valid {
			q.PDFGeneratedAt = &pdfGeneratedAt.Time
		}

		// Fetch attachments
		var attachments []models.FileAttachment
		db.Select(&attachments, `
			SELECT id, document_type, document_id, file_name, file_path, file_size, mime_type, uploaded_by, uploaded_at, description
			FROM file_attachments
			WHERE document_type = 'quotation' AND document_id = $1
		`, q.ID)
		q.Attachments = attachments

		quotes = append(quotes, q)
	}

	c.JSON(http.StatusOK, gin.H{"quotations": quotes})
}

// GetReceipts with attachments and items
func GetReceipts(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)

	rows, err := db.Query(`
		SELECT 
			r.id,
			COALESCE(r.reference_no, '') as receipt_number,
			r.receipt_date,
			COALESCE(c.name, 'Walk-in Customer') as customer_name,
			COALESCE(c.email, '') as customer_email,
			COALESCE(r.subtotal, 0) as subtotal,
			COALESCE(r.tax_rate, 0) as tax_rate,
			COALESCE(r.tax_amount, 0) as tax_amount,
			COALESCE(r.discount, 0) as discount,
			r.total,
			r.notes,
			r.status,
			r.pdf_path
		FROM receipts r
		LEFT JOIN customers c ON c.id = r.customer_id
		ORDER BY r.id DESC
	`)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	receipts := []map[string]interface{}{}

	for rows.Next() {
		var (
			id, receiptNumber, receiptDate, customerName, customerEmail string
			subtotal, taxRate, taxAmount, discount, total               float64
			notes, status, pdfPath                                      sql.NullString
		)

		err := rows.Scan(&id, &receiptNumber, &receiptDate, &customerName, &customerEmail,
			&subtotal, &taxRate, &taxAmount, &discount, &total, &notes, &status, &pdfPath)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get items for this receipt
		items := []map[string]interface{}{}
		itemRows, err := db.Query(`
			SELECT product_name, product_code, quantity, price, total
			FROM receipt_items
			WHERE receipt_id = $1
		`, id)

		if err == nil {
			for itemRows.Next() {
				var productName, productCode string
				var quantity int
				var price, itemTotal float64
				itemRows.Scan(&productName, &productCode, &quantity, &price, &itemTotal)
				items = append(items, map[string]interface{}{
					"name":         productName,
					"code":         productCode,
					"quantity":     quantity,
					"sellingPrice": price,
					"total":        itemTotal,
				})
			}
			itemRows.Close()
		}

		receipt := map[string]interface{}{
			"id":             id,
			"receipt_number": receiptNumber,
			"receipt_date":   receiptDate,
			"customer_name":  customerName,
			"customer_email": customerEmail,
			"subtotal":       subtotal,
			"tax_rate":       taxRate,
			"tax_amount":     taxAmount,
			"discount":       discount,
			"total":          total,
			"notes":          notes.String,
			"status":         status.String,
			"pdf_path":       pdfPath.String,
			"items":          items,
			"items_count":    len(items),
		}

		receipts = append(receipts, receipt)
	}

	c.JSON(http.StatusOK, gin.H{"receipts": receipts})
}

// UpdateInvoiceStatus updates invoice status
func UpdateInvoiceStatus(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	id := c.Param("id")

	var input struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	_, err := db.Exec(`
		UPDATE invoices SET status = $1 WHERE id = $2
	`, input.Status, id)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update invoice status"})
		return
	}

	c.JSON(200, gin.H{"message": "Invoice status updated"})
}

// UpdateQuotationStatus updates quotation status
func UpdateQuotationStatus(c *gin.Context) {
	db := c.MustGet("db").(*sqlx.DB)
	id := c.Param("id")

	var input struct {
		Status string `json:"status"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	_, err := db.Exec(`
		UPDATE quotations SET status = $1 WHERE id = $2
	`, input.Status, id)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update quotation status"})
		return
	}

	c.JSON(200, gin.H{"message": "Quotation status updated"})
}
