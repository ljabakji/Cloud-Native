package main

import (
	// "encoding/json"
	"fmt"
	"log"
	"net/http"
)

func main() {
	db := database{"shoes": 50, "socks": 5}
	mux := http.NewServeMux()
	mux.HandleFunc("/list", http.HandlerFunc(db.list))
	mux.HandleFunc("/list2",http.HandlerFunc(db.list))
	// mux.HandleFunc("/create", db.create)
	// mux.HandleFunc("/update", db.update)
	log.Fatal(http.ListenAndServe(":8000", mux))
	
	

}

type dollars float32

func (d dollars) String() string { return fmt.Sprintf("$%.2f", d) }

type database map[string]dollars

func (db database) list(w http.ResponseWriter, req *http.Request) {
	for item, price := range db {
		fmt.Fprintf(w, "%s: %s\n", item, price)
	}
}
func (db database) list2(w http.ResponseWriter, req *http.Request) {
	for item, price := range db {
		fmt.Fprintf(w, "%s: %s\n", item, price)
	}
}

func (db database) price(w http.ResponseWriter, req *http.Request) {
	item := req.URL.Query().Get("item")
	price, ok := db[item]
	if !ok {

		w.WriteHeader(http.StatusNotFound) // 404
		fmt.Fprintf(w, "no such item: %q\n", item)
		return

	}
	fmt.Fprintf(w, "%s\n", price)

}

// func (db database) create(w http.ResponseWriter, req *http.Request) {
// 	var newItem struct {
// 		Item  string  `json:"item"`
// 		Price float32 `json:"price"`
// 	}
// 	err := json.NewDecoder(req.Body).Decode(&newItem)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	if _, ok := db[newItem.Item]; ok {
// 		http.Error(w, "Item already exists", http.StatusBadRequest)
// 		return
// 	}

// 	db[newItem.Item] = dollars(newItem.Price)
// 	w.WriteHeader(http.StatusCreated)
// }

// func (db database) update(w http.ResponseWriter, req *http.Request) {
// 	item := req.URL.Query().Get("item")
// 	if _, ok := db[item]; !ok {
// 		http.Error(w, "Item not found", http.StatusNotFound)
// 		return
// 	}

// 	var updatedPrice float32
// 	err := json.NewDecoder(req.Body).Decode(&updatedPrice)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusBadRequest)
// 		return
// 	}

// 	db[item] = dollars(updatedPrice)
// }

// func (db database) delete(w http.ResponseWriter, req *http.Request) {
// 	item := req.URL.Query().Get("item")
// 	if _, ok := db[item]; !ok {

// 		http.Error(w, "Item not found", http.StatusNotFound)
// 		return
// 	}
// 	delete(db, item)
// 	fmt.Fprintf(w, "Deleted item :%s\n", item)

// }


func (db database) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/list":
		for item, price := range db {
			fmt.Fprintf(w, "%s: %s\n", item, price)
		}
	case "/list2":
		for item, price := range db {
			fmt.Fprintf(w, "%s: %s\n", item, price)
		}
	case "/price":
		item := req.URL.Query().Get("item")
		price, ok := db[item]
		if !ok {
			w.WriteHeader(http.StatusNotFound) // 404
			fmt.Fprintf(w, "no such item: %q\n", item)
			return
		}
		fmt.Fprintf(w, "%s\n", price)

	
	default:
		w.WriteHeader(http.StatusNotFound) //404
		fmt.Fprintf(w, "no such page: %s\n", req.URL)

	}
}
