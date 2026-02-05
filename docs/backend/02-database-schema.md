# Database Schema Documentation

## Overview

This document defines the MongoDB database schema for the Wedding Invitation system. MongoDB is chosen as a document database for its flexibility in handling semi-structured data, horizontal scalability, and native support for nested documents - perfect for wedding data with varying structures across different invitation themes.

## Collections Overview

| Collection | Description | Approximate Documents |
|------------|-------------|----------------------|
| `users` | Registered system users (couples creating invitations) | 1 per wedding couple |
| `weddings` | Wedding invitation data and configuration | 1 per wedding |
| `rsvps` | RSVP responses from guests | Multiple per wedding |
| `guests` | Guest list with contact information | Multiple per wedding |
| `media` | Uploaded images and files | Multiple per wedding |
| `analytics` | Page views and interaction events | Many per wedding |

## Collection: users

### Purpose

Stores registered user accounts for couples who create wedding invitations. Each user can own multiple weddings (e.g., for testing different themes).

### Go Struct Definition

```go
package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
    ID                primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
    Email             string               `bson:"email" json:"email" validate:"required,email"`
    PasswordHash      string               `bson:"password_hash" json:"-"` // Never expose in JSON
    FirstName         string               `bson:"first_name" json:"first_name" validate:"required,min=2,max=50"`
    LastName          string               `bson:"last_name" json:"last_name" validate:"required,min=2,max=50"`
    Phone             string               `bson:"phone,omitempty" json:"phone,omitempty" validate:"omitempty,e164"`
    EmailVerified     bool                 `bson:"email_verified" json:"email_verified"`
    EmailVerifiedAt   *time.Time           `bson:"email_verified_at,omitempty" json:"email_verified_at,omitempty"`
    ProfileImageURL   string               `bson:"profile_image_url,omitempty" json:"profile_image_url,omitempty" validate:"omitempty,url"`
    WeddingIDs        []primitive.ObjectID `bson:"wedding_ids" json:"wedding_ids"` // References to weddings
    CreatedAt         time.Time            `bson:"created_at" json:"created_at"`
    UpdatedAt         time.Time            `bson:"updated_at" json:"updated_at"`
    LastLoginAt       *time.Time           `bson:"last_login_at,omitempty" json:"last_login_at,omitempty"`
    Status            string               `bson:"status" json:"status" validate:"oneof=active inactive suspended"`
    PreferredLanguage string               `bson:"preferred_language,omitempty" json:"preferred_language,omitempty"`
    Timezone          string               `bson:"timezone,omitempty" json:"timezone,omitempty"`
}

// Validation tags explained:
// - required: Field must be provided
// - email: Must be valid email format
// - min/max: String length constraints
// - e164: International phone number format
// - url: Valid URL format
// - omitempty: Skip if empty
// - oneof: Must be one of the specified values
```

### Indexes

```javascript
// Unique email index - prevents duplicate accounts
db.users.createIndex({ "email": 1 }, { unique: true })

// Index on wedding_ids for fast lookups when listing user's weddings
db.users.createIndex({ "wedding_ids": 1 })

// Compound index for status and created_at for user listing queries
db.users.createIndex({ "status": 1, "created_at": -1 })

// TTL index on email verification (optional: auto-delete unverified accounts after 30 days)
// db.users.createIndex({ "created_at": 1 }, { expireAfterSeconds: 2592000, partialFilterExpression: { email_verified: false } })
```

### Validation Rules

```javascript
db.createCollection("users", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["email", "password_hash", "first_name", "last_name", "created_at", "updated_at", "status"],
            properties: {
                email: {
                    bsonType: "string",
                    pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
                },
                password_hash: {
                    bsonType: "string",
                    minLength: 60  // bcrypt hash length
                },
                first_name: {
                    bsonType: "string",
                    minLength: 2,
                    maxLength: 50
                },
                last_name: {
                    bsonType: "string",
                    minLength: 2,
                    maxLength: 50
                },
                status: {
                    enum: ["active", "inactive", "suspended"]
                },
                email_verified: {
                    bsonType: "bool"
                }
            }
        }
    }
})
```

## Collection: weddings

### Purpose

Core collection storing all wedding invitation data, including metadata, theme settings, event details, and configuration.

### Go Struct Definition

```go
package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// EventDetails represents wedding ceremony and reception info
type EventDetails struct {
    Title          string     `bson:"title" json:"title" validate:"required,max=100"`
    Date           time.Time  `bson:"date" json:"date" validate:"required"`
    Time           string     `bson:"time,omitempty" json:"time,omitempty"`
    VenueName      string     `bson:"venue_name" json:"venue_name" validate:"required,max=200"`
    VenueAddress   string     `bson:"venue_address" json:"venue_address" validate:"required,max=500"`
    VenueMapURL    string     `bson:"venue_map_url,omitempty" json:"venue_map_url,omitempty" validate:"omitempty,url"`
    DressCode      string     `bson:"dress_code,omitempty" json:"dress_code,omitempty"`
    AdditionalInfo string     `bson:"additional_info,omitempty" json:"additional_info,omitempty"`
}

// CoupleInfo represents bride and groom details
type CoupleInfo struct {
    Partner1 struct {
        FirstName     string `bson:"first_name" json:"first_name" validate:"required"`
        LastName      string `bson:"last_name" json:"last_name" validate:"required"`
        FullName      string `bson:"full_name" json:"full_name"`
        PhotoURL      string `bson:"photo_url,omitempty" json:"photo_url,omitempty"`
        SocialLinks   map[string]string `bson:"social_links,omitempty" json:"social_links,omitempty"`
    } `bson:"partner1" json:"partner1"`
    
    Partner2 struct {
        FirstName     string `bson:"first_name" json:"first_name" validate:"required"`
        LastName      string `bson:"last_name" json:"last_name" validate:"required"`
        FullName      string `bson:"full_name" json:"full_name"`
        PhotoURL      string `bson:"photo_url,omitempty" json:"photo_url,omitempty"`
        SocialLinks   map[string]string `bson:"social_links,omitempty" json:"social_links,omitempty"`
    } `bson:"partner2" json:"partner2"`
    
    Story      string `bson:"story,omitempty" json:"story,omitempty" validate:"omitempty,max=2000"`
    Engagement struct {
        Date   *time.Time `bson:"date,omitempty" json:"date,omitempty"`
        Story  string     `bson:"story,omitempty" json:"story,omitempty"`
        PhotoURL string   `bson:"photo_url,omitempty" json:"photo_url,omitempty"`
    } `bson:"engagement,omitempty" json:"engagement,omitempty"`
}

// ThemeSettings represents visual customization
type ThemeSettings struct {
    ThemeID         string            `bson:"theme_id" json:"theme_id" validate:"required"`
    PrimaryColor    string            `bson:"primary_color" json:"primary_color" validate:"omitempty,hexcolor"`
    SecondaryColor  string            `bson:"secondary_color" json:"secondary_color" validate:"omitempty,hexcolor"`
    BackgroundColor string            `bson:"background_color" json:"background_color" validate:"omitempty,hexcolor"`
    FontFamily      string            `bson:"font_family" json:"font_family"`
    CustomCSS       string            `bson:"custom_css,omitempty" json:"-"` // Never expose in API
    CustomSettings  map[string]interface{} `bson:"custom_settings,omitempty" json:"custom_settings,omitempty"`
}

// RSVPSettings configures RSVP form behavior
type RSVPSettings struct {
    Enabled           bool              `bson:"enabled" json:"enabled"`
    Deadline          *time.Time        `bson:"deadline,omitempty" json:"deadline,omitempty"`
    AllowPlusOne      bool              `bson:"allow_plus_one" json:"allow_plus_one"`
    MaxPlusOnes       int               `bson:"max_plus_ones" json:"max_plus_ones" validate:"min=0,max=5"`
    CollectEmail      bool              `bson:"collect_email" json:"collect_email"`
    CollectPhone      bool              `bson:"collect_phone" json:"collect_phone"`
    CollectDietary    bool              `bson:"collect_dietary" json:"collect_dietary"`
    DietaryOptions    []string          `bson:"dietary_options,omitempty" json:"dietary_options,omitempty"`
    CustomQuestions   []CustomQuestion  `bson:"custom_questions,omitempty" json:"custom_questions,omitempty"`
    ConfirmationEmail bool              `bson:"confirmation_email" json:"confirmation_email"`
    EmailTemplate     string            `bson:"email_template,omitempty" json:"email_template,omitempty"`
}

// CustomQuestion for RSVP forms
type CustomQuestion struct {
    ID          string   `bson:"id" json:"id"`
    Question    string   `bson:"question" json:"question" validate:"required,max=200"`
    Type        string   `bson:"type" json:"type" validate:"oneof=text textarea select checkbox radio"`
    Required    bool     `bson:"required" json:"required"`
    Options     []string `bson:"options,omitempty" json:"options,omitempty"`
    Order       int      `bson:"order" json:"order"`
}

// GalleryImage represents a photo in the gallery
type GalleryImage struct {
    ID          string    `bson:"id" json:"id"`
    URL         string    `bson:"url" json:"url"`
    ThumbnailURL string   `bson:"thumbnail_url" json:"thumbnail_url"`
    Caption     string    `bson:"caption,omitempty" json:"caption,omitempty"`
    Order       int       `bson:"order" json:"order"`
    UploadedAt  time.Time `bson:"uploaded_at" json:"uploaded_at"`
    FileSize    int64     `bson:"file_size" json:"file_size"`
}

// Wedding is the main collection document
type Wedding struct {
    ID               primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    UserID           primitive.ObjectID `bson:"user_id" json:"user_id"` // Reference to owner
    
    // URL and Access
    Slug             string             `bson:"slug" json:"slug" validate:"required,min=3,max=50,slug"`
    PasswordHash     string             `bson:"password_hash,omitempty" json:"-"` // For private weddings
    IsPublic         bool               `bson:"is_public" json:"is_public"`
    
    // Content
    Title            string             `bson:"title" json:"title" validate:"required,max=100"`
    Couple           CoupleInfo         `bson:"couple" json:"couple"`
    Event            EventDetails       `bson:"event" json:"event"`
    
    // Media
    CoverImageURL    string             `bson:"cover_image_url,omitempty" json:"cover_image_url,omitempty"`
    GalleryImages    []GalleryImage     `bson:"gallery_images,omitempty" json:"gallery_images,omitempty"`
    GalleryEnabled   bool               `bson:"gallery_enabled" json:"gallery_enabled"`
    
    // Settings
    Theme            ThemeSettings      `bson:"theme" json:"theme"`
    RSVP             RSVPSettings       `bson:"rsvp" json:"rsvp"`
    
    // Social/Sharing
    ShareMessage     string             `bson:"share_message,omitempty" json:"share_message,omitempty" validate:"omitempty,max=280"`
    
    // Status
    Status           string             `bson:"status" json:"status" validate:"oneof=draft published expired archived"`
    PublishedAt      *time.Time         `bson:"published_at,omitempty" json:"published_at,omitempty"`
    ExpiresAt        *time.Time         `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
    
    // Counts (denormalized for performance)
    RSVPCount        int                `bson:"rsvp_count" json:"rsvp_count"`
    GuestCount       int                `bson:"guest_count" json:"guest_count"`
    TotalAttending   int                `bson:"total_attending" json:"total_attending"`
    
    // Metadata
    CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
    LastViewedAt     *time.Time         `bson:"last_viewed_at,omitempty" json:"last_viewed_at,omitempty"`
    ViewCount        int64              `bson:"view_count" json:"view_count"`
}

