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

	router.HandleFunc("GET /notes", authMiddleware(s.handleGetNotes))
	router.HandleFunc("POST /notes", authMiddleware(s.handlePostNotes))
	router.HandleFunc("PATCH /notes", authMiddleware(s.handlePatchNotes))
	router.HandleFunc("DELETE /notes", authMiddleware(s.handleDeleteNotes))
	// router.HandleFunc("Get /notes", authMiddleware(s.handleSearchNotes)) wyszukiwanie

	router.HandleFunc("GET /categories", authMiddleware(s.handleGetCategories))
	router.HandleFunc("POST /categories", authMiddleware(s.handlePostCategories))
	router.HandleFunc("DELETE /categories", authMiddleware(s.handleDeleteCategories))
	router.HandleFunc("PATCH /categories", authMiddleware(s.handlePatchCategories))

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
	log.Printf("ERROR %d: %s", status, msg)
	sendResponse(w, status, ErrorResponse{Error: msg})
}
