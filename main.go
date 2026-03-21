package main

import (
	"fmt"
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
	// dbHost := os.Getenv("DB_HOST")
	// dbUser := os.Getenv("DB_USER")
	// dbPassword := os.Getenv("DB_PASSWORD")
	// dbName := os.Getenv("DB_NAME")
	// dbPort := os.Getenv("DB_PORT")

	dsn := os.Getenv("DATABASE_URL")

db, err := sqlx.Connect("postgres", dsn)
if err != nil {
    log.Fatal("Database connection failed:", err)
}

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()
	log.Println("Connected to DB")

	r := gin.Default()

	// CORS config
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"}, // your Nuxt dev URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Use(func(c *gin.Context) {
		c.Set("db", db.DB) // if using sqlx, db.DB is *sql.DB
		c.Next()
	})
	routes.SetupRoutes(r)
	r.Run(":8080")
}
