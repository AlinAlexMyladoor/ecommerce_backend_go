package services

	

import (
	"errors"

	"github.com/AlinAlexMyladoor/ecommerce-backend/models"
	"github.com/AlinAlexMyladoor/ecommerce-backend/utils"
	"gorm.io/gorm"
)

type AuthService struct {
	DB *gorm.DB
}

// Login handles authentication logic
func (s *AuthService) Login(email string, password string) (string, error) {

	var user models.User

	// 1️⃣ Find user by email
	if err := s.DB.Where("email = ?", email).First(&user).Error; err != nil {
		return "", errors.New("invalid email or password")
	}

	// 2️⃣ Compare password using model method
	if err := user.CheckPassword(password); err != nil {
		return "", errors.New("invalid email or password")
	}

	// 3️⃣ Generate JWT
	token, err := utils.Generatetoken(user.ID, user.Role)
	if err != nil {
		return "", err
	}

	return token, nil
}