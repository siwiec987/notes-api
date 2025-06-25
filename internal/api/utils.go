package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
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

func applyDateFilters(query *string, args *[]any, paramOperatorMap map[string]string) error {
	for param, operator := range paramOperatorMap {
		err := createDateFilter(query, args, param, operator)
		if err != nil {
			return err
		}
	}

	return nil
}

func applyLikeFilter(query *string, args *[]any, filter, columnName string) {
	if filter != "" {
		filter = "%" + filter + "%"
		*args = append(*args, filter)
		*query += fmt.Sprintf(" AND %s LIKE ?", columnName)
	}
}

func applyPagination(query *string, args *[]any, limitStr, offsetStr string) {
	limit := parseLimit(limitStr, 20)
	*args = append(*args, limit)
	*query += " LIMIT ?"
	
	offset := parseOffset(offsetStr, 0)
	*args = append(*args, offset)
	*query += " OFFSET ?"
}

func applySorting(query *string, sortBy, sortOrder string, allowedColumns []string, defaultColumn string) {
	if !slices.Contains(allowedColumns, sortBy) {
		sortBy = defaultColumn
	}

	if sortOrder != "ASC" && sortOrder != "DESC" {
		sortOrder = "DESC"
	}

	*query += fmt.Sprintf(" ORDER BY %s %s", sortBy, sortOrder)
}
