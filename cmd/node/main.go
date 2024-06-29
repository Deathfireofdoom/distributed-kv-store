package main

import (
	"log"
	"net/http"

	"github.com/Deathfireofdoom/distributed-kv-store/internal/api"
)

func main() {
	router := api.NewRouter()
	log.Println("Starting server")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
