# Wedding Invitation Backend - User Guide

This guide provides comprehensive instructions for using the Wedding Invitation Backend API. It's intended for frontend developers, integrators, and API consumers.

## 🚀 Getting Started

### Base URL
- **Development**: `http://localhost:8080/api/v1`
- **Production**: `https://api.yourdomain.com/api/v1`

### Authentication
All protected endpoints require a valid JWT access token:
```bash
Authorization: Bearer <access_token>
Content-Type: application/json
```

## 📚 API Usage Examples

### 1. User Registration & Login

#### Register a New User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "SecurePass123!",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "email": "john.doe@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "email_verified": false,
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

#### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "SecurePass123!"
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "token_type": "Bearer",
    "expires_in": 900,
    "user": {
      "id": "507f1f77bcf86cd799439011",
      "email": "john.doe@example.com",
      "first_name": "John",
      "last_name": "Doe"
    }
  }
}
```

#### Refresh Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

### 2. Wedding Management

#### Create a Wedding
```bash
curl -X POST http://localhost:8080/api/v1/weddings \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "John & Jane'\''s Wedding",
    "groom_name": "John Doe",
    "bride_name": "Jane Smith",
    "wedding_date": "2024-06-15T14:00:00Z",
    "venue": "Beautiful Garden Venue",
    "description": "Join us for our special day celebration",
    "slug": "john-jane-wedding-2024",
    "is_public": true,
    "is_published": true,
    "settings": {
      "allow_rsvp": true,
      "rsvp_deadline": "2024-06-01T23:59:59Z",
      "max_plus_ones": 2
    }
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439012",
    "title": "John & Jane's Wedding",
    "slug": "john-jane-wedding-2024",
    "groom_name": "John Doe",
    "bride_name": "Jane Smith",
    "wedding_date": "2024-06-15T14:00:00Z",
    "venue": "Beautiful Garden Venue",
    "description": "Join us for our special day celebration",
    "is_public": true,
    "is_published": true,
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

#### Get User's Weddings
```bash
curl -X GET http://localhost:8080/api/v1/weddings \
  -H "Authorization: Bearer <access_token>"
```

#### Get Wedding by ID
```bash
curl -X GET http://localhost:8080/api/v1/weddings/507f1f77bcf86cd799439012 \
  -H "Authorization: Bearer <access_token>"
```

### 3. Guest Management

#### Add a Single Guest
```bash
curl -X POST http://localhost:8080/api/v1/weddings/507f1f77bcf86cd799439012/guests \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Robert",
    "last_name": "Johnson",
    "email": "robert.johnson@example.com",
    "phone": "+1-555-0123",
    "address": {
      "street": "123 Main St",
      "city": "Anytown",
      "state": "CA",
      "zip_code": "12345",
      "country": "USA"
    },
    "relationship": "friend",
    "side": "groom",
    "allow_plus_one": true,
    "max_plus_ones": 1,
    "dietary_notes": "No specific restrictions",
    "vip": false,
    "notes": "College friend"
  }'
```

#### Bulk Add Guests
```bash
curl -X POST http://localhost:8080/api/v1/weddings/507f1f77bcf86cd799439012/guests/bulk \
  -H "Authorization: Bearer <access_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "guests": [
      {
        "first_name": "Alice",
        "last_name": "Brown",
        "email": "alice@example.com",
        "relationship": "family",
        "side": "bride"
      },
      {
        "first_name": "Charlie",
        "last_name": "Davis",
        "email": "charlie@example.com",
        "relationship": "friend",
        "side": "groom"
      }
    ]
  }'
```

#### Import Guests from CSV
```bash
curl -X POST http://localhost:8080/api/v1/weddings/507f1f77bcf86cd799439012/guests/import \
  -H "Authorization: Bearer <access_token>" \
  -F "file=@guests.csv"
```

**CSV Format:**
```csv
first_name,last_name,email,phone,relationship,side
Alice,Brown,alice@example.com,555-0123,family,bride
Charlie,Davis,charlie@example.com,555-0456,friend,groom
```

#### List Guests
```bash
curl -X GET "http://localhost:8080/api/v1/weddings/507f1f77bcf86cd799439012/guests?page=1&limit=20&side=groom" \
  -H "Authorization: Bearer <access_token>"
```

### 4. Public API (No Authentication Required)

#### Get Public Wedding by Slug
```bash
curl -X GET http://localhost:8080/api/v1/public/weddings/john-jane-wedding-2024
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439012",
    "title": "John & Jane's Wedding",
    "slug": "john-jane-wedding-2024",
    "groom_name": "John Doe",
    "bride_name": "Jane Smith",
    "wedding_date": "2024-06-15T14:00:00Z",
    "venue": "Beautiful Garden Venue",
    "description": "Join us for our special day celebration",
    "rsvp_settings": {
      "enabled": true,
      "deadline": "2024-06-01T23:59:59Z"
    }
  }
}
```

#### Submit Public RSVP
```bash
curl -X POST http://localhost:8080/api/v1/public/weddings/john-jane-wedding-2024/rsvp \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Sarah",
    "last_name": "Wilson",
    "email": "sarah.wilson@example.com",
    "phone": "+1-555-0789",
    "status": "attending",
    "attendance_count": 2,
    "plus_ones": [
      {
        "first_name": "Mike",
        "last_name": "Wilson",
        "relationship": "spouse"
      }
    ],
    "dietary_restrictions": "Vegetarian for Sarah, no restrictions for Mike",
    "additional_notes": "Looking forward to celebrating with you!",
    "custom_answers": [
      {
        "question_id": "song_request",
        "question": "What song would get you on the dance floor?",
        "answer": "September by Earth, Wind & Fire"
      }
    ]
  }'
```

