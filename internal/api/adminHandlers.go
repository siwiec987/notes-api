package api

import (
	"net/http"

	"github.com/siwiec987/notes-api/internal/migrations"
)

func (s *APIServer) handleSeed(w http.ResponseWriter, r *http.Request) {
	err := migrations.SeedData(s.db)
	if err != nil {
		sendError(w, http.StatusInternalServerError, "Failed to seed database: " + err.Error())
		return
	}
	sendResponse(w, http.StatusOK, map[string]string{"message": "Database seeded successfully"})
}