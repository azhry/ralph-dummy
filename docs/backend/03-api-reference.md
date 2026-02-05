# API Reference Documentation

## Overview

This document provides complete specifications for the Wedding Invitation REST API. All endpoints follow RESTful conventions and return JSON responses.

**Base URL:** `https://api.yourdomain.com/api/v1`

**Content-Type:** `application/json`

**Authentication:** Bearer Token (JWT) in `Authorization` header

---

## Response Format

### Success Response (200-201)

```json
{
  "success": true,
  "data": {
    // Response payload
  },
  "meta": {
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "uuid",
    "page": 1,
    "per_page": 20,
    "total": 100
  }
}
```

### Error Response (4xx-5xx)

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      {
        "field": "email",
        "message": "Email is required"
      }
    ]
  },
  "meta": {
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "uuid"
  }
}
```

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | OK - Success |
| 201 | Created - Resource created |
| 400 | Bad Request - Validation error |
| 401 | Unauthorized - Invalid or missing token |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource doesn't exist |
| 409 | Conflict - Duplicate or conflict |
| 422 | Unprocessable Entity - Business logic error |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error |

---

## Authentication Endpoints

### POST /auth/register

Register a new user account.

**Request:**

```json
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "first_name": "John",
  "last_name": "Doe"
}
```

**Validation Rules:**
- `email`: Required, valid email format, unique
- `password`: Required, min 8 chars, must contain uppercase, lowercase, number
- `first_name`: Required, min 2 chars, max 50 chars
- `last_name`: Required, min 2 chars, max 50 chars

**Response (201 Created):**

```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "email_verified": false,
    "created_at": "2024-01-15T10:30:00Z"
  },
  "meta": {
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "req-123"
  }
}
```

**Go Handler Example:**

```go
func (h *AuthHandler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Success: false,
            Error: ErrorDetail{
                Code:    "VALIDATION_ERROR",
                Message: err.Error(),
            },
        })
        return
    }
    
    user, err := h.authService.Register(c.Request.Context(), req)
    if err != nil {
        if errors.Is(err, ErrEmailExists) {
            c.JSON(http.StatusConflict, ErrorResponse{
                Success: false,
                Error: ErrorDetail{
                    Code:    "EMAIL_EXISTS",
                    Message: "Email already registered",
                },
            })
            return
        }
        
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Success: false,
            Error: ErrorDetail{
                Code:    "INTERNAL_ERROR",
                Message: "Failed to create user",
            },
        })
        return
    }
    
    c.JSON(http.StatusCreated, SuccessResponse{
        Success: true,
        Data:    user,
    })
}
```

**Error Responses:**

```json
// 400 Bad Request
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      {"field": "email", "message": "Invalid email format"},
      {"field": "password", "message": "Password must be at least 8 characters"}
    ]
  }
}

// 409 Conflict
{
  "success": false,
  "error": {
    "code": "EMAIL_EXISTS",
    "message": "Email already registered"
  }
}
```

---

### POST /auth/login

Authenticate user and receive JWT tokens.

**Request:**

```json
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "user": {
      "id": "507f1f77bcf86cd799439011",
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe"
    },
    "tokens": {
      "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
      "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
      "expires_in": 900
    }
  }
}
```

**Go Handler Example:**

```go
func (h *AuthHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Success: false,
            Error: ErrorDetail{
                Code:    "VALIDATION_ERROR",
                Message: err.Error(),
            },
        })
        return
    }
    
    result, err := h.authService.Login(c.Request.Context(), req)
    if err != nil {
        if errors.Is(err, ErrInvalidCredentials) {
            c.JSON(http.StatusUnauthorized, ErrorResponse{
                Success: false,
                Error: ErrorDetail{
                    Code:    "INVALID_CREDENTIALS",
                    Message: "Invalid email or password",
                },
            })
            return
        }
        
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Success: false,
            Error: ErrorDetail{
                Code:    "INTERNAL_ERROR",
                Message: "Login failed",
            },
        })
        return
    }
    
    // Set refresh token as httpOnly cookie
    c.SetCookie(
        "refresh_token",
        result.Tokens.RefreshToken,
        604800, // 7 days
        "/",
        "",
        true, // secure
        true, // httpOnly
    )
    
    c.JSON(http.StatusOK, SuccessResponse{
        Success: true,
        Data:    result,
    })
}
```

---

### POST /auth/refresh

Refresh access token using refresh token.

**Request:**
- Cookie: `refresh_token` (httpOnly)

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 900
  }
}
```

