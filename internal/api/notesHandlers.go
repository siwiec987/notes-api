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
	"github.com/siwiec987/notes-api/internal/validation"
)

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
	if content != "" {
		content = "%" + content + "%"
		args = append(args, content)
		query += " AND n.content LIKE ?"
	}
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

	err := validation.CreateDateFilters(&query, &args, paramOperator)
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
	query += " OFFSET ?"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "failed to fetch notes")
		return
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var note models.Note
		err := rows.Scan(&note.ID, &note.Content, &note.CreatedAt, &note.UpdatedAt, &note.Category.ID, &note.Category.Name)
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

	sendResponse(w, http.StatusOK, map[string]any{"notes": notes})
}

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

	sendResponse(w, http.StatusCreated, map[string]any{"inserted": len(notes)})
}

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
			return errors.New("failed to delete note(s)")
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

	sendResponse(w, http.StatusOK, map[string]any{"updated": len(notes)})
}
