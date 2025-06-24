package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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
