# Wedding Invitation Backend API Documentation

## Overview

This is the REST API documentation for the Wedding Invitation Backend system. The API provides endpoints for managing wedding invitations, guests, RSVPs, and user authentication.

**Base URL**: `https://api.yourdomain.com/api/v1`

## Authentication

The API uses JWT (JSON Web Token) authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

### Getting a Token

#### Register User
```http
POST /auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "first_name": "John",
  "last_name": "Doe"
}
```

#### Login
```http
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Response**:
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "token_type": "Bearer",
  "expires_in": 900,
  "user": {
    "id": "507f1f77bcf86cd799439011",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "status": "active"
  }
}
```

## API Endpoints

### Wedding Management

#### Create Wedding
```http
POST /weddings
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "John & Jane's Wedding",
  "slug": "john-jane-wedding",
  "description": "We invite you to celebrate our special day",
  "couple_info": {
    "partner1_name": "John Doe",
    "partner2_name": "Jane Smith",
    "partner1_role": "groom",
    "partner2_role": "bride"
  },
  "event_details": {
    "date": "2024-06-15T17:00:00Z",
    "venue_name": "Sunset Gardens",
    "venue_address": "123 Garden Lane, Springfield",
    "ceremony_time": "2024-06-15T17:00:00Z",
    "reception_time": "2024-06-15T19:00:00Z"
  },
  "theme_settings": {
    "primary_color": "#FF6B6B",
    "secondary_color": "#4ECDC4",
    "font_family": "Playfair Display",
    "background_style": "floral"
  }
}
```

#### Get Wedding by ID
```http
GET /weddings/{id}
```

#### Get Wedding by Slug
```http
GET /weddings/slug/{slug}
```

#### Update Wedding
```http
PUT /weddings/{id}
Authorization: Bearer <token>
Content-Type: application/json

{
  "title": "Updated Wedding Title",
  "description": "Updated description"
}
```

#### Delete Wedding
```http
DELETE /weddings/{id}
Authorization: Bearer <token>
```

#### Publish Wedding
```http
POST /weddings/{id}/publish
Authorization: Bearer <token>
```

#### Get User's Weddings
```http
GET /weddings?page=1&page_size=20&status=published&search=wedding
Authorization: Bearer <token>
```

#### List Public Weddings
```http
GET /public/weddings?page=1&page_size=20&search=garden
```

### Guest Management

#### Create Guest
```http
POST /weddings/{wedding_id}/guests
Authorization: Bearer <token>
Content-Type: application/json

{
  "first_name": "Alice",
  "last_name": "Johnson",
  "email": "alice@example.com",
  "phone": "+1234567890",
  "address": "123 Main St, City, State",
  "relationship": "friend",
  "invited_to": ["ceremony", "reception"],
  "plus_one_allowed": true,
  "dietary_restrictions": ["vegetarian"],
  "notes": "Guest of the bride"
}
```

#### Get Wedding Guests
```http
GET /weddings/{wedding_id}/guests?page=1&page_size=50&relationship=friend
Authorization: Bearer <token>
```

#### Import Guests from CSV
```http
POST /weddings/{wedding_id}/guests/import
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: <csv-file>
```

CSV Format:
```csv
first_name,last_name,email,phone,relationship,invited_to,plus_one_allowed,dietary_restrictions,notes
Alice,Johnson,alice@example.com,+1234567890,friend,"ceremony,reception",true,vegetarian,Guest of bride
Bob,Smith,bob@example.com,+1234567891,family,ceremony,false,,Brother of groom
```

#### Export Guests to CSV
```http
GET /weddings/{wedding_id}/guests/export
Authorization: Bearer <token>
```

### RSVP Management

#### Submit RSVP (Public)
```http
POST /public/weddings/{wedding_id}/rsvp
Content-Type: application/json

{
  "first_name": "Alice",
  "last_name": "Johnson",
  "email": "alice@example.com",
  "phone": "+1234567890",
  "attending": true,
  "attendance_count": 2,
  "plus_ones": [
    {
      "name": "Bob Johnson",
      "relationship": "spouse",
      "attending": true
    }
  ],
  "dietary_restrictions": ["vegetarian"],
  "custom_answers": {
    "meal_preference": "vegetarian",
    "song_request": "Perfect by Ed Sheeran"
  },
  "message": "Looking forward to celebrating with you!"
}
```

#### Get Wedding RSVPs
```http
GET /weddings/{wedding_id}/rsvps?page=1&page_size=50&attending=true
Authorization: Bearer <token>
```

