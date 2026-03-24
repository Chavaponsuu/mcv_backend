package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EnrollmentStatus string

const (
	StatusEnrolled  EnrollmentStatus = "enrolled"
	StatusDropped   EnrollmentStatus = "dropped"
	StatusCompleted EnrollmentStatus = "completed"
)

// Enrollment represents a student's enrollment in a course offering
type Enrollment struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        primitive.ObjectID `bson:"student_id" json:"student_id"`
	CourseOfferingID primitive.ObjectID `bson:"course_offering_id" json:"course_offering_id"`
	
	// Denormalized for faster queries
	SemesterID primitive.ObjectID `bson:"semester_id" json:"semester_id"`
	
	Status     EnrollmentStatus `bson:"status" json:"status"`
	Grade      string           `bson:"grade,omitempty" json:"grade,omitempty"`
	
	EnrolledAt time.Time `bson:"enrolled_at" json:"enrolled_at"`
	UpdatedAt  time.Time `bson:"updated_at" json:"updated_at"`
}



// Student represents a student in the system
