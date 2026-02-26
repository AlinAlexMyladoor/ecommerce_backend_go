package models

import "gorm.io/gorm"
type Order struct{
	gorm.Model
	UserID uint `json:"user_id"`
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	TotalAmount float64 `json:"total_amount"`
	Status string  `json:"status"` //pending,shipped,delivered,cancelled
	Items []Order_item `gorm:"foreignKey:OrderID" json:"order_items",omitempty` //one to many relationship

}