#### Update RSVP Status
```http
PUT /rsvps/{rsvp_id}/status
Authorization: Bearer <token>
Content-Type: application/json

{
  "status": "confirmed"
}
```

### File Upload

#### Upload File
```http
POST /protected/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: <image-file>
wedding_id: 507f1f77bcf86cd799439011
```

#### Get Presigned Upload URL
```http
GET /protected/upload/presigned?wedding_id=507f1f77bcf86cd799439011&content_type=image/jpeg
Authorization: Bearer <token>
```

### User Management

#### Get User Profile
```http
GET /users/profile
Authorization: Bearer <token>
```

#### Update User Profile
```http
PUT /users/profile
Authorization: Bearer <token>
Content-Type: application/json

{
  "first_name": "John",
  "last_name": "Smith",
  "phone": "+1234567890"
}
```

#### Change Password
```http
POST /auth/change-password
Authorization: Bearer <token>
Content-Type: application/json

{
  "current_password": "OldPassword123!",
  "new_password": "NewPassword456!"
}
```

#### Refresh Token
```http
POST /auth/refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Logout
```http
POST /auth/logout
Authorization: Bearer <token>
```

### Analytics

#### Get Wedding Analytics
```http
GET /weddings/{id}/analytics
Authorization: Bearer <token>
```

#### Track Page View
```http
POST /analytics/track/page-view
Content-Type: application/json

{
  "wedding_id": "507f1f77bcf86cd799439011",
  "page_url": "/weddings/john-jane-wedding",
  "user_agent": "Mozilla/5.0...",
  "ip_address": "192.168.1.1",
  "referrer": "https://google.com"
}
```

## Response Formats

### Success Response
```json
{
  "success": true,
  "data": {
    // Response data here
  },
  "message": "Operation successful"
}
```

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": {
      "field": "email",
      "reason": "Invalid email format"
    }
  }
}
```

### Pagination Response
```json
{
  "success": true,
  "data": {
    "items": [
      // Array of items
    ],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total_items": 100,
      "total_pages": 5,
      "has_next": true,
      "has_prev": false
    }
  }
}
```

## HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | OK - Request successful |
| 201 | Created - Resource created successfully |
| 400 | Bad Request - Invalid input data |
| 401 | Unauthorized - Authentication required |
| 403 | Forbidden - Permission denied |
| 404 | Not Found - Resource not found |
| 409 | Conflict - Resource already exists |
| 422 | Unprocessable Entity - Validation failed |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error - Server error |

## Rate Limiting

The API implements rate limiting to prevent abuse:

- **Authentication endpoints**: 5 requests per minute
- **General endpoints**: 100 requests per minute  
- **Upload endpoints**: 10 requests per minute
- **Analytics tracking**: 600 requests per minute

Rate limit headers are included in responses:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640995200
```

## File Upload Limits

- **Maximum file size**: 5MB per file
- **Maximum total size**: 20MB per request
- **Maximum files**: 10 files per request
- **Allowed formats**: JPEG, PNG, WebP

## CORS

The API supports Cross-Origin Resource Sharing for specified origins. Configure allowed origins in your deployment configuration.

## SDK Examples

### JavaScript/TypeScript
```typescript
const api = new WeddingAPI({
  baseURL: 'https://api.yourdomain.com/api/v1',
  token: 'your-jwt-token'
});

// Create a wedding
const wedding = await api.weddings.create({
  title: 'My Wedding',
  couple_info: {
    partner1_name: 'John',
    partner2_name: 'Jane'
  }
});

// Get public weddings
const publicWeddings = await api.public.weddings.list({
  page: 1,
  search: 'garden'
});
```

### Python
```python
from wedding_api import WeddingAPI

api = WeddingAPI(
    base_url='https://api.yourdomain.com/api/v1',
    token='your-jwt-token'
)

# Create a wedding
wedding = api.weddings.create({
    'title': 'My Wedding',
    'couple_info': {
        'partner1_name': 'John',
        'partner2_name': 'Jane'
    }
})
```

## Error Handling

Always check the response status and handle errors appropriately:

```javascript
try {
  const response = await api.weddings.create(weddingData);
  console.log('Wedding created:', response.data);
} catch (error) {
  if (error.response?.status === 422) {
    console.log('Validation error:', error.response.data.error.details);
  } else if (error.response?.status === 401) {
    console.log('Authentication required');
  } else {
    console.log('Server error:', error.message);
  }
}
```

## Support

For API support and questions:
- Documentation: https://docs.yourdomain.com
- Email: api-support@yourdomain.com
- Issues: https://github.com/your-org/wedding-invitation-backend/issues