# Event-Driven Notification System Architecture

## Overview

This document outlines the event-driven notification structure implemented in the MyCourseVille backend. The system allows decoupled communication between services through domain events and notifications.

## Architecture Components

### 1. **Domain Events** (`models/event.go`)

Events represent significant things that happen in the system:

```go
type DomainEvent struct {
    ID              primitive.ObjectID    // Unique event ID
    Type            EventType             // Event classification
    AggregateID     primitive.ObjectID    // Primary entity ID (userID, courseID, etc.)
    AggregateType   string                // Entity type (user, enrollment, etc.)
    ActorID         primitive.ObjectID    // Who triggered the event
    Data            map[string]interface{}// Event-specific data
    Metadata        map[string]interface{}// Additional context
    CreatedAt       time.Time
    ProcessedAt     *time.Time
}
```

**Supported Event Types:**

- `user.registered` - New user joins system
- `user.logged_in` - User authentication
- `enrollment.created` - Student enrolls in course
- `enrollment.dropped` - Student drops course
- `enrollment.canceled` - Enrollment canceled
- `grade.updated` - Grade posted to student
- `course.created` - New course added
- `course.updated` - Course information changed
- `semester.started` - New semester begins
- `semester.ended` - Semester concluded

### 2. **Notifications** (`models/notification.go`)

Notifications are user-facing messages generated from events:

```go
type Notification struct {
    ID          primitive.ObjectID
    UserID      primitive.ObjectID
    Type        NotificationType        // enrollment.confirmed, grade.posted, etc.
    Title       string                  // Short title
    Message     string                  // Detailed message
    Status      NotificationStatus      // unread, read, dismissed
    RelatedID   primitive.ObjectID      // References event/course/enrollment
    Data        map[string]interface{}  // Extra metadata for UI
    CreatedAt   time.Time
    ReadAt      *time.Time
    UpdatedAt   time.Time
}
```

**Notification Statuses:**

- `unread` - New notification
- `read` - User has viewed it
- `dismissed` - User dismissed it

### 3. **Event Bus** (`services/event_bus.go`)

Central pub/sub mechanism for publishing and subscribing to events:

```go
type EventBus struct {
    handlers    map[EventType][]EventHandler
    eventsColl  *mongo.Collection
}

// Methods:
Publish(ctx, event)               // Synchronously publish event
PublishAsync(ctx, event, onError)  // Asynchronously publish event
Subscribe(handler)                 // Register event handler
GetUnprocessedEvents(ctx, type)    // Retrieve pending events
MarkEventProcessed(ctx, eventID)   // Mark event as processed
```

### 4. **Notification Service** (`services/notification_service.go`)

Implements `EventHandler` interface to react to events:

```go
type NotificationService struct {
    Collection      *mongo.Collection
    PreferencesColl *mongo.Collection
}

// Event handling:
Handle(event)                      // Process event and create notification
SupportedEventTypes()              // Return handled event types

// Notification management:
GetUserNotifications(ctx, userID, limit)
MarkAsRead(ctx, notificationID)
MarkAllAsRead(ctx, userID)
DeleteNotification(ctx, notificationID)

// Preferences:
GetUserPreferences(ctx, userID)
SetUserPreferences(ctx, userID, prefs)
```

### 5. **Notification Handlers** (`handlers/notifications.go`)

REST API endpoints for user notification management:

```
GET  /api/notifications              - Get unread notifications
POST /api/notifications/:id/read     - Mark as read
POST /api/notifications/read-all     - Mark all as read
DELETE /api/notifications/:id        - Delete notification
GET  /api/notifications/preferences  - Get user preferences
POST /api/notifications/preferences  - Update preferences
```

## Data Flow

### Publishing an Event

```
Service/Handler
    ↓
EventBus.Publish(event)
    ↓
├─ Store event in MongoDB (audit trail)
    ↓
├─ Notify all registered handlers
    ├─ NotificationService.Handle()
    |   ├─ Check user preferences
    |   └─ Create Notification document
    └─ Other handlers...
```

### User Consuming Notifications

```
Frontend polls /api/notifications
    ↓
Handler calls NotificationService.GetUserNotifications()
    ↓
Query MongoDB for unread notifications
    ↓
Return to frontend with notification list
```

## Integration Guide

### 1. Initialize Event Bus in main.go

```go
// In main.go
import "mcv_backend/models"
import "mcv_backend/services"

func main() {
    // ... existing code ...

    // Initialize Event Bus
    eventBus := services.NewEventBus(config.DB)

    // Initialize Notification Service
    notifColl := config.DB.Collection("notifications")
    prefColl := config.DB.Collection("notification_preferences")
    notifService := services.NewNotificationService(notifColl, prefColl)

    // Subscribe notification service to events
    eventBus.Subscribe(notifService)

    // Pass eventBus and notifService to handlers/services that need them
    // Store in context or pass as parameters
}
```

