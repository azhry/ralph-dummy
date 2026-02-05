# Backend Architecture Overview

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [System Architecture Diagram](#system-architecture-diagram)
3. [Layer-by-Layer Breakdown](#layer-by-layer-breakdown)
4. [Technology Stack](#technology-stack)
5. [Why Clean Architecture](#why-clean-architecture)
6. [Data Flow Examples](#data-flow-examples)
7. [Key Design Decisions](#key-design-decisions)
8. [Scalability Considerations](#scalability-considerations)
9. [Technology Alternatives](#technology-alternatives)

---

## Executive Summary

The Wedding Invitation System is built on **Clean Architecture** principles, following the layered architecture pattern popularized by Robert C. Martin (Uncle Bob). This approach ensures:

- **Separation of Concerns**: Each layer has a distinct responsibility
- **Testability**: Business logic is isolated from external dependencies
- **Maintainability**: Easy to modify without affecting other layers
- **Flexibility**: Swap implementations without changing business logic
- **Independence**: Framework, UI, and database agnostic design

### Core Principles Applied

| Principle | Implementation |
|-----------|---------------|
| **Dependency Inversion** | Inner layers define interfaces; outer layers implement them |
| **Single Responsibility** | Each layer has one reason to change |
| **Open/Closed** | Extensible through new implementations, not modifications |
| **Testability** | Business logic tested without database or HTTP framework |

### Project Scope

The system supports:
- **Multi-tenant wedding management**: Each wedding is isolated
- **Guest RSVP handling**: Custom questions, dietary restrictions, plus-ones
- **Media management**: Photo uploads, galleries, CDN delivery
- **Real-time analytics**: RSVP tracking, page views, exports
- **Email notifications**: Confirmations, reminders, updates

---

## System Architecture Diagram

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT LAYER                                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Web App    │  │  Mobile App  │  │  Admin Panel │  │  Third Party │     │
│  │   (React)    │  │ (React Native│  │   (React)    │  │   APIs       │     │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘     │
└─────────┼─────────────────┼─────────────────┼─────────────────┼────────────┘
          │                 │                 │                 │
          └─────────────────┴────────┬────────┴─────────────────┘
                                     │
                              HTTP/HTTPS
                                     │
┌────────────────────────────────────┴─────────────────────────────────────────┐
│                           API GATEWAY LAYER                                   │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │                         Gin Router (HTTP)                                │ │
│  │  ┌──────────────┬──────────────┬──────────────┬──────────────┐         │ │
│  │  │ Rate Limit   │    CORS      │   Request    │   Request    │         │ │
│  │  │ Middleware   │  Middleware  │   Validation │    Logging   │         │ │
│  │  └──────────────┴──────────────┴──────────────┴──────────────┘         │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
└────────────────────────────────────┬─────────────────────────────────────────┘
                                     │
                              DTO Mapping
                                     │
┌────────────────────────────────────┴─────────────────────────────────────────┐
│                        APPLICATION LAYER                                      │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │                      Use Cases / Handlers                               │ │
│  │                                                                         │ │
│  │  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐            │ │
│  │  │ Wedding Handler│  │ Guest Handler  │  │RSVP Handler    │            │ │
│  │  │                │  │                │  │                │            │ │
│  │  │ • Create       │  │ • Import CSV   │  │ • Submit       │            │ │
│  │  │ • Update       │  │ • Add Guest    │  │ • Update       │            │ │
│  │  │ • Get Details  │  │ • Update Guest │  │ • Get Stats    │            │ │
│  │  │ • Delete       │  │ • Delete Guest │  │ • Export       │            │ │
│  │  └────────────────┘  └────────────────┘  └────────────────┘            │ │
│  │                                                                         │ │
│  │  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐            │ │
│  │  │  Auth Handler  │  │Upload Handler  │  │Analytics       │            │ │
│  │  │                │  │                │  │  Handler       │            │ │
│  │  │ • Register     │  │ • UploadImage  │  │ • Track View   │            │ │
│  │  │ • Login        │  │ • Get URL      │  │ • Get Reports  │            │ │
│  │  │ • Refresh      │  │ • Delete Image │  │ • Export Data  │            │ │
│  │  └────────────────┘  └────────────────┘  └────────────────┘            │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
└────────────────────────────────────┬─────────────────────────────────────────┘
                                     │
                             Domain Contracts
                                     │
┌────────────────────────────────────┴─────────────────────────────────────────┐
│                           DOMAIN LAYER                                        │
│                                                                               │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │                         Entities (Business Objects)                       │ │
│  │                                                                         │ │
│  │   type Wedding struct {              type Guest struct {                │ │
│  │       ID          primitive.ObjectID      ID          primitive.ObjectID  │ │
│  │       CoupleName  string                  WeddingID   primitive.ObjectID  │ │
│  │       Slug        string                  Name        string              │ │
│  │       EventDate   time.Time               Email       string              │ │
│  │       Venue       Venue                   Phone       string              │ │
│  │       Settings    Settings              RSVPStatus   RSVPStatus          │ │
│  │   }                                     Dietary      []string           │ │
│  │   }                                     ...                             │ │
│  │                                                                         │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                               │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │                    Repository Interfaces                                  │ │
│  │                                                                         │ │
│  │   type WeddingRepository interface {                                      │ │
│  │       Create(ctx context.Context, wedding *Wedding) error               │ │
│  │       GetByID(ctx context.Context, id primitive.ObjectID) (*Wedding, error)│ │
│  │       GetBySlug(ctx context.Context, slug string) (*Wedding, error)      │ │
│  │       Update(ctx context.Context, wedding *Wedding) error                │ │
│  │       Delete(ctx context.Context, id primitive.ObjectID) error           │ │
│  │       ListByUser(ctx context.Context, userID primitive.ObjectID) ([]*Wedding, error)│ │
│  │   }                                                                      │ │
│  │                                                                          │ │
│  │   type GuestRepository interface {                                       │ │
│  │       Create(ctx context.Context, guest *Guest) error                    │ │
│  │       GetByID(ctx context.Context, id primitive.ObjectID) (*Guest, error)│ │
│  │       ListByWedding(ctx context.Context, weddingID primitive.ObjectID,   │ │
│  │           filter GuestFilter) ([]*Guest, int64, error)                    │ │
│  │       UpdateRSVP(ctx context.Context, id primitive.ObjectID,               │ │
│  │           rsvp *RSVP) error                                               │ │
│  │       BulkImport(ctx context.Context, weddingID primitive.ObjectID,      │ │
│  │           guests []*Guest) error                                          │ │
│  │       Delete(ctx context.Context, id primitive.ObjectID) error            │ │
│  │   }                                                                      │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                               │
│  ┌─────────────────────────────────────────────────────────────────────────┐ │
│  │                      Business Rules / Domain Services                     │ │
│  │                                                                         │ │
│  │   type WeddingDomainService struct {                                     │ │
│  │       weddingRepo WeddingRepository                                     │ │
│  │       guestRepo   GuestRepository                                       │ │
│  │   }                                                                     │ │
│  │                                                                         │ │
│  │   func (s *WeddingDomainService) CanAddGuest(ctx context.Context,      │ │
│  │       weddingID primitive.ObjectID) error {                             │ │
│  │       // Business rule: Check if wedding allows more guests            │ │
│  │       // Check guest limit, wedding status, etc.                       │ │
│  │   }                                                                     │ │
│  │                                                                         │ │
│  │   func (s *WeddingDomainService) ValidateSlug(slug string) error {      │ │
│  │       // Business rule: Slug format, reserved words, uniqueness        │ │
│  │   }                                                                     │ │
│  └─────────────────────────────────────────────────────────────────────────┘ │
└────────────────────────────────────┬─────────────────────────────────────────┘
                                     │
                          Infrastructure Contracts
                                     │
┌────────────────────────────────────┴─────────────────────────────────────────┐
│                       INFRASTRUCTURE LAYER                                   │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────────┐│
│  │                    Repository Implementations                              ││
│  │                                                                          ││
│  │   type MongoWeddingRepository struct {                                   ││
│  │       collection *mongo.Collection                                      ││
│  │   }                                                                      ││
│  │                                                                          ││
│  │   func (r *MongoWeddingRepository) Create(ctx context.Context,           ││
│  │       wedding *domain.Wedding) error {                                   ││
│  │       _, err := r.collection.InsertOne(ctx, wedding)                   ││
│  │       return err                                                         ││
│  │   }                                                                      ││
│  │                                                                          ││
│  │   func (r *MongoWeddingRepository) GetBySlug(ctx context.Context,       ││
│  │       slug string) (*domain.Wedding, error) {                            ││
│  │       var wedding domain.Wedding                                         ││
│  │       err := r.collection.FindOne(ctx, bson.M{"slug": slug}).Decode(&wedding)││
│  │       return &wedding, err                                               ││
│  │   }                                                                      ││
│  └─────────────────────────────────────────────────────────────────────────┘│
│                                                                              │
│  ┌────────────────────────┐  ┌────────────────────────┐  ┌───────────────┐  │
│  │    MongoDB Database      │  │    File Storage        │  │    Email      │  │
│  │   ┌────────────────┐   │  │   ┌────────────────┐   │  │   Service     │  │
│  │   │  • Weddings    │   │  │   │   AWS S3       │   │  │   ┌────────┐  │  │
│  │   │  • Guests      │   │  │   │   Cloudflare   │   │  │   │SendGrid│  │  │
│  │   │  • RSVPs       │   │  │   │   R2           │   │  │   │AWS SES │  │  │
│  │   │  • Users       │   │  │   │   (CDN)        │   │  │   │Postmark│  │  │
│  │   │  • Analytics   │   │  │   └────────────────┘   │  │   └────────┘  │  │
│  │   └────────────────┘   │  │                        │  │               │  │
│  └────────────────────────┘  └────────────────────────┘  └───────────────┘  │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

### Dependency Flow

```
Outer Layers ────────────────────────────────► Inner Layers
     Depend on                                  Know nothing about outer

Clients ──► Gateway ──► Application ──► Domain ◄─── Infrastructure
    │          │            │             │            │
    │          │            │             │            │
    └──────────┴────────────┴─────────────┴────────────┘
                Dependency Rule: Only point INWARD
```

---

## Layer-by-Layer Breakdown

### 1. Client Layer

The entry points for all external interactions. This layer contains no business logic and is purely responsible for presenting data to users.

**Components:**

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Web Application** | React + Vite | Main wedding invitation interface |
| **Mobile Application** | React Native (optional) | Native mobile experience |
| **Admin Dashboard** | React + Tailwind | Wedding management interface |
| **API Consumers** | Any HTTP client | Third-party integrations |

**Key Characteristics:**
- Stateless: No session storage on client
- JWT-based authentication
- Responsive design for all devices
- Optimistic UI updates

### 2. API Gateway Layer

The HTTP interface that handles all incoming requests. Implements cross-cutting concerns before requests reach application logic.

**Middleware Stack (Gin):**

```go
func SetupRouter(cfg *config.Config) *gin.Engine {
    router := gin.New()
    
    // 1. Recovery - catch panics
    router.Use(gin.Recovery())
    
    // 2. Request ID - tracing
    router.Use(middleware.RequestID())
    
    // 3. Logging - structured logs
    router.Use(middleware.Logger())
    
    // 4. Rate Limiting - prevent abuse
    router.Use(middleware.RateLimiter(cfg.RateLimit))
    
    // 5. CORS - cross-origin requests
    router.Use(middleware.CORS(cfg.AllowedOrigins))
    
    // 6. Security Headers
    router.Use(middleware.SecurityHeaders())
    
    // 7. Request Validation
    router.Use(middleware.RequestValidation())
    
    return router
}
```

**Rate Limiting Strategy:**

| Endpoint Type | Limit | Window | Burst |
|---------------|-------|--------|-------|
| Authentication | 5 requests | 1 minute | 10 |
| API (Authenticated) | 100 requests | 1 minute | 150 |
| API (Anonymous) | 30 requests | 1 minute | 50 |
| File Upload | 10 requests | 5 minutes | 20 |
| RSVP Submission | 20 requests | 1 minute | 30 |

### 3. Application Layer

Contains the use cases of the application. Orchestrates the flow of data between the outer layers and the domain layer.

**Handler Example:**

```go
package handler

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    "wedding-invitation/internal/domain"
    "wedding-invitation/internal/usecase"
)

type WeddingHandler struct {
    weddingUseCase usecase.WeddingUseCase
}

func NewWeddingHandler(uc usecase.WeddingUseCase) *WeddingHandler {
    return &WeddingHandler{weddingUseCase: uc}
}

// CreateWedding godoc
// @Summary Create a new wedding
// @Tags weddings
// @Accept json
// @Produce json
// @Param wedding body dto.CreateWeddingRequest true "Wedding details"
// @Success 201 {object} dto.WeddingResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /weddings [post]
func (h *WeddingHandler) CreateWedding(c *gin.Context) {
    var req dto.CreateWeddingRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err))
        return
    }
    
    // Get authenticated user from context
    userID := c.GetString("user_id")
    
    // Call use case
    wedding, err := h.weddingUseCase.Create(c.Request.Context(), userID, req.ToDomain())
    if err != nil {
        c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
        return
    }
    
    c.JSON(http.StatusCreated, dto.NewWeddingResponse(wedding))
}

// GetWeddingBySlug godoc
// @Summary Get wedding by slug (public)
// @Tags weddings
// @Produce json
// @Param slug path string true "Wedding slug"
// @Success 200 {object} dto.PublicWeddingResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /weddings/{slug} [get]
func (h *WeddingHandler) GetWeddingBySlug(c *gin.Context) {
    slug := c.Param("slug")
    
    wedding, err := h.weddingUseCase.GetBySlug(c.Request.Context(), slug)
    if err != nil {
        if errors.Is(err, domain.ErrWeddingNotFound) {
            c.JSON(http.StatusNotFound, dto.NewErrorResponse(err))
            return
        }
        c.JSON(http.StatusInternalServerError, dto.NewErrorResponse(err))
        return
    }
    
    // Check if password protected
    if wedding.Settings.IsPasswordProtected && !isAuthenticatedForWedding(c, wedding.ID) {
        c.JSON(http.StatusForbidden, dto.ErrorResponse{Message: "Password required"})
        return
    }
    
    c.JSON(http.StatusOK, dto.NewPublicWeddingResponse(wedding))
}
```

**Use Case Example:**

```go
package usecase

type WeddingUseCase interface {
    Create(ctx context.Context, userID string, wedding *domain.Wedding) (*domain.Wedding, error)
    GetByID(ctx context.Context, id string) (*domain.Wedding, error)
    GetBySlug(ctx context.Context, slug string) (*domain.Wedding, error)
    Update(ctx context.Context, id string, updates *domain.Wedding) (*domain.Wedding, error)
    Delete(ctx context.Context, id string) error
    ListByUser(ctx context.Context, userID string, pagination Pagination) ([]*domain.Wedding, int64, error)
}

type weddingUseCase struct {
    weddingRepo    domain.WeddingRepository
    guestRepo      domain.GuestRepository
    slugService    domain.SlugService
    fileService    domain.FileService
    emailService   domain.EmailService
    cacheService   domain.CacheService
}

func (uc *weddingUseCase) Create(ctx context.Context, userID string, wedding *domain.Wedding) (*domain.Wedding, error) {
    // 1. Validate business rules
    if err := domain.ValidateWedding(wedding); err != nil {
        return nil, err
    }
    
    // 2. Generate unique slug
    slug, err := uc.slugService.GenerateUnique(ctx, wedding.CoupleName)
    if err != nil {
        return nil, err
    }
    wedding.Slug = slug
    
    // 3. Set ownership
    wedding.UserID = userID
    wedding.CreatedAt = time.Now()
    wedding.UpdatedAt = time.Now()
    
    // 4. Create in database
    if err := uc.weddingRepo.Create(ctx, wedding); err != nil {
        return nil, err
    }
    
    // 5. Send confirmation email
    uc.emailService.SendWeddingCreatedEmail(ctx, wedding)
    
    return wedding, nil
}
```

### 4. Domain Layer

The heart of the application. Contains business logic, entities, and repository interfaces. This layer has NO dependencies on external frameworks.

**Entity Example:**

```go
package domain

import (
    "time"
    "errors"
    "regexp"
    
    "go.mongodb.org/mongo-driver/bson/primitive"
)

var (
    ErrWeddingNotFound     = errors.New("wedding not found")
    ErrInvalidSlug         = errors.New("invalid slug format")
    ErrWeddingExpired      = errors.New("wedding has already occurred")
    ErrGuestLimitReached   = errors.New("guest limit reached")
    ErrInvalidWeddingDate  = errors.New("wedding date must be in the future")
    ErrReservedSlug        = errors.New("slug is reserved")
)

// Wedding represents a wedding invitation
type Wedding struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    UserID      primitive.ObjectID `bson:"user_id" json:"userId"`
    CoupleName  string             `bson:"couple_name" json:"coupleName"`
    Slug        string             `bson:"slug" json:"slug"`
    
    // Event Details
    EventDate   time.Time          `bson:"event_date" json:"eventDate"`
    Venue       Venue              `bson:"venue" json:"venue"`
    DressCode   string             `bson:"dress_code,omitempty" json:"dressCode,omitempty"`
    
    // Content
    Story       string             `bson:"story,omitempty" json:"story,omitempty"`
    Gallery     []Image            `bson:"gallery,omitempty" json:"gallery,omitempty"`
    CoverImage  Image              `bson:"cover_image,omitempty" json:"coverImage,omitempty"`
    
    // Settings
    Settings    WeddingSettings    `bson:"settings" json:"settings"`
    
    // RSVP Configuration
    RSVPConfig  RSVPConfiguration  `bson:"rsvp_config" json:"rsvpConfig"`
    
    // Timestamps
    CreatedAt   time.Time          `bson:"created_at" json:"createdAt"`
    UpdatedAt   time.Time          `bson:"updated_at" json:"updatedAt"`
}

// Venue represents wedding location details
type Venue struct {
    Name        string  `bson:"name" json:"name"`
    Address     string  `bson:"address" json:"address"`
    City        string  `bson:"city" json:"city"`
    State       string  `bson:"state" json:"state"`
    Country     string  `bson:"country" json:"country"`
    PostalCode  string  `bson:"postal_code" json:"postalCode"`
    Latitude    float64 `bson:"latitude,omitempty" json:"latitude,omitempty"`
    Longitude   float64 `bson:"longitude,omitempty" json:"longitude,omitempty"`
    MapsURL     string  `bson:"maps_url,omitempty" json:"mapsUrl,omitempty"`
}

// WeddingSettings contains wedding-level configuration
type WeddingSettings struct {
    IsPublic               bool   `bson:"is_public" json:"isPublic"`
    IsPasswordProtected    bool   `bson:"is_password_protected" json:"isPasswordProtected"`
    PasswordHash           string `bson:"password_hash,omitempty" json:"-"`
    AllowRSVP              bool   `bson:"allow_rsvp" json:"allowRsvp"`
    RSVPDeadline           *time.Time `bson:"rsvp_deadline,omitempty" json:"rsvpDeadline,omitempty"`
    GuestLimit             int    `bson:"guest_limit" json:"guestLimit"`
    AllowCustomQuestions   bool   `bson:"allow_custom_questions" json:"allowCustomQuestions"`
}

// RSVPConfiguration defines RSVP form fields
type RSVPConfiguration struct {
    CollectEmail        bool     `bson:"collect_email" json:"collectEmail"`
    CollectPhone        bool     `bson:"collect_phone" json:"collectPhone"`
    CollectDietary      bool     `bson:"collect_dietary" json:"collectDietary"`
    AllowPlusOne        bool     `bson:"allow_plus_one" json:"allowPlusOne"`
    PlusOneLimit        int      `bson:"plus_one_limit" json:"plusOneLimit"`
    CustomQuestions     []CustomQuestion `bson:"custom_questions" json:"customQuestions"`
}

// CustomQuestion for RSVP forms
type CustomQuestion struct {
    ID       string   `bson:"id" json:"id"`
    Question string   `bson:"question" json:"question"`
    Type     string   `bson:"type" json:"type"` // text, select, multiselect
    Required bool     `bson:"required" json:"required"`
    Options  []string `bson:"options,omitempty" json:"options,omitempty"`
}

// Business Logic Methods

func (w *Wedding) IsUpcoming() bool {
    return w.EventDate.After(time.Now())
}

func (w *Wedding) IsRSVPOpen() bool {
    if !w.Settings.AllowRSVP {
        return false
    }
    
    if w.Settings.RSVPDeadline != nil {
        return time.Now().Before(*w.Settings.RSVPDeadline)
    }
    
    return w.IsUpcoming()
}

func (w *Wedding) CanAddGuest(currentCount int) bool {
    if w.Settings.GuestLimit <= 0 {
        return true // No limit
    }
    return currentCount < w.Settings.GuestLimit
}

func (w *Wedding) VerifyPassword(password string) bool {
    if !w.Settings.IsPasswordProtected {
        return true
    }
    // Use bcrypt comparison
    return bcrypt.CompareHashAndPassword(
        []byte(w.Settings.PasswordHash), 
        []byte(password),
    ) == nil
}

// Validation Functions

func ValidateWedding(w *Wedding) error {
    if w.CoupleName == "" {
        return errors.New("couple name is required")
    }
    
    if w.EventDate.IsZero() {
        return ErrInvalidWeddingDate
    }
    
    if w.EventDate.Before(time.Now().AddDate(0, 0, -1)) {
        return ErrInvalidWeddingDate
    }
    
    if err := ValidateSlug(w.Slug); err != nil {
        return err
    }
    
    return nil
}

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

var reservedSlugs = map[string]bool{
    "admin": true, "api": true, "auth": true, "login": true,
    "logout": true, "register": true, "static": true, "public": true,
    "about": true, "contact": true, "help": true, "terms": true,
    "privacy": true, "wedding": true, "weddings": true,
}

func ValidateSlug(slug string) error {
    if slug == "" {
        return ErrInvalidSlug
    }
    
    if len(slug) < 3 || len(slug) > 50 {
        return ErrInvalidSlug
    }
    
    if !slugRegex.MatchString(slug) {
        return ErrInvalidSlug
    }
    
    if reservedSlugs[slug] {
        return ErrReservedSlug
    }
    
    return nil
}
```

**Repository Interface:**

```go
package domain

import "context"

// WeddingRepository defines the interface for wedding data access
// This is defined in the domain layer, implemented in infrastructure
type WeddingRepository interface {
    // Create inserts a new wedding
    Create(ctx context.Context, wedding *Wedding) error
    
    // GetByID retrieves a wedding by its ObjectID
    GetByID(ctx context.Context, id primitive.ObjectID) (*Wedding, error)
    
    // GetBySlug retrieves a wedding by its URL slug (public access)
    GetBySlug(ctx context.Context, slug string) (*Wedding, error)
    
    // GetByUserID retrieves all weddings for a specific user
    ListByUser(ctx context.Context, userID primitive.ObjectID, opts ListOptions) ([]*Wedding, int64, error)
    
    // Update modifies an existing wedding
    Update(ctx context.Context, wedding *Wedding) error
    
    // Delete removes a wedding by ID
    Delete(ctx context.Context, id primitive.ObjectID) error
    
    // Exists checks if a slug is already taken
    ExistsBySlug(ctx context.Context, slug string) (bool, error)
    
    // IncrementPageView atomically increments the view counter
    IncrementPageView(ctx context.Context, id primitive.ObjectID) error
}

// ListOptions for pagination and filtering
type ListOptions struct {
    Page     int
    PageSize int
    SortBy   string
    SortDesc bool
    Filters  map[string]interface{}
}

// GuestRepository defines the interface for guest data access
type GuestRepository interface {
    Create(ctx context.Context, guest *Guest) error
    CreateMany(ctx context.Context, guests []*Guest) error
    GetByID(ctx context.Context, id primitive.ObjectID) (*Guest, error)
    GetByEmail(ctx context.Context, weddingID primitive.ObjectID, email string) (*Guest, error)
    ListByWedding(ctx context.Context, weddingID primitive.ObjectID, opts ListOptions) ([]*Guest, int64, error)
    Update(ctx context.Context, guest *Guest) error
    UpdateRSVP(ctx context.Context, id primitive.ObjectID, rsvp *RSVP) error
    Delete(ctx context.Context, id primitive.ObjectID) error
    GetRSVPStats(ctx context.Context, weddingID primitive.ObjectID) (*RSVPStats, error)
}

// RSVPStats represents RSVP statistics
type RSVPStats struct {
    TotalGuests      int64
    Attending        int64
    NotAttending     int64
    Maybe            int64
    NotResponded     int64
    PlusOnes         int64
}
```

### 5. Infrastructure Layer

Contains concrete implementations of interfaces defined in the domain layer. This is where external dependencies live.

**MongoDB Repository Implementation:**

```go
package repository

import (
    "context"
    "time"
    
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    
    "wedding-invitation/internal/domain"
)

type mongoWeddingRepository struct {
    collection *mongo.Collection
}

func NewMongoWeddingRepository(db *mongo.Database) domain.WeddingRepository {
    return &mongoWeddingRepository{
        collection: db.Collection("weddings"),
    }
}

func (r *mongoWeddingRepository) Create(ctx context.Context, wedding *domain.Wedding) error {
    wedding.ID = primitive.NewObjectID()
    wedding.CreatedAt = time.Now()
    wedding.UpdatedAt = time.Now()
    
    _, err := r.collection.InsertOne(ctx, wedding)
    return err
}

func (r *mongoWeddingRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Wedding, error) {
    var wedding domain.Wedding
    err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&wedding)
    if err == mongo.ErrNoDocuments {
        return nil, domain.ErrWeddingNotFound
    }
    return &wedding, err
}

func (r *mongoWeddingRepository) GetBySlug(ctx context.Context, slug string) (*domain.Wedding, error) {
    var wedding domain.Wedding
    err := r.collection.FindOne(ctx, bson.M{"slug": slug}).Decode(&wedding)
    if err == mongo.ErrNoDocuments {
        return nil, domain.ErrWeddingNotFound
    }
    return &wedding, err
}

func (r *mongoWeddingRepository) ListByUser(ctx context.Context, userID primitive.ObjectID, opts domain.ListOptions) ([]*domain.Wedding, int64, error) {
    // Build filter
    filter := bson.M{"user_id": userID}
    
    // Apply additional filters
    for key, value := range opts.Filters {
        filter[key] = value
    }
    
    // Count total
    total, err := r.collection.CountDocuments(ctx, filter)
    if err != nil {
        return nil, 0, err
    }
    
    // Build options
    findOpts := options.Find()
    
    // Pagination
    if opts.PageSize > 0 {
        findOpts.SetLimit(int64(opts.PageSize))
        findOpts.SetSkip(int64((opts.Page - 1) * opts.PageSize))
    }
    
    // Sorting
    sortOrder := 1
    if opts.SortDesc {
        sortOrder = -1
    }
    findOpts.SetSort(bson.D{{Key: opts.SortBy, Value: sortOrder}})
    
    // Execute query
    cursor, err := r.collection.Find(ctx, filter, findOpts)
    if err != nil {
        return nil, 0, err
    }
    defer cursor.Close(ctx)
    
    var weddings []*domain.Wedding
    if err := cursor.All(ctx, &weddings); err != nil {
        return nil, 0, err
    }
    
    return weddings, total, nil
}

func (r *mongoWeddingRepository) Update(ctx context.Context, wedding *domain.Wedding) error {
    wedding.UpdatedAt = time.Now()
    
    filter := bson.M{"_id": wedding.ID}
    update := bson.M{"$set": wedding}
    
    result, err := r.collection.UpdateOne(ctx, filter, update)
    if err != nil {
        return err
    }
    
    if result.MatchedCount == 0 {
        return domain.ErrWeddingNotFound
    }
    
    return nil
}

func (r *mongoWeddingRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
    result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
    if err != nil {
        return err
    }
    
    if result.DeletedCount == 0 {
        return domain.ErrWeddingNotFound
    }
    
    return nil
}

func (r *mongoWeddingRepository) ExistsBySlug(ctx context.Context, slug string) (bool, error) {
    count, err := r.collection.CountDocuments(ctx, bson.M{"slug": slug})
    return count > 0, err
}

func (r *mongoWeddingRepository) IncrementPageView(ctx context.Context, id primitive.ObjectID) error {
    filter := bson.M{"_id": id}
    update := bson.M{"$inc": bson.M{"analytics.page_views": 1}}
    
    _, err := r.collection.UpdateOne(ctx, filter, update)
    return err
}
```

**File Storage Service (AWS S3):**

```go
package storage

import (
    "context"
    "fmt"
    "time"
    
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/s3"
    "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3FileService struct {
    client     *s3.Client
    bucket     string
    region     string
    cdnDomain  string
}

func NewS3FileService(cfg aws.Config, bucket, cdnDomain string) *S3FileService {
    return &S3FileService{
        client:    s3.NewFromConfig(cfg),
        bucket:    bucket,
        region:    cfg.Region,
        cdnDomain: cdnDomain,
    }
}

func (s *S3FileService) Upload(ctx context.Context, file FileUpload) (*UploadResult, error) {
    // Generate unique key
    key := fmt.Sprintf("weddings/%s/%s/%s", 
        file.WeddingID, 
        file.Type, 
        generateUniqueFilename(file.Filename),
    )
    
    // Upload to S3
    _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
        Bucket:      aws.String(s.bucket),
        Key:         aws.String(key),
        Body:        file.Reader,
        ContentType: aws.String(file.ContentType),
        Metadata: map[string]string{
            "wedding-id": file.WeddingID,
            "uploaded-by": file.UploadedBy,
        },
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to upload: %w", err)
    }
    
    // Generate URLs
    var url string
    if s.cdnDomain != "" {
        url = fmt.Sprintf("https://%s/%s", s.cdnDomain, key)
    } else {
        url = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key)
    }
    
    return &UploadResult{
        URL:      url,
        Key:      key,
        Size:     file.Size,
        MimeType: file.ContentType,
    }, nil
}

func (s *S3FileService) GetSignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
    presignClient := s3.NewPresignClient(s.client)
    
    req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(key),
    }, s3.WithPresignExpires(expiry))
    
    if err != nil {
        return "", err
    }
    
    return req.URL, nil
}

func (s *S3FileService) Delete(ctx context.Context, key string) error {
    _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(key),
    })
    return err
}
```

**Email Service:**

```go
package email

import (
    "bytes"
    "context"
    "html/template"
    
    "github.com/sendgrid/sendgrid-go"
    "github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridService struct {
    client    *sendgrid.Client
    fromEmail string
    templates map[string]string // Template IDs
}

func NewSendGridService(apiKey, fromEmail string) *SendGridService {
    return &SendGridService{
        client:    sendgrid.NewSendClient(apiKey),
        fromEmail: fromEmail,
        templates: map[string]string{
            "rsvp_confirmation": "d-xxxxx",
            "wedding_created":   "d-yyyyy",
            "guest_invitation":  "d-zzzzz",
        },
    }
}

func (s *SendGridService) SendRSVPConfirmation(ctx context.Context, guest *domain.Guest, wedding *domain.Wedding) error {
    from := mail.NewEmail("Wedding Invitation", s.fromEmail)
    to := mail.NewEmail(guest.Name, guest.Email)
    
    // Personalization data
    data := map[string]interface{}{
        "guest_name":    guest.Name,
        "couple_name":   wedding.CoupleName,
        "wedding_date":  wedding.EventDate.Format("January 2, 2006"),
        "venue_name":    wedding.Venue.Name,
        "rsvp_status":   guest.RSVP.Status,
        "wedding_url":   fmt.Sprintf("https://wedding.app/w/%s", wedding.Slug),
    }
    
    message := mail.NewV3Mail()
    message.SetTemplateID(s.templates["rsvp_confirmation"])
    message.SetFrom(from)
    
    personalization := mail.NewPersonalization()
    personalization.AddTos(to)
    for key, value := range data {
        personalization.SetDynamicTemplateData(key, value)
    }
    message.AddPersonalizations(personalization)
    
    _, err := s.client.SendWithContext(ctx, message)
    return err
}
```

---

## Technology Stack

### Complete Technology Stack

| Category | Technology | Version | Justification |
|----------|-----------|---------|---------------|
| **Language** | Go | 1.21+ | Native concurrency, fast compilation, single binary deployment, excellent standard library |
| **Web Framework** | Gin | v1.9+ | High performance, middleware support, easy routing, great documentation |
| **Database** | MongoDB | 6.0+ | Flexible schema for varying wedding structures, horizontal scaling, document-based matches domain model |
| **ODM/Driver** | mongo-driver | v1.13+ | Official driver, full MongoDB feature support, connection pooling |
| **Authentication** | JWT (golang-jwt) | v5+ | Stateless auth, widely adopted, RS256 for asymmetric keys |
| **Password Hashing** | bcrypt | - | Industry standard, adaptive cost factor, Go standard implementation |
| **Validation** | go-playground/validator | v10+ | Struct tag validation, custom validators, multilingual error messages |
| **File Storage** | AWS S3 / R2 | - | Reliable object storage, CDN integration, presigned URLs for security |
| **Email** | SendGrid / SES | - | High deliverability, templates, analytics, webhook support |
| **Cache** | Redis (optional) | 7.0+ | Session storage, rate limiting, hot data caching |
| **Logging** | Zap | - | High performance structured logging, multiple output formats |
| **Documentation** | Swaggo | v1.16+ | Auto-generated Swagger from Go comments, keeps docs in sync |
| **Testing** | Testify | - | Assertion helpers, mocking, test suites |
| **Container** | Docker | 24+ | Consistent environments, easy deployment, isolation |
| **Orchestration** | Docker Compose | - | Local development stack, service dependencies |
| **CI/CD** | GitHub Actions | - | Automated testing, building, deployment |
| **Monitoring** | Prometheus + Grafana | - | Metrics collection, visualization, alerting |

### Why This Stack?

#### Go vs Alternatives

| Aspect | Go | Node.js | Python | Java |
|--------|-----|---------|--------|-------|
| **Performance** | Excellent | Good | Moderate | Good |
| **Concurrency** | Goroutines (lightweight) | Event loop | Threads (GIL) | Threads |
| **Startup Time** | Fast | Fast | Slow | Slow |
| **Binary Size** | Single binary | Requires runtime | Requires runtime | Requires runtime |
| **Memory Usage** | Low | Moderate | High | High |
| **Learning Curve** | Low | Low | Low | High |
| **Wedding App Suit** | Best | Good | Good | Overkill |

**Verdict:** Go's combination of performance, simplicity, and deployment ease makes it ideal for a wedding invitation system that may need to scale and must be cost-effective to run.

#### MongoDB vs SQL

| Aspect | MongoDB | PostgreSQL | MySQL |
|--------|---------|-----------|-------|
| **Schema Flexibility** | High | Moderate | Low |
| **Horizontal Scaling** | Native | Complex | Complex |
| **Wedding Data Fit** | Excellent (documents) | Good | Good |
| **RSVP Variability** | Easy to handle | Requires migrations | Requires migrations |
| **Complex Queries** | Good | Excellent | Good |
| **JSON Support** | Native | Good (JSONB) | Moderate |

**Verdict:** MongoDB's document model naturally fits the varying structures of wedding data (different fields for different weddings) and allows for easy schema evolution without migrations.

#### Gin vs Other Frameworks

| Framework | Performance | Features | Learning Curve | Community |
|-----------|-------------|----------|----------------|-----------|
| **Gin** | Excellent | Rich | Low | Very Large |
| **Echo** | Excellent | Rich | Low | Large |
| **Fiber** | Excellent (fastest) | Moderate | Low | Growing |
| **Standard Library** | Good | Minimal | Medium | N/A |
| **Gorilla Mux** | Good | Moderate | Low | Stable |

**Verdict:** Gin strikes the best balance between performance, features, and community support. Its middleware system makes Clean Architecture implementation straightforward.

---

## Why Clean Architecture

### Benefits for Wedding Invitation System

#### 1. **Business Logic Protection**

Wedding domain rules (RSVP deadlines, guest limits, password protection) are isolated and testable without HTTP framework or database.

```go
// Can be tested without any infrastructure
func TestWedding_CanAddGuest(t *testing.T) {
    wedding := &domain.Wedding{
        Settings: domain.WeddingSettings{
            GuestLimit: 100,
        },
    }
    
    // Test pure business logic
    assert.True(t, wedding.CanAddGuest(50))  // Under limit
    assert.False(t, wedding.CanAddGuest(100)) // At limit
}
```

#### 2. **Framework Independence**

Can swap Gin for Fiber or standard library without touching business logic.

```go
// Handler depends on use case interface, not implementation
type WeddingHandler struct {
    useCase usecase.WeddingUseCase  // Interface, not concrete type
}

// Easy to test with mock
mockUseCase := &mocks.MockWeddingUseCase{}
handler := handler.NewWeddingHandler(mockUseCase)
```

#### 3. **Database Flexibility**

Can switch from MongoDB to PostgreSQL without changing application code.

```go
// Domain layer defines interface
type WeddingRepository interface {
    Create(ctx context.Context, wedding *Wedding) error
    // ...
}

// Infrastructure implements
// - MongoWeddingRepository (MongoDB)
// - PostgresWeddingRepository (PostgreSQL)
// - MemoryWeddingRepository (testing)
```

#### 4. **Testability**

| Test Type | Clean Architecture | Traditional MVC |
|-----------|-------------------|---------------|
| **Unit Tests** | Fast, no DB/HTTP needed | Often requires DB |
| **Integration Tests** | Test single layer | Complex setup |
| **E2E Tests** | Minimal infrastructure | Full stack |
| **Test Coverage** | Easy to achieve 80%+ | Harder to isolate |

#### 5. **Team Scalability**

Different team members can work on different layers simultaneously:
- **Frontend dev**: Works with API contracts (DTOs)
- **Backend dev**: Implements use cases
- **DevOps**: Swaps infrastructure implementations

### Trade-offs

| Aspect | Clean Architecture | Simple MVC |
|--------|-------------------|------------|
| **Initial Complexity** | Higher | Lower |
| **Boilerplate** | More interfaces | Less code |
| **Learning Curve** | Steeper | Gentle |
| **Long-term Maintenance** | Excellent | Degrades |
| **Refactoring Cost** | Low | High |

**Verdict:** For a wedding system that will evolve with features (analytics, payments, integrations), the upfront investment in Clean Architecture pays off in maintainability.

---

## Data Flow Examples

### Example 1: Creating a Wedding (Authenticated)

```
┌─────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ Client  │────▶│   Gateway   │────▶│  Handler    │────▶│  Use Case   │────▶│   Domain    │
└─────────┘     └─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
      │               │                   │                   │                   │
      │  POST /api/   │                   │                   │                   │
      │  weddings     │                   │                   │                   │
      │  JWT Token    │                   │                   │                   │
      │──────────────▶│                   │                   │                   │
      │               │                   │                   │                   │
      │               │  1. Auth MW       │                   │                   │
      │               │     Validate JWT  │                   │                   │
      │               │  2. Rate Limit    │                   │                   │
      │               │  3. Validate      │                   │                   │
      │               │     JSON body     │                   │                   │
      │               │──────────────────▶│                   │                   │
      │               │                   │                   │                   │
      │               │                   │  1. Bind DTO      │                   │
      │               │                   │  2. Get userID    │                   │
      │               │                   │     from context  │                   │
      │               │                   │  3. Call use case │                   │
      │               │                   │──────────────────▶│                   │
      │               │                   │                   │                   │
      │               │                   │                   │  1. Validate slug │
      │               │                   │                   │  2. Check unique  │
      │               │                   │                   │  3. Set defaults  │
      │               │                   │                   │──────────────────▶│
      │               │                   │                   │                   │
      │               │                   │                   │  Return wedding   │
      │               │                   │                   │◀──────────────────│
      │               │                   │                   │                   │
      │               │                   │  Return wedding   │                   │
      │               │                   │◀──────────────────│                   │
      │               │                   │                   │                   │
      │               │  Map to DTO       │                   │                   │
      │               │  JSON response    │                   │                   │
      │◀──────────────│                   │                   │                   │
      │               │                   │                   │                   │
      │               │                   │                   │                   │
      │               │     ┌─────────────┐                   │                   │
      │               │     │ Repository  │◀──────────────────│                   │
      │               │     │  Interface  │                   │                   │
      │               │     └──────┬──────┘                   │                   │
      │               │            │                          │                   │
      │               │            ▼                          │                   │
      │               │     ┌─────────────┐                   │                   │
      │               │     │ MongoDB     │                   │                   │
      │               │     │ Repository  │                   │                   │
      │               │     │  (infra)    │                   │                   │
      │               │     └─────────────┘                   │                   │
```

**Sequence Flow:**

```go
// 1. Client sends request
POST /api/weddings
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
    "coupleName": "Alice & Bob",
    "eventDate": "2024-12-25T18:00:00Z",
    "venue": {
        "name": "Grand Plaza Hotel",
        "address": "123 Main St",
        "city": "New York"
    }
}

// 2. Gateway middleware processes request
- AuthMiddleware: Validate JWT, extract userID
- RateLimitMiddleware: Check limit (100 req/min for authenticated users)
- ValidationMiddleware: Validate JSON schema

// 3. Handler receives validated request
func (h *WeddingHandler) CreateWedding(c *gin.Context) {
    var req dto.CreateWeddingRequest
    c.ShouldBindJSON(&req) // Already validated
    
    userID := c.GetString("user_id") // From JWT
    
    // Convert DTO to Domain
    wedding := req.ToDomain()
    
    // Call use case
    result, err := h.weddingUseCase.Create(c.Request.Context(), userID, wedding)
    // ...
}

// 4. Use case implements business logic
func (uc *weddingUseCase) Create(ctx context.Context, userID string, wedding *domain.Wedding) (*domain.Wedding, error) {
    // Business validation
    if err := domain.ValidateWedding(wedding); err != nil {
        return nil, err
    }
    
    // Generate unique slug
    slug, err := uc.slugService.GenerateUnique(ctx, wedding.CoupleName)
    if err != nil {
        return nil, err
    }
    wedding.Slug = slug
    
    // Set ownership
    wedding.UserID = mustParseObjectID(userID)
    
    // Persist
    if err := uc.weddingRepo.Create(ctx, wedding); err != nil {
        return nil, err
    }
    
    // Send notification (async)
    uc.emailService.SendWeddingCreatedEmail(ctx, wedding)
    
    return wedding, nil
}

// 5. Repository persists to MongoDB
func (r *mongoWeddingRepository) Create(ctx context.Context, wedding *domain.Wedding) error {
    wedding.ID = primitive.NewObjectID()
    wedding.CreatedAt = time.Now()
    
    _, err := r.collection.InsertOne(ctx, wedding)
    return err
}

// 6. Response returns to client
201 Created
Content-Type: application/json

{
    "id": "65a1b2c3d4e5f6a7b8c9d0e1",
    "slug": "alice-and-bob-wedding",
    "coupleName": "Alice & Bob",
    "eventDate": "2024-12-25T18:00:00Z",
    "venue": { ... },
    "settings": {
        "isPublic": false,
        "allowRSVP": true,
        "guestLimit": 0
    },
    "createdAt": "2024-01-15T10:30:00Z"
}
```

### Example 2: Guest Submitting RSVP (Public)

```
┌─────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ Browser │────▶│   Gateway   │────▶│  Handler    │────▶│  Use Case   │────▶│   Domain    │
└─────────┘     └─────────────┘     └─────────────┘     └─────────────┘     └─────────────┘
      │               │                   │                   │                   │
      │ POST          │                   │                   │                   │
      │ /w/{slug}     │                   │                   │                   │
      │ /rsvp         │                   │                   │                   │
      │──────────────▶│                   │                   │                   │
      │               │                   │                   │                   │
      │               │  1. Rate Limit    │                   │                   │
      │               │  (stricter: 10/min)│                   │                   │
      │               │  2. CORS          │                   │                   │
      │               │  3. Validate body │                   │                   │
      │               │──────────────────▶│                   │                   │
      │               │                   │                   │                   │
      │               │                   │  1. Find wedding  │                   │
      │               │                   │     by slug       │                   │
      │               │                   │  2. Check if      │                   │
      │               │                   │     RSVP open     │                   │
      │               │                   │  3. Validate      │                   │
      │               │                   │     RSVP data     │                   │
      │               │                   │  4. Check guest   │                   │
      │               │                   │     exists        │                   │
      │               │                   │  5. Update RSVP   │                   │
      │               │                   │──────────────────▶│                   │
      │               │                   │                   │                   │
      │               │                   │                   │  1. Validate      │
      │               │                   │                   │     responses     │
      │               │                   │                   │  2. Check         │
      │               │                   │                   │     deadline      │
      │               │                   │                   │  3. Update status │
      │               │                   │                   │◀──────────────────│
      │               │                   │                   │                   │
      │               │                   │  Return result    │                   │
      │               │                   │◀──────────────────│                   │
      │               │                   │                   │                   │
      │               │                   │                   │     ┌──────────┐  │
      │               │                   │                   │     │ Guest    │  │
      │               │                   │                   │     │ Repository│  │
      │               │                   │                   │     │ (update) │  │
      │               │                   │                   │     └────┬─────┘  │
      │               │                   │                   │          │        │
      │               │                   │                   │     ┌────▼─────┐    │
      │               │                   │                   │     │ Email    │    │
      │               │                   │                   │     │ Service  │    │
      │               │                   │                   │     │ (async)  │    │
      │               │                   │                   │     └──────────┘    │
      │               │                   │                   │                   │
      │               │  Return 200 OK      │                   │                   │
      │◀──────────────│                   │                   │                   │
```

**Code Flow:**

```go
// 1. Public RSVP endpoint - no auth required
POST /w/alice-and-bob-wedding/rsvp
Content-Type: application/json

{
    "name": "John Smith",
    "email": "john@example.com",
    "attending": "yes",
    "guestCount": 2,
    "dietary": ["vegetarian"],
    "customResponses": {
        "song_request": "Anything by The Beatles"
    }
}

// 2. Handler processes public RSVP
func (h *RSVPHandler) SubmitRSVP(c *gin.Context) {
    slug := c.Param("slug")
    
    var req dto.RSVPRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, dto.ErrorResponse{Message: err.Error()})
        return
    }
    
    // Submit RSVP through use case
    result, err := h.rsvpUseCase.Submit(c.Request.Context(), slug, req.ToDomain())
    if err != nil {
        switch {
        case errors.Is(err, domain.ErrWeddingNotFound):
            c.JSON(404, dto.ErrorResponse{Message: "Wedding not found"})
        case errors.Is(err, domain.ErrRSVPClosed):
            c.JSON(403, dto.ErrorResponse{Message: "RSVP period has ended"})
        default:
            c.JSON(500, dto.ErrorResponse{Message: "Failed to submit RSVP"})
        }
        return
    }
    
    c.JSON(200, dto.RSVPResponse{Message: "RSVP submitted successfully"})
}

// 3. Use case validates and processes
func (uc *rsvpUseCase) Submit(ctx context.Context, slug string, rsvp *domain.RSVP) error {
    // Get wedding by slug
    wedding, err := uc.weddingRepo.GetBySlug(ctx, slug)
    if err != nil {
        return err
    }
    
    // Check if RSVP is open
    if !wedding.IsRSVPOpen() {
        return domain.ErrRSVPClosed
    }
    
    // Validate RSVP against wedding config
    if err := validateRSVPAgainstConfig(rsvp, wedding.RSVPConfig); err != nil {
        return err
    }
    
    // Find existing guest or create new
    var guest *domain.Guest
    if rsvp.Email != "" {
        guest, _ = uc.guestRepo.GetByEmail(ctx, wedding.ID, rsvp.Email)
    }
    
    if guest == nil {
        // New guest - create
        guest = &domain.Guest{
            ID:        primitive.NewObjectID(),
            WeddingID: wedding.ID,
            Name:      rsvp.Name,
            Email:     rsvp.Email,
        }
        if err := uc.guestRepo.Create(ctx, guest); err != nil {
            return err
        }
    }
    
    // Update guest with RSVP
    guest.RSVP = *rsvp
    guest.RSVP.SubmittedAt = time.Now()
    
    if err := uc.guestRepo.UpdateRSVP(ctx, guest.ID, &guest.RSVP); err != nil {
        return err
    }
    
    // Send confirmation email (async)
    uc.emailService.SendRSVPConfirmation(ctx, guest, wedding)
    
    return nil
}
```

---

## Key Design Decisions

### 1. Repository Pattern with Interfaces

**Decision:** Define repository interfaces in domain layer, implement in infrastructure.

**Pros:**
- Database can be swapped without touching domain logic
- Easy to create in-memory repositories for testing
- Single Responsibility: domain defines contracts, infrastructure implements

**Cons:**
- More boilerplate code
- Slight performance overhead from interface dispatch

**Example:**

```go
// Domain layer - interface definition
type WeddingRepository interface {
    Create(ctx context.Context, wedding *Wedding) error
    GetByID(ctx context.Context, id primitive.ObjectID) (*Wedding, error)
    // ...
}

// Test - mock implementation
type MockWeddingRepository struct {
    weddings map[string]*Wedding
}

func (m *MockWeddingRepository) Create(ctx context.Context, w *Wedding) error {
    m.weddings[w.ID.Hex()] = w
    return nil
}

// Infrastructure - MongoDB implementation
type MongoWeddingRepository struct {
    collection *mongo.Collection
}

func (r *MongoWeddingRepository) Create(ctx context.Context, w *Wedding) error {
    _, err := r.collection.InsertOne(ctx, w)
    return err
}
```

### 2. DTO Layer for API Contracts

**Decision:** Separate domain entities from API request/response structures.

**Pros:**
- API can evolve independently from domain model
- Can expose different fields for different endpoints
- Prevents accidentally exposing sensitive data
- Versioning support

**Cons:**
- Mapping code between DTOs and entities
- Need to maintain two sets of structs

**Example:**

```go
// DTO for public wedding page (limited data)
type PublicWeddingResponse struct {
    Slug       string    `json:"slug"`
    CoupleName string    `json:"coupleName"`
    EventDate  time.Time `json:"eventDate"`
    Venue      VenueDTO  `json:"venue"`
    Story      string    `json:"story,omitempty"`
    Gallery    []ImageDTO `json:"gallery,omitempty"`
    // Note: No UserID, Settings, Analytics
}

// DTO for admin dashboard (full data)
type AdminWeddingResponse struct {
    ID          string            `json:"id"`
    UserID      string            `json:"userId"`
    CoupleName  string            `json:"coupleName"`
    Slug        string            `json:"slug"`
    EventDate   time.Time         `json:"eventDate"`
    Venue       VenueDTO          `json:"venue"`
    Settings    SettingsDTO       `json:"settings"`
    Analytics   AnalyticsDTO      `json:"analytics"`
    CreatedAt   time.Time         `json:"createdAt"`
    UpdatedAt   time.Time         `json:"updatedAt"`
    GuestCount  int64             `json:"guestCount"`
    RSVPStats   RSVPStatsDTO      `json:"rsvpStats"`
}
```

### 3. Use Case Layer for Business Operations

**Decision:** Introduce use case layer between handlers and domain.

**Pros:**
- Single place for business logic orchestration
- Reusable across different interfaces (HTTP, CLI, gRPC)
- Easy to unit test business flows
- Clear transaction boundaries

**Cons:**
- Additional layer to maintain
- May feel like overkill for simple CRUD

**Example:**

```go
// Use case coordinates multiple repositories and services
type WeddingUseCase struct {
    weddingRepo  WeddingRepository
    guestRepo    GuestRepository
    emailService EmailService
    fileService  FileService
}

func (uc *WeddingUseCase) CreateWedding(ctx context.Context, userID string, req CreateWeddingRequest) (*Wedding, error) {
    // 1. Generate unique slug
    slug, err := uc.generateUniqueSlug(ctx, req.CoupleName)
    if err != nil {
        return nil, err
    }
    
    // 2. Create wedding entity
    wedding := &Wedding{
        UserID:     userID,
        CoupleName: req.CoupleName,
        Slug:       slug,
        // ...
    }
    
    // 3. Persist
    if err := uc.weddingRepo.Create(ctx, wedding); err != nil {
        return nil, err
    }
    
    // 4. Send welcome email
    uc.emailService.SendWeddingCreatedEmail(ctx, wedding)
    
    return wedding, nil
}
```

### 4. Soft Deletes vs Hard Deletes

**Decision:** Use soft deletes with `deleted_at` timestamp.

**Pros:**
- Data recovery possible
- Analytics preserved (deleted weddings still count in reports)
- Referential integrity maintained
- Audit trail

**Cons:**
- Queries must filter out deleted records
- More storage used
- Complexity in data cleanup

**Implementation:**

```go
type Wedding struct {
    // ... fields ...
    DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"-"`
}

func (r *MongoWeddingRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
    // Soft delete
    filter := bson.M{"_id": id}
    update := bson.M{"$set": bson.M{"deleted_at": time.Now()}}
    
    _, err := r.collection.UpdateOne(ctx, filter, update)
    return err
}

func (r *MongoWeddingRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*Wedding, error) {
    // Always exclude deleted
    filter := bson.M{
        "_id":        id,
        "deleted_at": bson.M{"$exists": false},
    }
    
    var wedding Wedding
    err := r.collection.FindOne(ctx, filter).Decode(&wedding)
    return &wedding, err
}
```

### 5. Async Email Notifications

**Decision:** Send emails asynchronously using goroutines or message queue.

**Pros:**
- API responses aren't blocked by email delivery
- Retry logic can be implemented
- Can batch emails for efficiency
- Prevents timeout issues

**Cons:**
- Need to handle failures
- Requires monitoring
- Slight complexity increase

**Implementation:**

```go
func (uc *WeddingUseCase) Create(ctx context.Context, wedding *Wedding) (*Wedding, error) {
    // Save wedding first
    if err := uc.weddingRepo.Create(ctx, wedding); err != nil {
        return nil, err
    }
    
    // Send email asynchronously
    go func() {
        // Create new context with timeout for email
        emailCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        
        if err := uc.emailService.SendWeddingCreatedEmail(emailCtx, wedding); err != nil {
            // Log error, could retry or alert
            log.Printf("Failed to send email: %v", err)
        }
    }()
    
    return wedding, nil
}
```

### 6. Slug-based Public Access

**Decision:** Use human-readable slugs (`/w/alice-bob-wedding`) instead of IDs for public URLs.

**Pros:**
- SEO friendly
- Memorable and shareable
- Professional appearance
- Can be customized by users

**Cons:**
- Need to handle slug uniqueness
- Slug changes require redirects
- Reserved word conflicts

**Implementation:**

```go
func (s *SlugService) GenerateUnique(ctx context.Context, coupleName string) (string, error) {
    // Generate base slug
    base := slugify(coupleName) // "Alice & Bob" -> "alice-and-bob"
    
    // Check uniqueness
    slug := base + "-wedding"
    for i := 1; ; i++ {
        exists, err := s.weddingRepo.ExistsBySlug(ctx, slug)
        if err != nil {
            return "", err
        }
        if !exists {
            return slug, nil
        }
        // Append number if taken
        slug = fmt.Sprintf("%s-wedding-%d", base, i)
    }
}

func slugify(s string) string {
    // Lowercase, remove special chars, replace spaces with hyphens
    s = strings.ToLower(s)
    s = regexp.MustCompile(`[^a-z0-9\s]`).ReplaceAllString(s, "")
    s = strings.Join(strings.Fields(s), "-")
    return s
}
```

---

## Scalability Considerations

### Horizontal Scaling Strategy

```
                    ┌─────────────────┐
                    │   Load Balancer │
                    │   (Nginx/ALB)   │
                    └────────┬────────┘
                             │
           ┌─────────────────┼─────────────────┐
           │                 │                 │
    ┌──────▼──────┐   ┌──────▼──────┐   ┌──────▼──────┐
    │  Go API     │   │  Go API     │   │  Go API     │
    │  Server 1   │   │  Server 2   │   │  Server N   │
    │             │   │             │   │             │
    │  Stateless  │   │  Stateless  │   │  Stateless  │
    │  No Session │   │  No Session │   │  No Session │
    └──────┬──────┘   └──────┬──────┘   └──────┬──────┘
           │                 │                 │
           └─────────────────┼─────────────────┘
                             │
                    ┌────────┴────────┐
                    │   MongoDB       │
                    │   Replica Set   │
                    │                 │
                    │  Primary       │
                    │  + 2 Secondaries│
                    └────────┬────────┘
                             │
                    ┌────────┴────────┐
                    │   Redis         │
                    │   (Optional)    │
                    │                 │
                    │  • Sessions     │
                    │  • Rate Limit   │
                    │  • Cache        │
                    └─────────────────┘
```

### Database Scaling

| Strategy | When to Use | Implementation |
|----------|-------------|----------------|
| **Vertical Scaling** | < 10,000 weddings | Larger MongoDB instance |
| **Read Replicas** | High read load | Route reads to secondaries |
| **Sharding** | > 100,000 weddings | Shard by user_id or region |
| **Archiving** | Old weddings | Move >1 year to cold storage |

**Read Scaling Example:**

```go
// Repository with read preference
type MongoWeddingRepository struct {
    primary   *mongo.Collection  // For writes
    secondary *mongo.Collection  // For reads
}

func (r *MongoWeddingRepository) GetBySlug(ctx context.Context, slug string) (*Wedding, error) {
    // Use secondary for reads
    var wedding Wedding
    err := r.secondary.FindOne(ctx, bson.M{"slug": slug}).Decode(&wedding)
    return &wedding, err
}

func (r *MongoWeddingRepository) Create(ctx context.Context, wedding *Wedding) error {
    // Use primary for writes
    _, err := r.primary.InsertOne(ctx, wedding)
    return err
}
```

### Caching Strategy

| Data Type | Cache Location | TTL | Invalidation |
|-----------|---------------|-----|--------------|
| **Wedding Details** | Redis | 5 min | On update |
| **Public Wedding Page** | CDN (Cloudflare) | 1 hour | Manual purge |
| **RSVP Stats** | In-Memory | 1 min | On new RSVP |
| **Guest Lists** | Redis | 10 min | On guest change |
| **Slug Mappings** | Redis | 1 hour | On slug change |

```go
// Caching middleware for public wedding pages
func CachePublicWedding(ttl time.Duration) gin.HandlerFunc {
    return func(c *gin.Context) {
        slug := c.Param("slug")
        cacheKey := fmt.Sprintf("wedding:%s", slug)
        
        // Try cache first
        if cached, err := redis.Get(c, cacheKey); err == nil {
            c.Data(200, "application/json", cached)
            c.Abort()
            return
        }
        
        // Process request
        c.Next()
        
        // Cache successful responses
        if c.Writer.Status() == 200 {
            body := c.Writer.Body() // Custom response writer
            redis.Set(c, cacheKey, body, ttl)
        }
    }
}
```

### File Storage Scaling

| Scale | Strategy | Cost/Month |
|-------|----------|------------|
| < 1,000 images | Direct S3 | $5-20 |
| < 10,000 images | S3 + CloudFront CDN | $20-50 |
| < 100,000 images | Cloudflare R2 (no egress) | $50-100 |
| > 100,000 images | R2 + Image optimization | $100-300 |

**Optimization Strategy:**

```go
func (s *S3FileService) UploadOptimized(ctx context.Context, file FileUpload) (*UploadResult, error) {
    // 1. Validate and optimize image
    img, err := imaging.Decode(file.Reader)
    if err != nil {
        return nil, err
    }
    
    // 2. Generate multiple sizes
    sizes := map[string]int{
        "thumbnail": 150,
        "medium":    800,
        "large":     1920,
    }
    
    results := make(map[string]string)
    for name, width := range sizes {
        resized := imaging.Resize(img, width, 0, imaging.Lanczos)
        
        buf := new(bytes.Buffer)
        if err := imaging.Encode(buf, resized, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
            return nil, err
        }
        
        key := fmt.Sprintf("weddings/%s/%s/%s-%s.jpg", 
            file.WeddingID, file.Type, file.ID, name)
        
        _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
            Bucket:      aws.String(s.bucket),
            Key:         aws.String(key),
            Body:        bytes.NewReader(buf.Bytes()),
            ContentType: aws.String("image/jpeg"),
        })
        
        if err != nil {
            return nil, err
        }
        
        results[name] = s.getURL(key)
    }
    
    return &UploadResult{
        URLs: results,
    }, nil
}
```

### Rate Limiting at Scale

| Method | Pros | Cons | Best For |
|--------|------|------|----------|
| **In-Memory** | Fast, simple | Per-instance, not shared | Single instance |
| **Redis** | Shared, accurate | Network latency | Multi-instance |
| **Token Bucket** | Smooth traffic | Complex | APIs with bursts |

**Distributed Rate Limiting with Redis:**

```go
type RedisRateLimiter struct {
    client *redis.Client
}

func (r *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
    // Use Redis INCR with EXPIRE
    pipe := r.client.Pipeline()
    
    incr := pipe.Incr(ctx, fmt.Sprintf("rate:%s", key))
    pipe.Expire(ctx, fmt.Sprintf("rate:%s", key), window)
    
    _, err := pipe.Exec(ctx)
    if err != nil {
        return false, err
    }
    
    current := incr.Val()
    return current <= int64(limit), nil
}
```

---

## Technology Alternatives

### Alternative Approaches Considered

#### 1. Framework: Gin vs Fiber

**Alternative: Fiber**

```go
// Fiber implementation would look similar
app := fiber.New()

app.Post("/api/weddings", func(c *fiber.Ctx) error {
    var req dto.CreateWeddingRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(dto.ErrorResponse{Message: err.Error()})
    }
    
    wedding, err := weddingUseCase.Create(c.Context(), req.ToDomain())
    if err != nil {
        return c.Status(500).JSON(dto.ErrorResponse{Message: err.Error()})
    }
    
    return c.Status(201).JSON(dto.NewWeddingResponse(wedding))
})
```

| Aspect | Gin | Fiber |
|--------|-----|-------|
| **Performance** | 125k req/s | 180k req/s |
| **Memory Usage** | Moderate | Lower |
| **Maturity** | Very mature (2016) | Newer (2020) |
| **Ecosystem** | Extensive | Growing |
| **Clean Arch** | Excellent | Good |

**Decision:** Gin chosen for maturity and ecosystem. Could migrate to Fiber for performance gains later.

#### 2. Database: MongoDB vs PostgreSQL

**Alternative: PostgreSQL with JSONB**

```sql
-- PostgreSQL JSONB schema
CREATE TABLE weddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    couple_name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    event_data JSONB NOT NULL,  -- Flexible content
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW()
);

-- Query JSONB
SELECT * FROM weddings 
WHERE event_data->>'city' = 'New York';
```

| Aspect | MongoDB | PostgreSQL |
|--------|---------|------------|
| **Schema Flexibility** | Native | Good (JSONB) |
| **Query Power** | Good | Excellent |
| **Horizontal Scale** | Built-in | Requires effort |
| **Transactions** | Yes (multi-doc) | Excellent |
| **Hosting Cost** | $60-80/mo (Atlas) | $15-25/mo (RDS) |

**Decision:** MongoDB for natural document fit and horizontal scaling path.

#### 3. Authentication: JWT vs Session

**Alternative: Server-Side Sessions with Redis**

```go
// Session-based auth
func (h *AuthHandler) Login(c *gin.Context) {
    // Validate credentials
    user := validateCredentials(email, password)
    
    // Create session
    sessionID := generateSessionID()
    redis.Set(ctx, fmt.Sprintf("session:%s", sessionID), user.ID, 24*time.Hour)
    
    // Set cookie
    c.SetCookie("session_id", sessionID, 86400, "/", "", true, true)
    
    c.JSON(200, gin.H{"message": "Logged in"})
}
```

| Aspect | JWT | Session |
|--------|-----|---------|
| **Server Memory** | Stateless | Requires store |
| **Token Revocation** | Complex (blacklist) | Easy (delete key) |
| **Scale** | Excellent | Requires shared store |
| **Security** | Good (RS256) | Good |
| **Implementation** | Simple | More complex |

**Decision:** JWT for stateless scaling. Can add Redis blacklist for token revocation.

#### 4. File Storage: S3 vs Self-Hosted

**Alternative: Self-Hosted MinIO**

```go
// MinIO is S3-compatible
minioClient, err := minio.New("localhost:9000", &minio.Options{
    Creds:  credentials.NewStaticV4("access-key", "secret-key", ""),
    Secure: false,
})
```

| Aspect | S3/R2 | MinIO Self-Hosted |
|--------|-------|-------------------|
| **Cost** | $0.023/GB | Server cost only |
| **Maintenance** | None | Full management |
| **Scalability** | Infinite | Limited by server |
| **CDN** | Integrated | Need separate CDN |
| **Backup** | Automatic | Manual setup |

**Decision:** Cloudflare R2 for zero egress fees. MinIO for local development.

#### 5. Architecture: Clean vs MVC

**Alternative: Traditional MVC**

```go
// MVC approach - simpler but less flexible
// models/wedding.go - struct only
// controllers/wedding_controller.go - HTTP + business logic
// database/db.go - direct DB access

func CreateWedding(c *gin.Context) {
    var wedding Wedding
    c.BindJSON(&wedding)
    
    // Business logic mixed with HTTP
    if wedding.EventDate.Before(time.Now()) {
        c.JSON(400, gin.H{"error": "Invalid date"})
        return
    }
    
    // Direct DB call
    db.Collection("weddings").InsertOne(ctx, wedding)
    
    c.JSON(201, wedding)
}
```

| Aspect | Clean Architecture | MVC |
|--------|-------------------|-----|
| **Initial Development** | Slower | Faster |
| **Testing** | Excellent | Moderate |
| **Refactoring** | Easy | Hard |
| **Team Size** | Better for >2 devs | OK for solo |
| **Long-term** | Maintains quality | Tech debt grows |

**Decision:** Clean Architecture for long-term maintainability.

---

## Summary

This architecture provides a solid foundation for the Wedding Invitation System with the following key attributes:

| Attribute | Implementation |
|-----------|---------------|
| **Scalability** | Horizontal scaling ready, stateless design |
| **Maintainability** | Clean separation of concerns, testable layers |
| **Performance** | Efficient Go runtime, MongoDB, optional caching |
| **Security** | JWT auth, input validation, rate limiting, audit logs |
| **Flexibility** | Interface-based dependencies, swappable implementations |
| **Developer Experience** | Auto-generated docs, consistent patterns, clear structure |

### Next Steps

1. **Review Database Schema** - See [02-database-schema.md](./02-database-schema.md)
2. **API Design** - See [03-api-reference.md](./03-api-reference.md)
3. **Project Structure** - See [05-project-structure.md](./05-project-structure.md)
4. **Implementation Roadmap** - See [06-implementation-roadmap.md](./06-implementation-roadmap.md)

---

**Document Version:** 1.0  
**Last Updated:** 2024-01-15  
**Review Cycle:** Quarterly or on major feature addition
