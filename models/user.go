package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	
)

type User struct{
	gorm.Model//creates ID,CreatedAt,UpdatedAt,DeletedAt
	Email string`gorm:"uniqueIndex;not null"json:"email"`
	Password string`gorm:"not null"json:"-"`
	Role string`gorm:"default:user"json:"role"` //user or admin
	Orders []Order `gorm:"foreignKey:UserID" json:"orders",omitempty` //one to many relationship,orders history story akan.
}
func(u*User) HashPassword(Password string) error{//8address anu eduka
	bytes,err:=bcrypt.GenerateFromPassword([]byte(Password),14)
	if err!=nil{
		return err
	}
	u.Password=string(bytes)
	return nil

}
func(u*User)CheckPassword(inputPassword string) error{
	err:=bcrypt.CompareHashAndPassword([]byte(u.Password),[]byte(inputPassword))
	if err!=nil{
		return err
	}
	return nil 
}