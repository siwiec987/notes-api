package models

import (
	"time"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type ErrorsResponse struct {
	Errors map[string][]string `json:"errors"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginSuccessResponse struct {
	Token string `json:"token"`
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
	ID			 int        `json:"id"`
	Content		 string     `json:"content"`
	CreatedAt	 time.Time  `json:"created_at"`
	UpdatedAt	 time.Time  `json:"updated_at"`
	CategoryID	 int		`json:"category_id"`
	CategoryName string		`json:"category_name"`
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

type MultipleNotesResponse struct {
	Notes []Note `json:"notes"`
}

type MultipleCategoriesResponse struct {
	Categories []Category `json:"categories"`
}

type InsertedResponse struct {
	Inserted int `json:"inserted"`
}

type DeletedResponse struct {
	Deleted int `json:"deleted"`
}

type UpdatedResponse struct {
	Updated int `json:"updated"`
}