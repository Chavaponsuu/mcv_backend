package dto
import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"mcv_backend/models"
)
// ✅ ต้องมี bson tag ด้วย
type StudentCourseItem struct {
    EnrollmentID primitive.ObjectID `bson:"enrollment_id" json:"enrollment_id"`
    CourseCode   string             `bson:"course_code"   json:"course_code"`
    CourseTitle  string             `bson:"course_title"  json:"course_title"`
    Section      string             `bson:"section"       json:"section"`
    Instructor   string             `bson:"instructor"    json:"instructor"`
    Status       models.EnrollmentStatus             `bson:"status"        json:"status"`
    Grade        string             `bson:"grade"         json:"grade"`
    EnrolledAt   time.Time           `bson:"enrolled_at"   json:"enrolled_at"`
}
type StudentCoursesResponse struct {
    StudentID  string  `json:"student_id"`
    SemesterID primitive.ObjectID  `json:"semester_id"`
    Courses    []StudentCourseItem `json:"courses"`
    Total      int                 `json:"total"`
}