---

### POST /auth/logout

Logout user and invalidate tokens.

**Request:**
- Cookie: `refresh_token` (httpOnly)

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "message": "Logged out successfully"
  }
}
```

**Go Handler:**

```go
func (h *AuthHandler) Logout(c *gin.Context) {
    refreshToken, _ := c.Cookie("refresh_token")
    
    if refreshToken != "" {
        h.authService.Logout(c.Request.Context(), refreshToken)
    }
    
    // Clear cookie
    c.SetCookie("refresh_token", "", -1, "/", "", true, true)
    
    c.JSON(http.StatusOK, SuccessResponse{
        Success: true,
        Data: map[string]string{
            "message": "Logged out successfully",
        },
    })
}
```

---

### POST /auth/forgot-password

Request password reset email.

**Request:**

```json
{
  "email": "user@example.com"
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "message": "Password reset email sent"
  }
}
```

**Note:** Always returns 200 to prevent email enumeration attacks.

---

### GET /auth/me

Get current authenticated user profile.

**Authentication:** Required (Bearer Token)

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "email_verified": true,
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

**Go Handler:**

```go
func (h *AuthHandler) GetMe(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, ErrorResponse{
            Success: false,
            Error: ErrorDetail{
                Code:    "UNAUTHORIZED",
                Message: "Authentication required",
            },
        })
        return
    }
    
    user, err := h.userService.GetByID(c.Request.Context(), userID.(string))
    if err != nil {
        c.JSON(http.StatusNotFound, ErrorResponse{
            Success: false,
            Error: ErrorDetail{
                Code:    "USER_NOT_FOUND",
                Message: "User not found",
            },
        })
        return
    }
    
    c.JSON(http.StatusOK, SuccessResponse{
        Success: true,
        Data:    user,
    })
}
```

---

## Wedding Endpoints

### GET /weddings

List all weddings for the authenticated user.

**Authentication:** Required

**Query Parameters:**
- `page` (optional): Page number (default: 1)
- `per_page` (optional): Items per page (default: 20, max: 100)
- `status` (optional): Filter by status (`draft`, `published`, `archived`)

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "weddings": [
      {
        "id": "507f1f77bcf86cd799439012",
        "slug": "john-jane-wedding",
        "theme": "dark-romance",
        "groom_name": "John",
        "bride_name": "Jane",
        "wedding_date": "2024-06-15T00:00:00Z",
        "venue_name": "Garden Pavilion",
        "is_published": true,
        "created_at": "2024-01-10T08:00:00Z",
        "stats": {
          "total_rsvps": 45,
          "attending": 38,
          "declined": 7
        }
      }
    ]
  },
  "meta": {
    "page": 1,
    "per_page": 20,
    "total": 5,
    "total_pages": 1
  }
}
```

---

### POST /weddings

Create a new wedding invitation.

**Authentication:** Required

**Request:**

```json
{
  "theme": "dark-romance",
  "groom_name": "John",
  "bride_name": "Jane",
  "groom_role": "The Groom",
  "bride_role": "The Bride",
  "groom_bio": "A passionate architect...",
  "bride_bio": "An artist and creative soul...",
  "love_story": "We met under the cherry blossoms...",
  "wedding_date": "2024-06-15",
  "venue_name": "Garden Pavilion",
  "venue_address": "123 Botanical Gardens, City Center",
  "venue_map_url": "https://maps.google.com/...",
  "contact_email": "john.jane@example.com",
  "site_title": "John & Jane Wedding",
  "meta_description": "Join us for our special day",
  "rsvp_deadline": "2024-05-01",
  "allow_plus_one": true,
  "collect_dietary": true,
  "events": [
    {
      "title": "Ceremony",
      "description": "The main ceremony",
      "start_time": "2024-06-15T16:00:00Z",
      "location": "Garden Pavilion",
      "address": "123 Botanical Gardens"
    },
    {
      "title": "Reception",
      "description": "Dinner and dancing",
      "start_time": "2024-06-15T18:30:00Z",
      "end_time": "2024-06-15T23:00:00Z",
      "location": "Grand Ballroom",
      "address": "The Heritage Hotel, Downtown"
    }
  ],
  "custom_questions": [
    {
      "id": "song_request",
      "question": "What song would get you on the dance floor?",
      "type": "text",
      "required": false
    }
  ]
}
```

