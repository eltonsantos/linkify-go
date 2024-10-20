package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

type URLRequest struct {
	LongURL string `json:"long_url"`
}

type URLResponse struct {
	ShortURL string `json:"short_url"`
}

func generateShortID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./urls.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS urls (
        id TEXT PRIMARY KEY,
        long_url TEXT
    )`)
	if err != nil {
		log.Fatal(err)
	}
}

func saveURL(shortID, longURL string) error {
	_, err := db.Exec("INSERT INTO urls (id, long_url) VALUES (?, ?)", shortID, longURL)
	return err
}

func getURL(shortID string) (string, error) {
	var longURL string
	err := db.QueryRow("SELECT long_url FROM urls WHERE id = ?", shortID).Scan(&longURL)
	return longURL, err
}

func shortenURL(w http.ResponseWriter, r *http.Request) {
	var req URLRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.LongURL == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	shortID := generateShortID()
	err = saveURL(shortID, req.LongURL)
	if err != nil {
		http.Error(w, "Error saving URL", http.StatusInternalServerError)
		return
	}

	response := URLResponse{ShortURL: "http://localhost:8080/" + shortID}
	json.NewEncoder(w).Encode(response)
}

func redirectURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortID := vars["id"]

	longURL, err := getURL(shortID)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, longURL, http.StatusFound)
}

func main() {
	initDB()

	r := mux.NewRouter()

	r.HandleFunc("/s", shortenURL).Methods("POST")
	r.HandleFunc("/{id}", redirectURL).Methods("GET")

	fmt.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))

	// Usage:
	// Run go run main.go
	// and in another terminal run:
	// curl -X POST http://localhost:8080/s -d '{"long_url": "https://www.eltonmelosantos.com.br"}' -H "Content-Type: application/json"
}
