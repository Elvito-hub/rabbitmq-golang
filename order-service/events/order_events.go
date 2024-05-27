package events

import (
	"context"
	"encoding/json"
	"log"
	"order-service/connections"
	"order-service/models"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

// PublishOrderCreatedEvent publishes an order created event
func PublishOrderCreatedEvent(order *models.Order) error {
	eventData, err := json.Marshal(order)
	if err != nil {
		return err
	}

	ch, err := connections.RabitConn.Channel()

	if err != nil {
		return err
	}

	defer ch.Close()

	q, err := ch.QueueDeclare("order_created", true, false, false, false, nil)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	err = ch.PublishWithContext(
		ctx,
		"",
		q.Name,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        eventData,
		},
	)

	if err != nil {
		return err
	}

	log.Printf("Order created event published: %s", eventData)
	return nil
}
