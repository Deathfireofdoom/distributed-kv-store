package api

import (
	"encoding/json"
	"net/http"

	"github.com/Deathfireofdoom/distributed-kv-store/internal/kvstore"
	"github.com/Deathfireofdoom/distributed-kv-store/internal/models"
)

var store = kvstore.NewStore()

func PutHandler(w http.ResponseWriter, r *http.Request) {
	var req models.PutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	store.Put(req.Key, req.Value)

	resp := models.PutResponse{Success: true}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing key parameter", http.StatusBadRequest)
		return
	}

	value, ok := store.Get(key)
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

	store.Delete(req.Key)

	resp := models.DeleteResponse{Success: true}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
