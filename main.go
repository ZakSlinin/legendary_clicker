package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	_ "modernc.org/sqlite"
	"net/http"
	"strconv"
	"sync"
)

var (
	db *sql.DB
	mu sync.Mutex
)

func initDB() {
	var err error
	db, err = sql.Open("sqlite", "clicker.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
			balance INTEGER NOT NULL DEFAULT 0
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID int64 `json:"user_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	mu.Lock()
	defer mu.Unlock()

	_, err := db.Exec(`
		INSERT INTO users (id, balance) VALUES (?, 0)
		ON CONFLICT(id) DO NOTHING
	`, req.UserID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var balance int64
	db.QueryRow(`SELECT balance FROM users WHERE id = ?`, req.UserID).Scan(&balance)

	json.NewEncoder(w).Encode(map[string]int64{"balance": balance})
}

func tapHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID    int64 `json:"user_id"`
		TapCounts int64 `json:"tap_counts"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	mu.Lock()
	defer mu.Unlock()

	_, err := db.Exec(`INSERT INTO users (id, balance) VALUES (?, ?)
	ON CONFLICT(id) DO UPDATE SET balance = balance + ?
	`, req.UserID, req.TapCounts, req.TapCounts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var balance int64
	db.QueryRow(`SELECT balance FROM users WHERE id = ?`, req.UserID).Scan(&balance)

	json.NewEncoder(w).Encode(map[string]int64{"balance": balance})
}

func balanceHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)

	if err != nil {
		http.Error(w, "invalid user id", http.StatusInternalServerError)
		return
	}

	var balance int64
	err = db.QueryRow(`SELECT balance FROM users WHERE id = ?`, userID).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int64{"balance": balance})
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}


func main() {
	initDB()
 
 http.HandleFunc("/register", corsMiddleware(registerHandler))
	http.HandleFunc("/tap", corsMiddleware(tapHandler))
	http.HandleFunc("/balance", corsMiddleware(balanceHandler))

	
	log.Println("server listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