// Request/Response DTOs
type CreateWeddingRequest struct {
    Title         string         `json:"title" binding:"required"`
    Slug          string         `json:"slug" binding:"required,min=3,max=50"`
    Couple        CoupleInfo     `json:"couple" binding:"required"`
    Event         EventDetails   `json:"event" binding:"required"`
    Theme         ThemeSettings  `json:"theme" binding:"required"`
    RSVP          RSVPSettings   `json:"rsvp"`
    IsPublic      bool           `json:"is_public"`
}

type UpdateWeddingRequest struct {
    Title            string             `json:"title,omitempty"`
    Slug             string             `json:"slug,omitempty"`
    Couple           *CoupleInfo        `json:"couple,omitempty"`
    Event            *EventDetails      `json:"event,omitempty"`
    Theme            *ThemeSettings     `json:"theme,omitempty"`
    RSVP             *RSVPSettings      `json:"rsvp,omitempty"`
    ShareMessage     string             `json:"share_message,omitempty"`
    CoverImageURL    string             `json:"cover_image_url,omitempty"`
    IsPublic         *bool              `json:"is_public,omitempty"`
}
```

### Indexes

```javascript
// Unique slug index - critical for URL routing
db.weddings.createIndex({ "slug": 1 }, { unique: true })

// User ID index for listing user's weddings
db.weddings.createIndex({ "user_id": 1, "created_at": -1 })

// Status index for filtering
db.weddings.createIndex({ "status": 1 })

// Compound index for public listings with pagination
db.weddings.createIndex({ "status": 1, "published_at": -1 })

// Event date index for upcoming wedding queries
db.weddings.createIndex({ "event.date": 1 })

// TTL index to auto-archive expired weddings (optional)
// db.weddings.createIndex({ "expires_at": 1 }, { expireAfterSeconds: 0 })
```

### Validation Rules

```javascript
db.createCollection("weddings", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["user_id", "slug", "title", "couple", "event", "theme", "status", "created_at", "updated_at"],
            properties: {
                slug: {
                    bsonType: "string",
                    pattern: "^[a-z0-9]+(?:-[a-z0-9]+)*$",  // URL-safe slug
                    minLength: 3,
                    maxLength: 50
                },
                title: {
                    bsonType: "string",
                    maxLength: 100
                },
                status: {
                    enum: ["draft", "published", "expired", "archived"]
                },
                "event.date": {
                    bsonType: "date"
                },
                is_public: {
                    bsonType: "bool"
                },
                view_count: {
                    bsonType: "int",
                    minimum: 0
                }
            }
        }
    }
})
```

## Collection: rsvps

### Purpose

Stores guest RSVP responses with attendance status, plus-one information, dietary restrictions, and answers to custom questions.

### Go Struct Definition

```go
package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// PlusOneInfo for guests bringing additional people
type PlusOneInfo struct {
    FirstName string `bson:"first_name" json:"first_name"`
    LastName  string `bson:"last_name" json:"last_name"`
    Dietary   string `bson:"dietary,omitempty" json:"dietary,omitempty"`
}

// CustomAnswer stores responses to custom questions
type CustomAnswer struct {
    QuestionID string      `bson:"question_id" json:"question_id"`
    Question   string      `bson:"question" json:"question"`
    Answer     interface{} `bson:"answer" json:"answer"` // Can be string, []string, bool, etc.
}

