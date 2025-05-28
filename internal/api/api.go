package api

import (
	// "context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/siwiec987/notes-api/internal/migrations"
	"github.com/siwiec987/notes-api/internal/models"
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
	router.HandleFunc("DELETE /notes", authMiddleware(s.handleDeleteNotes))
	router.HandleFunc("/seed", s.handleSeed)

	log.Println("Running on port:", s.addr)

	http.ListenAndServe(s.addr, router)
}

func (s *APIServer) handleSeed(w http.ResponseWriter, r *http.Request) {
	err := migrations.SeedData(s.db)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	sendResponse(w, http.StatusOK, map[string]string{"message": "Database seeded successfully"})
}

func (s *APIServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	var user models.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Could not hash password")
		return
	}
	user.Password = string(hashedPassword)

	_, err = s.db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", user.Username, user.Email, user.Password)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Could not register user")
		return
	}

	sendResponse(w, http.StatusCreated, map[string]string{"message": "User registered successfully"})
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	var creds models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	var id int
	var hashedPassword string
	err = s.db.QueryRow("SELECT id, password FROM users WHERE username = ?", creds.Username).Scan(&id, &hashedPassword)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(creds.Password))
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	token, err := generateToken(id)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Could not generate token")
		return
	}

	sendResponse(w, http.StatusOK, map[string]string{"token": token})
}

func (s *APIServer) handleGetNotes(w http.ResponseWriter, r *http.Request) {
	val := r.Context().Value(userKey)
	userID, ok := val.(int)
	if !ok {
		sendError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	rows, err := s.db.Query(
		`SELECT n.id, n.content, n.created_at, n.category_id, c.name as category_name 
		FROM notes n
		JOIN categories c ON n.category_id = c.id 
		WHERE n.user_id = ?`, 
		userID)
	if err != nil {
		sendError(w, http.StatusNotFound, "Notes not found")
		return
	}

	var notes []models.Note
	for rows.Next() {
		var note models.Note
		err := rows.Scan(&note.ID, &note.Content, &note.CreatedAt, &note.Category.ID, &note.Category.Name)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		notes = append(notes, note)
	}

	err = rows.Err()
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	sendResponse(w, http.StatusOK, map[string]any{
		"user":  userID,
		"notes": notes,
	})
}

func (s *APIServer) handlePostNotes(w http.ResponseWriter, r *http.Request) {
	
}

func (s *APIServer) handleDeleteNotes(w http.ResponseWriter, r *http.Request) {

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
