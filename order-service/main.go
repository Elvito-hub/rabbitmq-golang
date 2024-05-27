package main

import (
	"context"
	"log"
	"order-service/connections"
	"order-service/db"

	"github.com/gin-gonic/gin"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
func main() {
	client, err := db.GetMongoClient()

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()
	server := gin.Default()

	conn, err := connections.Connect()

	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	server.GET("/orders", GetAllOrders)
	server.POST("/orders", CreateOrder)
	server.PUT("/orders/:id", UpdateOrder)
	server.GET("/orders/:id", GetOrderByID)

	server.Run(":1000")
}
