package main

import (
	"github.com/AlinAlexMyladoor/ecommerce-backend/Databse"
	"github.com/AlinAlexMyladoor/ecommerce-backend/Middleware"
	"github.com/AlinAlexMyladoor/ecommerce-backend/controller"
	"github.com/AlinAlexMyladoor/ecommerce-backend/workers"
	"github.com/gin-gonic/gin"
	
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	Databse.InitDB()
	Databse.ConnectRedis()
	Databse.ConnectRabbitMQ()
	Databse.ConnectTypesense()
	go workers.StartInvoiceWorker()
	
	r := gin.Default()
	_ = r.SetTrustedProxies(nil)
	// --- Public Routes ---

	r.POST("/signup", controller.Signup)
	r.POST("/login", controller.Login)
	r.GET("/products", controller.GetProducts) // Anyone can view products
	r.GET("/products/:id", controller.GetProductByID)
	r.GET("/search", controller.SearchProducts)

      

	// --- Protected Routes (User) ---
	// User must be logged in to access Cart
	userGroup := r.Group("/")
	userGroup.Use(middleware.AuthMiddleware())
	{
		userGroup.POST("/cart", controller.AddToCart)
		userGroup.GET("/cart", controller.GetCart)
		userGroup.POST("/orders", controller.CreateOrder)
	}

	// --- Protected Routes (Admin) ---
	// User must be logged in AND have role="admin" to change products
	adminGroup := r.Group("/admin")
	adminGroup.Use(middleware.AuthMiddleware(), middleware.RequireAdmin())
	{
		adminGroup.POST("/products", controller.CreateProduct)
		adminGroup.PUT("/products/:id", controller.UpdateProduct)
		adminGroup.DELETE("/products/:id", controller.DeleteProduct)
	}
	r.Run(":8080")
}