**Validation Rules:**
- `theme`: Required, must be valid theme ID
- `groom_name`: Required, 1-100 chars
- `bride_name`: Required, 1-100 chars
- `wedding_date`: Required, must be future date
- `contact_email`: Optional, valid email
- `events`: Optional array, max 10 events
- `custom_questions`: Optional array, max 5 questions

**Response (201 Created):**

```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439012",
    "slug": "john-jane-wedding",
    "theme": "dark-romance",
    "groom_name": "John",
    "bride_name": "Jane",
    "groom_role": "The Groom",
    "bride_role": "The Bride",
    "wedding_date": "2024-06-15T00:00:00Z",
    "venue_name": "Garden Pavilion",
    "contact_email": "john.jane@example.com",
    "is_published": false,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z",
    "public_url": "https://weddings.example.com/john-jane-wedding"
  }
}
```

---

### GET /weddings/:id

Get detailed information about a specific wedding.

**Authentication:** Required (must be owner)

**Parameters:**
- `id`: Wedding ID

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439012",
    "slug": "john-jane-wedding",
    "user_id": "507f1f77bcf86cd799439011",
    "theme": "dark-romance",
    "groom_name": "John",
    "bride_name": "Jane",
    "groom_role": "The Groom",
    "bride_role": "The Bride",
    "groom_bio": "A passionate architect...",
    "bride_bio": "An artist and creative soul...",
    "groom_photo_url": "https://cdn.example.com/photos/groom.jpg",
    "bride_photo_url": "https://cdn.example.com/photos/bride.jpg",
    "love_story": "We met under the cherry blossoms...",
    "wedding_date": "2024-06-15T00:00:00Z",
    "rsvp_deadline": "2024-05-01T00:00:00Z",
    "venue_name": "Garden Pavilion",
    "venue_address": "123 Botanical Gardens, City Center",
    "venue_map_url": "https://maps.google.com/...",
    "contact_email": "john.jane@example.com",
    "site_title": "John & Jane Wedding",
    "meta_description": "Join us for our special day",
    "is_published": true,
    "custom_domain": null,
    "password_protected": false,
    "allow_plus_one": true,
    "collect_dietary": true,
    "events": [
      {
        "title": "Ceremony",
        "description": "The main ceremony",
        "start_time": "2024-06-15T16:00:00Z",
        "location": "Garden Pavilion",
        "address": "123 Botanical Gardens"
      }
    ],
    "gallery_images": [
      "https://cdn.example.com/gallery/1.jpg",
      "https://cdn.example.com/gallery/2.jpg"
    ],
    "custom_questions": [],
    "stats": {
      "total_rsvps": 45,
      "attending": 38,
      "declined": 7,
      "pending": 12
    },
    "created_at": "2024-01-10T08:00:00Z",
    "updated_at": "2024-01-15T10:30:00Z",
    "published_at": "2024-01-12T09:00:00Z",
    "public_url": "https://weddings.example.com/john-jane-wedding"
  }
}
```

---

### PUT /weddings/:id

Update an existing wedding.

**Authentication:** Required (must be owner)

**Parameters:**
- `id`: Wedding ID

**Request:** Partial update (only changed fields)

```json
{
  "groom_bio": "Updated bio...",
  "venue_address": "456 New Address, City",
  "is_published": true
}
```

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439012",
    "slug": "john-jane-wedding",
    // ... updated wedding object
    "updated_at": "2024-01-15T14:00:00Z"
  }
}
```

---

### DELETE /weddings/:id

Delete a wedding and all associated data.

**Authentication:** Required (must be owner)

**Parameters:**
- `id`: Wedding ID

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "message": "Wedding deleted successfully"
  }
}
```

**Note:** This permanently deletes the wedding, RSVPs, guests, and associated files.

---

### POST /weddings/:id/publish

Publish a wedding to make it publicly accessible.

**Authentication:** Required (must be owner)

**Parameters:**
- `id`: Wedding ID

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439012",
    "is_published": true,
    "published_at": "2024-01-15T14:00:00Z",
    "public_url": "https://weddings.example.com/john-jane-wedding"
  }
}
```

---

### GET /weddings/:id/stats

Get statistics for a wedding.