type RSVP struct {
    ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    WeddingID       primitive.ObjectID `bson:"wedding_id" json:"wedding_id"`
    GuestID         *primitive.ObjectID `bson:"guest_id,omitempty" json:"guest_id,omitempty"` // Link to pre-registered guest
    
    // Guest Information (if not linked to pre-registered guest)
    FirstName       string             `bson:"first_name" json:"first_name" validate:"required,max=50"`
    LastName        string             `bson:"last_name" json:"last_name" validate:"required,max=50"`
    Email           string             `bson:"email,omitempty" json:"email,omitempty" validate:"omitempty,email,max=100"`
    Phone           string             `bson:"phone,omitempty" json:"phone,omitempty"`
    
    // RSVP Response
    Status          string             `bson:"status" json:"status" validate:"oneof=attending not-attending maybe"`
    AttendanceCount int                `bson:"attendance_count" json:"attendance_count" validate:"min=1"`
    
    // Plus Ones
    PlusOnes        []PlusOneInfo      `bson:"plus_ones,omitempty" json:"plus_ones,omitempty"`
    PlusOneCount    int                `bson:"plus_one_count" json:"plus_one_count" validate:"min=0,max=5"`
    
    // Dietary & Preferences
    DietaryRestrictions string         `bson:"dietary_restrictions,omitempty" json:"dietary_restrictions,omitempty"`
    DietarySelected []string           `bson:"dietary_selected,omitempty" json:"dietary_selected,omitempty"`
    AdditionalNotes string             `bson:"additional_notes,omitempty" json:"additional_notes,omitempty" validate:"omitempty,max=500"`
    
    // Custom Questions Answers
    CustomAnswers   []CustomAnswer     `bson:"custom_answers,omitempty" json:"custom_answers,omitempty"`
    
    // Metadata
    SubmittedAt     time.Time          `bson:"submitted_at" json:"submitted_at"`
    UpdatedAt       *time.Time         `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
    IPAddress       string             `bson:"ip_address,omitempty" json:"-"` // For spam prevention
    UserAgent       string             `bson:"user_agent,omitempty" json:"-"` // For analytics
    
    // Confirmation
    ConfirmationSent bool              `bson:"confirmation_sent" json:"confirmation_sent"`
    ConfirmationSentAt *time.Time        `bson:"confirmation_sent_at,omitempty" json:"confirmation_sent_at,omitempty"`
    
    // Internal tracking
    Source          string             `bson:"source" json:"source" validate:"oneof=web direct_link qr_code manual"`
    Notes           string             `bson:"notes,omitempty" json:"notes,omitempty"` // Admin notes
}

// RSVPStatus represents possible response statuses
type RSVPStatus string

const (
    RSVPAttending    RSVPStatus = "attending"
    RSVPNotAttending RSVPStatus = "not-attending"
    RSVPMaybe        RSVPStatus = "maybe"
)

// RSVPStatistics for dashboard display
type RSVPStatistics struct {
    TotalResponses    int            `json:"total_responses"`
    Attending         int            `json:"attending"`
    NotAttending      int            `json:"not_attending"`
    Maybe             int            `json:"maybe"`
    TotalGuests       int            `json:"total_guests"` // Including plus ones
    PlusOnesCount     int            `json:"plus_ones_count"`
    DietaryCounts     map[string]int `json:"dietary_counts"`
    SubmissionTrend   []DailyCount   `json:"submission_trend"`
}

type DailyCount struct {
    Date  string `json:"date"`
    Count int    `json:"count"`
}

// CreateRSVPRequest for new RSVP submissions
type CreateRSVPRequest struct {
    FirstName           string         `json:"first_name" binding:"required"`
    LastName            string         `json:"last_name" binding:"required"`
    Email               string         `json:"email,omitempty"`
    Phone               string         `json:"phone,omitempty"`
    Status              string         `json:"status" binding:"required,oneof=attending not-attending maybe"`
    AttendanceCount     int            `json:"attendance_count" binding:"required,min=1"`
    PlusOnes            []PlusOneInfo  `json:"plus_ones,omitempty"`
    DietaryRestrictions string         `json:"dietary_restrictions,omitempty"`
    DietarySelected     []string       `json:"dietary_selected,omitempty"`
    AdditionalNotes     string         `json:"additional_notes,omitempty"`
    CustomAnswers       []CustomAnswer `json:"custom_answers,omitempty"`
}

// UpdateRSVPRequest for modifying existing RSVPs
type UpdateRSVPRequest struct {
    Status              string         `json:"status,omitempty" binding:"omitempty,oneof=attending not-attending maybe"`
    AttendanceCount     int            `json:"attendance_count,omitempty" binding:"omitempty,min=1"`
    PlusOnes            []PlusOneInfo  `json:"plus_ones,omitempty"`
    DietaryRestrictions string         `json:"dietary_restrictions,omitempty"`
    AdditionalNotes     string         `json:"additional_notes,omitempty"`
    CustomAnswers       []CustomAnswer `json:"custom_answers,omitempty"`
}
```

### Indexes

```javascript
// Wedding ID index for fetching all RSVPs for a wedding
db.rsvps.createIndex({ "wedding_id": 1, "submitted_at": -1 })

// Guest name index for searching
db.rsvps.createIndex({ "wedding_id": 1, "last_name": 1, "first_name": 1 })

// Email index for duplicate prevention and searching
db.rsvps.createIndex({ "wedding_id": 1, "email": 1 })

// Status index for filtering
db.rsvps.createIndex({ "wedding_id": 1, "status": 1 })

// Compound index for statistics queries
db.rsvps.createIndex({ "wedding_id": 1, "status": 1, "submitted_at": -1 })

// Guest ID index for linking to guest list
db.rsvps.createIndex({ "guest_id": 1 })

// Unique constraint: one RSVP per email per wedding
db.rsvps.createIndex({ "wedding_id": 1, "email": 1 }, { 
    unique: true, 
    partialFilterExpression: { email: { $exists: true, $ne: "" } }
})
```

### Validation Rules

```javascript
db.createCollection("rsvps", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["wedding_id", "first_name", "last_name", "status", "attendance_count", "submitted_at"],
            properties: {
                wedding_id: {
                    bsonType: "objectId"
                },
                status: {
                    enum: ["attending", "not-attending", "maybe"]
                },
                attendance_count: {
                    bsonType: "int",
                    minimum: 1
                },
                plus_one_count: {
                    bsonType: "int",
                    minimum: 0,
                    maximum: 5
                },
                email: {
                    bsonType: "string",
                    pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
                },
                first_name: {
                    bsonType: "string",
                    maxLength: 50
                },
                last_name: {
                    bsonType: "string",
                    maxLength: 50
                }
            }
        }
    }
})
```

## Collection: guests

### Purpose

Pre-registered guest list with contact information, invitation status tracking, and RSVP linkage. Enables couples to track who they've invited and who has responded.

### Go Struct Definition

```go
package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// Address represents a physical address
type Address struct {
    Street     string `bson:"street,omitempty" json:"street,omitempty"`
    City       string `bson:"city,omitempty" json:"city,omitempty"`
    State      string `bson:"state,omitempty" json:"state,omitempty"`
    ZIPCode    string `bson:"zip_code,omitempty" json:"zip_code,omitempty"`
    Country    string `bson:"country,omitempty" json:"country,omitempty"`
}

// InvitationStatus tracks the invitation lifecycle
type InvitationStatus string

const (
    InvitationPending   InvitationStatus = "pending"
    InvitationSent      InvitationStatus = "sent"
    InvitationDelivered InvitationStatus = "delivered"
    InvitationOpened    InvitationStatus = "opened"
    InvitationBounced   InvitationStatus = "bounced"
)

// Guest represents a person invited to the wedding
type Guest struct {
    ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    WeddingID          primitive.ObjectID `bson:"wedding_id" json:"wedding_id"`
    
    // Personal Information
    FirstName          string             `bson:"first_name" json:"first_name" validate:"required,max=50"`
    LastName           string             `bson:"last_name" json:"last_name" validate:"required,max=50"`
    Email              string             `bson:"email,omitempty" json:"email,omitempty" validate:"omitempty,email,max=100"`
    Phone              string             `bson:"phone,omitempty" json:"phone,omitempty"`
    Address            *Address           `bson:"address,omitempty" json:"address,omitempty"`
    
    // Relationship
    Relationship       string             `bson:"relationship,omitempty" json:"relationship,omitempty"` // e.g., "family", "friend", "colleague"
    Side               string             `bson:"side,omitempty" json:"side,omitempty" validate:"omitempty,oneof=bride groom both"`
    
    // Grouping
    TableAssignment    string             `bson:"table_assignment,omitempty" json:"table_assignment,omitempty"`
    GroupID            string             `bson:"group_id,omitempty" json:"group_id,omitempty"` // For plus-one grouping
    
    // Invitation
    InvitedVia         string             `bson:"invited_via" json:"invited_via" validate:"oneof=digital print both"`
    InvitationStatus   InvitationStatus   `bson:"invitation_status" json:"invitation_status"`
    InvitationSentAt   *time.Time         `bson:"invitation_sent_at,omitempty" json:"invitation_sent_at,omitempty"`
    InvitationOpenedAt *time.Time         `bson:"invitation_opened_at,omitempty" json:"invitation_opened_at,omitempty"`
    
    // RSVP Linkage
    RSVPID             *primitive.ObjectID `bson:"rsvp_id,omitempty" json:"rsvp_id,omitempty"`
    RSVPStatus         string             `bson:"rsvp_status,omitempty" json:"rsvp_status,omitempty" validate:"omitempty,oneof=pending attending not-attending maybe"`
    RSVPSubmittedAt    *time.Time         `bson:"rsvp_submitted_at,omitempty" json:"rsvp_submitted_at,omitempty"`
    
    // Plus One Configuration
    AllowPlusOne       bool               `bson:"allow_plus_one" json:"allow_plus_one"`
    MaxPlusOnes        int                `bson:"max_plus_ones" json:"max_plus_ones" validate:"min=0,max=3"`
    PlusOneNames       []string           `bson:"plus_one_names,omitempty" json:"plus_one_names,omitempty"`
    
    // Special Requirements
    VIP                bool               `bson:"vip" json:"vip"`
    DietaryNotes       string             `bson:"dietary_notes,omitempty" json:"dietary_notes,omitempty"`
    AccessibilityNeeds string             `bson:"accessibility_needs,omitempty" json:"accessibility_needs,omitempty"`
    
    // Internal
    Notes              string             `bson:"notes,omitempty" json:"notes,omitempty" validate:"omitempty,max=1000"`
    ImportBatchID      string             `bson:"import_batch_id,omitempty" json:"-"` // For tracking CSV imports
    
    // Metadata
    CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
    CreatedBy          primitive.ObjectID `bson:"created_by" json:"created_by"` // User who added this guest
}

// GuestGroup for managing families or couples
type GuestGroup struct {
    ID          string             `bson:"_id,omitempty" json:"id"`
    WeddingID   primitive.ObjectID `bson:"wedding_id" json:"wedding_id"`
    Name        string             `bson:"name" json:"name"` // e.g., "The Smith Family"
    GuestIDs    []primitive.ObjectID `bson:"guest_ids" json:"guest_ids"`
    PrimaryGuestID primitive.ObjectID `bson:"primary_guest_id" json:"primary_guest_id"`
    CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
}

// GuestImportResult tracks CSV import results
type GuestImportResult struct {
    SuccessCount   int                `json:"success_count"`
    ErrorCount     int                `json:"error_count"`
    Errors         []ImportError      `json:"errors,omitempty"`
    BatchID        string             `json:"batch_id"`
    DuplicatesFound int               `json:"duplicates_found"`
}

type ImportError struct {
    Row    int    `json:"row"`
    Field  string `json:"field"`
    Value  string `json:"value"`
    Error  string `json:"error"`
}

// CreateGuestRequest for adding new guests
type CreateGuestRequest struct {
    FirstName          string    `json:"first_name" binding:"required"`
    LastName           string    `json:"last_name" binding:"required"`
    Email              string    `json:"email,omitempty"`
    Phone              string    `json:"phone,omitempty"`
    Address            *Address  `json:"address,omitempty"`
    Relationship       string    `json:"relationship,omitempty"`
    Side               string    `json:"side,omitempty"`
    AllowPlusOne       bool      `json:"allow_plus_one"`
    MaxPlusOnes        int       `json:"max_plus_ones,omitempty"`
    DietaryNotes       string    `json:"dietary_notes,omitempty"`
    VIP                bool      `json:"vip"`
    Notes              string    `json:"notes,omitempty"`
}

// BulkImportGuestsRequest for CSV uploads
type BulkImportGuestsRequest struct {
    Guests []CreateGuestRequest `json:"guests" binding:"required,min=1,max=500"`
}
```

### Indexes

```javascript
// Wedding ID index for listing guests
db.guests.createIndex({ "wedding_id": 1, "last_name": 1, "first_name": 1 })

// Email index for duplicate detection within wedding
db.guests.createIndex({ "wedding_id": 1, "email": 1 })

// RSVP linkage index
db.guests.createIndex({ "rsvp_id": 1 })

// Side index for filtering by bride/groom
db.guests.createIndex({ "wedding_id": 1, "side": 1 })

// Invitation status for follow-up queries
db.guests.createIndex({ "wedding_id": 1, "invitation_status": 1 })

// Group ID for family queries
db.guests.createIndex({ "group_id": 1 })

// RSVP status for tracking pending responses
db.guests.createIndex({ "wedding_id": 1, "rsvp_status": 1 })

// Unique constraint on email per wedding (partial - only when email exists)
db.guests.createIndex({ "wedding_id": 1, "email": 1 }, { 
    unique: true,
    partialFilterExpression: { email: { $exists: true, $ne: "" } }
})
```

### Validation Rules

```javascript
db.createCollection("guests", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["wedding_id", "first_name", "last_name", "created_at", "updated_at", "created_by"],
            properties: {
                wedding_id: {
                    bsonType: "objectId"
                },
                email: {
                    bsonType: "string",
                    pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
                },
                side: {
                    enum: ["bride", "groom", "both"]
                },
                rsvp_status: {
                    enum: ["pending", "attending", "not-attending", "maybe"]
                },
                max_plus_ones: {
                    bsonType: "int",
                    minimum: 0,
                    maximum: 3
                },
                first_name: {
                    bsonType: "string",
                    maxLength: 50
                },
                last_name: {
                    bsonType: "string",
                    maxLength: 50
                }
            }
        }
    }
})
```

## Collection: media

### Purpose

Stores metadata for uploaded images and files, including storage locations, thumbnails, and processing status.

### Go Struct Definition

```go
package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// MediaType represents the type of media
type MediaType string

const (
    MediaTypeImage    MediaType = "image"
    MediaTypeVideo    MediaType = "video"
    MediaTypeAudio    MediaType = "audio"
    MediaTypeDocument MediaType = "document"
)

// MediaPurpose describes how the media is used
type MediaPurpose string

const (
    MediaPurposeCover    MediaPurpose = "cover"
    MediaPurposeGallery  MediaPurpose = "gallery"
    MediaPurposePartner1 MediaPurpose = "partner1"
    MediaPurposePartner2 MediaPurpose = "partner2"
    MediaPurposeStory    MediaPurpose = "story"
    MediaPurposeGeneral  MediaPurpose = "general"
)

// ProcessingStatus tracks image processing state
type ProcessingStatus string

const (
    ProcessingPending    ProcessingStatus = "pending"
    ProcessingComplete   ProcessingStatus = "complete"
    ProcessingFailed     ProcessingStatus = "failed"
    ProcessingOptimizing ProcessingStatus = "optimizing"
)

// Media represents uploaded files and their metadata
type Media struct {
    ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    WeddingID         primitive.ObjectID `bson:"wedding_id" json:"wedding_id"`
    UploadedBy        primitive.ObjectID `bson:"uploaded_by" json:"uploaded_by"`
    
    // File Information
    OriginalFilename  string             `bson:"original_filename" json:"original_filename"`
    FileSize          int64              `bson:"file_size" json:"file_size"`
    MimeType          string             `bson:"mime_type" json:"mime_type"`
    Extension         string             `bson:"extension" json:"extension"`
    Checksum          string             `bson:"checksum" json:"checksum"` // SHA256 hash
    
    // Storage
    StorageProvider   string             `bson:"storage_provider" json:"storage_provider"` // s3, r2, gcs
    StorageBucket     string             `bson:"storage_bucket" json:"-"` // Never expose
    StorageKey        string             `bson:"storage_key" json:"-"`    // Never expose
    StorageRegion     string             `bson:"storage_region,omitempty" json:"-"`
    
    // URLs (CDN)
    OriginalURL       string             `bson:"original_url" json:"original_url"`
    ThumbnailURL      string             `bson:"thumbnail_url,omitempty" json:"thumbnail_url,omitempty"`
    OptimizedURL      string             `bson:"optimized_url,omitempty" json:"optimized_url,omitempty"`
    WebPURL           string             `bson:"webp_url,omitempty" json:"webp_url,omitempty"`
    
    // Image Properties
    Width             int                `bson:"width,omitempty" json:"width,omitempty"`
    Height            int                `bson:"height,omitempty" json:"height,omitempty"`
    AspectRatio       float64            `bson:"aspect_ratio,omitempty" json:"aspect_ratio,omitempty"`
    Format            string             `bson:"format,omitempty" json:"format,omitempty"` // jpeg, png, webp
    ColorSpace        string             `bson:"color_space,omitempty" json:"color_space,omitempty"`
    DominantColor     string             `bson:"dominant_color,omitempty" json:"dominant_color,omitempty"` // Hex color
    
    // Categorization
    Type              MediaType          `bson:"type" json:"type"`
    Purpose           MediaPurpose       `bson:"purpose" json:"purpose"`
    Category          string             `bson:"category,omitempty" json:"category,omitempty"`
    
    // Gallery/Display Settings
    Caption           string             `bson:"caption,omitempty" json:"caption,omitempty" validate:"omitempty,max=200"`
    AltText           string             `bson:"alt_text,omitempty" json:"alt_text,omitempty" validate:"omitempty,max=200"`
    DisplayOrder      int                `bson:"display_order" json:"display_order"`
    IsFeatured        bool               `bson:"is_featured" json:"is_featured"`
    
    // Processing
    ProcessingStatus  ProcessingStatus   `bson:"processing_status" json:"processing_status"`
    ProcessingError   string             `bson:"processing_error,omitempty" json:"-"`
    Versions          []MediaVersion     `bson:"versions,omitempty" json:"versions,omitempty"`
    
    // EXIF Data (for photos)
    EXIF              *EXIFData          `bson:"exif,omitempty" json:"exif,omitempty"`
    
    // Access Control
    IsPublic          bool               `bson:"is_public" json:"is_public"`
    
    // Metadata
    UploadedAt        time.Time          `bson:"uploaded_at" json:"uploaded_at"`
    UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at"`
    LastAccessedAt    *time.Time         `bson:"last_accessed_at,omitempty" json:"last_accessed_at,omitempty"`
    AccessCount       int64              `bson:"access_count" json:"access_count"`
}

