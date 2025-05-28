package models

import (
	"time"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type NotePostRequest struct {
	Content    string `json:"content"`
	CategoryID int    `json:"category_id"`
}

type Note struct {
	ID         int       `json:"id"`
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"created_at"`
	Category  struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"category"`
}