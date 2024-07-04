package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/Deathfireofdoom/distributed-kv-store/internal/models"
	"github.com/Deathfireofdoom/distributed-kv-store/internal/raft"
)

func PutHandler(w http.ResponseWriter, r *http.Request, node *raft.RaftNode) { 
	node.PutHandler(w, r)
}

func GetHandler(w http.ResponseWriter, r *http.Request, node *raft.RaftNode) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	store := node.GetStore()
	value, ok := store.Get(key)
	log.Printf("key %s got value %s", key, value)
	log.Printf("%s", store)
	if !ok {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	resp := models.GetResponse{Value: value, Found: true}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	var req models.DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		http.Error(w, "empty key", http.StatusBadRequest)
		return
	}

	resp := models.DeleteResponse{Success: true}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