// MediaVersion represents different formats/sizes
type MediaVersion struct {
    Name      string `bson:"name" json:"name"`           // "thumbnail", "optimized", "webp"
    Width     int    `bson:"width" json:"width"`
    Height    int    `bson:"height" json:"height"`
    URL       string `bson:"url" json:"url"`
    FileSize  int64  `bson:"file_size" json:"file_size"`
    Format    string `bson:"format" json:"format"`
    Quality   int    `bson:"quality" json:"quality"`     // 0-100
}

// EXIFData extracted from image metadata
type EXIFData struct {
    CameraMake     string     `bson:"camera_make,omitempty" json:"camera_make,omitempty"`
    CameraModel    string     `bson:"camera_model,omitempty" json:"camera_model,omitempty"`
    DateTaken      *time.Time `bson:"date_taken,omitempty" json:"date_taken,omitempty"`
    FNumber        string     `bson:"f_number,omitempty" json:"f_number,omitempty"`
    ExposureTime   string     `bson:"exposure_time,omitempty" json:"exposure_time,omitempty"`
    ISO            int        `bson:"iso,omitempty" json:"iso,omitempty"`
    FocalLength    string     `bson:"focal_length,omitempty" json:"focal_length,omitempty"`
    GPSLatitude    float64    `bson:"gps_latitude,omitempty" json:"gps_latitude,omitempty"`
    GPSLongitude   float64    `bson:"gps_longitude,omitempty" json:"gps_longitude,omitempty"`
    Orientation    int        `bson:"orientation,omitempty" json:"orientation,omitempty"`
}

// UploadMediaRequest represents a file upload request
type UploadMediaRequest struct {
    File     []byte         `form:"file" binding:"required"`
    Purpose  MediaPurpose   `form:"purpose" binding:"required"`
    Caption  string         `form:"caption,omitempty"`
    Category string         `form:"category,omitempty"`
}

// MediaUploadResult returned after successful upload
type MediaUploadResult struct {
    Media   *Media `json:"media"`
    Success bool   `json:"success"`
    Message string `json:"message,omitempty"`
}
```

### Indexes

```javascript
// Wedding ID index for listing media
db.media.createIndex({ "wedding_id": 1, "display_order": 1 })

// Purpose index for filtering by usage
db.media.createIndex({ "wedding_id": 1, "purpose": 1, "display_order": 1 })

// Checksum index for duplicate detection
db.media.createIndex({ "checksum": 1 })

// Processing status for tracking
db.media.createIndex({ "processing_status": 1, "uploaded_at": -1 })

// Type index
db.media.createIndex({ "wedding_id": 1, "type": 1 })

// Featured index
db.media.createIndex({ "wedding_id": 1, "is_featured": 1 })

// TTL index for cleaning up unprocessed media after 24 hours
db.media.createIndex({ "uploaded_at": 1 }, { 
    expireAfterSeconds: 86400,
    partialFilterExpression: { processing_status: "pending" }
})
```

### Validation Rules

```javascript
db.createCollection("media", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["wedding_id", "uploaded_by", "original_filename", "file_size", "mime_type", "type", "purpose", "storage_provider", "uploaded_at", "updated_at"],
            properties: {
                wedding_id: {
                    bsonType: "objectId"
                },
                file_size: {
                    bsonType: "long",
                    minimum: 0,
                    maximum: 50000000  // 50MB max
                },
                mime_type: {
                    bsonType: "string",
                    pattern: "^(image|video|audio|application)/.*$"
                },
                type: {
                    enum: ["image", "video", "audio", "document"]
                },
                purpose: {
                    enum: ["cover", "gallery", "partner1", "partner2", "story", "general"]
                },
                processing_status: {
                    enum: ["pending", "complete", "failed", "optimizing"]
                },
                width: {
                    bsonType: "int",
                    minimum: 1,
                    maximum: 10000
                },
                height: {
                    bsonType: "int",
                    minimum: 1,
                    maximum: 10000
                },
                is_public: {
                    bsonType: "bool"
                }
            }
        }
    }
})
```

## Collection: analytics

### Purpose

Stores analytics data including page views, RSVP submissions, and user interactions. Uses a time-series pattern for efficient querying of aggregated data.

### Go Struct Definition

```go
package models

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

// EventType represents different analytics events
type EventType string

const (
    EventPageView       EventType = "page_view"
    EventRSVPSubmit     EventType = "rsvp_submit"
    EventGalleryView    EventType = "gallery_view"
    EventImageClick     EventType = "image_click"
    EventLinkClick      EventType = "link_click"
    EventMapOpen        EventType = "map_open"
    EventShareClick     EventType = "share_click"
    EventTimeOnPage     EventType = "time_on_page"
    EventScrollDepth    EventType = "scroll_depth"
    EventError          EventType = "error"
)

// AnalyticsEvent represents a single tracked event
type AnalyticsEvent struct {
    ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    WeddingID       primitive.ObjectID `bson:"wedding_id" json:"wedding_id"`
    
    // Event Type
    Type            EventType          `bson:"type" json:"type"`
    
    // Session/User Info
    SessionID       string             `bson:"session_id" json:"session_id"`       // Anonymous session
    VisitorID       string             `bson:"visitor_id" json:"visitor_id"`       // Hashed visitor identifier
    UserID          *primitive.ObjectID `bson:"user_id,omitempty" json:"user_id,omitempty"` // If logged in
    GuestID         *primitive.ObjectID `bson:"guest_id,omitempty" json:"guest_id,omitempty"` // If linked to guest
    
    // Event Details
    PagePath        string             `bson:"page_path" json:"page_path"`         // e.g., "/", "/gallery"
    PageTitle       string             `bson:"page_title,omitempty" json:"page_title,omitempty"`
    Referrer        string             `bson:"referrer,omitempty" json:"referrer,omitempty"`
    
    // Engagement Metrics
    Duration        int                `bson:"duration,omitempty" json:"duration,omitempty"` // Time in seconds
    ScrollDepth     int                `bson:"scroll_depth,omitempty" json:"scroll_depth,omitempty"` // Percentage 0-100
    
    // Device Info
    UserAgent       string             `bson:"user_agent" json:"-"`                  // Not exposed in API
    DeviceType      string             `bson:"device_type" json:"device_type"`       // desktop, mobile, tablet
    Browser         string             `bson:"browser" json:"browser"`
    OS              string             `bson:"os" json:"os"`
    ScreenResolution string            `bson:"screen_resolution,omitempty" json:"screen_resolution,omitempty"`
    
    // Geographic Info
    IPAddress       string             `bson:"ip_address" json:"-"`                  // Not exposed
    Country         string             `bson:"country,omitempty" json:"country,omitempty"`
    City            string             `bson:"city,omitempty" json:"city,omitempty"`
    Region          string             `bson:"region,omitempty" json:"region,omitempty"`
    
    // Custom Properties
    Metadata        map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
    
    // Timing
    Timestamp       time.Time          `bson:"timestamp" json:"timestamp"`
    Date            string             `bson:"date" json:"date"`                     // YYYY-MM-DD for indexing
    Hour            int                `bson:"hour" json:"hour"`                     // 0-23 for time-based queries
    DayOfWeek       int                `bson:"day_of_week" json:"day_of_week"`       // 0=Sunday
    
    // Processing
    Processed       bool               `bson:"processed" json:"processed"`
    ProcessedAt     *time.Time         `bson:"processed_at,omitempty" json:"processed_at,omitempty"`
}

