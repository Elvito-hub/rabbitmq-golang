package main

import (
	"context"
	"inventory/connections"
	"inventory/db"
	"inventory/events"
	"log"

	"github.com/gin-gonic/gin"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
func main() {
	client, err := db.GetMongoClient()

	conn, err := connections.Connect()

	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = client.Disconnect(context.Background()); err != nil {
			log.Fatal(err)
		}
	}()
	server := gin.Default()

	server.POST("/products", CreateProduct)
	server.PUT("/products/:id", UpdateProduct)
	server.GET("/products", GetAllProducts)
	server.GET("/products/:id", GetProductByID)

	events.ConsumeOrderCreatedEvent()

	server.Run(":1001")
}
