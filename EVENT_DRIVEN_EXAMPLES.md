# Event-Driven Notification Integration Examples

## Quick Start Example

### 1. Update main.go

```go
package main

import (
	"mcv_backend/config"
	"mcv_backend/handlers"
	"mcv_backend/middleware"
	"mcv_backend/models"
	"mcv_backend/services"
	"github.com/gorilla/mux"
)

func main() {
	// ... existing database connection ...
	config.ConnectDB()
	defer config.DisconnectDB()

	// ===== NEW: Initialize Event-Driven System =====
	eventBus := services.NewEventBus(config.DB)

	notifColl := config.DB.Collection("notifications")
	prefColl := config.DB.Collection("notification_preferences")
	notifService := services.NewNotificationService(notifColl, prefColl)

	// Subscribe notification service to events
	eventBus.Subscribe(notifService)

	// You can add more handlers here:
	// emailService := services.NewEmailService(...)
	// eventBus.Subscribe(emailService)

	// ===== Setup Router =====
	router := mux.NewRouter()

	// Public routes
	public := router.PathPrefix("/api").Subrouter()
	{
		public.HandleFunc("/auth/register", handlers.RegisterHandler).Methods("POST")
		public.HandleFunc("/auth/login", handlers.LoginHandler).Methods("POST")
	}

	// Protected routes
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	{
		protected.HandleFunc("/student/me", handlers.StudentInfoHandler).Methods("GET")
		protected.HandleFunc("/student/course", handlers.GetCoursesHandler).Methods("GET")
		protected.HandleFunc("/student/me/courses", handlers.GetStudentCoursesHandler).Methods("GET")

		// ===== NEW: Notification routes =====
		protected.HandleFunc("/notifications", handlers.GetNotificationsHandler(notifService)).Methods("GET")
		protected.HandleFunc("/notifications/{id}/read", handlers.MarkNotificationAsReadHandler(notifService)).Methods("POST")
		protected.HandleFunc("/notifications/read-all", handlers.MarkAllNotificationsReadHandler(notifService)).Methods("POST")
		protected.HandleFunc("/notifications/{id}", handlers.DeleteNotificationHandler(notifService)).Methods("DELETE")
		protected.HandleFunc("/notifications/preferences", handlers.GetNotificationPreferencesHandler(notifService)).Methods("GET")
		protected.HandleFunc("/notifications/preferences", handlers.UpdateNotificationPreferencesHandler(notifService)).Methods("POST")
	}

	// Store eventBus in a global or context for access in handlers
	// Option 1: Store in handler package
	handlers.SetEventBus(eventBus)

	// Option 2: Wrap handlers with eventBus
	// (see examples below)

	http.ListenAndServe(":8080", router)
}
```

### 2. Update Auth Service to Publish Events

```go
// services/auth.go

package services

import (
	"mcv_backend/models"
	"context"
)

var globalEventBus *EventBus // Add this for global access

// SetEventBus sets the global event bus
func SetEventBus(bus *EventBus) {
	globalEventBus = bus
}

func (as *AuthService) Register(ctx context.Context, req *RegisterRequest) (*models.User, error) {
	// ... existing registration logic ...

	// Create user and student...
	user, err := as.createUserAndStudent(ctx, req)
	if err != nil {
		return nil, err
	}

	// ===== NEW: Publish registration event =====
	event := &models.DomainEvent{
		Type:          models.EventUserRegistered,
		AggregateID:   user.ID,
		AggregateType: "user",
		ActorID:       user.ID,
		Data: map[string]interface{}{
			"email": user.Email,
			"name":  req.Name,
		},
	}

	// Use async to avoid blocking registration
	globalEventBus.PublishAsync(ctx, event, func(err error) {
		if err != nil {
			log.Printf("Error publishing registration event: %v", err)
		}
	})

	return user, nil
}

func (as *AuthService) Login(ctx context.Context, req *LoginRequest, userID string) (*TokenResponse, error) {
	// ... existing login logic ...

	// ===== NEW: Publish login event =====
	id, _ := primitive.ObjectIDFromHex(userID)
	event := &models.DomainEvent{
		Type:          models.EventUserLoggedIn,
		AggregateID:   id,
		AggregateType: "user",
		ActorID:       id,
		Data: map[string]interface{}{
			"email": req.Email,
			"timestamp": time.Now(),
		},
	}

	globalEventBus.PublishAsync(ctx, event, nil)

	return tokens, nil
}
```

### 3. Publish Events from Enrollment Handler