### 5. File Upload

#### Upload Wedding Photo
```bash
curl -X POST http://localhost:8080/api/v1/upload \
  -H "Authorization: Bearer <access_token>" \
  -F "file=@wedding-photo.jpg" \
  -F "wedding_id=507f1f77bcf86cd799439012" \
  -F "type=wedding_photo" \
  -F "caption=Our engagement photo"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439013",
    "filename": "wedding-photo.jpg",
    "original_name": "our-engagement.jpg",
    "mime_type": "image/jpeg",
    "size": 2048576,
    "url": "http://localhost:8080/uploads/wedding-photo.jpg",
    "thumbnails": {
      "small": "http://localhost:8080/uploads/thumbnails/small_wedding-photo.jpg",
      "medium": "http://localhost:8080/uploads/thumbnails/medium_wedding-photo.jpg",
      "large": "http://localhost:8080/uploads/thumbnails/large_wedding-photo.jpg"
    }
  }
}
```

### 6. Analytics

#### Track Page View (Client-side)
```bash
curl -X POST http://localhost:8080/api/v1/analytics/track/page-view \
  -H "Content-Type: application/json" \
  -d '{
    "wedding_id": "507f1f77bcf86cd799439012",
    "page": "invitation",
    "referrer": "https://facebook.com",
    "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
    "ip_address": "192.168.1.100",
    "device_id": "device_unique_id"
  }'
```

#### Get Wedding Analytics
```bash
curl -X GET http://localhost:8080/api/v1/weddings/507f1f77bcf86cd799439012/analytics \
  -H "Authorization: Bearer <access_token>"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "wedding_id": "507f1f77bcf86cd799439012",
    "page_views": {
      "total": 1250,
      "unique": 890,
      "by_page": {
        "invitation": 750,
        "details": 320,
        "rsvp": 180
      }
    },
    "rsvps": {
      "total": 45,
      "attending": 38,
      "not_attending": 5,
      "maybe": 2,
      "pending": 3
    },
    "conversion_rate": 5.1
  }
}
```

## 🔄 Common Workflows

### Complete Wedding Setup Workflow

1. **Register User Account**
2. **Login and Get Access Token**
3. **Create Wedding**
4. **Upload Wedding Photos**
5. **Add Guests (Manual or CSV Import)**
6. **Publish Wedding**
7. **Share Public URL with Guests**

### Guest RSVP Workflow

1. **Guest visits public wedding URL**
2. **Views wedding details**
3. **Submits RSVP form**
4. **Receives confirmation (if email configured)**
5. **Wedding couple tracks RSVPs**

## 📱 Error Handling

### Common Error Codes

| Code | Description | Example |
|------|-------------|---------|
| `VALIDATION_ERROR` | Invalid input data | Missing required fields |
| `UNAUTHORIZED` | Invalid or missing token | Token expired |
| `FORBIDDEN` | Insufficient permissions | Accessing other's data |
| `NOT_FOUND` | Resource doesn't exist | Invalid wedding ID |
| `CONFLICT` | Duplicate resource | Email already registered |
| `RATE_LIMITED` | Too many requests | API rate limit exceeded |

### Error Response Format
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      {
        "field": "email",
        "message": "Email is required and must be valid"
      }
    ]
  },
  "meta": {
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "req-123"
  }
}
```

## 🛠 Development Tips

### Testing with Postman

1. **Import Environment Variables**
   - `baseUrl`: `http://localhost:8080/api/v1`
   - `accessToken`: Store after login

2. **Common Headers**
   - `Authorization`: `Bearer {{accessToken}}`
   - `Content-Type`: `application/json`

3. **Test Sequence**
   - Register → Login → Create Wedding → Add Guests → Test RSVP

### Using curl Scripts

Create a script file with multiple commands:
```bash
#!/bin/bash

BASE_URL="http://localhost:8080/api/v1"
EMAIL="test@example.com"
PASSWORD="TestPass123!"

echo "Registering user..."
curl -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\",\"first_name\":\"Test\",\"last_name\":\"User\"}"

echo "Logging in..."
TOKEN=$(curl -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" | jq -r '.data.access_token')

echo "Creating wedding..."
curl -X POST "$BASE_URL/weddings" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Wedding","groom_name":"John","bride_name":"Jane","wedding_date":"2024-06-15T14:00:00Z"}'
```

## 📊 Rate Limits

| Endpoint Category | Requests per Minute | Burst |
|-------------------|---------------------|-------|
| Authentication | 5 | 3 |
| Default | 60 | 10 |
| File Upload | 10 | 5 |
| RSVP | 20 | 10 |
| Analytics | 100 | 20 |

Rate limit headers are included in responses:
```bash
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 59
X-RateLimit-Reset: 1642248699
```

## 🔒 Security Considerations

- **HTTPS Required** in production
- **JWT Tokens** expire after 15 minutes
- **Refresh Tokens** expire after 7 days
- **Input Validation** on all endpoints
- **File Upload Limits**: 5MB per file, 20MB total
- **CORS** configured for allowed origins

## 📞 Support

For API support:
- Check the [API Documentation](docs/backend/03-api-reference.md)
- Review [Database Schema](docs/backend/02-database-schema.md)
- Test with the provided examples
- Check server logs for detailed error information

---

**Last Updated**: 2026-02-07  
**API Version**: v1.0.0  
**Status**: Production Ready