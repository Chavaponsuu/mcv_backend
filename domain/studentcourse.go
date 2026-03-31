// ✅ ต้องมี bson tag ด้วย'
package domain
import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"mcv_backend/models"
)


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
type CourseItem struct {
	ID         string `json:"id"`
	CourseCode string `json:"course_code"`
	Title      string `json:"title"`
	Credits    int    `json:"credits"`
}