// DailyAggregatedStats represents pre-aggregated daily metrics
type DailyAggregatedStats struct {
    ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    WeddingID       primitive.ObjectID `bson:"wedding_id" json:"wedding_id"`
    Date            string             `bson:"date" json:"date"`                     // YYYY-MM-DD
    
    // Page Views
    PageViews       int64              `bson:"page_views" json:"page_views"`
    UniqueVisitors  int64              `bson:"unique_visitors" json:"unique_visitors"`
    Sessions        int64              `bson:"sessions" json:"sessions"`
    
    // Device Breakdown
    DeviceDesktop   int64              `bson:"device_desktop" json:"device_desktop"`
    DeviceMobile    int64              `bson:"device_mobile" json:"device_mobile"`
    DeviceTablet    int64              `bson:"device_tablet" json:"device_tablet"`
    
    // Engagement
    AvgTimeOnPage   float64            `bson:"avg_time_on_page" json:"avg_time_on_page"`
    AvgScrollDepth  float64            `bson:"avg_scroll_depth" json:"avg_scroll_depth"`
    BounceRate      float64            `bson:"bounce_rate" json:"bounce_rate"`
    
    // RSVP Stats
    RSVPsSubmitted  int64              `bson:"rsvps_submitted" json:"rsvps_submitted"`
    RSVPsAttending  int64              `bson:"rsvps_attending" json:"rsvps_attending"`
    RSVPsDeclined   int64              `bson:"rsvps_declined" json:"rsvps_declined"`
    
    // Top Pages
    TopPages        []PageStats        `bson:"top_pages" json:"top_pages"`
    
    // Geographic
    TopCountries    []GeoStats         `bson:"top_countries" json:"top_countries"`
    
    // Time-based
    HourlyBreakdown []HourlyStat       `bson:"hourly_breakdown" json:"hourly_breakdown"`
    
    // Referrers
    TopReferrers    []ReferrerStats    `bson:"top_referrers" json:"top_referrers"`
    
    CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
    UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

type PageStats struct {
    PagePath   string `bson:"page_path" json:"page_path"`
    PageViews  int64  `bson:"page_views" json:"page_views"`
    UniqueVisitors int64 `bson:"unique_visitors" json:"unique_visitors"`
}

type GeoStats struct {
    Country string `bson:"country" json:"country"`
    Count   int64  `bson:"count" json:"count"`
}

type HourlyStat struct {
    Hour  int   `bson:"hour" json:"hour"`
    Count int64 `bson:"count" json:"count"`
}

type ReferrerStats struct {
    Referrer string `bson:"referrer" json:"referrer"`
    Count    int64  `bson:"count" json:"count"`
}

// TrackEventRequest for submitting analytics events
type TrackEventRequest struct {
    Type            EventType          `json:"type" binding:"required"`
    PagePath        string             `json:"page_path" binding:"required"`
    SessionID       string             `json:"session_id" binding:"required"`
    Duration        int                `json:"duration,omitempty"`
    ScrollDepth     int                `json:"scroll_depth,omitempty"`
    Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// AnalyticsDashboardResponse for admin dashboard
type AnalyticsDashboardResponse struct {
    TotalPageViews    int64                  `json:"total_page_views"`
    UniqueVisitors    int64                  `json:"unique_visitors"`
    TotalRSVPs        int64                  `json:"total_rsvps"`
    ConversionRate    float64                `json:"conversion_rate"`
    DailyStats        []DailyAggregatedStats `json:"daily_stats"`
    TopPages          []PageStats            `json:"top_pages"`
    DeviceBreakdown   map[string]int64       `json:"device_breakdown"`
    GeographicData    []GeoStats             `json:"geographic_data"`
    HourlyPattern     []HourlyStat           `json:"hourly_pattern"`
    TrendData         []TrendPoint           `json:"trend_data"`
}

type TrendPoint struct {
    Date  string  `json:"date"`
    Value int64   `json:"value"`
}
```

### Indexes

```javascript
// Time-series index for event queries
db.analytics.createIndex({ "wedding_id": 1, "timestamp": -1 })

// Date-based index for aggregation queries
db.analytics.createIndex({ "wedding_id": 1, "date": -1 })

// Event type index for filtering
db.analytics.createIndex({ "wedding_id": 1, "type": 1, "timestamp": -1 })

// Session ID for session-based analytics
db.analytics.createIndex({ "wedding_id": 1, "session_id": 1 })

// Hour index for time-of-day analysis
db.analytics.createIndex({ "wedding_id": 1, "hour": 1 })

// Daily stats unique index
db.analytics_daily.createIndex({ "wedding_id": 1, "date": 1 }, { unique: true })

// TTL index for raw events (keep for 90 days, then aggregate and delete)
db.analytics.createIndex({ "timestamp": 1 }, { expireAfterSeconds: 7776000 })
```

### Validation Rules

```javascript
db.createCollection("analytics", {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["wedding_id", "type", "session_id", "timestamp", "date", "hour", "day_of_week"],
            properties: {
                wedding_id: {
                    bsonType: "objectId"
                },
                type: {
                    enum: ["page_view", "rsvp_submit", "gallery_view", "image_click", "link_click", "map_open", "share_click", "time_on_page", "scroll_depth", "error"]
                },
                session_id: {
                    bsonType: "string",
                    minLength: 10,
                    maxLength: 100
                },
                timestamp: {
                    bsonType: "date"
                },
                date: {
                    bsonType: "string",
                    pattern: "^\\d{4}-\\d{2}-\\d{2}$"
                },
                hour: {
                    bsonType: "int",
                    minimum: 0,
                    maximum: 23
                },
                day_of_week: {
                    bsonType: "int",
                    minimum: 0,
                    maximum: 6
                },
                scroll_depth: {
                    bsonType: "int",
                    minimum: 0,
                    maximum: 100
                },
                processed: {
                    bsonType: "bool"
                }
            }
        }
    }
})
```

## Database Relationships

### Relationship Strategy

MongoDB is schemaless, but we follow these patterns for consistency:

#### Reference Pattern (Preferred for 1:N and N:M)

Use when:
- Data can exist independently
- Need to query from either side
- Data is frequently updated separately
- Avoiding document size limits (16MB)

**Examples:**
- `wedding.user_id` → `users._id`
- `rsvp.wedding_id` → `weddings._id`
- `guest.wedding_id` → `weddings._id`
- `media.wedding_id` → `weddings._id`

#### Embedded Pattern (Use for 1:1 and small 1:N)

Use when:
- Data is always accessed together
- No need to query embedded data independently
- Updates are atomic and infrequent
- Staying within 16MB document limit

**Examples:**
- `wedding.couple` - Always needed with wedding
- `wedding.event` - Core wedding data
- `wedding.theme` - Always loaded with wedding
- `wedding.gallery_images[]` - Small arrays, < 100 items
- `rsvp.plus_ones[]` - Small arrays, < 5 items
- `rsvp.custom_answers[]` - Always needed with RSVP

#### Hybrid Approach

Some collections use both patterns:

```go
// Wedding has embedded summary data AND references

// EMBEDDED (for fast reads):
type Wedding struct {
    // ... other fields
    
    // Denormalized counts (updated via triggers/background jobs)
    RSVPCount      int `bson:"rsvp_count" json:"rsvp_count"`
    GuestCount     int `bson:"guest_count" json:"guest_count"`
    TotalAttending int `bson:"total_attending" json:"total_attending"`
    ViewCount      int64 `bson:"view_count" json:"view_count"`
}

// Full data is in separate collections and joined in application layer
```

### Relationship Summary

```
users (1) ────────< (N) weddings
    │                    │
    │                    ├───< (N) rsvps
    │                    ├───< (N) guests
    │                    ├───< (N) media
    │                    └───< (N) analytics
    │
    └───< (N) wedding_ids (embedded array of references)

weddings (1) ──────< (N) rsvps
weddings (1) ──────< (N) guests
weddings (1) ──────< (N) media
weddings (1) ──────< (N) analytics

guests (0..1) ────── (0..1) rsvps (bidirectional optional reference)
```

## Schema Validation Rules

### MongoDB JSON Schema Validation

All collections use MongoDB's native JSON Schema validation for data integrity:

```javascript
// Validation levels and actions
db.runCommand({
    collMod: "weddings",
    validator: { ... },
    validationLevel: "strict",   // strict | moderate | off
    validationAction: "error"    // error | warn
})
```

### Validation Levels

- **strict**: All inserts and updates must pass validation (default for production)
- **moderate**: Updates to existing invalid documents are allowed
- **off**: No validation performed

### Validation Actions

- **error**: Reject documents that don't pass validation
- **warn**: Log warnings but allow documents

### Common Validation Patterns

```javascript
// Required fields
required: ["field1", "field2"]

// String constraints
{
    bsonType: "string",
    minLength: 2,
    maxLength: 100,
    pattern: "^[a-zA-Z0-9]+$"
}

// Numeric constraints
{
    bsonType: "int",
    minimum: 0,
    maximum: 100
}

// Enum values
{
    enum: ["value1", "value2", "value3"]
}

// Date validation
{
    bsonType: "date"
}

// ObjectId reference
{
    bsonType: "objectId"
}

// Array validation
{
    bsonType: "array",
    minItems: 0,
    maxItems: 100,
    items: {
        bsonType: "object",
        required: ["name"],
        properties: {
            name: { bsonType: "string" }
        }
    }
}
```

### Application-Level Validation (Go)

Use struct tags with `go-playground/validator`:

```go
import "github.com/go-playground/validator/v10"

var validate = validator.New()

func (w *Wedding) Validate() error {
    return validate.Struct(w)
}

// Custom validators
validate.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
    slug := fl.Field().String()
    matched, _ := regexp.MatchString(`^[a-z0-9]+(?:-[a-z0-9]+)*$`, slug)
    return matched
})
```

## Sample Queries

### 1. Find Wedding by Slug

```go
package repositories

