package api

import (
	"net/http"

	"github.com/Deathfireofdoom/distributed-kv-store/internal/raft"
	"github.com/gorilla/mux"
)

func NewRouter(node *raft.RaftNode) *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/store", func(w http.ResponseWriter, r *http.Request) { PutHandler(w, r, node) }).Methods("POST")
	router.HandleFunc("/store", func(w http.ResponseWriter, r *http.Request) { GetHandler(w, r, node) }).Methods("GET").Queries("key", "{key}")
	router.HandleFunc("/store", DeleteHandler).Methods("DELETE")
	return router
}
