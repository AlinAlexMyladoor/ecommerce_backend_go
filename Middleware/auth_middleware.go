package middleware

import (
	"net/http"
	"strings"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("jhfskgyevchvevcy")


func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")
		if authHeader == ""{
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return secretKey, nil
        })

		if err != nil || !token.Valid {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            return
        }

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
            c.Set("user_id", claims["user_id"]) // Crucial: We use this later in Cart
            c.Set("role", claims["role"])
        }

        c.Next()
    }
}

// RBAC: Middleware to check if user is Admin
func RequireAdmin() gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get("role")
        if !exists || role != "admin" {
            c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admins only"})
            return
        }
        c.Next()
    }
}


// package middleware
// //2 middle ware,for login anonu nokan and 1 for admin or user
// //middelware use akumbbo c.next use akanm.
// import (
// 	"net/http"
// 	"strings"
// 	"fmt"
// 	"github.com/gin-gonic/gin"
// 	"github.com/golang-jwt/jwt/v5"
// )

// var secretKey = []byte("super-secret-key-change-me")
// //gin is a framework like (express in nodejs) in golang

// //funct type anu gin.HandlerFunc

// func AuthMiddleware() gin.HandlerFunc {//user authenticate cheyyan-jwt authenticate cheyyan-string
// 	return func(c *gin.Context) {//request,response
// 		//header ninnu token edukkan bearer variable for authization,Authorization: Bearer eyJhbGciOiJIUzI1NiIs..., authorization,bearer,jwt token
// 		//authorizationil ula headeril token save akan in aithhreader

// 		authHeader := c.GetHeader("Authorization")
// 		if authHeader == ""{
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
// 			return
// 		}
// //separet bearer and jwt token,split btween uathheader and space
// 		parts := strings.Split(authHeader, " ")
// 		//partsil bearer and jwt toke

// 		if len(parts) != 2 || parts[0] != "Bearer" {
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
// 			return
// 		}

// 		tokenString := parts[1]
// 		//kitiya token,crt anonu nokan funct
// 		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
//             if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
//                 return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
//             }
//             return secretKey, nil
// 			//crt anki secret key return,nil means no error
// 			//token will be store in token and err
// 			//token error indeki,tokenil nil and err ee function
//         })

// 		if err != nil || !token.Valid {
//             c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
//             return
//         }
// //user id,role oke store avum in claims,request store akum before going to controller
// 		if claims, ok := token.Claims.(jwt.MapClaims); ok {
//             c.Set("user_id", claims["user_id"]) // Crucial: We use this later in Cart
//             c.Set("role", claims["role"])
//         }

//         c.Next()//controlleril pokan
//     }
// }

// // RBAC: Middleware to check if user is Admin
// func RequireAdmin() gin.HandlerFunc {
//     return func(c *gin.Context) {
//         role, exists := c.Get("role")
// 		//role indeki roll store akum,,illenki existsindavila.
// 		//2nd mdlware anu,first mdwlareil set cheytha role edukkan
//         if !exists || role != "admin" {
//             c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admins only"})
//             return
//         }
//         c.Next()//controlleril pokan
//     }
// }