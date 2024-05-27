package main

import (
	"fmt"
	"log"
	"net/http"
	"order-service/events"
	"order-service/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetAllOrders(c *gin.Context) {
	queryParams := c.MustGet("queryParams").(map[string]string)
	page, _ := strconv.Atoi(queryParams["page"])
	count, _ := strconv.Atoi(queryParams["count"])
	search := queryParams["search"]

	orders, totalElements, totalPages, err := models.GetOrders(page, count, search)
	if err != nil {
		fmt.Println("error occurred")
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to get orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orders":        orders,
		"totalElements": totalElements,
		"totalPages":    totalPages,
	})
}

func CreateOrder(c *gin.Context) {
	var order models.Order
	err := c.ShouldBindJSON(&order)

	fmt.Println(order)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse data"})
		return
	}

	newOrder, err := order.SaveOrder()

	fmt.Println("we finished saving the order")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not save order"})
		return
	}

	// Publish order created event
	err = events.PublishOrderCreatedEvent(newOrder)
	if err != nil {
		log.Println("Failed to publish order created event:", err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Order created",
		"order":   newOrder,
	})
}

func UpdateOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse order ID"})
		return
	}

	var updatedOrder models.Order
	err := c.ShouldBindJSON(&updatedOrder)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Could not parse data"})
		return
	}

	err = updatedOrder.UpdateOrder(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Could not update order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order updated successfully",
		"order":   updatedOrder,
	})
}

func GetOrderByID(c *gin.Context) {
	orderID := c.Param("id")
	order, err := models.GetOrderByID(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}