import (
    "context"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

func (r *WeddingRepository) FindBySlug(ctx context.Context, slug string) (*models.Wedding, error) {
    var wedding models.Wedding
    
    err := r.collection.FindOne(ctx, bson.M{
        "slug": slug,
        "status": bson.M{"$in": []string{"published", "draft"}},
    }).Decode(&wedding)
    
    if err == mongo.ErrNoDocuments {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    
    return &wedding, nil
}
```

### 2. Get RSVPs for a Wedding

```go
func (r *RSVPRepository) GetByWedding(ctx context.Context, weddingID primitive.ObjectID, filter RSVPFilter) ([]models.RSVP, error) {
    // Build query
    query := bson.M{"wedding_id": weddingID}
    
    // Apply filters
    if filter.Status != "" {
        query["status"] = filter.Status
    }
    if filter.Search != "" {
        query["$or"] = []bson.M{
            {"first_name": bson.M{"$regex": filter.Search, "$options": "i"}},
            {"last_name": bson.M{"$regex": filter.Search, "$options": "i"}},
            {"email": bson.M{"$regex": filter.Search, "$options": "i"}},
        }
    }
    
    // Sort options
    sort := bson.M{"submitted_at": -1}
    if filter.SortBy != "" {
        direction := -1
        if filter.SortOrder == "asc" {
            direction = 1
        }
        sort = bson.M{filter.SortBy: direction}
    }
    
    // Pagination
    skip := int64((filter.Page - 1) * filter.PageSize)
    limit := int64(filter.PageSize)
    
    // Execute query
    cursor, err := r.collection.Find(ctx, query, 
        options.Find().
            SetSort(sort).
            SetSkip(skip).
            SetLimit(limit),
    )
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var rsvps []models.RSVP
    if err := cursor.All(ctx, &rsvps); err != nil {
        return nil, err
    }
    
    return rsvps, nil
}
```

### 3. List User's Weddings

```go
func (r *WeddingRepository) GetByUser(ctx context.Context, userID primitive.ObjectID, status string) ([]models.Wedding, error) {
    query := bson.M{"user_id": userID}
    
    if status != "" && status != "all" {
        query["status"] = status
    }
    
    cursor, err := r.collection.Find(ctx, query,
        options.Find().
            SetSort(bson.M{"created_at": -1}),
    )
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var weddings []models.Wedding
    if err := cursor.All(ctx, &weddings); err != nil {
        return nil, err
    }
    
    return weddings, nil
}

// Alternative: Using the embedded array in users collection
func (r *UserRepository) GetUserWeddings(ctx context.Context, userID primitive.ObjectID) ([]models.Wedding, error) {
    // First get user to get wedding_ids
    var user models.User
    err := r.collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
    if err != nil {
        return nil, err
    }
    
    // Then fetch all weddings
    cursor, err := r.weddingCollection.Find(ctx, bson.M{
        "_id": bson.M{"$in": user.WeddingIDs},
    })
    if err != nil {
        return nil, err
    }
    
    var weddings []models.Wedding
    if err := cursor.All(ctx, &weddings); err != nil {
        return nil, err
    }
    
    return weddings, nil
}
```

### 4. Analytics Aggregation

```go
// Get RSVP statistics for a wedding
func (r *RSVPRepository) GetStatistics(ctx context.Context, weddingID primitive.ObjectID) (*models.RSVPStatistics, error) {
    pipeline := mongo.Pipeline{
        // Match wedding
        {{Key: "$match", Value: bson.M{"wedding_id": weddingID}}},
        
        // Group by status
        {{Key: "$group", Value: bson.M{
            "_id": "$status",
            "count": bson.M{"$sum": 1},
            "totalGuests": bson.M{"$sum": "$attendance_count"},
            "plusOnes": bson.M{"$sum": "$plus_one_count"},
        }}},
        
        // Group all results
        {{Key: "$group", Value: bson.M{
            "_id": nil,
            "statuses": bson.M{"$push": bson.M{
                "k": "$_id",
                "v": bson.M{
                    "count": "$count",
                    "totalGuests": "$totalGuests",
                    "plusOnes": "$plusOnes",
                },
            }},
            "totalResponses": bson.M{"$sum": "$count"},
            "totalGuests": bson.M{"$sum": "$totalGuests"},
            "totalPlusOnes": bson.M{"$sum": "$plusOnes"},
        }}},
    }
    
    cursor, err := r.collection.Aggregate(ctx, pipeline)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var result struct {
        Statuses       []struct {
            K string `bson:"k"`
            V struct {
                Count       int `bson:"count"`
                TotalGuests int `bson:"totalGuests"`
                PlusOnes    int `bson:"plusOnes"`
            } `bson:"v"`
        } `bson:"statuses"`
        TotalResponses int `bson:"totalResponses"`
        TotalGuests    int `bson:"totalGuests"`
        TotalPlusOnes  int `bson:"totalPlusOnes"`
    }
    
    if !cursor.Next(ctx) {
        // Return empty stats
        return &models.RSVPStatistics{}, nil
    }
    
    if err := cursor.Decode(&result); err != nil {
        return nil, err
    }
    
    // Build response
    stats := &models.RSVPStatistics{
        TotalResponses: result.TotalResponses,
        TotalGuests:    result.TotalGuests,
        PlusOnesCount:  result.TotalPlusOnes,
    }
    
    for _, s := range result.Statuses {
        switch s.K {
        case "attending":
            stats.Attending = s.V.Count
        case "not-attending":
            stats.NotAttending = s.V.Count
        case "maybe":
            stats.Maybe = s.V.Count
        }
    }
    
    return stats, nil
}

// Get daily submission trend
func (r *RSVPRepository) GetDailyTrend(ctx context.Context, weddingID primitive.ObjectID, days int) ([]models.DailyCount, error) {
    pipeline := mongo.Pipeline{
        // Match wedding and date range
        {{Key: "$match", Value: bson.M{
            "wedding_id": weddingID,
            "submitted_at": bson.M{
                "$gte": time.Now().AddDate(0, 0, -days),
            },
        }}},
        
        // Group by day
        {{Key: "$group", Value: bson.M{
            "_id": bson.M{
                "$dateToString": bson.M{
                    "format": "%Y-%m-%d",
                    "date":  "$submitted_at",
                },
            },
            "count": bson.M{"$sum": 1},
        }}},
        
        // Sort by date
        {{Key: "$sort", Value: bson.M{"_id": 1}}},
        
        // Project to final format
        {{Key: "$project", Value: bson.M{
            "_id":   0,
            "date":  "$_id",
            "count": "$count",
        }}},
    }
    
    cursor, err := r.collection.Aggregate(ctx, pipeline)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var trend []models.DailyCount
    if err := cursor.All(ctx, &trend); err != nil {
        return nil, err
    }
    
    return trend, nil
}

// Get analytics dashboard data
func (r *AnalyticsRepository) GetDashboard(ctx context.Context, weddingID primitive.ObjectID, startDate, endDate time.Time) (*models.AnalyticsDashboardResponse, error) {
    // Get page views
    pageViews, _ := r.getPageViews(ctx, weddingID, startDate, endDate)
    
    // Get unique visitors
    uniqueVisitors, _ := r.getUniqueVisitors(ctx, weddingID, startDate, endDate)
    
    // Get RSVPs
    rsvpStats, _ := r.rsvpRepo.GetStatistics(ctx, weddingID)
    
    // Get device breakdown
    devices, _ := r.getDeviceBreakdown(ctx, weddingID, startDate, endDate)
    
    // Get daily stats
    dailyStats, _ := r.getDailyStats(ctx, weddingID, startDate, endDate)
    
    // Calculate conversion rate
    var conversionRate float64
    if pageViews > 0 {
        conversionRate = float64(rsvpStats.TotalResponses) / float64(pageViews) * 100
    }
    
    return &models.AnalyticsDashboardResponse{
        TotalPageViews: pageViews,
        UniqueVisitors: uniqueVisitors,
        TotalRSVPs:    int64(rsvpStats.TotalResponses),
        ConversionRate: conversionRate,
        DeviceBreakdown: devices,
        DailyStats:    dailyStats,
    }, nil
}
```

## Data Migration Strategy

### Migration Philosophy

1. **Backward Compatibility**: Always maintain backward compatibility during migrations
2. **Zero Downtime**: Use blue-green deployment patterns
3. **Rollback Plan**: Always have a rollback strategy
4. **Data Integrity**: Validate data before, during, and after migration

### Migration Types

#### 1. Schema Migration (Adding/Modifying Fields)

```go
// migration_001_add_wedding_cover_image.go
package migrations

import (
    "context"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

func Migrate001_AddWeddingCoverImage(ctx context.Context, db *mongo.Database) error {
    collection := db.Collection("weddings")
    
    // Add new field with default value
    _, err := collection.UpdateMany(ctx, 
        bson.M{"cover_image_url": bson.M{"$exists": false}},
        bson.M{"$set": bson.M{"cover_image_url": ""}},
    )
    
    return err
}
```

#### 2. Data Transformation Migration

```go
// migration_002_normalize_guest_names.go
package migrations

import (
    "context"
    "strings"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

func Migrate002_NormalizeGuestNames(ctx context.Context, db *mongo.Database) error {
    collection := db.Collection("guests")
    
    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        return err
    }
    defer cursor.Close(ctx)
    
    for cursor.Next(ctx) {
        var guest models.Guest
        if err := cursor.Decode(&guest); err != nil {
            return err
        }
        
        // Normalize names
        _, err := collection.UpdateOne(ctx,
            bson.M{"_id": guest.ID},
            bson.M{"$set": bson.M{
                "first_name": strings.TrimSpace(strings.Title(strings.ToLower(guest.FirstName))),
                "last_name":  strings.TrimSpace(strings.Title(strings.ToLower(guest.LastName))),
            }},
        )
        if err != nil {
            return err
        }
    }
    
    return cursor.Err()
}
```

#### 3. Collection Split Migration

```go
// migration_003_split_rsvp_plus_ones.go
func Migrate003_SplitRSVPPlusOnes(ctx context.Context, db *mongo.Database) error {
    // Create new collection
    plusOnesCollection := db.Collection("rsvp_plus_ones")
    
    // Migrate data from embedded to separate collection
    rsvpCollection := db.Collection("rsvps")
    
    cursor, err := rsvpCollection.Find(ctx, bson.M{
        "plus_ones": bson.M{"$exists": true, "$ne": bson.A{}},
    })
    if err != nil {
        return err
    }
    defer cursor.Close(ctx)
    
    for cursor.Next(ctx) {
        var rsvp models.RSVP
        if err := cursor.Decode(&rsvp); err != nil {
            return err
        }
        
        // Insert plus ones to new collection
        for _, po := range rsvp.PlusOnes {
            _, err := plusOnesCollection.InsertOne(ctx, bson.M{
                "rsvp_id":    rsvp.ID,
                "wedding_id": rsvp.WeddingID,
                "first_name": po.FirstName,
                "last_name":  po.LastName,
                "dietary":    po.Dietary,
            })
            if err != nil {
                return err
            }
        }
    }
    
    // After verification, drop embedded field
    // _, err = rsvpCollection.UpdateMany(ctx, bson.M{}, bson.M{"$unset": bson.M{"plus_ones": ""}})
    
    return nil
}
```

### Migration Runner

```go
// migrations/runner.go
package migrations

import (
    "context"
    "fmt"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
)

type Migration struct {
    Version int
    Name    string
    Run     func(context.Context, *mongo.Database) error
}

var migrations = []Migration{
    {1, "Add wedding cover image", Migrate001_AddWeddingCoverImage},
    {2, "Normalize guest names", Migrate002_NormalizeGuestNames},
    // Add more migrations here
}

func RunMigrations(ctx context.Context, db *mongo.Database) error {
    // Create migration tracking collection
    migrationCollection := db.Collection("migrations")
    
    for _, m := range migrations {
        // Check if already run
        var existing bson.M
        err := migrationCollection.FindOne(ctx, bson.M{"version": m.Version}).Decode(&existing)
        if err == nil {
            fmt.Printf("Migration %d (%s) already run, skipping\n", m.Version, m.Name)
            continue
        }
        
        // Run migration
        fmt.Printf("Running migration %d: %s\n", m.Version, m.Name)
        if err := m.Run(ctx, db); err != nil {
            return fmt.Errorf("migration %d failed: %w", m.Version, err)
        }
        
        // Record migration
        _, err = migrationCollection.InsertOne(ctx, bson.M{
            "version":   m.Version,
            "name":      m.Name,
            "applied_at": time.Now(),
        })
        if err != nil {
            return err
        }
        
        fmt.Printf("Migration %d completed successfully\n", m.Version)
    }
    
    return nil
}

// Rollback specific migration
func RollbackMigration(ctx context.Context, db *mongo.Database, version int) error {
    // Implementation depends on migration
    // Each migration should have a corresponding rollback function
    return fmt.Errorf("rollback not implemented for version %d", version)
}
```

### Best Practices

1. **Always run migrations in transactions** (if replica set available):

```go
session, err := client.StartSession()
if err != nil {
    return err
}
defer session.EndSession(ctx)

_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
    // Run migrations within transaction
    return nil, RunMigrations(sessCtx, db)
})
```

2. **Test migrations on staging data first**
3. **Keep migrations idempotent** (safe to run multiple times)
4. **Log all migration operations**
5. **Create data backups before major migrations**
6. **Document breaking changes** in API versioning

## MongoDB Connection Configuration

### Connection String Format

```
Standard:
mongodb://username:password@host1:port1,host2:port2/database?options

SRV (DNS seed list):
mongodb+srv://username:password@cluster.mongodb.net/database?options
```

### Recommended Connection Options

```go
package database

import (
    "context"
    "fmt"
    "time"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/mongo/readpref"
)

// Config holds MongoDB configuration
type Config struct {
    URI            string
    Database       string
    MaxPoolSize    uint64
    MinPoolSize    uint64
    MaxConnIdleTime time.Duration
    ConnectTimeout  time.Duration
    SocketTimeout   time.Duration
    ServerSelectionTimeout time.Duration
    HeartbeatInterval time.Duration
}

// DefaultConfig returns production-ready defaults
func DefaultConfig() *Config {
    return &Config{
        URI:                    getEnv("MONGODB_URI", "mongodb://localhost:27017"),
        Database:               getEnv("MONGODB_DATABASE", "wedding_invitations"),
        MaxPoolSize:            100,
        MinPoolSize:            10,
        MaxConnIdleTime:        30 * time.Minute,
        ConnectTimeout:         10 * time.Second,
        SocketTimeout:          5 * time.Second,
        ServerSelectionTimeout: 30 * time.Second,
        HeartbeatInterval:      10 * time.Second,
    }
}

// Connect establishes database connection with all configurations
func Connect(ctx context.Context, cfg *Config) (*mongo.Database, error) {
    // Parse URI
    clientOptions := options.Client().ApplyURI(cfg.URI)
    
    // Connection pool settings
    clientOptions.SetMaxPoolSize(cfg.MaxPoolSize)
    clientOptions.SetMinPoolSize(cfg.MinPoolSize)
    clientOptions.SetMaxConnIdleTime(cfg.MaxConnIdleTime)
    
    // Timeout settings
    clientOptions.SetConnectTimeout(cfg.ConnectTimeout)
    clientOptions.SetSocketTimeout(cfg.SocketTimeout)
    clientOptions.SetServerSelectionTimeout(cfg.ServerSelectionTimeout)
    clientOptions.SetHeartbeatInterval(cfg.HeartbeatInterval)
    
    // Retry writes (for replica sets)
    clientOptions.SetRetryWrites(true)
    
    // Compression
    clientOptions.SetCompressors([]string{"zstd", "snappy", "zlib"})
    
    // Read/Write concerns
    clientOptions.SetReadConcern(readconcern.Majority())
    clientOptions.SetWriteConcern(writeconcern.New(
        writeconcern.WMajority(),
        writeconcern.J(true),
        writeconcern.WTimeout(10*time.Second),
    ))
    
    // Read preference (primary for writes, can use secondaryPreferred for reads)
    clientOptions.SetReadPreference(readpref.Primary())
    
    // Create client
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        return nil, fmt.Errorf("failed to create client: %w", err)
    }
    
    // Verify connection
    ctx, cancel := context.WithTimeout(ctx, cfg.ConnectTimeout)
    defer cancel()
    
    if err := client.Ping(ctx, readpref.Primary()); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    
    db := client.Database(cfg.Database)
    
    // Create indexes
    if err := createIndexes(ctx, db); err != nil {
        return nil, fmt.Errorf("failed to create indexes: %w", err)
    }
    
    return db, nil
}

