package main

import (

	"github.com/AlinAlexMyladoor/ecommerce-backend/controller"
	"github.com/AlinAlexMyladoor/ecommerce-backend/Databse"
	"github.com/AlinAlexMyladoor/ecommerce-backend/Middleware"
	"github.com/gin-gonic/gin"
)



func main() {
	Databse.InitDB()
	r := gin.Default()//gin.locker = error 500 convert akan,or 400,default use akiyondu 
	// --- Public Routes ---
    //admin,sign up,sign p avathrou

	r.POST("/signup", controller.Signup)
    r.POST("/login", controller.Login)
    r.GET("/products", controller.GetProducts)       // Anyone can view products

    r.GET("/products/:id", controller.GetProductByID)


    // --- Protected Routes (User) ---
    // User must be logged in to access Cart
    userGroup := r.Group("/")//prefix in url     
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