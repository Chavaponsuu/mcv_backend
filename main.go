package main

import (
	"log"
	"net/http"
	"os"

	"mcv_backend/config"
	"mcv_backend/handlers"
	"mcv_backend/middleware"
	"mcv_backend/services"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Connect to MongoDB
	config.ConnectDB()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := mux.NewRouter()
	courseService := services.NewCourseService(config.DB.Collection("courses"))
	courseHandler := &handlers.CourseHandler{
		Service: courseService,
	}

	// Public auth endpoints
	router.HandleFunc("/api/auth/register", handlers.RegisterHandler).Methods("POST")
	router.HandleFunc("/api/auth/login", handlers.LoginHandler).Methods("POST")
	router.HandleFunc("/api/student/course", courseHandler.GetAllCourses).Methods("GET")
	
	// Protected endpoints
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	protected.HandleFunc("/student/me", handlers.GetMeHandler).Methods("GET")
	protected.HandleFunc("/student/me/courses", handlers.GetUserCourses).Methods("GET")
	
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
