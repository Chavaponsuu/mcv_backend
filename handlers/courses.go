package handlers

import (
	"encoding/json"
	"net/http"
	"context"
    "time"

	"mcv_backend/models"
	"mcv_backend/services"
	"mcv_backend/dto"

	"go.mongodb.org/mongo-driver/bson/primitive"
	
)
type CourseHandler struct {
	Service *services.CourseService
}

// GetUserCourses handles GET /api/student/me/courses?semester_id=xxx
// Fetches courses through the relational model: Student → Enrollment → CourseOffering → Course
func (h *CourseHandler) GetUserCourses(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from context (set by auth middleware)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

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
	student, err := services.GetStudentByUserID(ctx, userID)
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
	courses, err := services.GetStudentCourses(ctx, userID, semesterID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to fetch courses: "+err.Error())
		return
	}

	response := &dto.StudentCoursesResponse{
		StudentID: student.StudentID,
		Courses:   courses,
		Total:     len(courses),
	}
	if semesterID != nil {
		response.SemesterID = *semesterID
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (h *CourseHandler) GetAllCourses(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	
	result, err := h.Service.GetAllCourses(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := &dto.CoursesResponse{
		Courses: result,
		Total: len(result),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCourseByID handles GET /api/courses/:course_id
// Fetches a specific course by its ID
func (h *CourseHandler) GetCourseByID(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Extract course ID from URL path
	courseID := r.PathValue("course_id")
	if courseID == "" {
		respondWithError(w, http.StatusBadRequest, "Course ID is required")
		return
	}

	// Fetch course by ID from service
	course, err := h.Service.GetCourseByID(ctx, courseID)
	if err != nil {
		if err.Error() == "course not found" {
			respondWithError(w, http.StatusNotFound, "Course not found")
		} else {
			respondWithError(w, http.StatusInternalServerError, "Failed to fetch course: "+err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, course)
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