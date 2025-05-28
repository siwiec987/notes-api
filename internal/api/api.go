package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
)

type APIServer struct {
	addr string
	db *sql.DB
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewAPIServer(addr string, db *sql.DB) *APIServer {
	return &APIServer{
		addr: addr,
		db: db,
	}
}

func (s *APIServer) Run() {
	router := http.NewServeMux()

	router.HandleFunc("POST /register", s.handleRegister)
	router.HandleFunc("POST /login", s.handleLogin)
	router.HandleFunc("POST /notes", authMiddleware(s.handlePostNotes))
	router.HandleFunc("GET /notes", authMiddleware(s.handleGetNotes))
	router.HandleFunc("DELETE /notes/{id}", authMiddleware(s.handleDeleteNotes))
	router.HandleFunc("/seed", s.handleSeed)

	log.Println("Running on port:", s.addr)

	http.ListenAndServe(s.addr, router)
}

func sendResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Println("Error encoding response:", err)
	}
}

func sendError(w http.ResponseWriter, status int, msg string) {
	sendResponse(w, status, ErrorResponse{Error: msg})
}
