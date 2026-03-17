package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Course represents a course in MyCourseVille
type Course struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CourseID    string             `bson:"course_id" json:"course_id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Instructor  string             `bson:"instructor" json:"instructor"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// MyCourseVilleResponse represents the API response structure
type MyCourseVilleResponse struct {
	Data []Course `json:"data"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
