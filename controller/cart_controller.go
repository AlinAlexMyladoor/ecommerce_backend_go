package controller
//add to cart and get cart items

import (
    "net/http"
    "github.com/AlinAlexMyladoor/ecommerce-backend/Databse"
    "github.com/AlinAlexMyladoor/ecommerce-backend/models"
    "github.com/gin-gonic/gin"
)

type CartInput struct {
    ProductID uint `json:"product_id" binding:"required"`
    Quantity  int  `json:"quantity" binding:"required"`
}

func AddToCart(c *gin.Context) {
    // 1. Get User ID from Context (set by Middleware)

    userIDRaw, exists := c.Get("user_id")

    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
        return
    }
    // Type assertion because c.Get returns interface{}
    userID := uint(userIDRaw.(float64)) // JWT numbers often parse as float64

    var input CartInput
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 2. Check if product exists
    var product models.Product//pruduct id nokiu indeki &product alenki error
    if err := Databse.DB.First(&product, input.ProductID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }
	

    // 3. Add or Update Cart Item
    var cartItem models.CartItem
    result := Databse.DB.Where("user_id = ? AND product_id = ?", userID, input.ProductID).First(&cartItem)

    if result.Error == nil {
        // Item exists, update quantity
        cartItem.Quantity += input.Quantity
        Databse.DB.Save(&cartItem)
    } else {
        // Create new
        newItem := models.CartItem{
            UserID:    userID,
            ProductID: input.ProductID,
            Quantity:  input.Quantity,
        }
        Databse.DB.Create(&newItem)
    }

    c.JSON(http.StatusOK, gin.H{"message": "Item added to cart"})
}

func GetCart(c *gin.Context) {
    userIDRaw, _ := c.Get("user_id")
    userID := uint(userIDRaw.(float64))

    var items []models.CartItem
    // Preload fetches the associated Product data
    Databse.DB.Preload("Product").Where("user_id = ?", userID).Find(&items)

    c.JSON(http.StatusOK, items)
}