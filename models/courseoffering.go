package models

import (
	
	"go.mongodb.org/mongo-driver/bson/primitive"
)
type CourseOffering struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	CourseID   primitive.ObjectID `bson:"course_id"`
	SemesterID primitive.ObjectID `bson:"semester_id"`

	Section    string `bson:"section"`
	Instructor string `bson:"instructor"`
}