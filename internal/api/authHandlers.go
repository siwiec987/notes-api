package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/siwiec987/notes-api/internal/models"
	"github.com/siwiec987/notes-api/internal/validation"
	"golang.org/x/crypto/bcrypt"
)

func (s *APIServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	var user models.RegisterRequest
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if errs := validation.ValidateInput(user); len(errs) > 0 {
		sendError(w, http.StatusBadRequest, errs)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "could not hash password")
		return
	}
	user.Password = string(hashedPassword)

	_, err = s.db.Exec("INSERT INTO users (username, email, password) VALUES (?, ?, ?)", user.Username, user.Email, user.Password)
	if err != nil {
		sendError(w, http.StatusConflict, "username or email already exists")
		return
	}

	sendResponse(w, http.StatusCreated, map[string]string{"message": "user registered successfully"})
}

func (s *APIServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	var creds models.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var id int
	var hashedPassword string
	err = s.db.QueryRow("SELECT id, password FROM users WHERE username = ?", creds.Username).Scan(&id, &hashedPassword)
	if err == sql.ErrNoRows {
		sendError(w, http.StatusUnauthorized, "incorrect username or password")
		return
	}
	if err != nil {
		sendError(w, http.StatusInternalServerError, "database error")
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(creds.Password))
	if err != nil {
		sendError(w, http.StatusUnauthorized, "incorrect username or password")
		return
	}

	token, err := generateToken(id)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to generate authentication token")
		return
	}

	sendResponse(w, http.StatusOK, map[string]string{"token": token})
}