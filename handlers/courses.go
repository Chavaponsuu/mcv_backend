package handlers

import (
	"encoding/json"
	"net/http"

	"mcv_backend/models"
	"mcv_backend/services"

	"go.mongodb.org/mongo-driver/bson/primitive"
)
type CourseHandler struct {
	Service *services.CourseService
}
// GetUserCourses handles GET /api/courses?semester_id=xxx
// Fetches courses through the relational model: Student → Enrollment → CourseOffering → Course
func GetUserCourses(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context (set by auth middleware)
	userIDStr, ok := r.Context().Value("user_id").(string)
	if !ok || userIDStr == "" {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	// Get student record for this user
	student, err := services.GetStudentByUserID(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Student profile not found")
		return
	}

	// Get semester_id from query params (optional)
	var semesterID *primitive.ObjectID
	if semesterIDStr := r.URL.Query().Get("semester_id"); semesterIDStr != "" {
		sid, err := primitive.ObjectIDFromHex(semesterIDStr)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid semester ID")
			return
		}
		semesterID = &sid
	}

	// Fetch courses through relational model.
	// Enrollments are keyed by the student document id.
	response, err := services.GetStudentCourses(r.Context(), student.ID, semesterID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch courses: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (h *CourseHandler) GetAllCourses(w http.ResponseWriter, r *http.Request) {

	result, err := h.Service.GetAllCourses(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
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