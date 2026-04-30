package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	
	"mcv_backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// EventBus manages event publishing and subscription
type EventBus struct {
	handlers map[models.EventType][]models.EventHandler
	mu       sync.RWMutex
	eventsColl *mongo.Collection
}

// NewEventBus creates a new event bus instance
func NewEventBus(db *mongo.Database) *EventBus {
	return &EventBus{
		handlers: make(map[models.EventType][]models.EventHandler),
		eventsColl: db.Collection("domain_events"),
	}
}

// Subscribe registers an event handler for specific event types
func (eb *EventBus) Subscribe(handler models.EventHandler) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	for _, eventType := range handler.SupportedEventTypes() {
		eb.handlers[eventType] = append(eb.handlers[eventType], handler)
	}
}

// Publish publishes an event to all registered handlers and stores it
func (eb *EventBus) Publish(ctx context.Context, event *models.DomainEvent) error {
	// Generate ID if not set
	if event.ID == primitive.NilObjectID {
		event.ID = primitive.NewObjectID()
	}

	// Set creation timestamp if not set
	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}
	fmt.Println("🚀 PUBLISH:", event.Type)

	// Store event in database for audit trail
	if eb.eventsColl != nil {
		_, err := eb.eventsColl.InsertOne(ctx, event)
		if err != nil {
			fmt.Printf("failed to store event: %v\n", err)
			// Don't fail the entire publish operation if storage fails
		}
	}

	// Notify handlers synchronously
	eb.mu.RLock()
	handlers := eb.handlers[event.Type]
	eb.mu.RUnlock()

	var errs []error
	for _, handler := range handlers {
		if err := handler.Handle(event); err != nil {
			errs = append(errs, err)
			fmt.Printf("error handling event %s: %v\n", event.Type, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("one or more handlers failed: %v", errs)
	}

	return nil
}

// PublishAsync publishes an event asynchronously (non-blocking)
func (eb *EventBus) PublishAsync(ctx context.Context, event *models.DomainEvent, onError func(error)) {
	go func() {
		if err := eb.Publish(ctx, event); err != nil {
			if onError != nil {
				onError(err)
			}
		}
	}()
}

// GetUnprocessedEvents retrieves events that haven't been processed yet
func (eb *EventBus) GetUnprocessedEvents(ctx context.Context, eventType models.EventType) ([]models.DomainEvent, error) {
	if eb.eventsColl == nil {
		return nil, fmt.Errorf("events collection not configured")
	}

	query := bson.M{"type": eventType, "processed_at": nil}
	cursor, err := eb.eventsColl.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []models.DomainEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}

// MarkEventProcessed marks an event as processed
func (eb *EventBus) MarkEventProcessed(ctx context.Context, eventID primitive.ObjectID) error {
	if eb.eventsColl == nil {
		return fmt.Errorf("events collection not configured")
	}

	now := time.Now()
	_, err := eb.eventsColl.UpdateByID(ctx, eventID, bson.M{
		"$set": bson.M{
			"processed_at": now,
		},
	})
	return err
}
