package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"mcv_backend/models"
	"mcv_backend/services"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AssignmentHandler struct {
	AssignmentService *services.AssignmentService
	CourseService     *services.CourseService
}

func (h *AssignmentHandler) CreateAssignment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		CourseID    string    `json:"course_id"`
		Title       string    `json:"title"`
		Description string    `json:"description"`
		DueDate     time.Time `json:"due_date"`
		UserIDs     []string  `json:"user_ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	courseID, err := primitive.ObjectIDFromHex(req.CourseID)
	if err != nil {
		http.Error(w, "invalid course_id", http.StatusBadRequest)
		return
	}

	// Fetch course name from course service
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	course, err := h.CourseService.GetCourseByID(ctx, req.CourseID)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to fetch course: %v", err), http.StatusNotFound)
		return
	}

	var userIDs []primitive.ObjectID
	for _, id := range req.UserIDs {
		uid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, "invalid user_id", http.StatusBadRequest)
			return
		}
		userIDs = append(userIDs, uid)
	}

	assignment := &models.Assignment{
		CourseID:    courseID,
		Title:       course.Title,
		Description: req.Description,
		DueDate:    req.DueDate,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = h.AssignmentService.CreateAssignment(r.Context(), assignment, userIDs, course)
	if err != nil {
		http.Error(w, "failed to create assignment", http.StatusInternalServerError)
		return
	}

	// response
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "assignment created",
		"id":           assignment.ID,
		"course_name":  course.Title,
		"course_code":  course.CourseCode,
	})
}
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
