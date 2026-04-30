package services

import (
	"context"
	"fmt"
	"time"

	"mcv_backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NotificationService handles notification creation and management
type NotificationService struct {
	Collection *mongo.Collection
	// PreferencesColl *mongo.Collection
}

// NewNotificationService creates a new notification service
func NewNotificationService(notifColl, prefColl *mongo.Collection) *NotificationService {
	return &NotificationService{
		Collection: notifColl,
		// PreferencesColl: prefColl,
	}
}

// Handle processes domain events and creates notifications
func (ns *NotificationService) Handle(event *models.DomainEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	fmt.Println("🔥 HANDLE CALLED:", event.Type)
	// Create notification based on event type
	fmt.Println("MAPPING EVENT:", event.Type)

notification := ns.mapEventToNotification(event)
if notification == nil {
	fmt.Println("❌ mapping returned nil")
	return nil
}
fmt.Println("✅ notification created:", notification)
	// Check user preferences
	// prefs, err := ns.GetUserPreferences(ctx, notification.UserID.Hex())
	// if err == nil && !ns.isNotificationEnabled(prefs, event.Type) {
	// 	return nil // User has disabled this notification type
	// }

	// Insert notification
		_, err := ns.Collection.InsertOne(ctx, notification)
	if err != nil {
		return fmt.Errorf("failed to insert notification: %w", err)
	}
	fmt.Println("✅ INSERT SUCCESS")

	return nil
}

// SupportedEventTypes returns the event types this handler supports
func (ns *NotificationService) SupportedEventTypes() []models.EventType {
	return []models.EventType{
		models.EventAssignmentCreated,
		models.EventAssignmentDeadline,
	}
}

// mapEventToNotification converts domain events to notifications
func (ns *NotificationService) mapEventToNotification(event *models.DomainEvent) *models.Notification {
	switch event.Type {
	case models.EventAssignmentCreated:
		assignmentName := "Assignment"
		if name, ok := event.Data["name"].(string); ok {
			assignmentName = name
		}
		courseName := "Course"
		if name, ok := event.Data["course_name"].(string); ok {
			courseName = name
		}
		imageUrl := ""
		if url, ok := event.Data["imageUrl"].(string); ok {
			imageUrl = url
		}
		return &models.Notification{
			UserID:      event.AggregateID,
			Type:        models.TypeAssignmentCreated,
			Title:       courseName,
			Message:     fmt.Sprintf("New assignment: %s", assignmentName),
			Status:      models.NotificationUnread,
			AssigmentID: event.ID,
			ImageUrl: imageUrl,
			CreatedAt:   time.Now(),
			DeadlineAt: event.DeadlineAt,

		}

	case models.EventAssignmentDeadline:
		assignmentName := "Assignment"
		if name, ok := event.Data["name"].(string); ok {
			assignmentName = name
		}
		deadline := "upcoming"
		if d, ok := event.Data["deadline"].(string); ok {
			deadline = d
		}
		return &models.Notification{
			UserID:    event.AggregateID,
			Type:      models.TypeAssignmentDeadline,
			Title:     "Assignment Deadline",
			Message:   fmt.Sprintf("%s is due %s", assignmentName, deadline),
			Status:    models.NotificationUnread,
			AssigmentID: event.ID,
			CreatedAt: time.Now(),
		}

	default:
		return nil
	}
}

// isNotificationEnabled checks if user has enabled this notification type
func (ns *NotificationService) isNotificationEnabled(prefs *models.NotificationPreference, eventType models.EventType) bool {
	if prefs == nil {
		return true // Default to enabled
	}

	switch eventType {
	case models.EventAssignmentCreated, models.EventAssignmentDeadline:
		return prefs.EnableAssignmentNotifications
	default:
		return true
	}
}

// GetUserNotifications retrieves unread notifications for a user
func (ns *NotificationService) GetUserNotifications(ctx context.Context, userID primitive.ObjectID, limit int64) ([]models.Notification, error) {
	filter := bson.M{
		"user_id": userID,
		"status": models.NotificationUnread,
	}

	opts := options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(limit)
	cursor, err := ns.Collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notifications []models.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		return nil, err
	}

	return notifications, nil
}

// MarkAsRead marks a notification as read
func (ns *NotificationService) MarkAsRead(ctx context.Context, notificationID string) error {
	objID, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		return err
	}

	now := time.Now()
	_, err = ns.Collection.UpdateByID(ctx, objID, bson.M{
		"$set": bson.M{
			"status": models.NotificationRead,
			"read_at": now,
			"updated_at": now,
		},
	})
	return err
}

// MarkAllAsRead marks all unread notifications as read for a user
func (ns *NotificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	filter := bson.M{
		"user_id": userID,
		"status": models.NotificationUnread,
	}

	update := bson.M{
		"$set": bson.M{
			"status": models.NotificationRead,
			"read_at": time.Now(),
			"updated_at": time.Now(),
		},
	}

	_, err := ns.Collection.UpdateMany(ctx, filter, update)
	return err
}

// DeleteNotification deletes a notification
func (ns *NotificationService) DeleteNotification(
	ctx context.Context,
	notificationID string,
) error {

	objID, err := primitive.ObjectIDFromHex(notificationID)
	if err != nil {
		return err
	}

	result, err := ns.Collection.DeleteOne(ctx, bson.M{
		"_id": objID,
	})
	if err != nil {
		return err
	}

	// 🔥 เช็คว่า delete จริงไหม
	if result.DeletedCount == 0 {
		return fmt.Errorf("notification not found")
	}

	return nil
}
// // GetUserPreferences retrieves notification preferences for a user
// func (ns *NotificationService) GetUserPreferences(ctx context.Context, userID string) (*models.NotificationPreference, error) {
// 	objID, err := primitive.ObjectIDFromHex(userID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var prefs models.NotificationPreference
// 	err = ns.PreferencesColl.FindOne(ctx, bson.M{"user_id": objID}).Decode(&prefs)
// 	if err == mongo.ErrNoDocuments {
// 		return nil, nil // No preferences set, use defaults
// 	}
// 	return &prefs, err
// }

// SetUserPreferences updates notification preferences for a user
// func (ns *NotificationService) SetUserPreferences(ctx context.Context, userID string, prefs *models.NotificationPreference) error {
// 	objID, err := primitive.ObjectIDFromHex(userID)
// 	if err != nil {
// 		return err
// 	}

// 	prefs.UserID = objID
// 	prefs.UpdatedAt = time.Now()
// 	if prefs.ID == primitive.NilObjectID {
// 		prefs.ID = primitive.NewObjectID()
// 		prefs.CreatedAt = time.Now()
// 	}

// 	opts := options.Update().SetUpsert(true)
// 	_, err = ns.PreferencesColl.UpdateByID(ctx, prefs.ID, bson.M{"$set": prefs}, opts)
// 	return err
// }