```go
// handlers/enrollment.go (NEW FILE)

package handlers

type EnrollCourseRequest struct {
	CourseOfferingID string `json:"course_offering_id" binding:"required"`
	SemesterID       string `json:"semester_id" binding:"required"`
}

func EnrollCourseHandler(enrollService *services.EnrollmentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		var req EnrollCourseRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Create enrollment
		enrollment, err := enrollService.CreateEnrollment(c, userID.(string), req.CourseOfferingID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "enrollment failed"})
			return
		}

		// ===== NEW: Publish enrollment event =====
		userObjID, _ := primitive.ObjectIDFromHex(userID.(string))
		semesterObjID, _ := primitive.ObjectIDFromHex(req.SemesterID)
		courseObjID, _ := primitive.ObjectIDFromHex(req.CourseOfferingID)

		event := &models.DomainEvent{
			Type:          models.EventEnrollmentCreated,
			AggregateID:   userObjID,
			AggregateType: "enrollment",
			ActorID:       userObjID,
			Data: map[string]interface{}{
				"enrollment_id":     enrollment.ID.Hex(),
				"course_offering_id": req.CourseOfferingID,
				"semester_id":       req.SemesterID,
				"enrolled_at":       time.Now(),
			},
		}

		// Publish and let notification service handle it
		eventBus.Publish(c, event)

		c.JSON(http.StatusCreated, enrollment)
	}
}

func DropCourseHandler(enrollService *services.EnrollmentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		enrollmentID := c.Param("enrollment_id")
		userID, _ := c.Get("user_id")

		err := enrollService.DropEnrollment(c, enrollmentID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "drop failed"})
			return
		}

		// ===== NEW: Publish enrollment dropped event =====
		userObjID, _ := primitive.ObjectIDFromHex(userID.(string))
		enrollObjID, _ := primitive.ObjectIDFromHex(enrollmentID)

		event := &models.DomainEvent{
			Type:          models.EventEnrollmentDropped,
			AggregateID:   userObjID,
			AggregateType: "enrollment",
			ActorID:       userObjID,
			Data: map[string]interface{}{
				"enrollment_id": enrollmentID,
				"dropped_at":    time.Now(),
			},
		}

		eventBus.PublishAsync(c, event, nil)

		c.JSON(http.StatusOK, gin.H{"message": "course dropped"})
	}
}
```

### 4. Publish Grade Update Events

```go
// handlers/grades.go (NEW FILE)

package handlers

type UpdateGradeRequest struct {
	EnrollmentID string `json:"enrollment_id" binding:"required"`
	Grade        string `json:"grade" binding:"required,max=3"`
}

func UpdateGradeHandler(gradeService *services.GradeService, enrollService *services.EnrollmentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpdateGradeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Get enrollment to find student
		enrollment, _ := enrollService.GetEnrollment(c, req.EnrollmentID)

		// Update grade
		err := gradeService.UpdateGrade(c, req.EnrollmentID, req.Grade)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "grade update failed"})
			return
		}

		// ===== NEW: Publish grade update event =====
		event := &models.DomainEvent{
			Type:          models.EventGradeUpdated,
			AggregateID:   enrollment.UserID,
			AggregateType: "enrollment",
			ActorID:       enrollment.UserID, // Could be instructor
			Data: map[string]interface{}{
				"enrollment_id": req.EnrollmentID,
				"grade":        req.Grade,
				"updated_at":   time.Now(),
			},
		}

		eventBus.PublishAsync(c, event, nil)

		c.JSON(http.StatusOK, gin.H{"message": "grade updated"})
	}
}
```

### 5. Custom Event Handler Example (Email Notifications)

```go
// services/email_notification_handler.go (EXAMPLE)

package services

import (
	"fmt"
	"mcv_backend/models"
)

type EmailNotificationHandler struct {
	// email service dependencies
}

func (enh *EmailNotificationHandler) Handle(event *models.DomainEvent) error {
	switch event.Type {
	case models.EventEnrollmentCreated:
		return enh.sendEnrollmentConfirmation(event)
	case models.EventGradeUpdated:
		return enh.sendGradeNotification(event)
	}
	return nil
}

func (enh *EmailNotificationHandler) SupportedEventTypes() []models.EventType {
	return []models.EventType{
		models.EventEnrollmentCreated,
		models.EventGradeUpdated,
	}
}

func (enh *EmailNotificationHandler) sendEnrollmentConfirmation(event *models.DomainEvent) error {
	// Get user email from database
	// Send confirmation email
	fmt.Printf("Sending enrollment confirmation email for user %s\n", event.AggregateID)
	return nil
}

func (enh *EmailNotificationHandler) sendGradeNotification(event *models.DomainEvent) error {
	grade := event.Data["grade"].(string)
	fmt.Printf("Sending grade notification: %s\n", grade)
	return nil
}
```

## Database Indexes

Create these indexes for optimal performance:

```javascript
// MongoDB shell commands

// Notifications collection
db.notifications.createIndex({ user_id: 1, status: 1, created_at: -1 });
db.notifications.createIndex({ created_at: -1 });

// Notification preferences collection
db.notification_preferences.createIndex({ user_id: 1 }, { unique: true });

// Domain events collection (audit trail)
db.domain_events.createIndex({ type: 1, created_at: -1 });
db.domain_events.createIndex({ aggregate_id: 1 });
db.domain_events.createIndex({ created_at: -1 });
```

## API Usage Examples

### Get Notifications

```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/notifications?limit=10
```

Response:

```json
{
  "notifications": [
    {
      "_id": "...",
      "user_id": "...",
      "type": "enrollment.confirmed",
      "title": "Enrollment Confirmed",
      "message": "You have successfully enrolled in a course",
      "status": "unread",
      "created_at": "2024-04-13T10:30:00Z"
    }
  ],
  "count": 1
}
```

### Mark as Read

```bash
curl -X POST -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/notifications/{notification_id}/read
```

### Get/Update Preferences

```bash
# Get preferences
curl -H "Authorization: Bearer <token>" \
  http://localhost:8080/api/notifications/preferences

# Update preferences
curl -X POST -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "enable_enrollment": true,
    "enable_grades": true,
    "enable_course_updates": false,
    "enable_announcements": false
  }' \
  http://localhost:8080/api/notifications/preferences
```

## Error Handling

The system is resilient to handler failures:

```go
// If one handler fails, others still execute
eventBus.Publish(ctx, event)

// Errors are logged but don't block notification
// Check logs for failures:
// "error handling event enrollment.created: connection refused"
```
