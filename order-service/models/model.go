package models

import (
	"context"
	"errors"
	"fmt"
	"math"
	"order-service/db"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	ProductID string  `json:"productId" bson:"productId"`
	Quantity  int     `json:"quantity" bson:"quantity"`
	Price     float64 `json:"price" bson:"price"`
}

func (o *Order) SaveOrder() (*Order, error) {
	var orderCollec = db.MongoClient.Database("ordersystem").Collection("orders")
	newDoc := bson.M{
		"items":       o.Items,
		"totalAmount": o.TotalAmount,
		"status":      o.Status,
		"createdAt":   time.Now(),
		"updatedAt":   time.Now(),
	}

	fmt.Println(newDoc, "the new dox")

	result, err := orderCollec.InsertOne(context.TODO(), newDoc)
	if err != nil {
		return nil, err
	}
	o.ID = result.InsertedID.(primitive.ObjectID)
	return o, nil
}

func GetOrderByID(orderID string) (*Order, error) {
	orderObjectID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return nil, errors.New("invalid ID")
	}

	var orderCollec = db.MongoClient.Database("ordersystem").Collection("orders")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var order *Order
	filter := bson.M{
		"_id": orderObjectID,
	}

	err = orderCollec.FindOne(ctx, filter).Decode(&order)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("order not found")
		}
		return nil, errors.New("could not find order")
	}
	return order, nil
}

func (o *Order) UpdateOrder(orderID string) error {
	var orderCollec = db.MongoClient.Database("ordersystem").Collection("orders")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(orderID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{}

	if o.CustomerID != "" {
		update["customerId"] = o.CustomerID
	}
	if len(o.Items) > 0 {
		update["items"] = o.Items
	}
	if o.TotalAmount != 0 {
		update["totalAmount"] = o.TotalAmount
	}
	if o.Status != "" {
		update["status"] = o.Status
	}
	update["updatedAt"] = time.Now()

	if len(update) == 0 {
		return nil
	}

	updateQuery := bson.M{"$set": update}

	_, err = orderCollec.UpdateOne(ctx, filter, updateQuery)
	if err != nil {
		return err
	}

	return nil
}

func GetOrders(page, pageSize int, search string) ([]*Order, int64, int64, error) {
	var orderCollec = db.MongoClient.Database("ordersystem").Collection("orders")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var ordersList []*Order

	// Calculate the skip value based on the page number and page size
	skip := (page - 1) * pageSize

	// Set up options for limiting and skipping results
	options := options.Find().
		SetLimit(int64(pageSize)).
		SetSkip(int64(skip))

	filter := bson.D{}

	if search != "" {
		filter = append(filter, bson.E{"customerId", primitive.Regex{Pattern: "(?i).*" + search + ".*", Options: ""}})
	}

	var cursor *mongo.Cursor
	var err error

	cursor, err = orderCollec.Find(ctx, filter, options)
	if err != nil {
		fmt.Println(err)
		return ordersList, 0, 0, errors.New("unable to fetch data")
	}
	if err = cursor.All(context.TODO(), &ordersList); err != nil {
		fmt.Println(err)
		return ordersList, 0, 0, errors.New("unable to fetch data")
	}

	// Get the total number of elements without skipping and limiting
	totalElements, err := orderCollec.CountDocuments(ctx, filter)

	fmt.Println(totalElements, "total elements")

	if err != nil {
		fmt.Println(err)
		return ordersList, 0, 0, errors.New("unable to fetch total elements")
	}

	totalPages := int64(math.Ceil(float64(totalElements) / float64(pageSize)))

	if err != nil {
		fmt.Println(err)
		return ordersList, 0, 0, errors.New("unable to fetch data")
	}

	return ordersList, totalElements, totalPages, nil
}
