package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

// NewRouter sets up the HTTP routes
func NewRouter(h *Handler) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/tasks", h.CreateTask).Methods("POST")
	// Health check
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	}).Methods("GET")

	return r
}
