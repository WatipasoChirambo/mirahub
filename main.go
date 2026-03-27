package main

import (
	"log"
	"mirahub/routes"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	// Build DSN
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		host := os.Getenv("DB_HOST")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		name := os.Getenv("DB_NAME")
		port := os.Getenv("DB_PORT")

		if host == "" {
			log.Fatal("No database configuration provided")
		}

		dsn = "postgres://" + user + ":" + password +
			"@" + host + ":" + port + "/" + name + "?sslmode=disable"
	}

	// Connect to DB
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	log.Println("Connected to DB")

	// Gin router
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Make db available in handlers via context
	r.Use(func(c *gin.Context) {
		c.Set("db", db.DB) // <- db.DB is *sql.DB
		c.Next()
	})

	// PASS db to routes
	routes.SetupRoutes(r, db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server running on port", port)
	r.Run(":" + port)
}