**Authentication:** Required (must be owner)

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "rsvps": {
      "total": 45,
      "attending": 38,
      "declined": 7,
      "pending": 12,
      "total_guests": 76
    },
    "views": {
      "total": 342,
      "unique_visitors": 156,
      "last_7_days": 45,
      "last_30_days": 120
    },
    "events": {
      "ceremony": {
        "attending": 38,
        "capacity": 50
      },
      "reception": {
        "attending": 35,
        "capacity": 80
      }
    },
    "dietary_restrictions": {
      "vegetarian": 5,
      "vegan": 2,
      "gluten_free": 3,
      "nut_allergy": 1
    }
  }
}
```

---

## Public Routes

### GET /public/weddings/:slug

View a public wedding invitation (no authentication required).

**Parameters:**
- `slug`: Wedding URL slug

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "slug": "john-jane-wedding",
    "theme": "dark-romance",
    "groom_name": "John",
    "bride_name": "Jane",
    "groom_role": "The Groom",
    "bride_role": "The Bride",
    "groom_bio": "A passionate architect...",
    "bride_bio": "An artist and creative soul...",
    "groom_photo_url": "https://cdn.example.com/photos/groom.jpg",
    "bride_photo_url": "https://cdn.example.com/photos/bride.jpg",
    "love_story": "We met under the cherry blossoms...",
    "wedding_date": "2024-06-15T00:00:00Z",
    "venue_name": "Garden Pavilion",
    "venue_address": "123 Botanical Gardens, City Center",
    "venue_map_url": "https://maps.google.com/...",
    "contact_email": "john.jane@example.com",
    "site_title": "John & Jane Wedding",
    "meta_description": "Join us for our special day",
    "events": [
      {
        "title": "Ceremony",
        "description": "The main ceremony",
        "start_time": "2024-06-15T16:00:00Z",
        "location": "Garden Pavilion",
        "address": "123 Botanical Gardens"
      }
    ],
    "gallery_images": [
      "https://cdn.example.com/gallery/1.jpg",
      "https://cdn.example.com/gallery/2.jpg"
    ],
    "allow_plus_one": true,
    "collect_dietary": true,
    "custom_questions": [
      {
        "id": "song_request",
        "question": "What song would get you on the dance floor?",
        "type": "text",
        "required": false
      }
    ],
    "rsvp_deadline": "2024-05-01T00:00:00Z",
    "rsvp_status": "open"
  }
}
```

**Error Responses:**

```json
// 404 - Wedding not found or not published
{
  "success": false,
  "error": {
    "code": "WEDDING_NOT_FOUND",
    "message": "Wedding not found or not yet published"
  }
}

// 403 - Password protected
{
  "success": false,
  "error": {
    "code": "PASSWORD_REQUIRED",
    "message": "This wedding is password protected"
  }
}
```

---

### POST /public/weddings/:slug/rsvp

Submit an RSVP for a public wedding (no authentication required).

**Parameters:**
- `slug`: Wedding URL slug

**Request:**

```json
{
  "name": "Alice Smith",
  "email": "alice@example.com",
  "phone": "+1-555-0123",
  "attending": true,
  "number_of_guests": 2,
  "plus_one_name": "Bob Smith",
  "dietary_restrictions": "Vegetarian, no nuts",
  "message": "So excited for your big day!",
  "song_request": "September by Earth, Wind & Fire",
  "custom_answers": {
    "song_request": "September by Earth, Wind & Fire"
  }
}
```

**Validation Rules:**
- `name`: Required, 1-100 chars
- `email`: Optional but recommended, valid email
- `phone`: Optional
- `attending`: Required, boolean
- `number_of_guests`: Required if attending, min 1, max 10
- `plus_one_name`: Optional, required if number_of_guests > 1
- `dietary_restrictions`: Optional, max 500 chars
- `message`: Optional, max 1000 chars

**Response (201 Created):**

```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439020",
    "wedding_id": "507f1f77bcf86cd799439012",
    "name": "Alice Smith",
    "email": "alice@example.com",
    "attending": true,
    "number_of_guests": 2,
    "plus_one_name": "Bob Smith",
    "submitted_at": "2024-01-15T14:30:00Z",
    "confirmation_sent": false
  }
}
```

**Go Handler Example:**

