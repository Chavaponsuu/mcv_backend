package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Course represents a course definition (not a specific offering)
type Course struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CourseCode  string             `bson:"course_code" json:"course_code"` // e.g., "2301107"
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Credits     int                `bson:"credits" json:"credits"`
	
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
