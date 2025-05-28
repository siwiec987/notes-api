package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/siwiec987/notes-api/internal/migrations"
	"github.com/siwiec987/notes-api/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func (s *APIServer) handleSeed(w http.ResponseWriter, r *http.Request) {
	err := migrations.SeedData(s.db)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to seed database: " + err.Error())
		return
	}
	sendResponse(w, http.StatusOK, map[string]string{"message": "Database seeded successfully"})
}

func (s *APIServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	var user models.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
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
		sendError(w, http.StatusConflict, "Username or email already exists")
		return
	}

	sendResponse(w, http.StatusCreated, map[string]string{"message": "User registered successfully"})
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	var creds models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var id int
	var hashedPassword string
	err = s.db.QueryRow("SELECT id, password FROM users WHERE username = ?", creds.Username).Scan(&id, &hashedPassword)
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Incorrect username or password")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(creds.Password))
	if err != nil {
		sendError(w, http.StatusUnauthorized, "Incorrect username or password")
		return
	}

	token, err := generateToken(id)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to generate authentication token")
		return
	}

	sendResponse(w, http.StatusOK, map[string]string{"token": token})
}

func (s *APIServer) handleGetNotes(w http.ResponseWriter, r *http.Request) {
	val := r.Context().Value(userKey)
	userID, ok := val.(int)
	if !ok {
		sendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	rows, err := s.db.Query(
		`SELECT n.id, n.content, n.created_at, n.category_id, c.name as category_name 
		FROM notes n
		JOIN categories c ON n.category_id = c.id 
		WHERE n.user_id = ?`, 
		userID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to fetch notes")
		return
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var note models.Note
		err := rows.Scan(&note.ID, &note.Content, &note.CreatedAt, &note.Category.ID, &note.Category.Name)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "Failed to fetch notes")
			return
		}
		notes = append(notes, note)
	}

	err = rows.Err()
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Error reading notes from database")
		return
	}

	sendResponse(w, http.StatusOK, map[string]any{
		"user":  userID,
		"notes": notes,
	})
}

func (s *APIServer) handlePostNotes(w http.ResponseWriter, r *http.Request) {
	val := r.Context().Value(userKey)
	userID, ok := val.(int)
	if !ok {
		sendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var notes []models.NotePostRequest
	err := json.NewDecoder(r.Body).Decode(&notes)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	for _, note := range notes {
		_, err := s.db.Exec(
			`INSERT INTO notes (content, category_id, user_id) 
			VALUES (?, ? ,?)`, note.Content, note.CategoryID, userID)
		if err != nil {
			sendError(w, http.StatusBadRequest, "Invalid note data or category does not exist")
			return
		}
	}

	sendResponse(w, http.StatusOK, notes)
}

func (s *APIServer) handleDeleteNotes(w http.ResponseWriter, r *http.Request) {
	val := r.Context().Value(userKey)
	userID, ok := val.(int)
	if !ok {
		sendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	noteID := r.PathValue("id")
	if noteID == "" {
		sendError(w, http.StatusBadRequest, "Note ID is required")
		return
	}

	res, err := s.db.Exec("DELETE FROM notes WHERE user_id = ? AND id = ?", userID, noteID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to delete note")
		return
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		sendError(w, http.StatusNotFound, "Note not found")
		return
	}

	sendResponse(w, http.StatusOK, fmt.Sprintf("Note with id: %s deleted", noteID))
}