// createIndexes sets up all required indexes
func createIndexes(ctx context.Context, db *mongo.Database) error {
    // Users indexes
    _, err := db.Collection("users").Indexes().CreateMany(ctx, []mongo.IndexModel{
        {
            Keys:    bson.D{{Key: "email", Value: 1}},
            Options: options.Index().SetUnique(true),
        },
        {Keys: bson.D{{Key: "wedding_ids", Value: 1}}},
        {Keys: bson.D{{Key: "status", Value: 1}, {Key: "created_at", Value: -1}}},
    })
    if err != nil {
        return fmt.Errorf("users indexes: %w", err)
    }
    
    // Weddings indexes
    _, err = db.Collection("weddings").Indexes().CreateMany(ctx, []mongo.IndexModel{
        {
            Keys:    bson.D{{Key: "slug", Value: 1}},
            Options: options.Index().SetUnique(true),
        },
        {Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
        {Keys: bson.D{{Key: "status", Value: 1}}},
        {Keys: bson.D{{Key: "event.date", Value: 1}}},
    })
    if err != nil {
        return fmt.Errorf("weddings indexes: %w", err)
    }
    
    // RSVP indexes
    _, err = db.Collection("rsvps").Indexes().CreateMany(ctx, []mongo.IndexModel{
        {Keys: bson.D{{Key: "wedding_id", Value: 1}, {Key: "submitted_at", Value: -1}}},
        {Keys: bson.D{{Key: "wedding_id", Value: 1}, {Key: "status", Value: 1}}},
        {
            Keys: bson.D{{Key: "wedding_id", Value: 1}, {Key: "email", Value: 1}},
            Options: options.Index().SetUnique(true).SetPartialFilterExpression(
                bson.M{"email": bson.M{"$exists": true, "$ne": ""}},
            ),
        },
    })
    if err != nil {
        return fmt.Errorf("rsvps indexes: %w", err)
    }
    
    // Guests indexes
    _, err = db.Collection("guests").Indexes().CreateMany(ctx, []mongo.IndexModel{
        {Keys: bson.D{{Key: "wedding_id", Value: 1}, {Key: "last_name", Value: 1}, {Key: "first_name", Value: 1}}},
        {Keys: bson.D{{Key: "rsvp_id", Value: 1}}},
        {
            Keys: bson.D{{Key: "wedding_id", Value: 1}, {Key: "email", Value: 1}},
            Options: options.Index().SetUnique(true).SetPartialFilterExpression(
                bson.M{"email": bson.M{"$exists": true, "$ne": ""}},
            ),
        },
    })
    if err != nil {
        return fmt.Errorf("guests indexes: %w", err)
    }
    
    // Media indexes
    _, err = db.Collection("media").Indexes().CreateMany(ctx, []mongo.IndexModel{
        {Keys: bson.D{{Key: "wedding_id", Value: 1}, {Key: "display_order", Value: 1}}},
        {Keys: bson.D{{Key: "checksum", Value: 1}}},
        {Keys: bson.D{{Key: "processing_status", Value: 1}, {Key: "uploaded_at", Value: -1}}},
    })
    if err != nil {
        return fmt.Errorf("media indexes: %w", err)
    }
    
    // Analytics indexes
    _, err = db.Collection("analytics").Indexes().CreateMany(ctx, []mongo.IndexModel{
        {Keys: bson.D{{Key: "wedding_id", Value: 1}, {Key: "timestamp", Value: -1}}},
        {Keys: bson.D{{Key: "wedding_id", Value: 1}, {Key: "date", Value: -1}}},
        {
            Keys:    bson.D{{Key: "timestamp", Value: 1}},
            Options: options.Index().SetExpireAfterSeconds(7776000), // 90 days TTL
        },
    })
    if err != nil {
        return fmt.Errorf("analytics indexes: %w", err)
    }
    
    return nil
}

// Graceful shutdown
func Disconnect(client *mongo.Client, timeout time.Duration) error {
    ctx, cancel := context.WithTimeout(context.Background(), timeout)
    defer cancel()
    return client.Disconnect(ctx)
}
```

### Connection Health Monitoring

```go
package database

import (
    "context"
    "time"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/readpref"
)

type HealthChecker struct {
    client  *mongo.Client
    timeout time.Duration
}

func NewHealthChecker(client *mongo.Client, timeout time.Duration) *HealthChecker {
    return &HealthChecker{client: client, timeout: timeout}
}

func (h *HealthChecker) Check(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, h.timeout)
    defer cancel()
    return h.client.Ping(ctx, readpref.Primary())
}

func (h *HealthChecker) GetStats() mongoDBStats {
    // Get connection pool statistics
    return mongoDBStats{
        Connections: h.client.NumberSessionsInProgress(),
    }
}
```

### Environment-Specific Configurations

```yaml
# config/development.yaml
mongodb:
  uri: "mongodb://localhost:27017"
  database: "wedding_invitations_dev"
  max_pool_size: 10
  min_pool_size: 1

# config/production.yaml
mongodb:
  uri: "mongodb+srv://user:pass@cluster.mongodb.net"
  database: "wedding_invitations"
  max_pool_size: 100
  min_pool_size: 10
  retry_writes: true
  read_preference: "primaryPreferred"
  write_concern: "majority"
```

## Repository Pattern Implementation

### Base Repository

```go
package repositories

