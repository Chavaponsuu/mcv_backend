package dto
import (
	
    "go.mongodb.org/mongo-driver/bson/primitive"
    "mcv_backend/domain"
	
)

type StudentCoursesResponse struct {
    StudentID  string  `json:"student_id"`
    SemesterID primitive.ObjectID  `json:"semester_id"`
    Courses    []domain.StudentCourseItem `json:"courses"`
    Total      int                 `json:"total"`
}





