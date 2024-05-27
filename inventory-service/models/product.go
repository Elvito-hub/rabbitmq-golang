package models

import (
	"context"
	"errors"
	"fmt"
	"inventory/db"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Product struct {
	ID          primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Price       float64            `json:"price"`
	Quantity    int                `json:"quantity"`
}

func (p *Product) SaveProduct() (*Product, error) {
	var productCollec = db.MongoClient.Database("ordersystem").Collection("products")
	newDoc := bson.M{
		"name":        p.Name,
		"description": p.Description,
		"price":       p.Price,
		"quantity":    p.Quantity,
	}

	result, err := productCollec.InsertOne(context.TODO(), newDoc)
	if err != nil {
		return nil, err
	}
	p.ID = result.InsertedID.(primitive.ObjectID)
	return p, nil
}

func GetProductByID(productID string) (*Product, error) {
	productObjectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return nil, errors.New("invalid ID")
	}

	var productCollec = db.MongoClient.Database("ordersystem").Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var product *Product
	filter := bson.M{
		"_id": productObjectID,
	}

	err = productCollec.FindOne(ctx, filter).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("product not found")
		}
		return nil, errors.New("could not find product")
	}
	return product, nil
}

func (p *Product) UpdateProduct(productID string) error {
	var productCollec = db.MongoClient.Database("ordersystem").Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": objectID}
	update := bson.M{}

	if p.Name != "" {
		update["name"] = p.Name
	}
	if p.Description != "" {
		update["description"] = p.Description
	}
	if p.Price != 0 {
		update["price"] = p.Price
	}
	if p.Quantity != 0 {
		update["quantity"] = p.Quantity
	}

	if len(update) == 0 {
		return nil
	}

	updateQuery := bson.M{"$set": update}

	_, err = productCollec.UpdateOne(ctx, filter, updateQuery)
	if err != nil {
		return err
	}

	return nil
}

func GetProducts(page, pageSize int, search string) ([]*Product, int64, int64, error) {
	var productCollec = db.MongoClient.Database("ordersystem").Collection("products")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var productsList []*Product

	// Calculate the skip value based on the page number and page size
	skip := (page - 1) * pageSize

	// Set up options for limiting and skipping results
	options := options.Find().
		SetLimit(int64(pageSize)).
		SetSkip(int64(skip))

	filter := bson.D{}

	if search != "" {
		filter = append(filter, bson.E{"name", primitive.Regex{Pattern: "(?i).*" + search + ".*", Options: ""}})
	}

	var cursor *mongo.Cursor
	var err error

	cursor, err = productCollec.Find(ctx, filter, options)
	if err != nil {
		fmt.Println(err)
		return productsList, 0, 0, errors.New("unable to fetch data")
	}
	if err = cursor.All(context.TODO(), &productsList); err != nil {
		fmt.Println(err)
		return productsList, 0, 0, errors.New("unable to fetch data")
	}

	// Get the total number of elements without skipping and limiting
	totalElements, err := productCollec.CountDocuments(ctx, filter)

	fmt.Println(totalElements, "total elements")

	if err != nil {
		fmt.Println(err)
		return productsList, 0, 0, errors.New("unable to fetch total elements")
	}

	totalPages := int64(math.Ceil(float64(totalElements) / float64(pageSize)))

	if err != nil {
		fmt.Println(err)
		return productsList, 0, 0, errors.New("unable to fetch data")
	}

	return productsList, totalElements, totalPages, nil
}
