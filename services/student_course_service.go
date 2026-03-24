package services

import (
	"context"
	"fmt"

	"mcv_backend/config"
	"mcv_backend/dto"
	"mcv_backend/models"
	"time"
	
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"    
)
type CourseService struct {
	Collection *mongo.Collection
}
func NewCourseService(col *mongo.Collection) *CourseService {
	return &CourseService{
		Collection: col,
	}
}
// GetStudentCourses fetches a student's enrolled courses through the relational model:
// Student → Enrollment → CourseOffering → Course + Semester
func GetStudentCourses(ctx context.Context, userID primitive.ObjectID, semesterID *primitive.ObjectID) (*dto.StudentCoursesResponse, error) {
	enrollmentsColl := config.DB.Collection("enrollments")
	
	// Build query
	query := bson.M{
		"user_id": userID,
		"status":     models.StatusEnrolled,
	}
	if semesterID != nil {
		query["semester_id"] = *semesterID
	}

	pipeline := mongo.Pipeline{
    {{Key: "$match", Value: query}},
    {{Key: "$lookup", Value: bson.M{
        "from":         "course_offerings",
        "localField":   "course_offering_id",
        "foreignField": "_id",
        "as":           "offering",
    }}},
    {{Key: "$unwind", Value: "$offering"}},
    {{Key: "$lookup", Value: bson.M{
        "from":         "courses",
        "localField":   "offering.course_id",
        "foreignField": "_id",
        "as":           "course",
    }}},
    {{Key: "$unwind", Value: "$course"}},
    {{Key: "$lookup", Value: bson.M{
        "from":         "semesters",
        "localField":   "semester_id",
        "foreignField": "_id",
        "as":           "semester",
    }}},
    {{Key: "$unwind", Value: "$semester"}},
    {{Key: "$project", Value: bson.M{
        "enrollment_id": "$_id",
        "course_code":   "$course.course_id",
        "course_title":  "$course.title",
        "section":       "$offering.section",
        "instructor":    "$offering.instructor",
        "status":        "$status",
        "grade":         "$grade",
        "enrolled_at":   "$enrolled_at",
    }}},
}
	
	cursor, err := enrollmentsColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate enrollments: %w", err)
	}
	defer cursor.Close(ctx)

	var courses []dto.StudentCourseItem
	if err := cursor.All(ctx, &courses); err != nil {
		return nil, fmt.Errorf("failed to decode courses: %w", err)
	}

	// Determine which semester to return
	var responseSemesterID primitive.ObjectID
	if semesterID != nil {
		responseSemesterID = *semesterID
	} else if len(courses) > 0 {
		// Get semester from first enrollment
		var enrollment models.Enrollment
		if err := enrollmentsColl.FindOne(ctx, bson.M{"user_id": userID}).Decode(&enrollment); err == nil {
			responseSemesterID = enrollment.SemesterID
		}
	}
	var student *models.Student
	

	student, err = GetStudentByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch student by user_id: %w", err)
	}

	return &dto.StudentCoursesResponse{
		StudentID:  student.StudentID,
		SemesterID: responseSemesterID,
		Courses:    courses,
		Total:      len(courses),
	}, nil
}

// GetStudentByUserID retrieves a student record by user ID
func GetStudentByUserID(ctx context.Context, userID primitive.ObjectID) (*models.Student, error) {
	studentsColl := config.DB.Collection("students")
	
	var student models.Student
	err := studentsColl.FindOne(ctx, bson.M{"user_id": userID}).Decode(&student)
	if err != nil {
		return nil, fmt.Errorf("student not found for user: %w", err)
	}
	
	return &student, nil
}
func (s *CourseService) GetAllCourses(ctx context.Context) (*dto.CoursesResponse, error) {

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := s.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var courses []models.Course

	if err := cursor.All(ctx, &courses); err != nil {
		return nil, err
	}

	// map → DTO
	items := make([]dto.CourseItem, 0)

	for _, c := range courses {
		items = append(items, dto.CourseItem{
			ID:         c.ID.Hex(),
			CourseCode: c.CourseCode,
			Title:      c.Title,
			Credits:    c.Credits,
		})
	}

	return &dto.CoursesResponse{
		Courses: items,
		Total:   len(items),
	}, nil
}
