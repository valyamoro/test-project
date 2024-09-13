package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type Item struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

var db *sql.DB

func initDB() {
	connStr := "user=postgres password=root dbname=test_project sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully connected to database.")
}

func createItem(w http.ResponseWriter, r *http.Request) {
	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `INSERT INTO items (title) VALUES ($1) RETURNING id`
	if err := db.QueryRow(query, item.Title).Scan(&item.ID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

func getItem(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	var item Item
	query := `SELECT id, title FROM items WHERE id=$1`
	if err := db.QueryRow(query, id).Scan(&item.ID, &item.Title); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "Item not found", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(item)
}

func updateItem(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	var item Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `UPDATE items SET title = $1 WHERE id = $2`
	if _, err := db.Exec(query, item.Title, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode("Item updated successfully")
}

func deleteItem(w http.ResponseWriter, r *http.Request) {
	query := `DELETE FROM items WHERE id = $1`
	id := r.URL.Query().Get("id")
	if _, err := db.Exec(query, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Item deleted successfully")
}

func main() {
	initDB()

	http.HandleFunc("/create", createItem)
	http.HandleFunc("/read", getItem)
	http.HandleFunc("/update", updateItem)
	http.HandleFunc("/delete", deleteItem)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("Server is running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
