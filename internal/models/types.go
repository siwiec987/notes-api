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

type NotePatchRequest struct {
	ID         int     `json:"id"`
	Content    *string `json:"content"`
	CategoryID *int    `json:"category_id"`
}

type Note struct {
	ID        int       `json:"id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Category  Category  `json:"category"`
}

type Category struct {
	ID			int			`json:"id"`
	Name 		string		`json:"name"`
	CreatedAt	time.Time	`json:"created_at"`
	UpdatedAt	time.Time	`json:"updated_at"`
}

type CategoryPostRequest struct {
	Name string `json:"name"`
}

type CategoryPatchRequest struct {
	ID	 int	 `json:"id"`
	Name *string `json:"name"`
}