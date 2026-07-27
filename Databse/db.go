package Databse

import (
	"fmt"
	"log"
	"os"

	"github.com/AlinAlexMyladoor/ecommerce-backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)
//env var like use case,os.getenv
var DB *gorm.DB

func InitDB() {

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB.AutoMigrate(
		&models.Product{},
		&models.User{},
		&models.CartItem{},
		&models.Order{},
		&models.Order_item{},
	)

	log.Println("Database connected and migrated!")
}