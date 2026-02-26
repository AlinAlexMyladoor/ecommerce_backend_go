package services

import (
	"fmt"
	"github.com/AlinAlexMyladoor/ecommerce-backend/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderService struct {
	DB *gorm.DB
}

type OrderItemInput struct {
	ProductID uint
	Quantity  int
}

func (s *OrderService) CreateOrder(userID uint, items []OrderItemInput) error {

	return s.DB.Transaction(func(tx *gorm.DB) error {

		// 1️⃣ Create order
		order := models.Order{
			UserID: userID,
			Status: "created",
		}

		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		var total float64

		// 2️⃣ Process items
		for _, it := range items {

			var product models.Product

			// Row-level lock
			if err := tx.
				Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&product, it.ProductID).Error; err != nil {
				return err
			}

			// Stock validation
			if product.Stock < it.Quantity {
				return fmt.Errorf("product %d out of stock", product.ID)
			}

			// Create order item
			item := models.Order_item{
				OrderID:   order.ID,
				ProductID: product.ID,
				Quantity:  it.Quantity,
				Price:     product.Price,
			}

			if err := tx.Create(&item).Error; err != nil {
				return err
			}

			// Deduct stock
			product.Stock -= it.Quantity
			if err := tx.Save(&product).Error; err != nil {
				return err
			}

			total += product.Price * float64(it.Quantity)
		}

		// Update total
		return tx.Model(&order).
			Update("total_amount", total).Error
	})
}