package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NotificationStatus represents the state of a notification
type NotificationStatus string

const (
	NotificationUnread NotificationStatus = "unread"
	NotificationRead   NotificationStatus = "read"
	NotificationDismissed NotificationStatus = "dismissed"
)

// NotificationType categorizes different notification events
type NotificationType string

const (
	TypeAssignmentCreated    NotificationType = "assignment.created"
	TypeAssignmentDeadline   NotificationType = "assignment.deadline"
)

// Notification represents a user notification
type Notification struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UserID    primitive.ObjectID `bson:"user_id"`
	Type      NotificationType   `bson:"type"`
	Title     string             `bson:"title"`
	Message   string             `bson:"message"`
	Status    NotificationStatus `bson:"status"`
	AssigmentID primitive.ObjectID `bson:"related_id,omitempty"` // courseID, enrollmentID, etc.
	ImageUrl string `bson:"imageUrl"`
	// Data      map[string]interface{} `bson:"data,omitempty"`   // Extra metadata
	CreatedAt time.Time          `bson:"created_at"`
	// ReadAt    *time.Time         `bson:"read_at,omitempty"`
	// UpdatedAt time.Time          `bson:"updated_at"`
	DeadlineAt time.Time `bson:"deadline_at"`
}

// NotificationPreference represents user notification settings
type NotificationPreference struct {
	ID                     primitive.ObjectID `bson:"_id,omitempty"`
	UserID                 primitive.ObjectID `bson:"user_id"`
	EnableAssignmentNotifications bool        `bson:"enable_assignment_notifications"`
	CreatedAt              time.Time          `bson:"created_at"`
	UpdatedAt              time.Time          `bson:"updated_at"`
}
