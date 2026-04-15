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

		public.POST("/contact", handlers.ContactHandler)
		public.POST("/quote", handlers.QuoteHandler)

		public.GET("/products", handlers.GetProducts)

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

		public.GET("/vehicles", handlers.GetVehicles)
		public.GET("/vehicle/:id", handlers.GetVehicle)

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
		protected.POST("/vehicle", handlers.CreateVehicle)

		// Products
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

		protected.GET("/customers", handlers.GetCustomers)
		protected.GET("/customers/:id", handlers.GetCustomerByID)
		protected.POST("/customers", handlers.CreateCustomer)
		protected.PUT("/customers/:id", handlers.UpdateCustomer)
		protected.DELETE("/customers/:id", handlers.DeleteCustomer)

		protected.GET("/orders", handlers.GetOrders)

		// Vehicle attachments for products
		protected.POST("/products/:id/vehicles", handlers.AttachVehicle(db))
		protected.DELETE("/products/:id/vehicles", handlers.DetachVehicle(db))

		// ========== NEW: File Attachment Routes ==========

		// Upload file for documents (invoice, quotation, receipt, order)
		protected.POST("/upload/:type/:id", handlers.UploadFileAttachment)

		// Get all attachments for a document
		protected.GET("/attachments/:type/:id", handlers.GetDocumentAttachments)

		// Delete a specific attachment
		protected.DELETE("/attachments/:attachment_id", handlers.DeleteFileAttachment)

		// ========== NEW: Document Status Update Routes ==========

		// Update invoice status (draft, sent, paid, overdue, cancelled)
		protected.PUT("/invoices/:id/status", handlers.UpdateInvoiceStatus)

		// Update quotation status (draft, sent, accepted, rejected, expired)
		protected.PUT("/quotations/:id/status", handlers.UpdateQuotationStatus)

		// ========== NEW: Enhanced Document Routes with File Support ==========

		// Get single invoice with attachments
		protected.GET("/invoices/:id", handlers.GetInvoiceByID)

		// Get single quotation with attachments
		protected.GET("/quotations/:id", handlers.GetQuotationByID)

		// Get single receipt with attachments
		protected.GET("/receipts/:id", handlers.GetReceiptByID)

		// Generate PDF for invoice
		protected.POST("/invoices/:id/generate-pdf", handlers.GenerateInvoicePDF)

		// Generate PDF for quotation
		protected.POST("/quotations/:id/generate-pdf", handlers.GenerateQuotationPDF)

		// Send invoice via email
		protected.POST("/invoices/:id/send-email", handlers.SendInvoiceEmail)

		// Send quotation via email
		protected.POST("/quotations/:id/send-email", handlers.SendQuotationEmail)
	}
}
