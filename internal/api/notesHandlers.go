package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/siwiec987/notes-api/internal/database"
	"github.com/siwiec987/notes-api/internal/models"
)

//	@Summary		Get notes
//	@Description	Get a list of notes with optional filters, pagination and sorting.
//	@Tags			notes
//	@Produce		json
//	@Param			content				query		string	false	"Filter notes by content (case-insensitive)"
//	@Param			category_id			query		int		false	"Filter by category ID"
//	@Param			created_at_start	query		string	false	"Filter by created_at >= (format: 2006-01-02 15:04:05)"
//	@Param			created_at_end		query		string	false	"Filter by created_at <= (format: 2006-01-02 15:04:05)"
//	@Param			updated_at_start	query		string	false	"Filter by updated_at >= (format: 2006-01-02 15:04:05)"
//	@Param			updated_at_end		query		string	false	"Filter by updated_at <= (format: 2006-01-02 15:04:05)"
//	@Param			limit				query		int		false	"Number of notes to return"
//	@Param			offset				query		int		false	"Number of notes to skip"
//	@Param			sort_by				query		string	false	"Sort by column (id, content, created_at, updated_at, category_id, default: updated_at)"
//	@Param			sort_order			query		string	false	"Sort order (ASC or DESC)"
//	@Success		200					{object}	models.MultipleNotesResponse
//	@Failure		400					{object}	models.ErrorResponse
//	@Failure		500					{object}	models.ErrorResponse
//	@Security		BearerAuth
//	@Router			/notes [get]
func (s *APIServer) handleGetNotes(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	content := strings.ToLower(r.URL.Query().Get("content"))
	categoryIDStr := r.URL.Query().Get("category_id")
	createdAtStart := r.URL.Query().Get("created_at_start")
	createdAtEnd := r.URL.Query().Get("created_at_end")
	updatedAtStart := r.URL.Query().Get("updated_at_start")
	updatedAtEnd := r.URL.Query().Get("updated_at_end")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	sortBy := strings.ToLower((r.URL.Query().Get("sort_by")))
	sortOrder := strings.ToUpper((r.URL.Query().Get("sort_order")))

	query := `
		SELECT n.id, n.content, n.created_at, n.updated_at, n.category_id, c.name as category_name 
		FROM notes n
		JOIN categories c ON n.category_id = c.id 
		WHERE n.user_id = ?
	`
	var args []any
	args = append(args, userID)

	applyLikeFilter(&query, &args, content, "n.content")

	if categoryIDStr != "" {
		categoryID, err := strconv.Atoi(categoryIDStr)
		if err != nil {
			sendError(w, http.StatusBadRequest, "invalid category_id")
			return
		}
		args = append(args, categoryID)
		query += " AND n.category_id = ?"
	}

	paramOperator := map[string]string{
		createdAtStart: "n.created_at >=",
		createdAtEnd:   "n.created_at <=",
		updatedAtStart: "n.updated_at >=",
		updatedAtEnd:   "n.updated_at <=",
	}
	err := applyDateFilters(&query, &args, paramOperator)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	columnNames, err := database.GetColumnNamesForTable(s.db, "notes")
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	applySorting(&query, sortBy, sortOrder, columnNames, "updated_at")
	applyPagination(&query, &args, limitStr, offsetStr)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to fetch notes")
		return
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var note models.Note
		err := rows.Scan(&note.ID, &note.Content, &note.CreatedAt, &note.UpdatedAt, &note.CategoryID, &note.CategoryName)
		if err != nil {
			sendError(w, http.StatusInternalServerError, "failed to fetch notes")
			return
		}
		notes = append(notes, note)
	}

	err = rows.Err()
	if err != nil {
		sendError(w, http.StatusInternalServerError, "error reading notes from database")
		return
	}

	sendResponse(w, http.StatusOK, models.MultipleNotesResponse{Notes: notes})
}

