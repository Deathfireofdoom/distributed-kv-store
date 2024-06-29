package api

import (
	"github.com/gorilla/mux"
)

func NewRouter() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/store", PutHandler).Methods("POST")
	router.HandleFunc("/store", GetHandler).Methods("GET").Queries("key", "{key}")
	router.HandleFunc("/store", DeleteHandler).Methods("DELETE")
	return router
}
