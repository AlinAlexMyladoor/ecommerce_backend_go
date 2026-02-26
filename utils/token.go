package utils
import (
	"time"
	"github.com/golang-jwt/jwt/v5"
)
var secretKey = []byte("jhfskgyevchvevcy")
func Generatetoken(userId uint,role string)(string,error){
	claims:=jwt.MapClaims{"user_id": userId,"esp": time.Now().Add(time.Hour*24).Unix(),
	"role":role,
	//token create alumbo pokenda data

}
token :=jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
signedToken,err:=token.SignedString(secretKey)
if err!=nil{
	return "", err
}
return signedToken, nil
}
