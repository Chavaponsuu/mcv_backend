# Authentication API Documentation

## Overview
The auth service provides JWT-based authentication with user registration, login, and protected endpoints.

## Endpoints

### 1. Register User
**POST** `/api/auth/register`

Creates a new user account.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response (201 Created):**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "email": "user@example.com"
}
```

**Error Responses:**
- `400 Bad Request`: Invalid request body or validation error
- `409 Conflict`: User already exists

### 2. Login
**POST** `/api/auth/login`

Authenticates a user and returns JWT tokens.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response (200 OK):**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_in": 3600
}
```

**Error Responses:**
- `400 Bad Request`: Invalid request body
- `401 Unauthorized`: Invalid credentials

### 3. Get Current User
**GET** `/api/auth/me`

Returns the authenticated user's information.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response (200 OK):**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "email": "user@example.com"
}
```

**Error Responses:**
- `401 Unauthorized`: Missing or invalid token
- `404 Not Found`: User not found

## Protected Endpoints

All endpoints under `/api/*` (except `/api/auth/register` and `/api/auth/login`) require authentication.

Include the JWT token in the Authorization header:
```
Authorization: Bearer <access_token>
```

## Security Features

- Passwords are hashed using bcrypt
- JWT tokens with configurable expiry
- Access tokens (short-lived, default 1 hour)
- Refresh tokens (long-lived, 7 days)
- Middleware-based route protection

## Environment Variables

```env
JWT_SECRET=your-secret-key
JWT_EXPIRY=3600
```
