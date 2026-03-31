package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"mcv_backend/dto"
	"mcv_backend/services"
)

// GetSemesterIDByYearAndTerm handles GET /api/student/me/semester?year=2024&term=1
// Fetches semester ID by year and term
func GetSemesterIDByYearAndTerm(w http.ResponseWriter, r *http.Request) {

	// Extract year and term from query parameters
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	yearStr := r.URL.Query().Get("year")
	termStr := r.URL.Query().Get("term")

	if yearStr == "" || termStr == "" {
		respondWithError(w, http.StatusBadRequest, "year and term parameters are required")
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid year parameter")
		return
	}

	term, err := strconv.Atoi(termStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid term parameter")
		return
	}

	// Validate term (1, 2, or 3)
	if term < 1 || term > 3 {
		respondWithError(w, http.StatusBadRequest, "Term must be 1 (first), 2 (second), or 3 (summer)")
		return
	}

	// Get semester ID from service
	semesterID, err := services.GetSemesterIDByYearAndTerm(ctx, year, term)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Semester not found: "+err.Error())
		return
	}

	// Return semester ID response
	response := dto.SemesterItem{
		ID:   semesterID.Hex(), // convert ObjectID -> string
		Year: year,
		Term: term,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func GetAllSemesters(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// call service
	semesters, err := services.GetAllSemesters(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}


	items := make([]dto.SemesterItem, 0, len(semesters))

	for _, s := range semesters {
		items = append(items, dto.SemesterItem{
			ID:   s.ID.Hex(),
			Year: s.Year,
			Term: s.Term,
		})
	}

	response := dto.SemestersResponse{
		Semesters: items,
		Total:     len(items),
	}

	// response json
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
