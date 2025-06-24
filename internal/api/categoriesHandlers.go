package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/siwiec987/notes-api/internal/database"
	"github.com/siwiec987/notes-api/internal/models"
	"github.com/siwiec987/notes-api/internal/validation"
)

func (s *APIServer) handleGetCategories(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	name := strings.ToLower(r.URL.Query().Get("name"))
	createdAtStart := r.URL.Query().Get("created_at_start")
	createdAtEnd := r.URL.Query().Get("created_at_end")
	updatedAtStart := r.URL.Query().Get("updated_at_start")
	updatedAtEnd := r.URL.Query().Get("updated_at_end")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	query := `
		SELECT id, name, created_at, updated_at
		FROM categories
		WHERE user_id = ?
	`

	var args []any
	args = append(args, userID)
	if name != "" {
		name = "%" + name + "%"
		args = append(args, name)
		query += " AND name LIKE ?"
	}

	paramOperator := map[string]string{
		createdAtStart: "created_at >=",
		createdAtEnd:   "created_at <=",
		updatedAtStart: "updated_at >=",
		updatedAtEnd:   "updated_at <=",
	}

	err := validation.CreateDateFilters(&query, &args, paramOperator)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	limit := parseLimit(limitStr, 20)
	args = append(args, limit)
	query += " LIMIT ?"

	offset := parseOffset(offsetStr, 0)
	args = append(args, offset)
	query += " OFFSET ?"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to fetch categories")
		return
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var category models.Category
		err := rows.Scan(&category.ID, &category.Name, &category.CreatedAt, &category.UpdatedAt)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "failed to fetch categories")
			return
		}
		categories = append(categories, category)
	}

	err = rows.Err()
	if err != nil {
		sendError(w, http.StatusInternalServerError, "error reading categories from database")
		return
	}

	sendResponse(w, http.StatusOK, map[string]any{"categories": categories})
}

func (s *APIServer) handlePostCategories(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	var categories []models.CategoryPostRequest
	err := json.NewDecoder(r.Body).Decode(&categories)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(categories) == 0 {
		sendError(w, http.StatusBadRequest, "no categories provided")
		return
	}

	err = database.WithTransaction(s.db, func(tx *sql.Tx) error {
		for _, category := range categories {
			category.Name = strings.ToLower(category.Name)
			_, err := tx.Exec(
				`INSERT INTO categories (name, user_id) 
				VALUES (?, ?)`, category.Name, userID)
			if err != nil {
				return errors.New("invalid category data")
			}
		}

		return nil
	})
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendResponse(w, http.StatusCreated, map[string]any{"inserted": len(categories)})
}

func (s *APIServer) handleDeleteCategories(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	var categoryIDs []int
	err := json.NewDecoder(r.Body).Decode(&categoryIDs)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(categoryIDs) == 0 {
		sendError(w, http.StatusBadRequest, "no IDs provided")
		return
	}

	placeholders, args := database.BuildQueryArgs(userID, categoryIDs)

	categoriesExist, err := database.DoRecordsExistForUser(s.db, "categories", userID, categoryIDs)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to verify category existence")
		return
	}
	if !categoriesExist {
		sendError(w, http.StatusNotFound, "one or more category not found")
		return
	}

	query := fmt.Sprintf("DELETE FROM categories WHERE user_id = ? AND id IN (%s)", placeholders)

	var res sql.Result
	err = database.WithTransaction(s.db, func(tx *sql.Tx) error {
		res, err = tx.Exec(query, args...)
		if err != nil {
			return errors.New("failed to delete categories")
		}
		return nil
	})
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if res == nil {
		sendError(w, http.StatusInternalServerError, "no result from transaction")
		return
	}

	affected, err := res.RowsAffected()
	if err != nil {
		sendError(w, http.StatusInternalServerError, "could not determine affected rows")
		return
	}
	sendResponse(w, http.StatusOK, map[string]any{"deleted": affected})
}

func (s *APIServer) handlePatchCategories(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	var categories []models.CategoryPatchRequest
	err := json.NewDecoder(r.Body).Decode(&categories)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(categories) == 0 {
		sendError(w, http.StatusBadRequest, "no categories provided")
		return
	}

	var categoryIDs []int
	for _, category := range categories {
		if category.Name != nil {
			*category.Name = strings.ToLower(*category.Name)
		}
		categoryIDs = append(categoryIDs, category.ID)
	}

	exist, err := database.DoRecordsExistForUser(s.db, "categories", userID, categoryIDs)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to verify categories existence")
		return
	}
	if !exist {
		sendError(w, http.StatusNotFound, "one or more categories not found")
		return
	}

	toExecute := func(tx *sql.Tx) error {
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
				return fmt.Errorf("failed to update category with id %d", category.ID)
			}
		}

		return nil
	}

	err = database.WithTransaction(s.db, toExecute)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendResponse(w, http.StatusOK, map[string]any{"updated": len(categories)})
}
