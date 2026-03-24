package services

import (
	"context"
	"errors"
	"os"
	"time"

	"mcv_backend/config"
	"mcv_backend/models"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
	"strconv"
)

var (
	ErrUserExists       = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound     = errors.New("user not found")
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// HashPassword generates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a password with its hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateTokens creates access and refresh JWT tokens
func GenerateTokens(userID primitive.ObjectID, email string) (string, string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", "", errors.New("JWT_SECRET not configured")
	}

	// Access token (short-lived)
	expirySeconds := 3600 // Default 1 hour
	if expiry := os.Getenv("JWT_EXPIRY"); expiry != "" {
		// Parse expiry from env if needed
		if parsed, err := strconv.Atoi(expiry); err == nil {
			expirySeconds = parsed
	}
	}

	accessClaims := Claims{
		UserID: userID.Hex(),
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expirySeconds) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", err
	}

	// Refresh token (long-lived)
	refreshClaims := Claims{
		UserID: userID.Hex(),
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (*Claims, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET not configured")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// Register creates a new user account
func Register(ctx context.Context, req models.RegisterRequest) (*models.User, error) {
	collection := config.DB.Collection("users")
	studentCollection := config.DB.Collection("students")

	// Check if user already exists
	var existingUser models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		return nil, ErrUserExists
	}
	if err != mongo.ErrNoDocuments {
		return nil, err
	}

	// Hash password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := models.User{
		ID:        primitive.NewObjectID(),
		Email:     req.Email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	student := models.Student{
    UserID: user.ID,
    StudentID: req.StudentID,
    Name: req.Name,
    Email: req.Email,
    Faculty: req.Faculty,
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
}
studentCollection.InsertOne(ctx, student)

	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// Login authenticates a user and returns tokens
func Login(ctx context.Context, req models.LoginRequest) (*models.AuthResponse, error) {
	collection := config.DB.Collection("users")

	// Find user by email
	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check password
	if !CheckPasswordHash(req.Password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, refreshToken, err := GenerateTokens(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	expirySeconds := 3600
	if expiry := os.Getenv("JWT_EXPIRY"); expiry != "" {
		// Parse expiry from env if needed
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expirySeconds,
	}, nil
}

// GetUserByID retrieves a user by their ID
func GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	collection := config.DB.Collection("users")

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}
