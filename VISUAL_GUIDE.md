# Wedding Invitation Backend - Visual Guide

This guide provides visual representations of the API structure, data flows, and common workflows to help you understand how the system works.

## 🏗️ System Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Mobile App    │    │  Third Party    │
│   (React/Vue)   │    │   (iOS/Android) │    │   Integrations  │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────┴─────────────┐
                    │   Wedding Invitation     │
                    │      Backend API         │
                    │   (Go + Gin + MongoDB)   │
                    └─────────────┬─────────────┘
                                 │
          ┌──────────────────────┼──────────────────────┐
          │                      │                      │
    ┌─────┴─────┐        ┌───────┴───────┐      ┌───────┴───────┐
    │ MongoDB   │        │   File Storage│      │   Email      │
    │ Database  │        │   (S3/R2)     │      │  (SendGrid)  │
    └───────────┘        └───────────────┘      └───────────────┘
```

## 🔄 Authentication Flow

```
User Registration/Login
         │
         ▼
┌─────────────────┐
│   Email/Password│
│   Validation    │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│   bcrypt        │
│   Password      │
│   Hashing       │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│   JWT Token     │
│   Generation    │
│   (Access +     │
│    Refresh)     │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│   Token Return  │
│   to Client     │
└─────────────────┘
```

## 📊 Data Models Relationship

```
┌─────────────────┐
│      User       │
├─────────────────┤
│ • id            │
│ • email         │
│ • first_name    │
│ • last_name     │
│ • created_at    │
└─────────┬───────┘
          │ (creates)
          ▼
┌─────────────────┐     ┌─────────────────┐
│     Wedding     │────▶│      Guest      │
├─────────────────┤     ├─────────────────┤
│ • id            │     │ • id            │
│ • title         │     │ • wedding_id    │
│ • slug          │     │ • first_name    │
│ • couple_info   │     │ • last_name     │
│ • event_details │     │ • email         │
│ • theme_settings│     │ • relationship  │
│ • created_by    │     │ • side          │
└─────────┬───────┘     └─────────┬───────┘
          │ (has)                  │ (submits)
          ▼                        ▼
┌─────────────────┐     ┌─────────────────┐
│      RSVP       │◀────│      Media      │
├─────────────────┤     ├─────────────────┤
│ • id            │     │ • id            │
│ • wedding_id    │     │ • wedding_id    │
│ • guest_id      │     │ • filename      │
│ • status        │     │ • file_type     │
│ • attendance    │     │ • url           │
│ • dietary_info  │     │ • size          │
│ • custom_answers│     │ • created_at    │
└─────────────────┘     └─────────────────┘
```

## 🌐 API Endpoint Structure

```
/api/v1/
├── auth/                    # Authentication
│   ├── POST /register       # User registration
│   ├── POST /login          # User login
│   ├── POST /refresh        # Token refresh
│   └── POST /logout         # User logout
│
├── weddings/                # Wedding Management
│   ├── GET /weddings        # List user weddings
│   ├── POST /weddings       # Create wedding
│   ├── GET /weddings/:id    # Get wedding details
│   ├── PUT /weddings/:id    # Update wedding
│   └── DELETE /weddings/:id # Delete wedding
│
├── guests/                  # Guest Management
│   ├── GET /weddings/:id/guests
│   ├── POST /weddings/:id/guests
│   ├── PUT /guests/:id
│   ├── DELETE /guests/:id
│   └── POST /guests/bulk
│
├── rsvps/                   # RSVP Management
│   ├── GET /weddings/:id/rsvps
│   ├── POST /weddings/:id/rsvps
│   ├── PUT /rsvps/:id
│   └── GET /rsvps/:id/stats
│
├── upload/                  # File Uploads
│   ├── POST /upload         # Single file
│   ├── POST /upload/multiple # Multiple files
│   └── POST /upload/presign  # Get presigned URL
│
├── analytics/               # Analytics & Tracking
│   ├── POST /track/page-view
│   ├── POST /track/rsvp-submission
│   ├── GET /weddings/:id/analytics
│   └── GET /system/analytics
│
└── public/                  # Public Endpoints (No auth)
    ├── GET /weddings/:slug
    └── POST /weddings/:slug/rsvp
