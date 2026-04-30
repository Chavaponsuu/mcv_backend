package handlers

import (
	"net/http"
	"strconv"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"mcv_backend/models"
	"mcv_backend/services"
	// "encoding/json"
)

// GetNotificationsHandler retrieves user's unread notifications
func GetNotificationsHandler(notifService *services.NotificationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID := r.Context().Value("user_id").(string)
		if userID =="" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "unauthorized",
			})
			return
		}

		limit := int64(20)
		if l := r.URL.Query().Get("limit"); l != "" {
			if parsed, err := strconv.ParseInt(l, 10, 64); err == nil && parsed > 0 && parsed <= 100 {
				limit = parsed
			}
		}
		userObjID, err := primitive.ObjectIDFromHex(userID)
if err != nil {
	http.Error(w, "invalid user_id format", http.StatusBadRequest)
	return
}
		notifications, err := notifService.GetUserNotifications(r.Context(), userObjID, limit)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to fetch notifications",
			})
			return
		}

		if notifications == nil {
			notifications = []models.Notification{}
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"notifications": notifications,
			"count": len(notifications),
		})
	}
}


// MarkNotificationAsReadHandler marks a single notification as read
func MarkNotificationAsReadHandler(notifService *services.NotificationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id := r.URL.Path[len("/notifications/"):]

		if err := notifService.MarkAsRead(r.Context(), id); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid notification ID",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "notification marked as read",
		})
	}
}
// MarkAllNotificationsReadHandler marks all notifications as read for a user
func MarkAllNotificationsReadHandler(notifService *services.NotificationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID := r.Context().Value("user_id")
		if userID == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "unauthorized",
			})
			return
		}

		if err := notifService.MarkAllAsRead(r.Context(), userID.(string)); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to mark notifications as read",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "all notifications marked as read",
		})
	}
}

// DeleteNotificationHandler deletes a notification
func DeleteNotificationHandler(notifService *services.NotificationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		id := r.URL.Path[len("/notifications/"):]

		if err := notifService.DeleteNotification(r.Context(), id); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "invalid notification ID",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"message": "notification deleted",
		})
	}
}

// GetNotificationPreferencesHandler retrieves user's notification preferences
// func GetNotificationPreferencesHandler(notifService *services.NotificationService) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {

// 		userID := r.Context().Value("user_id")
// 		if userID == nil {
// 			writeJSON(w, http.StatusUnauthorized, map[string]string{
// 				"error": "unauthorized",
// 			})
// 			return
// 		}

// 		// prefs, err := notifService.GetUserPreferences(r.Context(), userID.(string))
// 		if err != nil {
// 			writeJSON(w, http.StatusInternalServerError, map[string]string{
// 				"error": "failed to fetch preferences",
// 			})
// 			return
// 		}

// 		// if prefs == nil {
// 		// 	prefs = &models.NotificationPreference{
// 		// 		EnableAssignmentNotifications: true,
// 		// 	}
// 		// }

// 		writeJSON(w, http.StatusOK, prefs)
// 	}
// }
// UpdateNotificationPreferencesHandler updates user's notification preferences
// func UpdateNotificationPreferencesHandler(notifService *services.NotificationService) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {

// 		userID := r.Context().Value("user_id")
// 		if userID == nil {
// 			writeJSON(w, http.StatusUnauthorized, map[string]string{
// 				"error": "unauthorized",
// 			})
// 			return
// 		}

// 		var prefs models.NotificationPreference

// 		if err := json.NewDecoder(r.Body).Decode(&prefs); err != nil {
// 			writeJSON(w, http.StatusBadRequest, map[string]string{
// 				"error": err.Error(),
// 			})
// 			return
// 		}

// 		if err := notifService.SetUserPreferences(r.Context(), userID.(string), &prefs); err != nil {
// 			writeJSON(w, http.StatusInternalServerError, map[string]string{
// 				"error": "failed to update preferences",
// 			})
// 			return
// 		}

// 		writeJSON(w, http.StatusOK, map[string]interface{}{
// 			"message":     "preferences updated successfully",
// 			"preferences": prefs,
// 		})
// 	}
// }
