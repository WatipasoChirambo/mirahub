package routes

import (
	"mirahub/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	public := r.Group("/api")
	{
		public.POST("/auth/register", handlers.Register)
		public.POST("/auth/login", handlers.Login)
		public.POST("/auth/logout", handlers.Logout)

		public.GET("/products", handlers.GetProducts)
		public.GET("/categories", handlers.GetCategories)
		public.GET("/suppliers", handlers.GetSuppliers)
		public.GET("/warehouses", handlers.GetWarehouses)
	}

	api := r.Group("/api")
	api.Use(handlers.AuthMiddleware())
	{
		//products
		api.POST("/products", handlers.CreateProduct)
		api.PUT("/products/:id", handlers.UpdateProduct)
		api.DELETE("/products/:id", handlers.DeleteProduct)

		// Categories
		api.POST("/categories", handlers.CreateCategory)
		api.PUT("/categories/:id", handlers.UpdateCategory)
		api.DELETE("/categories/:id", handlers.DeleteCategory)

		// Suppliers
		api.POST("/suppliers", handlers.CreateSupplier)
		api.PUT("/suppliers/:id", handlers.UpdateSupplier)
		api.DELETE("/suppliers/:id", handlers.DeleteSupplier)

		// Warehouses
		api.POST("/warehouses", handlers.CreateWarehouse)
		api.PUT("/warehouses/:id", handlers.UpdateWarehouse)
		api.DELETE("/warehouses/:id", handlers.DeleteWarehouse)

		// Sales
		api.POST("/sales", handlers.CreateSale)
		api.GET("/sales", handlers.GetSales)

		// Invoices
		api.POST("/invoices", handlers.CreateInvoice)
		api.GET("/invoices", handlers.GetInvoices)

		// Quotations
		api.POST("/quotations", handlers.CreateQuotation)
		api.GET("/quotations", handlers.GetQuotations)

		// Receipts
		api.POST("/receipts", handlers.CreateReceipt)
		api.GET("/receipts", handlers.GetReceipts)
	}
}
