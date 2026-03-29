package routes

import (
	"mirahub/handlers"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// SetupRoutes sets up all API routes
func SetupRoutes(r *gin.Engine, db *sqlx.DB) {
	// Public routes
	public := r.Group("/api")
	{
		public.POST("/auth/register", handlers.Register)
		public.POST("/auth/login", handlers.Login)
		public.POST("/auth/logout", handlers.Logout)

		// Sales
		// public.POST("/sales", handlers.CreateSale)
		// public.GET("/sales", handlers.GetSales)

		// Products
		public.POST("/products", handlers.CreateProduct)
		public.PUT("/products/:id", handlers.UpdateProduct)
		public.DELETE("/products/:id", handlers.DeleteProduct)

		// Categories
		public.POST("/categories", handlers.CreateCategory)
		public.PUT("/categories/:id", handlers.UpdateCategory)
		public.DELETE("/categories/:id", handlers.DeleteCategory)

		// Suppliers
		public.POST("/suppliers", handlers.CreateSupplier)
		public.PUT("/suppliers/:id", handlers.UpdateSupplier)
		public.DELETE("/suppliers/:id", handlers.DeleteSupplier)

		// Warehouses
		public.POST("/warehouses", handlers.CreateWarehouse)
		public.PUT("/warehouses/:id", handlers.UpdateWarehouse)
		public.DELETE("/warehouses/:id", handlers.DeleteWarehouse)

		// Invoices
		public.POST("/invoices", handlers.CreateInvoice)
		public.GET("/invoices", handlers.GetInvoices)

		// Quotations
		public.POST("/quotations", handlers.CreateQuotation)
		public.GET("/quotations", handlers.GetQuotations)

		// Receipts
		public.POST("/receipts", handlers.CreateReceipt)
		public.GET("/receipts", handlers.GetReceipts)

		public.POST("/seed-all", func(c *gin.Context) {
			handlers.SeedAll(c, db)
		})

		public.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Welcome to Mirahub API"})
		})

		// Wrap SeedProducts to pass db
		public.POST("/seed-products", func(c *gin.Context) {
			handlers.SeedProducts(c, db)
		})

		public.GET("/products", handlers.GetProducts)
		public.GET("/categories", handlers.GetCategories)
		public.GET("/suppliers", handlers.GetSuppliers)
		public.GET("/warehouses", handlers.GetWarehouses)
	}

	// Authenticated routes
	api := r.Group("/api")
	// 🔒 Protected routes
	protected := api.Group("/")
	protected.Use(handlers.AuthMiddleware())
	{
		protected.POST("/sales", handlers.CreateSale)
		protected.GET("/sales", handlers.GetSales)
	}
}