import (
    "context"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// BaseRepository provides common CRUD operations
type BaseRepository struct {
    collection *mongo.Collection
}

func NewBaseRepository(collection *mongo.Collection) *BaseRepository {
    return &BaseRepository{collection: collection}
}

// FindByID finds a document by ObjectID
func (r *BaseRepository) FindByID(ctx context.Context, id primitive.ObjectID, result interface{}) error {
    return r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(result)
}

// FindOne finds a single document matching the filter
func (r *BaseRepository) FindOne(ctx context.Context, filter bson.M, result interface{}) error {
    return r.collection.FindOne(ctx, filter).Decode(result)
}

// FindMany finds multiple documents
func (r *BaseRepository) FindMany(ctx context.Context, filter bson.M, opts *options.FindOptions, results interface{}) error {
    cursor, err := r.collection.Find(ctx, filter, opts)
    if err != nil {
        return err
    }
    defer cursor.Close(ctx)
    return cursor.All(ctx, results)
}

// Create inserts a new document
func (r *BaseRepository) Create(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
    return r.collection.InsertOne(ctx, document)
}

// CreateMany inserts multiple documents
func (r *BaseRepository) CreateMany(ctx context.Context, documents []interface{}) (*mongo.InsertManyResult, error) {
    return r.collection.InsertMany(ctx, documents)
}

// Update updates a single document
func (r *BaseRepository) Update(ctx context.Context, id primitive.ObjectID, update bson.M) (*mongo.UpdateResult, error) {
    return r.collection.UpdateOne(ctx, 
        bson.M{"_id": id}, 
        bson.M{"$set": update, "$currentDate": bson.M{"updated_at": true}},
    )
}

// UpdateMany updates multiple documents
func (r *BaseRepository) UpdateMany(ctx context.Context, filter bson.M, update bson.M) (*mongo.UpdateResult, error) {
    return r.collection.UpdateMany(ctx, filter, update)
}

// Delete removes a document
func (r *BaseRepository) Delete(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
    return r.collection.DeleteOne(ctx, bson.M{"_id": id})
}

// DeleteMany removes multiple documents
func (r *BaseRepository) DeleteMany(ctx context.Context, filter bson.M) (*mongo.DeleteResult, error) {
    return r.collection.DeleteMany(ctx, filter)
}

// Count returns the count of documents matching the filter
func (r *BaseRepository) Count(ctx context.Context, filter bson.M) (int64, error) {
    return r.collection.CountDocuments(ctx, filter)
}

// Exists checks if any document matches the filter
func (r *BaseRepository) Exists(ctx context.Context, filter bson.M) (bool, error) {
    count, err := r.collection.CountDocuments(ctx, filter, options.Count().SetLimit(1))
    return count > 0, err
}

// Aggregate performs an aggregation pipeline
func (r *BaseRepository) Aggregate(ctx context.Context, pipeline mongo.Pipeline, results interface{}) error {
    cursor, err := r.collection.Aggregate(ctx, pipeline)
    if err != nil {
        return err
    }
    defer cursor.Close(ctx)
    return cursor.All(ctx, results)
}

// WithTransaction executes operations within a transaction
func (r *BaseRepository) WithTransaction(ctx context.Context, fn func(sessCtx mongo.SessionContext) error) error {
    session, err := r.collection.Database().Client().StartSession()
    if err != nil {
        return err
    }
    defer session.EndSession(ctx)
    
    _, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
        return nil, fn(sessCtx)
    })
    
    return err
}
```

### Concrete Repositories

```go
package repositories

// WeddingRepository handles wedding-related database operations
type WeddingRepository struct {
    *BaseRepository
}

func NewWeddingRepository(db *mongo.Database) *WeddingRepository {
    return &WeddingRepository{
        BaseRepository: NewBaseRepository(db.Collection("weddings")),
    }
}

// FindBySlug finds a wedding by its unique slug
func (r *WeddingRepository) FindBySlug(ctx context.Context, slug string) (*models.Wedding, error) {
    var wedding models.Wedding
    err := r.FindOne(ctx, bson.M{"slug": slug}, &wedding)
    if err == mongo.ErrNoDocuments {
        return nil, nil
    }
    if err != nil {
        return nil, err
    }
    return &wedding, nil
}

// GetByUser retrieves all weddings owned by a user
func (r *WeddingRepository) GetByUser(ctx context.Context, userID primitive.ObjectID) ([]models.Wedding, error) {
    var weddings []models.Wedding
    opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
    err := r.FindMany(ctx, bson.M{"user_id": userID}, opts, &weddings)
    return weddings, err
}

// CheckSlugExists verifies if a slug is already taken
func (r *WeddingRepository) CheckSlugExists(ctx context.Context, slug string) (bool, error) {
    return r.Exists(ctx, bson.M{"slug": slug})
}

// UpdateStats updates denormalized statistics
func (r *WeddingRepository) UpdateStats(ctx context.Context, weddingID primitive.ObjectID, stats WeddingStats) error {
    _, err := r.Update(ctx, weddingID, bson.M{
        "rsvp_count":      stats.RSVPCount,
        "guest_count":     stats.GuestCount,
        "total_attending": stats.TotalAttending,
        "view_count":      stats.ViewCount,
    })
    return err
}

// IncrementViewCount atomically increments the view counter
func (r *WeddingRepository) IncrementViewCount(ctx context.Context, weddingID primitive.ObjectID) error {
    _, err := r.collection.UpdateOne(ctx,
        bson.M{"_id": weddingID},
        bson.M{
            "$inc": bson.M{"view_count": 1},
            "$currentDate": bson.M{"last_viewed_at": true},
        },
    )
    return err
}
```

```go
// RSVPRepository handles RSVP operations
type RSVPRepository struct {
    *BaseRepository
}

func NewRSVPRepository(db *mongo.Database) *RSVPRepository {
    return &RSVPRepository{
        BaseRepository: NewBaseRepository(db.Collection("rsvps")),
    }
}

// GetByWedding retrieves RSVPs for a wedding with filtering and pagination
func (r *RSVPRepository) GetByWedding(ctx context.Context, weddingID primitive.ObjectID, filter RSVPFilter) (*PaginatedResult[models.RSVP], error) {
    query := bson.M{"wedding_id": weddingID}
    
    if filter.Status != "" {
        query["status"] = filter.Status
    }
    
    // Get total count
    total, err := r.Count(ctx, query)
    if err != nil {
        return nil, err
    }
    
    // Calculate pagination
    skip := int64((filter.Page - 1) * filter.PageSize)
    limit := int64(filter.PageSize)
    
    // Fetch results
    var rsvps []models.RSVP
    opts := options.Find().
        SetSort(bson.D{{Key: "submitted_at", Value: -1}}).
        SetSkip(skip).
        SetLimit(limit)
    
    err = r.FindMany(ctx, query, opts, &rsvps)
    if err != nil {
        return nil, err
    }
    
    return &PaginatedResult[models.RSVP]{
        Data:       rsvps,
        Total:      total,
        Page:       filter.Page,
        PageSize:   filter.PageSize,
        TotalPages: (total + int64(filter.PageSize) - 1) / int64(filter.PageSize),
    }, nil
}

// GetStatistics returns RSVP statistics for a wedding
func (r *RSVPRepository) GetStatistics(ctx context.Context, weddingID primitive.ObjectID) (*models.RSVPStatistics, error) {
    pipeline := mongo.Pipeline{
        {{Key: "$match", Value: bson.M{"wedding_id": weddingID}}},
        {{Key: "$group", Value: bson.M{
            "_id": "$status",
            "count": bson.M{"$sum": 1},
            "attendance": bson.M{"$sum": "$attendance_count"},
        }}},
    }
    
    var results []struct {
        Status     string `bson:"_id"`
        Count      int    `bson:"count"`
        Attendance int    `bson:"attendance"`
    }
    
    if err := r.Aggregate(ctx, pipeline, &results); err != nil {
        return nil, err
    }
    
    stats := &models.RSVPStatistics{}
    for _, r := range results {
        stats.TotalResponses += r.Count
        switch r.Status {
        case "attending":
            stats.Attending = r.Count
            stats.TotalGuests += r.Attendance
        case "not-attending":
            stats.NotAttending = r.Count
        case "maybe":
            stats.Maybe = r.Count
        }
    }
    
    return stats, nil
}

// CheckDuplicate verifies if an RSVP already exists for this email
func (r *RSVPRepository) CheckDuplicate(ctx context.Context, weddingID primitive.ObjectID, email string) (bool, error) {
    if email == "" {
        return false, nil
    }
    return r.Exists(ctx, bson.M{
        "wedding_id": weddingID,
        "email":      email,
    })
}
```

### Repository Factory

```go
package repositories

import "go.mongodb.org/mongo-driver/mongo"

// RepositoryFactory creates and manages all repositories
type RepositoryFactory struct {
    db *mongo.Database
}

func NewRepositoryFactory(db *mongo.Database) *RepositoryFactory {
    return &RepositoryFactory{db: db}
}

func (f *RepositoryFactory) Users() *UserRepository {
    return NewUserRepository(f.db)
}

func (f *RepositoryFactory) Weddings() *WeddingRepository {
    return NewWeddingRepository(f.db)
}

func (f *RepositoryFactory) RSVPs() *RSVPRepository {
    return NewRSVPRepository(f.db)
}

func (f *RepositoryFactory) Guests() *GuestRepository {
    return NewGuestRepository(f.db)
}

func (f *RepositoryFactory) Media() *MediaRepository {
    return NewMediaRepository(f.db)
}

func (f *RepositoryFactory) Analytics() *AnalyticsRepository {
    return NewAnalyticsRepository(f.db)
}
```

### Common Patterns

```go
// Pagination support
type PaginatedResult[T any] struct {
    Data       []T   `json:"data"`
    Total      int64 `json:"total"`
    Page       int   `json:"page"`
    PageSize   int   `json:"page_size"`
    TotalPages int64 `json:"total_pages"`
}

// Filter structures
type RSVPFilter struct {
    WeddingID primitive.ObjectID
    Status    string
    Search    string
    SortBy    string
    SortOrder string
    Page      int
    PageSize  int
}

func DefaultRSVPFilter(weddingID primitive.ObjectID) RSVPFilter {
    return RSVPFilter{
        WeddingID: weddingID,
        Page:      1,
        PageSize:  20,
        SortBy:    "submitted_at",
        SortOrder: "desc",
    }
}
```

### Error Handling

```go
package repositories

import (
    "errors"
    "go.mongodb.org/mongo-driver/mongo"
)

var (
    ErrNotFound       = errors.New("document not found")
    ErrDuplicate      = errors.New("duplicate document")
    ErrValidation     = errors.New("validation failed")
    ErrTransaction    = errors.New("transaction failed")
)

func mapMongoError(err error) error {
    if err == mongo.ErrNoDocuments {
        return ErrNotFound
    }
    
    if mongo.IsDuplicateKeyError(err) {
        return ErrDuplicate
    }
    
    return err
}
```

## Summary

This database schema provides:

- **6 collections** with clear separation of concerns
- **Comprehensive Go structs** with proper BSON tags and validation
- **Optimized indexes** for common query patterns
- **Document validation** at the database level
- **Practical examples** for all CRUD operations
- **Migration strategy** for evolving schemas
- **Production-ready** connection configuration
- **Repository pattern** for clean data access layer

The schema is designed for a wedding invitation system with:
- Scalable document storage for varying invitation themes
- Efficient querying for public wedding pages
- Comprehensive RSVP and guest management
- Analytics tracking for insights
- Media handling with optimization
- Security through proper access patterns

---

**Document Version:** 1.0  
**Last Updated:** 2024-01-15  
**MongoDB Version:** 6.0+  
**Go Driver:** mongo-driver v1.13+