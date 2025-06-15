package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/siwiec987/notes-api/internal/models"
)

func (s *APIServer) handleGetCategories(w http.ResponseWriter, r *http.Request) {
	val := r.Context().Value(userKey)
	userID, ok := val.(int)
	if !ok {
		sendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	rows, err := s.db.Query(
		`SELECT id, name
		FROM categories
		WHERE user_id = ?`,
		userID)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to fetch categories")
		return
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "Failed to fetch categories")
			return
		}
		categories = append(categories, category)
	}

	err = rows.Err()
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Error reading categories from database")
		return
	}

	sendResponse(w, http.StatusOK, map[string]any{"categories": categories})
}

func (s *APIServer) handlePostCategories(w http.ResponseWriter, r *http.Request) {
	val := r.Context().Value(userKey)
	userID, ok := val.(int)
	if !ok {
		sendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var categories []models.CategoryPostRequest
	err := json.NewDecoder(r.Body).Decode(&categories)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if len(categories) == 0 {
		sendError(w, http.StatusBadRequest, "No categories provided")
		return
	}

	tx, err := s.db.Begin()
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	for _, category := range categories {
		_, err := tx.Exec(
			`INSERT INTO categories (name, user_id) 
			VALUES (?, ?)`, category.Name, userID)
		if err != nil {
			tx.Rollback()
			sendError(w, http.StatusBadRequest, "Invalid category data")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	sendResponse(w, http.StatusCreated, map[string]any{"inserted": len(categories)})
}

func (s *APIServer) handleDeleteCategories(w http.ResponseWriter, r *http.Request) {
	val := r.Context().Value(userKey)
	userID, ok := val.(int)
	if !ok {
		sendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var categoryIDs []int
	err := json.NewDecoder(r.Body).Decode(&categoryIDs)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if len(categoryIDs) == 0 {
		sendError(w, http.StatusBadRequest, "No IDs provided")
		return
	}

	placeholders := make([]string, len(categoryIDs))
	args := make([]any, len(categoryIDs)+1)
	args[0] = userID
	for i, id := range categoryIDs {
		placeholders[i] = "?"
		args[i+1] = id
	}

	categoriesExist, err := s.doCategoriesExist(userID, categoryIDs)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to verify category existence")
		return
	}
	if !categoriesExist {
		sendError(w, http.StatusNotFound, "One or more category not found")
		return
	}

	query := fmt.Sprintf("DELETE FROM categories WHERE user_id = ? AND id IN (%s)", strings.Join(placeholders, ","))

	tx, err := s.db.Begin()
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	res, err := tx.Exec(query, args...)
	if err != nil {
		tx.Rollback()
		sendError(w, http.StatusInternalServerError, "Failed to delete categories")
		return
	}

	if err := tx.Commit(); err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	affected, err := res.RowsAffected()
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Could not determine affected rows")
		return
	}
	sendResponse(w, http.StatusOK, map[string]any{"deleted": affected})
}

func (s *APIServer) handlePatchCategories(w http.ResponseWriter, r *http.Request) {
	val := r.Context().Value(userKey)
	userID, ok := val.(int)
	if !ok {
		sendError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var categories []models.CategoryPatchRequest
	err := json.NewDecoder(r.Body).Decode(&categories)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if len(categories) == 0 {
		sendError(w, http.StatusBadRequest, "No categories provided")
		return
	}

	var categoryIDs []int
	for _, category := range categories {
		categoryIDs = append(categoryIDs, category.ID)
	}

	exist, err := s.doNotesExist(userID, categoryIDs)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to verify categories existence")
		return
	}
	if !exist {
		sendError(w, http.StatusNotFound, "One or more categories not found")
		return
	}

	tx, err := s.db.Begin()
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to start transaction")
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	for _, category := range categories {
		var setClauses []string
		var args []any

		if category.Name != nil {
			setClauses = append(setClauses, "name = ?")
			args = append(args, *category.Name)
		}
		if len(setClauses) == 0 {
			continue
		}

		query := fmt.Sprintf(
			`UPDATE categories 
			SET %s 
			WHERE user_id = ? 
			AND id = ?`,
			strings.Join(setClauses, ","))

		args = append(args, userID, category.ID)

		_, err = tx.Exec(query, args...)
		if err != nil {
			tx.Rollback()
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update category with id %d", category.ID))
			return
		}
	}

	if err := tx.Commit(); err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	sendResponse(w, http.StatusOK, map[string]any{"updated": len(categories)})
}

func (s *APIServer) doCategoriesExist(userID int, categoriesIDs []int) (bool, error) {
	placeholders := make([]string, len(categoriesIDs))
	args := make([]any, len(categoriesIDs)+1)
	args[0] = userID
	for i, id := range categoriesIDs {
		placeholders[i] = "?"
		args[i+1] = id
	}

	queryCheck := fmt.Sprintf(`
    SELECT COUNT(*) FROM categories 
    WHERE user_id = ? AND id IN (%s)`, strings.Join(placeholders, ","))

	var count int
	err := s.db.QueryRow(queryCheck, args...).Scan(&count)
	return count == len(categoriesIDs), err
}
