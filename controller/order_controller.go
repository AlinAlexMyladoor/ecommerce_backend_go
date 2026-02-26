// controllers/order_controller.go
//transaction-oneki full ayitu nadakum alenki onum nadakila
package controller

import (
	"fmt"
	"net/http"
	"strings"
	"github.com/AlinAlexMyladoor/ecommerce-backend/Databse"
	"github.com/AlinAlexMyladoor/ecommerce-backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

)

// OrderItemInput represents one item in the posted order
//oronu ayitu store akum
type OrderItemInput struct {
    ProductID uint `json:"product_id" binding:"required"`
    Quantity  int  `json:"quantity" binding:"required,min=1"`
}

// OrderCreateInput is the body for creating an order
//dive means oro prouct check cheyan,lisintinte ulil uladine check cehyum
type OrderCreateInput struct {//list
    Items []OrderItemInput `json:"items" binding:"required,dive"`
    // optionally you can accept shipping info, etc.
}

// POST /orders
func CreateOrder(c *gin.Context) {
    // get user_id from the middleware (assuming middleware sets it)
    userIDRaw, exists := c.Get("user_id")
    if !exists {
        JSONError(c, http.StatusUnauthorized, "user not found in context")
        return
    }
    userID := uint(userIDRaw.(float64))//int formatil store akum,convert cheyan

    var input OrderCreateInput
    if err := c.ShouldBindJSON(&input); err != nil {
        JSONError(c, http.StatusBadRequest, "invalid request body: "+err.Error())
        return
    }

    db := Databse.DB//access cheyn

    // Use a DB transaction
    err := db.Transaction(func(tx *gorm.DB) error {
        // create order record
        order := models.Order{
            UserID: userID,
            Status: "created",
        }
        if err := tx.Create(&order).Error; err != nil {
            return err
        }

        var total float64

        // For each item: check stock, create order_item, update product stock
        for _, it := range input.Items {//items listil oro products indavum,itil-product store cheyum,oro items,product id,quantity store cheyum
            var product models.Product
            if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, it.ProductID).Error; err != nil {//same product 2 peru orde cheydal,first  use akitu 
                return err // product not found -> rollback
            }

            if product.Stock < it.Quantity {
                return fmt.Errorf("product %d out of stock", product.ID)
            }

            price := product.Price//ipolathe price store cheyum,order itemil price store cheyum,price update cheyyan kazhiyum
            item := models.Order_item{//ororo item
                OrderID:   order.ID,
                ProductID: product.ID,
                Quantity:  it.Quantity,
                Price:     price,
            }
            if err := tx.Create(&item).Error; err != nil {
                return err
            }//create cheyyum

            // deduct stock
            product.Stock = product.Stock - it.Quantity
            if err := tx.Save(&product).Error; err != nil {//update cheyyum,save cheyyum
                return err
            }

            total += price * float64(it.Quantity)
        }

        // update order total
        if err := tx.Model(&order).Update("total_amount", total).Error; err != nil {
            return err
        }

        // success -> commit
        return nil//idoke full transactionil anu.err alenki nil,err aneki full roll back-elam normal aka,nil aneki suceess
    })
    if err != nil {
        // if your error text includes "out of stock", convert to 400
        if strings.Contains(err.Error(), "out of stock") {
            JSONError(c, http.StatusBadRequest, err.Error())
            return
        }
        JSONError(c, http.StatusInternalServerError, "failed to create order: "+err.Error())
        return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "order created"})
}

// Helper function to send consistent JSON error responses
func JSONError(c *gin.Context, status int, message string) {
    c.JSON(status, gin.H{
        "error": message,
    })
}