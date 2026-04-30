package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AssignmentStatus string

const (
	AssignmentOpen     AssignmentStatus = "open"
	AssignmentClosed   AssignmentStatus = "closed"
	AssignmentUpcoming AssignmentStatus = "upcoming"
)

// Assignment represents a course assignment
type Assignment struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CourseID primitive.ObjectID `bson:"course_id" json:"course_id"`
	
	Title       string    `bson:"title" json:"title"`
	Description string    `bson:"description" json:"description"`
	DueDate     time.Time `bson:"due_date" json:"due_date"`
	OpenDate    time.Time `bson:"open_date" json:"open_date"`
	MaxScore    int       `bson:"max_score" json:"max_score"`
	
	Status AssignmentStatus `bson:"status" json:"status"`
	
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// Submission represents a student's assignment submissio