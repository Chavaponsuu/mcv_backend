package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// EventType represents different domain events
type EventType string

const (
	EventAssignmentCreated   EventType = "assignment.created"
	EventAssignmentDeadline  EventType = "assignment.deadline"
)

// DomainEvent represents an event that occurred in the domain
type DomainEvent struct {
	ID        primitive.ObjectID         `bson:"_id,omitempty"`
	Type      EventType                  `bson:"type"`
	AggregateID primitive.ObjectID       `bson:"aggregate_id"` // The primary entity ID (userID, courseID, etc.)
	AggregateType string                 `bson:"aggregate_type"` // Type of entity (user, enrollment, etc.)
	ActorID   primitive.ObjectID         `bson:"actor_id,omitempty"` // Who triggered the event
	Data      map[string]interface{}     `bson:"data"` // Event-specific data
	Metadata  map[string]interface{}     `bson:"metadata,omitempty"` // Additional context
	CreatedAt time.Time                  `bson:"created_at"`
	ProcessedAt *time.Time               `bson:"processed_at,omitempty"`
	DeadlineAt time.Time `bson:"deadline_at"`
}

// EventHandler defines the interface for event subscribers
type EventHandler interface {
	Handle(event *DomainEvent) error
	SupportedEventTypes() []EventType
}
