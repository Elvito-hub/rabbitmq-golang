package main

import (
	"fmt"
	"inventory/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAllProducts(c *gin.Context) {
	queryParams := c.MustGet("queryParams").(map[string]string)
	page, _ := strconv.Atoi(queryParams["page"])
	count, _ := strconv.Atoi(queryParams["count"])
	search := queryParams["search"]

	products, totalElements, totalPages, err := models.GetProducts(page, count, search)
	if err != nil {
		fmt.Println("error occurred")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to get products"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products":      products,
		"totalElements": totalElements,
		"totalPages":    totalPages,
	})
}

func CreateProduct(c *gin.Context) {
	var product models.Product
	err := c.ShouldBindJSON(&product)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse data"})
		return
	}

	newProduct, err := product.SaveProduct()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not save product"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Product created",
		"product": newProduct,
	})
}

func UpdateProduct(c *gin.Context) {
	productID := c.Param("id")
	if productID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse product ID"})
		return
	}

	var updatedProduct models.Product
	err := c.ShouldBindJSON(&updatedProduct)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse data"})
		return
	}

	err = updatedProduct.UpdateProduct(productID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not update product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product updated successfully",
		"product": updatedProduct,
	})
}

func GetProductByID(c *gin.Context) {
	productID := c.Param("id")
	product, err := models.GetProductByID(productID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, product)
}