```

## 🔄 Common Workflows

### 1. Wedding Creation Workflow

```
1. User Login
   │
   ▼
2. Create Wedding
   │   - Title, slug, description
   │   - Couple information
   │   - Event details (date, venue)
   │   - Theme settings
   │
   ▼
3. Upload Media
   │   - Couple photos
   │   - Venue images
   │   - Gallery images
   │
   ▼
4. Configure RSVP
   │   - Enable/disable RSVP
   │   - Set deadline
   │   - Custom questions
   │
   ▼
5. Publish Wedding
   │   - Make public
   │   - Generate sharing link
   │   - Send invitations
```

### 2. Guest Management Workflow

```
1. Add Guests
   │   - Manual entry
   │   - CSV import
   │   - Bulk operations
   │
   ▼
2. Send Invitations
   │   - Email notifications
   │   - Personalized messages
   │   - Tracking delivery
   │
   ▼
3. Track RSVPs
   │   - Real-time updates
   │   - Status changes
   │   - Dietary restrictions
   │
   ▼
4. Manage Guest List
   │   - Update information
   │   - Add plus-ones
   │   - Export reports
```

### 3. Public RSVP Workflow

```
Guest Access
   │
   ▼
┌─────────────────┐
│ View Wedding    │
│ - Public page   │
│ - Event details │
│ - Photos        │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│ Submit RSVP     │
│ - Personal info │
│ - Attendance    │
│ - Plus-ones     │
│ - Custom Q's    │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│ Confirmation   │
│ - Email sent    │
│ - Reference #  │
│ - Edit link     │
└─────────────────┘
```

## 📈 Analytics Flow

```
User Interactions
   │
   ▼
┌─────────────────┐
│ Event Tracking │
│ - Page views    │
│ - RSVP starts  │
│ - Completions   │
│ - Abandonments  │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│ Data Storage    │
│ - MongoDB       │
│ - Time series   │
│ - Aggregated    │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│ Analytics API   │
│ - Reports       │
│ - Insights      │
│ - Export data   │
└─────────────────┘
```

## 🔒 Security Layers

```
┌─────────────────┐
│   CORS          │
│   Headers       │
│   Validation    │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│   Rate          │
│   Limiting      │
│   (Redis)       │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│   JWT           │
│   Authentication│
│   Authorization │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│   Input         │
│   Sanitization   │
│   XSS Protection│
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│   Business      │
│   Logic         │
│   Validation    │
└─────────────────┘
```

## 📁 File Upload Process

```
Client Upload
   │
   ▼
┌─────────────────┐
│   File          │
│   Validation    │
│ - Size check    │
│ - Type check    │
│ - Scan for      │
│   malware       │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│   Storage       │
│   Processing    │
│ - Generate      │
│   thumbnails    │
│ - Optimize      │
│ - Store in      │
│   S3/R2         │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│   Database      │
│   Record        │
│ - File metadata │
│ - URLs          │
│ - Associations  │
└─────────────────┘
```

## 🚀 Deployment Architecture

```
┌─────────────────┐
│   Load          │
│   Balancer      │
│   (Nginx)       │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│   Application   │
│   Servers       │
│   (Go API)      │
│   - Multiple    │
│     instances   │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│   Database      │
│   Cluster       │
│   (MongoDB)     │
│   - Replica set │
│   - Backups     │
└─────────┬───────┘
          │
          ▼
┌─────────────────┐
│   Cache         │
│   Layer         │
│   (Redis)       │
│   - Sessions    │
│   - Rate limits │
└─────────────────┘
```

## 📊 Response Format Standards

### Success Response
```json
{
  "success": true,
  "data": { ... },
  "message": "Operation completed successfully",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Error Response
```json
{
  "success": false,
  "error": "Validation failed",
  "details": {
    "field": "email",
    "message": "Invalid email format"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Paginated Response
```json
{
  "success": true,
  "data": [ ... ],
  "pagination": {
    "page": 1,
    "size": 20,
    "total": 100,
    "pages": 5
  }
}
```

---

This visual guide should help you understand the system architecture and data flows. For detailed API specifications, see the API Documentation.