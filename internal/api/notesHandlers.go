package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/siwiec987/notes-api/internal/models"
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
			sendError(w, http.StatusBadRequest, "Invalid category_id")
			return
		}
		args = append(args, categoryID)
		query += " AND n.category_id = ?"
	}
	if createdAtStart != "" {
		if !isDateCorrect(createdAtStart) {
			sendError(w, http.StatusBadRequest, "Invalid date format")
			return
		}
		args = append(args, createdAtStart)
		query += " AND n.created_at >= ?"
	}
	if createdAtEnd != "" {
		if !isDateCorrect(createdAtEnd) {
			sendError(w, http.StatusBadRequest, "Invalid date format")
			return
		}
		args = append(args, createdAtEnd)
		query += " AND n.created_at <= ?"
	}
	if updatedAtStart != "" {
		if !isDateCorrect(updatedAtStart) {
			sendError(w, http.StatusBadRequest, "Invalid date format")
			return
		}
		args = append(args, updatedAtStart)
		query += " AND n.updated_at >= ?"
	}
	if updatedAtEnd != "" {
		if !isDateCorrect(updatedAtEnd) {
			sendError(w, http.StatusBadRequest, "Invalid date format")
			return
		}
		args = append(args, updatedAtEnd)
		query += " AND n.updated_at <= ?"
	}

	limit := 20
	if limitStr != "" {
		parsed, err := strconv.Atoi(limitStr)
		if err == nil && parsed > 0 {
			limit = parsed
		}
	}
	args = append(args, limit)
	query += " LIMIT ?"
	
	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err == nil && offset > 0{
			args = append(args, offset)
			query += " OFFSET ?"
		}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to fetch notes")
		return
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var note models.Note
		err := rows.Scan(&note.ID, &note.Content, &note.CreatedAt, &note.UpdatedAt, &note.Category.ID, &note.Category.Name)
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

	sendResponse(w, http.StatusOK, map[string]any { "notes": notes })
}

func (s *APIServer) handlePostNotes(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)
	var notes []models.NotePostRequest
	err := json.NewDecoder(r.Body).Decode(&notes)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if len(notes) == 0 {
		sendError(w, http.StatusBadRequest, "No notes provided")
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

	for _, note := range notes {
		_, err := tx.Exec(
			`INSERT INTO notes (content, category_id, user_id) 
			VALUES (?, ? ,?)`, note.Content, note.CategoryID, userID)
		if err != nil {
			tx.Rollback()
			sendError(w, http.StatusBadRequest, "Invalid note data")
			return
		}
	}

	if err := tx.Commit(); err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	sendResponse(w, http.StatusCreated, map[string]any { "inserted": len(notes) })
}

func (s *APIServer) handleDeleteNotes(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	var noteIDs []int
	err := json.NewDecoder(r.Body).Decode(&noteIDs)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if len(noteIDs) == 0 {
		sendError(w, http.StatusBadRequest, "No IDs provided")
		return
	}

	placeholders := make([]string, len(noteIDs))
	args := make([]any, len(noteIDs) + 1)
	args[0] = userID
	for i, id := range noteIDs {
		placeholders[i] = "?"
		args[i + 1] = id
	}

	notesExist, err := s.doNotesExist(userID, noteIDs)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to verify notes existence")
		return
	}
	if !notesExist {
		sendError(w, http.StatusNotFound, "One or more notes not found")
		return
	}

	query := fmt.Sprintf("DELETE FROM notes WHERE user_id = ? AND id IN (%s)", strings.Join(placeholders, ","))

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
		sendError(w, http.StatusInternalServerError, "Failed to delete note(s)")
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
	sendResponse(w, http.StatusOK, map[string]any { "deleted": affected })
}

func (s *APIServer) handlePatchNotes(w http.ResponseWriter, r *http.Request) {
	userID := getUserID(r)

	var notes []models.NotePatchRequest
	err := json.NewDecoder(r.Body).Decode(&notes)
	if err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if len(notes) == 0 {
		sendError(w, http.StatusBadRequest, "No notes provided")
		return
	}

	var noteIDs []int
	for _, note := range notes {
		noteIDs = append(noteIDs, note.ID)
	}

	exist, err := s.doNotesExist(userID, noteIDs)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to verify notes existence")
		return
	}
	if !exist {
		sendError(w, http.StatusNotFound, "One or more notes not found")
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
			tx.Rollback()
			sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update note with id %d", note.ID))
			return
		}
	}

	if err := tx.Commit(); err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	sendResponse(w, http.StatusOK, map[string]any { "updated": len(notes) })
}

func (s *APIServer) doNotesExist(userID int, noteIDs []int) (bool, error) {
	placeholders := make([]string, len(noteIDs))
	args := make([]any, len(noteIDs)+1)
	args[0] = userID
	for i, id := range noteIDs {
		placeholders[i] = "?"
		args[i+1] = id
	}

	queryCheck := fmt.Sprintf(`
    SELECT COUNT(*) FROM notes 
    WHERE user_id = ? AND id IN (%s)`, strings.Join(placeholders, ","))

	var count int
	err := s.db.QueryRow(queryCheck, args...).Scan(&count)
	return count == len(noteIDs), err
} 
