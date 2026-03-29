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

		public.POST("/orders", handlers.CreateOrder)

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

		// public.GET("/products", handlers.GetProducts)
		// public.GET("/categories", handlers.GetCategories)
		// public.GET("/suppliers", handlers.GetSuppliers)
		// public.GET("/warehouses", handlers.GetWarehouses)
	}

	// Authenticated routes
	api := r.Group("/api")
	// 🔒 Protected routes
	protected := api.Group("/")
	protected.Use(handlers.AuthMiddleware())
	{
		// Sales
		protected.POST("/sales", handlers.CreateSale)
		protected.GET("/sales", handlers.GetSales)

		// Products
		protected.GET("/products", handlers.GetProducts)
		protected.POST("/products", handlers.CreateProduct)
		protected.PUT("/products/:id", handlers.UpdateProduct)
		protected.DELETE("/products/:id", handlers.DeleteProduct)

		// Categories
		protected.GET("/categories", handlers.GetCategories)
		protected.POST("/categories", handlers.CreateCategory)
		protected.PUT("/categories/:id", handlers.UpdateCategory)
		protected.DELETE("/categories/:id", handlers.DeleteCategory)

		// Suppliers
		protected.GET("/suppliers", handlers.GetSuppliers)
		protected.POST("/suppliers", handlers.CreateSupplier)
		protected.PUT("/suppliers/:id", handlers.UpdateSupplier)
		protected.DELETE("/suppliers/:id", handlers.DeleteSupplier)

		// Warehouses
		protected.GET("/warehouses", handlers.GetWarehouses)
		protected.POST("/warehouses", handlers.CreateWarehouse)
		protected.PUT("/warehouses/:id", handlers.UpdateWarehouse)
		protected.DELETE("/warehouses/:id", handlers.DeleteWarehouse)

		protected.POST("/customers", handlers.CreateCustomer)
		protected.GET("/customers", handlers.GetCustomers)

		protected.GET("/orders", handlers.GetOrders)
	}
}
