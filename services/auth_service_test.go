package services
//testing is a framework
//fataf-error indeki excution nirkum

import (
	"testing"
	"github.com/AlinAlexMyladoor/ecommerce-backend/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Setup in-memory DB for testing
func setupTestDB() *gorm.DB {

	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	db.AutoMigrate(&models.User{})

	return db
}

// Helper to create mock user
func createMockUser(db *gorm.DB, password string) models.User {

	user := models.User{
		Email: "test@example.com",
		Role:  "user",
	}

	user.HashPassword(password)

	db.Create(&user)

	return user
}

func TestLoginSuccess(t *testing.T) {

	db := setupTestDB()

	user := createMockUser(db, "password123")

	service := AuthService{DB: db}

	token, err := service.Login(user.Email, "password123")

	if err != nil {
		t.Fatalf("expected success but got error: %v", err)
	}

	if token == "" {
		t.Errorf("expected token but got empty string")
	}
}


func TestLoginWrongPassword(t *testing.T) {

	db := setupTestDB()

	user := createMockUser(db, "password123")

	service := AuthService{DB: db}

	_, err := service.Login(user.Email, "wrongpass")

	if err == nil {
		t.Errorf("expected error but got nil")
	}
}


func TestLoginUserNotFound(t *testing.T) {

	db := setupTestDB()

	service := AuthService{DB: db}

	_, err := service.Login("nouser@test.com", "password123")

	if err == nil {
		t.Errorf("expected error but got nil")
	}
}