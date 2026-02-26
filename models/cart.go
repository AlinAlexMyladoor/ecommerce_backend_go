package models

import "gorm.io/gorm"

type CartItem struct {
    gorm.Model
    UserID    uint    `json:"user_id"`
    ProductID uint    `json:"product_id"`
    Product   Product `gorm:"foreignKey:ProductID"` // Relation
    Quantity  int     `json:"quantity"`
}