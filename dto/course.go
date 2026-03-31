package dto

import (

	"mcv_backend/domain"
)
type CoursesResponse struct {
	Courses []domain.CourseItem `json:"courses"`
	Total   int          `json:"total"`
}