//	@Summary		Create notes
//	@Description	Create one or more notes for the current user.
//	@Tags			notes
//	@Accept			json
//	@Produce		json
//	@Param			notes	body		[]models.NotePostRequest	true	"List of notes to create"
//	@Success		201		{object}	models.InsertedResponse		"Number of inserted notes"
//	@Failure		400		{object}	models.ErrorResponse
//	@Failure		500		{object}	models.ErrorResponse
//	@Security		BearerAuth
//	@Router			/notes [post]
func (s *APIServer) handlePostNotes(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	var notes []models.NotePostRequest
	err := json.NewDecoder(r.Body).Decode(&notes)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(notes) == 0 {
		sendError(w, http.StatusBadRequest, "no notes provided")
		return
	}

	err = database.WithTransaction(s.db, func(tx *sql.Tx) error {
		for _, note := range notes {
			_, err := tx.Exec(
				`INSERT INTO notes (content, category_id, user_id) 
				VALUES (?, ? ,?)`, note.Content, note.CategoryID, userID)
			if err != nil {
				return errors.New("invalid note data")
			}
		}

		return nil
	})
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendResponse(w, http.StatusCreated, models.InsertedResponse{Inserted: len(notes)})
}

//	@Summary		Delete notes
//	@Description	Delete one or more notes by their IDs.
//	@Tags			notes
//	@Accept			json
//	@Produce		json
//	@Param			ids	body		[]int					true	"List of note IDs to delete"
//	@Success		200	{object}	models.DeletedResponse	"Number of deleted notes"
//	@Failure		400	{object}	models.ErrorResponse
//	@Failure		404	{object}	models.ErrorResponse
//	@Failure		500	{object}	models.ErrorResponse
//	@Security		BearerAuth
//	@Router			/notes [delete]
func (s *APIServer) handleDeleteNotes(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	var noteIDs []int
	err := json.NewDecoder(r.Body).Decode(&noteIDs)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(noteIDs) == 0 {
		sendError(w, http.StatusBadRequest, "no IDs provided")
		return
	}

	placeholders, args := database.BuildQueryArgs(userID, noteIDs)

	notesExist, err := database.DoRecordsExistForUser(s.db, "notes", userID, noteIDs)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to verify notes existence")
		return
	}
	if !notesExist {
		sendError(w, http.StatusNotFound, "one or more notes not found")
		return
	}

	query := fmt.Sprintf("DELETE FROM notes WHERE user_id = ? AND id IN (%s)", placeholders)

	var res sql.Result
	err = database.WithTransaction(s.db, func(tx *sql.Tx) error {
		res, err = tx.Exec(query, args...)
		if err != nil {
			return errors.New("failed to delete notes")
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

//	@Summary		Update notes
//	@Description	Update content and/or category of one or multiple notes.
//	@Tags			notes
//	@Accept			json
//	@Produce		json
//	@Param			notes	body		[]models.NotePatchRequest	true	"List of notes to update"
//	@Success		200		{object}	models.UpdatedResponse		"Number of updated notes"
//	@Failure		400		{object}	models.ErrorResponse
//	@Failure		404		{object}	models.ErrorResponse
//	@Failure		500		{object}	models.ErrorResponse
//	@Security		BearerAuth
//	@Router			/notes [patch]
func (s *APIServer) handlePatchNotes(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	var notes []models.NotePatchRequest
	err := json.NewDecoder(r.Body).Decode(&notes)
	if err != nil {
		sendError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if len(notes) == 0 {
		sendError(w, http.StatusBadRequest, "no notes provided")
		return
	}

	var noteIDs []int
	for _, note := range notes {
		noteIDs = append(noteIDs, note.ID)
	}

	exist, err := database.DoRecordsExistForUser(s.db, "notes", userID, noteIDs)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to verify notes existence")
		return
	}
	if !exist {
		sendError(w, http.StatusNotFound, "one or more notes not found")
		return
	}

	toExecute := func(tx *sql.Tx) error {
		for _, note := range notes {
			var setClauses []string
			var args []any

			if note.Content != nil {
				setClauses = append(setClauses, "content = ?")
				args = append(args, *note.Content)
			}
			if note.CategoryID != nil {
				setClauses = append(setClauses, "category_id = ?")
				args = append(args, *note.CategoryID)
			}
			if len(setClauses) == 0 {
				continue
			}

			query := fmt.Sprintf(
				`UPDATE notes 
				SET %s 
				WHERE user_id = ? 
				AND id = ?`,
				strings.Join(setClauses, ","))

			args = append(args, userID, note.ID)

			_, err = tx.Exec(query, args...)
			if err != nil {
				return fmt.Errorf("failed to update note with id %d", note.ID)
			}
		}

		return nil
	}

	err = database.WithTransaction(s.db, toExecute)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	sendResponse(w, http.StatusOK, models.UpdatedResponse{Updated: len(notes)})
}
