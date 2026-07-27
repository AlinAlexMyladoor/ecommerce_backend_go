package utils
import (
	"time"
	"github.com/golang-jwt/jwt/v5"
)
var secretKey = []byte("jhfskgyevchvevcy")

func GenerateToken(userId uint, role string) (string, error) {
	// 1. Define claims (payload)
	claims := jwt.MapClaims{
		"user_id": userId,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), 
	}

	// 2. Create token with claims and signing method
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 3. Sign the token with our secret key
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}