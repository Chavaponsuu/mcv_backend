package handlers

import (
	"encoding/json"
	"net/http"

	"mcv_backend/models"
	"mcv_backend/services"
)

// GetUserCourses handles GET /api/courses
func GetUserCourses(w http.ResponseWriter, r *http.Request) {
	// Extract access token from Authorization header
	token := r.Header.Get("Authorization")
	if token == "" {
		respondWithError(w, http.StatusUnauthorized, "Missing authorization token")
		return
	}

	// Remove "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Fetch courses from MyCourseVille API
	courses, err := services.FetchUserCourses(token)
	if err != nil {
		if err == services.ErrUnauthorized {
			respondWithError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}
		if err == services.ErrServiceUnavailable {
			respondWithError(w, http.StatusServiceUnavailable, "MyCourseVille API unavailable")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch courses")
		return
	}

	respondWithJSON(w, http.StatusOK, courses)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, models.ErrorResponse{Error: message})
}
