package api

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/swaggo/http-swagger"
	_ "github.com/siwiec987/notes-api/docs"
)

type APIServer struct {
	addr string
	db *sql.DB
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

	router.HandleFunc("GET /categories", authMiddleware(s.handleGetCategories))
	router.HandleFunc("POST /categories", authMiddleware(s.handlePostCategories))
	router.HandleFunc("DELETE /categories", authMiddleware(s.handleDeleteCategories))
	router.HandleFunc("PATCH /categories", authMiddleware(s.handlePatchCategories))

	// router.HandleFunc("/seed", s.handleSeed)

	router.Handle("GET /docs/", httpSwagger.WrapHandler)

	log.Println("Running on port:", s.addr)

	http.ListenAndServe(s.addr, router)
}