### 2. Publish Events from Services

```go
// Example: In auth service when user registers
func (as *AuthService) Register(ctx context.Context, req *RegisterRequest) error {
    // ... registration logic ...

    // Publish event
    event := &models.DomainEvent{
        Type:          models.EventUserRegistered,
        AggregateID:   userID,
        AggregateType: "user",
        ActorID:       userID,
        Data: map[string]interface{}{
            "email": req.Email,
            "name":  req.Name,
        },
    }

    eventBus.Publish(ctx, event) // or PublishAsync for non-blocking
}
```

### 3. Publish Events from Enrollment

```go
// In enrollment handler or service
event := &models.DomainEvent{
    Type:          models.EventEnrollmentCreated,
    AggregateID:   userID,
    AggregateType: "enrollment",
    ActorID:       userID,
    Data: map[string]interface{}{
        "course_id":    courseID,
        "course_name":  courseName,
        "semester_id":  semesterID,
    },
}

eventBus.Publish(ctx, event)
```

### 4. Add Notification Routes

```go
// In main.go router setup
protected := router.Group("/api")
protected.Use(AuthMiddleware)
{
    protected.GET("/notifications", NotificationsHandler(notifService))
    protected.POST("/notifications/:id/read", MarkAsReadHandler(notifService))
    protected.POST("/notifications/read-all", MarkAllAsReadHandler(notifService))
    protected.DELETE("/notifications/:id", DeleteNotificationHandler(notifService))
    protected.GET("/notifications/preferences", GetPreferencesHandler(notifService))
    protected.POST("/notifications/preferences", UpdatePreferencesHandler(notifService))
}
```

## MongoDB Collections

```
notifications/
  - Stores user notifications
  - Indexed by: user_id, status, created_at

notification_preferences/
  - Stores per-user notification settings
  - Indexed by: user_id

domain_events/
  - Audit trail of all domain events
  - Indexed by: type, created_at, aggregate_id
  - Used for replay and debugging
```

## Event Handling Patterns

### Synchronous Handling

```go
eventBus.Publish(ctx, event) // Blocks until all handlers complete
```

### Asynchronous Handling

```go
eventBus.PublishAsync(ctx, event, func(err error) {
    if err != nil {
        log.Printf("Event handling failed: %v", err)
    }
})
```

### Custom Event Handlers

Implement the `EventHandler` interface:

```go
type CustomHandler struct {
    // ... dependencies ...
}

func (h *CustomHandler) Handle(event *models.DomainEvent) error {
    // Process event
    return nil
}

func (h *CustomHandler) SupportedEventTypes() []models.EventType {
    return []models.EventType{
        models.EventEnrollmentCreated,
        models.EventGradeUpdated,
    }
}

// Subscribe
eventBus.Subscribe(&customHandler)
```

## User Preferences

Users can control which notifications they receive:

```go
prefs := &models.NotificationPreference{
    EnableEnrollment:    true,
    EnableGrades:       true,
    EnableCourseUpdates: true,
    EnableAnnouncements: false,
}

notifService.SetUserPreferences(ctx, userID, prefs)
```

## Events Audit Trail

All events are stored in MongoDB for audit purposes:

```go
// Retrieve unprocessed events
events, _ := eventBus.GetUnprocessedEvents(ctx, models.EventEnrollmentCreated)

// Mark event as processed
eventBus.MarkEventProcessed(ctx, eventID)
```

## Future Enhancements

1. **Message Queues**: Use RabbitMQ/Redis for distributed event handling
2. **Email Notifications**: Extend NotificationService to send emails
3. **Push Notifications**: Integrate with Firebase/OneSignal
4. **Event Replay**: Replay events from database for recovery
5. **Event Sourcing**: Store all state changes as events
6. **Webhook Support**: Allow external systems to subscribe to events
7. **Event Filtering**: Advanced query builder for event subscriptions
8. **Rate Limiting**: Prevent notification spam to users

## Testing

```go
// Mock EventBus for testing
mockBus := &MockEventBus{}
mockBus.OnPublish = func(event *DomainEvent) error {
    // Assert event properties
    return nil
}

// Test event publishing
eventBus.Publish(ctx, event)

// Test notification creation
notifications, _ := notifService.GetUserNotifications(ctx, userID, 10)
assert.Equal(t, 1, len(notifications))
```
