package services

import (
	"context"
	"fmt"

	"mcv_backend/config"
	"mcv_backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetSemesterIDByYearAndTerm retrieves the semester ID by year and term
func GetSemesterIDByYearAndTerm(ctx context.Context, year int, term int) (*primitive.ObjectID, error) {
	semestersColl := config.DB.Collection("semesters")

	// Query by year and term
	query := bson.M{
		"year": year,
		"term": term,
	}

	var semester models.Semester
	err := semestersColl.FindOne(ctx, query).Decode(&semester)
	if err != nil {
		return nil, fmt.Errorf("semester not found for year %d, term %d: %w", year, term, err)
	}

	return &semester.ID, nil
}

// GetSemesterIDByYearAndTerm retrieves the semester ID by year and term
func GetAllSemesters(ctx context.Context) ([]models.Semester, error) {
	semestersColl := config.DB.Collection("semesters")

	cursor, err := semestersColl.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var semesters []models.Semester
	if err := cursor.All(ctx, &semesters); err != nil {
		return nil, err
	}

	return semesters, nil
}