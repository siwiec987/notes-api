package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func sendResponse(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		log.Println("Error encoding response:", err)
	}
}

func sendError(w http.ResponseWriter, status int, msg any) {
	log.Printf("ERROR %d: %v", status, msg)

	switch msg.(type) {
	case string:
		sendResponse(w, status, map[string]any{"error": msg})
	default:
		sendResponse(w, status, map[string]any{"errors": msg})
	}
}

func getUserID(r *http.Request) int {
	val := r.Context().Value(userKey)
	userID, ok := val.(int)
	if !ok {
		panic("userID not found in context - authMiddleware must be used")
	}
	return userID
}

func parseLimit(limitStr string, defaultLimit int) int {
	limit := defaultLimit

	if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
		limit = parsed
	}

	return limit
}

func parseOffset(offsetStr string, defaultOffset int) int {
	offset := defaultOffset

	if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed > 0 {
		offset = parsed
	}

	return offset
}

func isDateCorrect(s string) bool {
	_, err := time.Parse("2006-01-02 15:04:05", s)
	return err == nil
}

func createDateFilter(query *string, args *[]any, param, operator string) error {
	if param != "" {
		if !isDateCorrect(param) {
			return errors.New("invalid date format, example: 2025-06-23 22:49:00")
		}

		*query += fmt.Sprintf(" AND %s ?", operator)
		*args = append(*args, param)
	}
	return nil
}

func createDateFilters(query *string, args *[]any, paramOperatorMap map[string]string) error {
	for param, operator := range paramOperatorMap {
		err := createDateFilter(query, args, param, operator)
		if err != nil {
			return err
		}
	}

	return nil
}

func withTransaction(db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.Begin()
	if err != nil {
		return errors.New("failed to start transaction")
	}
	defer tx.Rollback()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.New("failed to commit transaction")
	}

	return nil
}

func buildQueryArgs(userID int, ids []int) (string, []any) {
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids) + 1)
	args[0] = userID
	for i, id := range ids {
		placeholders[i] = "?"
		args[i + 1] = id
	}

	return strings.Join(placeholders, ","), args
}
