package models

import "gorm.io/gorm"
 type Order_item struct{
gorm.Model
OrderID uint `json:"order_id"`

ProductID uint `json:"product_id"`
Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`//one to many relationship
Quantity int 	`json:"quantity"`
Price float64 `json:"price"`//order products many to many

	
}