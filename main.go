package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type Link struct {
	Original string `json:"original" db:"original"`
	Key      string `json:"short" db:"short"`
}

type LinkReq struct {
	Original string `json:"original" db:"original"`
}

func main() {
	db, _ := sqlx.Connect("mysql", "root:root@tcp(127.0.0.1:9001)/shortener")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		links := []Link{}
		db.Select(&links, "SELECT original, short FROM map")

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(links)
	})

	mux.HandleFunc("GET /{key}", func(w http.ResponseWriter, r *http.Request) {
		key := r.PathValue("key")

		var link Link
		db.Get(&link, "SELECT original, short FROM map WHERE short = ?", key)

		http.Redirect(w, r, link.Original, http.StatusTemporaryRedirect)
	})

	mux.HandleFunc("POST /", func(w http.ResponseWriter, r *http.Request) {
		var req LinkReq
		json.NewDecoder(r.Body).Decode(&req)

		randomString := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

		seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

		key := make([]byte, 6)
		for i := range key {
			key[i] = randomString[seededRand.Intn(len(randomString))]
		}
		shortKey := string(key)

		db.Query("INSERT INTO map (original, short) VALUES (?, ?)", req.Original, shortKey)

		w.Header().Add("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Link{
			Original: req.Original,
			Key:      shortKey,
		})
	})

	log.Println("Running on port 9000")
	http.ListenAndServe(":9000", mux)
}
