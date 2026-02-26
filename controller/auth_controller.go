//gin is framework
//arrange-arrange cchyn,act-run,assert-verify--test pattern
package controller
import (
	"net/http"
	"github.com/AlinAlexMyladoor/ecommerce-backend/Databse" // Import the new DB package
	"github.com/AlinAlexMyladoor/ecommerce-backend/models"
	
	"github.com/gin-gonic/gin"//json
	"github.com/AlinAlexMyladoor/ecommerce-backend/services"
)
type AuthInput struct {
	Email    string `json:"email" binding:"required,email"`//binding-mandatory,emil anonu check cheyan.
	Password string `json:"password" binding:"required"`
}




func Signup(c *gin.Context){//request access cheyyan c
	var input AuthInput//bind cheayyan input

	if err := c.ShouldBindJSON(&input); err != nil {//json email,email map akum captial email store akum
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user1 := models.User{
		Email: input.Email,
	}//input email,user object store cheyunu

	if err := user1.HashPassword(input.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}
	
	if err := Databse.DB.Create(&user1).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user already exist"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "user created"})
}



func Login(c *gin.Context) {
	var input AuthInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service := services.AuthService{
		DB: Databse.DB,
	}

	token, err := service.Login(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
