package events

import (
	"encoding/json"
	"fmt"
	"inventory/connections"
	"inventory/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CustomerID  string             `json:"customerId" bson:"customerId"`
	Items       []OrderItem        `json:"items" bson:"items"`
	TotalAmount float64            `json:"totalAmount" bson:"totalAmount"`
	Status      string             `json:"status" bson:"status"`
	CreatedAt   time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt" bson:"updatedAt"`
}

type OrderItem struct {
	ProductID string `json:"productId" bson:"productId"`
	Quantity  int    `json:"quantity" bson:"quantity"`
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
func ConsumeOrderCreatedEvent() {
	ch, err := connections.RabitConn.Channel()
	if err != nil {
		failOnError(err, "Failed to open a channel")
	}

	defer ch.Close()

	q, err := ch.QueueDeclare("order_created", true, false, false, false, nil)
	if err != nil {
		failOnError(err, "Failed to declare a queue")
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)

	failOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)

			// Parse the order data from the message body
			var order Order
			err := json.Unmarshal(d.Body, &order)
			if err != nil {
				log.Printf("Failed to parse order data: %v", err)
				continue
			}

			// Process the order and update inventory
			err = processOrder(order)
			if err != nil {
				log.Printf("Failed to process order: %v", err)
			}
		}
	}()

	log.Printf("Waiting for messages")
	<-forever

}

func processOrder(order Order) error {

	fmt.Print(order)
	// Check items and reduce stock for each item in the order
	for _, item := range order.Items {

		fmt.Println(item)
		//Retrieve the product from the inventory
		product, err := models.GetProductByID(item.ProductID)
		if err != nil {
			return fmt.Errorf("failed to retrieve product %s: %v", item.ProductID, err)
		}

		// Check if the product has sufficient stock
		if product.Quantity < item.Quantity {
			return fmt.Errorf("insufficient stock for product %s", item.ProductID)
		}

		// Reduce the stock quantity
		product.Quantity -= item.Quantity

		// Update the product in the inventory
		err = product.UpdateProduct(product.ID.Hex())
		if err != nil {
			return fmt.Errorf("failed to update product %s: %v", item.ProductID, err)
		}

		log.Printf("Stock reduced for product %s. New quantity: %d", item.ProductID, product.Quantity)
	}

	return nil
}
