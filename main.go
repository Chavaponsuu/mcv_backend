package main

import (
	"log"
	"net/http"
	"os"

	"mcv_backend/config"
	"mcv_backend/handlers"
	"mcv_backend/middleware"
	"mcv_backend/services"
	// "mcv_backend/models"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	config.ConnectDB()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := mux.NewRouter()

	// ===== Services =====
	courseService := services.NewCourseService(
		config.DB.Collection("courses"),
	)

	courseHandler := &handlers.CourseHandler{
		Service: courseService,
	}

	// ===== EventBus (อยู่ใน services package) =====
	eventBus := services.NewEventBus(config.DB)

assignmentService := &services.AssignmentService{
	Collection: config.DB.Collection("assignments"),
	EventBus:   eventBus,
}

assignmentHandler := &handlers.AssignmentHandler{
	AssignmentService: assignmentService,
	CourseService:     courseService,
}

notificationService := &services.NotificationService{
	Collection: config.DB.Collection("notifications"),
}
eventBus.Subscribe(notificationService)
	// ===== Routes =====

	router.HandleFunc(
		"/api/assignments",
		assignmentHandler.CreateAssignment,
	).Methods("POST")

	router.HandleFunc("/api/auth/register", handlers.RegisterHandler).Methods("POST")
	router.HandleFunc("/api/auth/login", handlers.LoginHandler).Methods("POST")

	router.HandleFunc("/api/student/course", courseHandler.GetAllCourses).Methods("GET")
	router.HandleFunc("/api/student/semester", handlers.GetAllSemesters).Methods("GET")

	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	
	protected.HandleFunc("/student/me", handlers.GetMeHandler).Methods("GET")
	protected.HandleFunc("/student/me/courses", courseHandler.GetUserCourses).Methods("GET")
	protected.HandleFunc("/student/me/semester", handlers.GetSemesterIDByYearAndTerm).Methods("GET")
	protected.HandleFunc("/student/me/signature", handlers.GenerateSignatureHandler).Methods("POST")
		protected.HandleFunc("/student/me/notifications" , handlers.GetNotificationsHandler(notificationService)).Methods("GET")
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}