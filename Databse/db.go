package Databse

import (
	"log"

	"github.com/AlinAlexMyladoor/ecommerce-backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Global DB variable (Exported)
var DB *gorm.DB

func InitDB() {
	// Moved your connection string here
	dsn := "host=localhost user=postgres password=postgres dbname=ecommerce port=5432 sslmode=disable"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// AutoMigrate both Product and User
	DB.AutoMigrate(&models.Product{}, &models.User{}, &models.CartItem{}, &models.Order{}, &models.Order_item{})
	log.Println("Database connected and migrated!")
}
