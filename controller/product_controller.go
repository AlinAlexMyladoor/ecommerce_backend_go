package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/AlinAlexMyladoor/ecommerce-backend/Databse" // Import the new DB package
	"github.com/AlinAlexMyladoor/ecommerce-backend/models"
	"github.com/gin-gonic/gin"
)

//pagination-sort akitu page wise splitting
// GET /products
func GetProducts(c *gin.Context) {//c ilu reponse,request varum
    db := Databse.DB

    // parse query params
    //page parameter indeki,aa page pagestr store akum,page ilenki deafault 1 ayie dukum,default query
    pageStr := c.DefaultQuery("page", "1")//default query=if pathil ilneki empty ayrilum,query=if pathil indeki searchil store akum,urlil epolum string ayrikum
    limitStr := c.DefaultQuery("limit", "10")//default query=if pathil ilneki empty ayrilum,query=if pathil indeki searchil store akum,urlil epolum string ayrikum
    sort := c.DefaultQuery("sort", "created_desc")
    search := c.Query("search")

    page, _ := strconv.Atoi(pageStr)//string to int convert akum,default query,string conersion alphabet to integer,uderscore means ignore
    limit, _ := strconv.Atoi(limitStr)
    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 10
    }
    offset := (page - 1) * limit//offset eviduna satrt akendedenu

    var products []models.Product//type anu models.product,databseinu varundu store akum
    query := db.Model(&models.Product{})//all products kitan query=selext * equivalnet anu query
//search
    if search != "" {
        // simple search on name/description
        //like is a string 
        like := "%" + strings.ToLower(search) + "%"
        query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", like, like)
    }

    // sorting
    switch sort {
    case "price_asc":
        query = query.Order("price ASC")
    case "price_desc":
        query = query.Order("price DESC")
    case "name_asc":
        query = query.Order("name ASC")
    case "name_desc":
        query = query.Order("name DESC")
    default:
        query = query.Order("created_at DESC")
    }

    var total int64
    if err := query.Count(&total).Error; err != nil {
        JSONError(c, http.StatusInternalServerError, "failed to count products")
        return
    }

    if err := query.Offset(offset).Limit(limit).Find(&products).Error; err != nil {
        JSONError(c, http.StatusInternalServerError, "failed to fetch products")
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "ok": true,
        "data": products,
        "meta": gin.H{
            "page":  page,
            "limit": limit,
            "total": total,//no of products
        },
    })
}

// GET /products/:id
func GetProductByID(c *gin.Context) {
    id := c.Param("id")
    var product models.Product
    
    if err := Databse.DB.First(&product, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }   
    c.JSON(http.StatusOK, product)
}

// POST /products (Admin Only)
func CreateProduct(c *gin.Context) {
    var newProduct models.Product

    if err := c.ShouldBindJSON(&newProduct); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    result := Databse.DB.Create(&newProduct)
    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, newProduct)
}

// PUT /products/:id (Admin Only)
func UpdateProduct(c *gin.Context) {
    id := c.Param("id")
    var product models.Product
    
    // Check if product exists
    if err := Databse.DB.First(&product, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }

    // Bind new data
    var updatedData models.Product
    if err := c.ShouldBindJSON(&updatedData); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Update
    Databse.DB.Model(&product).Updates(updatedData)
    c.JSON(http.StatusOK, product)
}

// DELETE /products/:id (Admin Only)
func DeleteProduct(c *gin.Context) {
    id := c.Param("id")
    var product models.Product
    
    if err := Databse.DB.First(&product, id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }

    Databse.DB.Delete(&product)
    c.JSON(http.StatusOK, gin.H{"message": "Product deleted"})
}