package connections

import (
	"github.com/rabbitmq/amqp091-go"
)

var (
	RabitConn *amqp091.Connection
)

func Connect() (*amqp091.Connection, error) {
	Conn, err := amqp091.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		return nil, err
	}
	RabitConn = Conn
	return Conn, nil
}
