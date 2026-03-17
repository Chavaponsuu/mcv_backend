# MyCourseVille Backend

Go-based backend service for MyCourseVille mobile application.

## Setup

```bash
# Install dependencies
go mod download

# Run the server
go run main.go
```

## Environment Variables

- `PORT`: Server port (default: 8080)

## API Endpoints

### GET /api/courses
Fetch user's enrolled courses.

**Headers:**
- `Authorization: Bearer <access_token>`

**Response (200 OK):**
```json
[
  {
    "course_id": "12345",
    "title": "Introduction to Computer Science",
    "description": "Learn the fundamentals of CS",
    "instructor": "Dr. Smith"
  }
]
```

**Error Responses:**
- `401 Unauthorized`: Invalid or expired token
- `503 Service Unavailable`: MyCourseVille API unavailable
