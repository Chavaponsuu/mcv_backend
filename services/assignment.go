package services
import (
	"time"
	"mcv_backend/models"
	"mcv_backend/domain"
	// "mcv_backend/services"
	"context"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AssignmentService struct {
	Collection *mongo.Collection
	EventBus   *EventBus // ใช้ของคุณ
}

func (as *AssignmentService) CreateAssignment(
	ctx context.Context, assignment *models.Assignment, userIDs []primitive.ObjectID, courseInfo *domain.CourseItem,
) error {
	assignment.ID = primitive.NewObjectID()
	assignment.CreatedAt = time.Now()
	_, err := as.Collection.InsertOne(ctx, assignment)
	if err != nil {
		return err
	}
	for _, uid := range userIDs {
		event := &models.DomainEvent{
			Type:          models.EventAssignmentCreated,
			AggregateID:   uid,
			AggregateType: "user",
			Data: map[string]interface{}{
				"name":            assignment.Title,
				"assignment_id":   assignment.ID.Hex(),
				"course_id": assignment.CourseID.Hex(),
				"course_name":     courseInfo.Title,
				"imageUrl":        courseInfo.ImageUrl,
			},
			CreatedAt:  assignment.CreatedAt,
			DeadlineAt: assignment.DueDate,
		}
		as.EventBus.PublishAsync(ctx, event, nil)
	}
	return nil
}

