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
)

//	@Summary		Get categories
//	@Description	Get a list of categories with optional filters, pagination and sorting.
//	@Tags			categories
//	@Produce		json
//	@Param			name				query		string	false	"Filter by category name (case-insensitive)"
//	@Param			created_at_start	query		string	false	"Filter by created_at >= (format: 2006-01-02 15:04:05)"
//	@Param			created_at_end		query		string	false	"Filter by created_at <= (format: 2006-01-02 15:04:05)"
//	@Param			updated_at_start	query		string	false	"Filter by updated_at >= (format: 2006-01-02 15:04:05)"
//	@Param			updated_at_end		query		string	false	"Filter by updated_at <= (format: 2006-01-02 15:04:05)"
//	@Param			limit				query		int		false	"Number of categories to return"
//	@Param			offset				query		int		false	"Number of categories to skip"
//	@Param			sort_by				query		string	false	"Sort by field (id, name, created_at, updated_at, default: updated_at)"
//	@Param			sort_order			query		string	false	"Sort order (ASC or DESC)"
//	@Success		200					{object}	models.MultipleCategoriesResponse
//	@Failure		400					{object}	models.ErrorResponse
//	@Failure		500					{object}	models.ErrorResponse
//	@Security		BearerAuth
//	@Router			/categories [get]
func (s *APIServer) handleGetCategories(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	name := strings.ToLower(r.URL.Query().Get("name"))
	createdAtStart := r.URL.Query().Get("created_at_start")
	createdAtEnd := r.URL.Query().Get("created_at_end")
	updatedAtStart := r.URL.Query().Get("updated_at_start")
	updatedAtEnd := r.URL.Query().Get("updated_at_end")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	sortBy := strings.ToLower((r.URL.Query().Get("sort_by")))
	sortOrder := strings.ToUpper((r.URL.Query().Get("sort_order")))

	query := `
		SELECT id, name, created_at, updated_at
		FROM categories
		WHERE user_id = ?
	`

	var args []any
	args = append(args, userID)

	applyLikeFilter(&query, &args, name, "name")

	paramOperator := map[string]string{
		createdAtStart: "created_at >=",
		createdAtEnd:   "created_at <=",
		updatedAtStart: "updated_at >=",
		updatedAtEnd:   "updated_at <=",
	}
	err := applyDateFilters(&query, &args, paramOperator)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	columnNames, err := database.GetColumnNamesForTable(s.db, "categories")
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	applySorting(&query, sortBy, sortOrder, columnNames, "updated_at")
	applyPagination(&query, &args, limitStr, offsetStr)

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

	sendResponse(w, http.StatusOK, models.MultipleCategoriesResponse{Categories: categories})
}

//	@Summary		Create categories
//	@Description	Create one or more categories for the current user.
//	@Tags			categories
//	@Accept			json
//	@Produce		json
//	@Param			categories	body		[]models.CategoryPostRequest	true	"List of categories to create"
//	@Success		201			{object}	models.InsertedResponse			"Number of inserted categories"
//	@Failure		400			{object}	models.ErrorResponse
//	@Failure		500			{object}	models.ErrorResponse
//	@Security		BearerAuth
//	@Router			/categories [post]
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

	sendResponse(w, http.StatusCreated, models.InsertedResponse{Inserted: len(categories)})
}

//	@Summary		Delete categories
//	@Description	Delete one or more categories by their IDs.
//	@Tags			categories
//	@Accept			json
//	@Produce		json
//	@Param			ids	body		[]int					true	"List of category IDs to delete"
//	@Success		200	{object}	models.DeletedResponse	"Number of deleted categories"
//	@Failure		400	{object}	models.ErrorResponse	
//	@Failure		404	{object}	models.ErrorResponse
//	@Failure		500	{object}	models.ErrorResponse
//	@Security		BearerAuth
//	@Router			/categories [delete]
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
	sendResponse(w, http.StatusOK, models.DeletedResponse{Deleted: int(affected)})
}

//	@Summary		Update categories
//	@Description	Update name of one or multiple categories.
//	@Tags			categories
//	@Accept			json
//	@Produce		json
//	@Param			categories	body		[]models.CategoryPatchRequest	true	"List of categories with updated data"
//	@Success		200			{object}	models.UpdatedResponse			"Number of updated categories"
//	@Failure		400			{object}	models.ErrorResponse
//	@Failure		404			{object}	models.ErrorResponse
//	@Failure		500			{object}	models.ErrorResponse
//	@Security		BearerAuth
//	@Router			/categories [patch]
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

	sendResponse(w, http.StatusOK, models.UpdatedResponse{Updated: len(categories)})
}
