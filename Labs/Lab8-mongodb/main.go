// Example use of Go mongo-driver
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	mongodbEndpoint = "mongodb://172.17.0.2:27017" // Find this from the Mongo container
)

type dollars float32

func (d dollars) String() string { return fmt.Sprintf("$%.2f", d) }

type Field struct {
	ID        primitive.ObjectID `bson:"_id"`
	Item      string             `bson:"item"`
	Price     dollars            `bson:"price"`
	Category  string             `bson:"category"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

type database struct {
	collection *mongo.Collection
}

func main() {
	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodbEndpoint))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	// Select database and collection
	collection := client.Database("store").Collection("items")
	db := database{collection: collection}

	// Initialize router
	router := http.NewServeMux()

	// Map the handlers
	router.HandleFunc("/list", db.list)
	router.HandleFunc("/price", db.price)
	router.HandleFunc("/create", db.create)
	router.HandleFunc("/read", db.read)
	router.HandleFunc("/update", db.update)
	router.HandleFunc("/delete", db.delete)

	// Start server
	log.Println("Server started on port 8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
func (db *database) list(w http.ResponseWriter, r *http.Request) {
	// Get items from the collection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := db.collection.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// Retrieve the items from the cursor
	var items []Field
	for cursor.Next(ctx) {
		var item Field
		if err := cursor.Decode(&item); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		items = append(items, item)
	}

	// Write the items
	for _, item := range items {
		fmt.Fprintf(w, "%s: %s\n", item.Item, item.Price)
	}
}

func (db *database) price(w http.ResponseWriter, req *http.Request) {
	// Get the item from the query parameter
	item := req.URL.Query().Get("item")

	// Take item from the database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result Field
	err := db.collection.FindOne(ctx, bson.M{"item": item}).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusNotFound) // 404
			fmt.Fprintf(w, "no such item: %q\n", item)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write item price 
	fmt.Fprintf(w, "%s\n", result.Price)
}

func (db *database) create(w http.ResponseWriter, req *http.Request) {
	item := req.URL.Query().Get("item")
	newPrice := req.URL.Query().Get("price")

	price, err := strconv.ParseFloat(newPrice, 32)
	// Parsing Failure
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		fmt.Fprintf(w, "invalid price: %q\n", newPrice)
		return
	}

	// Check if item exists
	var existingItem Field
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.collection.FindOne(ctx, bson.M{"item": item}).Decode(&existingItem)
	if err == nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		fmt.Fprintf(w, "item already exists: %s\n", item)
		return
	} else if err != mongo.ErrNoDocuments {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create item 
	newItem := Field{
		Item:      item,
		Price:     dollars(price),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert item to database
	_, err = db.collection.InsertOne(ctx, newItem)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "create item: %s, price: %.2f\n", item, price)
}

func (db *database) read(w http.ResponseWriter, req *http.Request) {
	db.list(w, req)
}

func (db *database) update(w http.ResponseWriter, req *http.Request) {
	item := req.URL.Query().Get("item")
	newPrice := req.URL.Query().Get("price")

	price, err := strconv.ParseFloat(newPrice, 32)
	// Parsing Failure
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		fmt.Fprintf(w, "invalid price: %q\n", newPrice)
		return
	}

	// Check if item exists
	var existingItem Field
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.collection.FindOne(ctx, bson.M{"item": item}).Decode(&existingItem)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusBadRequest) // 400
			fmt.Fprintf(w, "item does not exist: %s\n", item)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update item price
	_, err = db.collection.UpdateOne(ctx,
		bson.M{"item": item},
		bson.M{"$set": bson.M{"price": dollars(price), "updated_at": time.Now()}})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "update item: %s, price: %.2f\n", item, price)
}

func (db *database) delete(w http.ResponseWriter, req *http.Request) {
	item := req.URL.Query().Get("item")

	// Check if item exists
	var existingItem Field
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := db.collection.FindOne(ctx, bson.M{"item": item}).Decode(&existingItem)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			w.WriteHeader(http.StatusBadRequest) // 400
			fmt.Fprintf(w, "item does not exist: %s\n", item)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete item from database
	_, err = db.collection.DeleteOne(ctx, bson.M{"item": item})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "delete item: %s\n", item)
}