```go
func (h *PublicHandler) SubmitRSVP(c *gin.Context) {
    slug := c.Param("slug")
    
    wedding, err := h.weddingService.GetBySlug(c.Request.Context(), slug)
    if err != nil {
        c.JSON(http.StatusNotFound, ErrorResponse{
            Success: false,
            Error: ErrorDetail{
                Code:    "WEDDING_NOT_FOUND",
                Message: "Wedding not found",
            },
        })
        return
    }
    
    // Check if wedding is published
    if !wedding.IsPublished {
        c.JSON(http.StatusNotFound, ErrorResponse{
            Success: false,
            Error: ErrorDetail{
                Code:    "WEDDING_NOT_PUBLISHED",
                Message: "Wedding not found",
            },
        })
        return
    }
    
    var req CreateRSVPRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Success: false,
            Error: ErrorDetail{
                Code:    "VALIDATION_ERROR",
                Message: err.Error(),
            },
        })
        return
    }
    
    // Capture metadata
    req.IPAddress = c.ClientIP()
    req.UserAgent = c.Request.UserAgent()
    
    rsvp, err := h.rsvpService.Create(c.Request.Context(), wedding.ID, req)
    if err != nil {
        if errors.Is(err, ErrDuplicateRSVP) {
            c.JSON(http.StatusConflict, ErrorResponse{
                Success: false,
                Error: ErrorDetail{
                    Code:    "RSVP_EXISTS",
                    Message: "You have already submitted an RSVP for this wedding",
                },
            })
            return
        }
        
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Success: false,
            Error: ErrorDetail{
                Code:    "INTERNAL_ERROR",
                Message: "Failed to submit RSVP",
            },
        })
        return
    }
    
    // Track analytics
    go h.analyticsService.TrackRSVP(wedding.ID, rsvp.ID)
    
    c.JSON(http.StatusCreated, SuccessResponse{
        Success: true,
        Data:    rsvp,
    })
}
```

---

## RSVP Management Endpoints

### GET /weddings/:id/rsvps

List all RSVPs for a wedding.

**Authentication:** Required (must be owner)

**Query Parameters:**
- `page`: Page number (default: 1)
- `per_page`: Items per page (default: 50)
- `status`: Filter by status (`attending`, `declined`, `all`)
- `search`: Search by name or email

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "rsvps": [
      {
        "id": "507f1f77bcf86cd799439020",
        "name": "Alice Smith",
        "email": "alice@example.com",
        "phone": "+1-555-0123",
        "attending": true,
        "number_of_guests": 2,
        "plus_one_name": "Bob Smith",
        "dietary_restrictions": "Vegetarian",
        "message": "So excited!",
        "song_request": "September",
        "custom_answers": {
          "song_request": "September"
        },
        "submitted_at": "2024-01-15T14:30:00Z",
        "confirmation_sent": true,
        "confirmation_sent_at": "2024-01-15T14:31:00Z"
      }
    ]
  },
  "meta": {
    "page": 1,
    "per_page": 50,
    "total": 45,
    "total_pages": 1
  }
}
```

---

### GET /weddings/:id/rsvps/export

Export RSVPs to CSV or Excel.

**Authentication:** Required (must be owner)

**Query Parameters:**
- `format`: `csv` or `excel` (default: csv)

**Response:** File download with appropriate Content-Type header

---

## Guest List Endpoints

### GET /weddings/:id/guests

List guest list for a wedding.

**Authentication:** Required (must be owner)

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "guests": [
      {
        "id": "507f1f77bcf86cd799439030",
        "name": "Alice Smith",
        "email": "alice@example.com",
        "phone": "+1-555-0123",
        "group": "family",
        "invited": true,
        "invitation_sent": true,
        "rsvp_status": "attending",
        "has_plus_one": true,
        "plus_one_name": "Bob Smith",
        "table_assignment": "Table 3",
        "notes": "Vegetarian meal"
      }
    ]
  }
}
```

---

### POST /weddings/:id/guests/bulk

Bulk import guests from CSV.

**Authentication:** Required (must be owner)

**Request:**
- Content-Type: `multipart/form-data`
- File field: `file` (CSV file)

CSV Format:
```csv
name,email,phone,group,has_plus_one
Alice Smith,alice@example.com,+1-555-0123,family,true
Bob Johnson,bob@example.com,+1-555-0124,friends,false
```

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "imported": 50,
    "skipped": 2,
    "errors": [
      {
        "row": 3,
        "error": "Invalid email format"
      }
    ]
  }
}
```

---

## File Upload Endpoints

### POST /weddings/:id/upload/couple

Upload couple photos.

**Authentication:** Required (must be owner)

**Request:**
- Content-Type: `multipart/form-data`
- Fields:
  - `groom_photo`: File (optional)
  - `bride_photo`: File (optional)

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "groom_photo_url": "https://cdn.example.com/couple/groom-abc123.jpg",
    "bride_photo_url": "https://cdn.example.com/couple/bride-xyz789.jpg"
  }
}
```

---

### POST /weddings/:id/upload/gallery

Upload gallery images.

**Authentication:** Required (must be owner)

