package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

func isDateCorrect(s string) bool {
	date, err := time.Parse("2006-01-02 15:04", s)
	fmt.Println(date)
	fmt.Println(err)
	return err == nil
}

func getUserID(r *http.Request) int {
	val := r.Context().Value(userKey)
	userID, ok := val.(int)
	if !ok {
		panic("userID not found in context - authMiddleware must be used")
	}
	return userID
}