**Request:**
- Content-Type: `multipart/form-data`
- Field: `images[]` (multiple files, max 10)

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "uploaded": 5,
    "urls": [
      "https://cdn.example.com/gallery/img1-abc123.jpg",
      "https://cdn.example.com/gallery/img2-def456.jpg"
    ]
  }
}
```

---

## Analytics Endpoints

### GET /weddings/:id/analytics/views

Get page view statistics.

**Authentication:** Required (must be owner)

**Query Parameters:**
- `start_date`: ISO date (optional)
- `end_date`: ISO date (optional)

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "total_views": 342,
    "unique_visitors": 156,
    "by_date": [
      {
        "date": "2024-01-15",
        "views": 25,
        "unique": 12
      }
    ],
    "by_referrer": [
      {
        "source": "instagram",
        "count": 89
      },
      {
        "source": "facebook",
        "count": 45
      }
    ],
    "by_country": [
      {
        "country": "US",
        "count": 156
      }
    ]
  }
}
```

---

### GET /weddings/:id/analytics/rsvps

Get RSVP statistics.

**Authentication:** Required (must be owner)

**Response (200 OK):**

```json
{
  "success": true,
  "data": {
    "total": 45,
    "attending": 38,
    "declined": 7,
    "by_date": [
      {
        "date": "2024-01-15",
        "rsvps": 5,
        "attending": 4
      }
    ],
    "response_time_avg": 2.5
  }
}
```

---

## Rate Limiting

All endpoints have rate limiting to prevent abuse:

| Endpoint Category | Limit | Window |
|------------------|-------|--------|
| Authentication | 5 | 1 minute |
| Public API | 100 | 1 minute |
| Authenticated API | 1000 | 1 minute |
| File Uploads | 10 | 1 minute |

When rate limit is exceeded:

```json
{
  "success": false,
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests. Please try again later.",
    "retry_after": 60
  }
}
```

---

## Pagination

List endpoints support pagination with these query parameters:

- `page`: Page number (1-based, default: 1)
- `per_page`: Items per page (default: 20, max: 100)

Pagination info is returned in the `meta` field:

```json
{
  "meta": {
    "page": 2,
    "per_page": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

---

## Error Codes Reference

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Input validation failed |
| `INVALID_CREDENTIALS` | 401 | Wrong email or password |
| `UNAUTHORIZED` | 401 | Missing or invalid token |
| `FORBIDDEN` | 403 | Insufficient permissions |
| `USER_NOT_FOUND` | 404 | User doesn't exist |
| `WEDDING_NOT_FOUND` | 404 | Wedding doesn't exist |
| `WEDDING_NOT_PUBLISHED` | 404 | Wedding exists but not published |
| `EMAIL_EXISTS` | 409 | Email already registered |
| `SLUG_EXISTS` | 409 | Slug already taken |
| `RSVP_EXISTS` | 409 | RSVP already submitted |
| `UNPROCESSABLE_ENTITY` | 422 | Business logic error |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |
| `INTERNAL_ERROR` | 500 | Server error |

---

## Go Handler Examples

### Generic Handler Structure

```go
package handler

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
)

// Response structures
type SuccessResponse struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type ErrorResponse struct {
    Success bool        `json:"success"`
    Error   ErrorDetail `json:"error"`
    Meta    *Meta       `json:"meta,omitempty"`
}

type ErrorDetail struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

type Meta struct {
    Timestamp  string `json:"timestamp"`
    RequestID  string `json:"request_id"`
    Page       int    `json:"page,omitempty"`
    PerPage    int    `json:"per_page,omitempty"`
    Total      int    `json:"total,omitempty"`
    TotalPages int    `json:"total_pages,omitempty"`
}

// BaseHandler provides common functionality
type BaseHandler struct {
    logger *zap.Logger
}

func (h *BaseHandler) respondWithError(c *gin.Context, status int, code, message string) {
    c.JSON(status, ErrorResponse{
        Success: false,
        Error: ErrorDetail{
            Code:    code,
            Message: message,
        },
        Meta: &Meta{
            Timestamp: time.Now().UTC().Format(time.RFC3339),
            RequestID: c.GetString("request_id"),
        },
    })
}

func (h *BaseHandler) respondWithSuccess(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, SuccessResponse{
        Success: true,
        Data:    data,
        Meta: &Meta{
            Timestamp: time.Now().UTC().Format(time.RFC3339),
            RequestID: c.GetString("request_id"),
        },
    })
}
```

---

**Version:** 1.0  
**Last Updated:** 2024-01-15  
**API Version:** v1
