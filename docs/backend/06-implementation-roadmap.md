# Backend Implementation Roadmap

**Wedding Invitation System - 6-Week Development Plan**

---

## Executive Overview

| Phase | Duration | Focus | Deliverables |
|-------|----------|-------|--------------|
| Phase 1 | Week 1-2 | Foundation | Project setup, auth, basic wedding CRUD |
| Phase 2 | Week 3 | Core Features | Complete wedding management, slugs, public API, RSVP |
| Phase 3 | Week 4 | Guest Management | Guest CRUD, CSV import, RSVP management, emails |
| Phase 4 | Week 5 | Advanced Features | Analytics, rate limiting, security, email verification |
| Phase 5 | Week 6 | Deployment | Production setup, CI/CD, monitoring, backups |

---

## Phase Dependencies

```
Week 1-2: [PHASE 1: Foundation]
    |
    |- Docker Compose
    |- MongoDB connection
    |- Configuration
    |- Auth system
    |- Basic wedding CRUD
         |
         v
Week 3: [PHASE 2: Core Features]
    |
    |- Complete wedding CRUD
    |- Slug generation
    |- Public API
    |- RSVP submission
    |- File upload infrastructure
         |
         v
Week 4: [PHASE 3: Guest Management]
    |
    |- Guest list CRUD
    |- Bulk CSV import
    |- Complete RSVP management
    |- Email notifications
         |
         v
Week 5: [PHASE 4: Advanced Features]
    |
    |- Analytics tracking
    |- Rate limiting
    |- Email verification
    |- Password reset
    |- Security hardening
         |
         v
Week 6: [PHASE 5: Deployment]
    |
    |- Production Docker
    |- CI/CD pipeline
    |- Monitoring setup
    |- Backup strategy
```

---

## Phase 1: Foundation (Week 1-2)

### Goals
- Establish project structure and development environment
- Implement authentication and authorization
- Create basic wedding CRUD operations
- Set up database layer

### Week 1 Deliverables

| File/Component | Path | Purpose | Priority |
|----------------|------|---------|----------|
| **Project Initialization** | | | |
| `go.mod` | `/go.mod` | Go module definition | Must Have |
| `main.go` | `/cmd/api/main.go` | Application entry point | Must Have |
| **Configuration** | | | |
| `config.go` | `/internal/config/config.go` | Environment-based configuration | Must Have |
| `.env.example` | `/.env.example` | Environment template | Must Have |
| **Database** | | | |
| `database.go` | `/pkg/database/mongodb.go` | MongoDB connection manager | Must Have |
| `mongo_repository.go` | `/internal/repository/mongodb/base.go` | Base MongoDB repository | Must Have |
| **Domain Layer** | | | |
| `user.go` | `/internal/domain/models/user.go` | User entity | Must Have |
| `wedding.go` | `/internal/domain/models/wedding.go` | Wedding entity | Must Have |
| `repository.go` | `/internal/domain/repository/interfaces.go` | Repository interfaces | Must Have |
| **Middleware** | | | |
| `auth.go` | `/internal/middleware/auth.go` | JWT authentication middleware | Must Have |
| `logger.go` | `/internal/middleware/logger.go` | Request logging | Should Have |

### Week 2 Deliverables

| File/Component | Path | Purpose | Priority |
|----------------|------|---------|----------|
| **Authentication** | | | |
| `auth_handler.go` | `/internal/handler/auth.go` | Auth HTTP handlers | Must Have |
| `auth_service.go` | `/internal/service/auth.go` | Auth business logic | Must Have |
| `jwt.go` | `/internal/utils/jwt.go` | JWT utilities | Must Have |
| `password.go` | `/internal/utils/password.go` | Password hashing | Must Have |
| **User Management** | | | |
| `user_repository.go` | `/internal/repository/mongodb/user.go` | User data access | Must Have |
| `user_service.go` | `/internal/service/user.go` | User business logic | Must Have |
| **Wedding Basics** | | | |
| `wedding_repository.go` | `/internal/repository/mongodb/wedding.go` | Wedding data access | Must Have |
| `wedding_handler.go` | `/internal/handler/wedding.go` | Wedding HTTP handlers | Must Have |
| **Docker** | | | |
| `docker-compose.yml` | `/docker-compose.yml` | Local development stack | Must Have |
| `Dockerfile` | `/Dockerfile` | Production build | Must Have |
| `Dockerfile.dev` | `/Dockerfile.dev` | Development build | Should Have |

### Implementation Tasks

#### Task 1.1: Project Setup

```bash
# Initialize Go module
go mod init wedding-invitation-backend

# Create project structure
mkdir -p cmd/api internal/{config,domain/{models,repository},handler,middleware,service,utils,dto} pkg/database

# Add core dependencies
go get github.com/gin-gonic/gin
go get go.mongodb.org/mongo-driver/mongo
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt
go get github.com/go-playground/validator/v10
go get github.com/spf13/viper
go get go.uber.org/zap
```

#### Task 1.2: Configuration Management

```go
// internal/config/config.go
package config

import (
    "time"
    "github.com/spf13/viper"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Auth     AuthConfig
    Storage  StorageConfig
    Email    EmailConfig
}

type ServerConfig struct {
    Port           string        `mapstructure:"PORT"`
    Environment    string        `mapstructure:"APP_ENV"`
    AllowedOrigins []string      `mapstructure:"ALLOWED_ORIGINS"`
    ReadTimeout    time.Duration `mapstructure:"SERVER_READ_TIMEOUT"`
    WriteTimeout   time.Duration `mapstructure:"SERVER_WRITE_TIMEOUT"`
}

type DatabaseConfig struct {
    URI      string `mapstructure:"MONGODB_URI"`
    Database string `mapstructure:"MONGODB_DATABASE"`
    Timeout  int    `mapstructure:"MONGODB_TIMEOUT_SECONDS"`
}

type AuthConfig struct {
    JWTSecret        string        `mapstructure:"JWT_SECRET"`
    JWTRefreshSecret string        `mapstructure:"JWT_REFRESH_SECRET"`
    AccessTokenTTL   time.Duration `mapstructure:"JWT_ACCESS_TTL"`
    RefreshTokenTTL  time.Duration `mapstructure:"JWT_REFRESH_TTL"`
    BcryptCost       int           `mapstructure:"BCRYPT_COST"`
}

type StorageConfig struct {
    Provider  string `mapstructure:"STORAGE_PROVIDER"`
    Region    string `mapstructure:"AWS_REGION"`
    Bucket    string `mapstructure:"S3_BUCKET_NAME"`
    AccessKey string `mapstructure:"AWS_ACCESS_KEY_ID"`
    SecretKey string `mapstructure:"AWS_SECRET_ACCESS_KEY"`
    CDNURL    string `mapstructure:"CDN_URL"`
}

type EmailConfig struct {
    Provider string `mapstructure:"EMAIL_PROVIDER"`
    APIKey   string `mapstructure:"SENDGRID_API_KEY"`
    From     string `mapstructure:"EMAIL_FROM"`
}

func Load() (*Config, error) {
    viper.SetDefault("PORT", "8080")
    viper.SetDefault("APP_ENV", "development")
    viper.SetDefault("MONGODB_TIMEOUT_SECONDS", 10)
    viper.SetDefault("JWT_ACCESS_TTL", "15m")
    viper.SetDefault("JWT_REFRESH_TTL", "168h")
    viper.SetDefault("BCRYPT_COST", 12)
    
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath("./config")
    viper.AddConfigPath(".")
    
    viper.AutomaticEnv()
    
    if err := viper.ReadInConfig(); err != nil {
        // Config file not required - env vars can be used
    }
    
    var cfg Config
    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}

func (c *Config) IsDevelopment() bool {
    return c.Server.Environment == "development"
}

func (c *Config) IsProduction() bool {
    return c.Server.Environment == "production"
}
```

#### Task 1.3: MongoDB Connection

```go
// pkg/database/mongodb.go
package database

import (
    "context"
    "fmt"
    "time"
    
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/mongo/readpref"
    "wedding-invitation-backend/internal/config"
)

type MongoDB struct {
    Client   *mongo.Client
    Database *mongo.Database
}

func NewMongoDB(cfg *config.DatabaseConfig) (*MongoDB, error) {
    ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Timeout)*time.Second)
    defer cancel()
    
    clientOptions := options.Client().ApplyURI(cfg.URI)
    
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
    }
    
    ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := client.Ping(ctx, readpref.Primary()); err != nil {
        return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
    }
    
    return &MongoDB{
        Client:   client,
        Database: client.Database(cfg.Database),
    }, nil
}

func (m *MongoDB) Close(ctx context.Context) error {
    return m.Client.Disconnect(ctx)
}

func (m *MongoDB) Collection(name string) *mongo.Collection {
    return m.Database.Collection(name)
}

func (m *MongoDB) EnsureIndexes(ctx context.Context) error {
    users := m.Collection("users")
    if _, err := users.Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys:    map[string]interface{}{"email": 1},
        Options: options.Index().SetUnique(true),
    }); err != nil {
        return fmt.Errorf("failed to create users email index: %w", err)
    }
    
    weddings := m.Collection("weddings")
    if _, err := weddings.Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys:    map[string]interface{}{"slug": 1},
        Options: options.Index().SetUnique(true),
    }); err != nil {
        return fmt.Errorf("failed to create weddings slug index: %w", err)
    }
    
    if _, err := weddings.Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys: map[string]interface{}{"user_id": 1, "created_at": -1},
    }); err != nil {
        return fmt.Errorf("failed to create weddings user_id index: %w", err)
    }
    
    return nil
}
```

#### Task 1.4: Docker Compose Configuration

```yaml
# docker-compose.yml
version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile.dev
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=development
      - PORT=8080
      - MONGODB_URI=mongodb://mongodb:27017/wedding_invitations
      - MONGODB_DATABASE=wedding_invitations
      - JWT_SECRET=dev-secret-key-change-in-production
      - JWT_REFRESH_SECRET=dev-refresh-secret-change-in-production
      - ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
    volumes:
      - .:/app
      - /app/vendor
    depends_on:
      - mongodb
      - redis
    networks:
      - wedding-network

  mongodb:
    image: mongo:6.0
    ports:
      - "27017:27017"
    environment:
      - MONGO_INITDB_DATABASE=wedding_invitations
    volumes:
      - mongodb_data:/data/db
    networks:
      - wedding-network

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - wedding-network

  mongo-express:
    image: mongo-express:latest
    ports:
      - "8081:8081"
    environment:
      - ME_CONFIG_MONGODB_URL=mongodb://mongodb:27017/
      - ME_CONFIG_BASICAUTH_USERNAME=admin
      - ME_CONFIG_BASICAUTH_PASSWORD=admin
    depends_on:
      - mongodb
    networks:
      - wedding-network

volumes:
  mongodb_data:
  redis_data:

networks:
  wedding-network:
    driver: bridge
```

### Testing Requirements (Phase 1)

| Test Type | Coverage Target | Test Files |
|-----------|-----------------|------------|
| Unit Tests | 70% | All services, utilities |
| Integration Tests | MongoDB connectivity | Repository layer |
| E2E Tests | Auth endpoints | `/tests/e2e/auth_test.go` |

**Run Tests:**
```bash
# Unit tests
go test ./internal/... -v

# Integration tests
go test ./tests/integration/... -v

# Coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Definition of Done (Phase 1)

- [ ] Project structure created with all required directories
- [ ] Docker Compose runs successfully (`docker-compose up`)
- [ ] MongoDB connection established and indexes created
- [ ] Configuration loads from environment variables and files
- [ ] User can register, login, and refresh tokens
- [ ] JWT middleware protects authenticated routes
- [ ] Basic wedding CRUD endpoints working
- [ ] All unit tests pass with >70% coverage
- [ ] API documentation generated (Swagger)

---

## Phase 2: Core Features (Week 3)

### Goals
- Complete wedding management functionality
- Implement slug generation system
- Create public API for viewing weddings
- Implement RSVP submission
- Set up file upload infrastructure

### Deliverables Checklist

| File/Component | Path | Purpose | Priority |
|----------------|------|---------|----------|
| **Slug Service** | | | |
| `slug_service.go` | `/internal/service/slug.go` | Unique slug generation | Must Have |
| **Wedding Service** | | | |
| `wedding_service.go` | `/internal/service/wedding.go` | Wedding business logic | Must Have |
| **Complete Wedding Handlers** | | | |
| `wedding_handlers.go` | `/internal/handler/wedding.go` | Full wedding CRUD | Must Have |
| `public_handlers.go` | `/internal/handler/public.go` | Public wedding viewing | Must Have |
| **RSVP** | | | |
| `rsvp.go` | `/internal/domain/models/rsvp.go` | RSVP entity | Must Have |
| `rsvp_repository.go` | `/internal/repository/mongodb/rsvp.go` | RSVP data access | Must Have |
| `rsvp_service.go` | `/internal/service/rsvp.go` | RSVP business logic | Must Have |
| `rsvp_handler.go` | `/internal/handler/rsvp.go` | RSVP endpoints | Must Have |
| **File Upload** | | | |
| `upload_handler.go` | `/internal/handler/upload.go` | File upload endpoints | Must Have |
| `storage_service.go` | `/internal/service/storage.go` | Storage abstraction | Must Have |
| `local_storage.go` | `/pkg/storage/local.go` | Local file storage | Should Have |
| **DTOs** | | | |
| `wedding_dto.go` | `/internal/dto/wedding.go` | Wedding request/response | Must Have |
| `rsvp_dto.go` | `/internal/dto/rsvp.go` | RSVP request/response | Must Have |
| `upload_dto.go` | `/internal/dto/upload.go` | Upload request/response | Must Have |

### Implementation Tasks

#### Task 2.1: Slug Generation Service

```go
// internal/service/slug.go
package service

import (
    "context"
    "fmt"
    "regexp"
    "strings"
    "time"
    "crypto/rand"
    "wedding-invitation-backend/internal/domain/repository"
)

type SlugService interface {
    GenerateUnique(ctx context.Context, coupleName string) (string, error)
    IsAvailable(ctx context.Context, slug string) (bool, error)
}

type slugService struct {
    weddingRepo repository.WeddingRepository
}

func NewSlugService(weddingRepo repository.WeddingRepository) SlugService {
    return &slugService{weddingRepo: weddingRepo}
}

func (s *slugService) GenerateUnique(ctx context.Context, coupleName string) (string, error) {
    baseSlug := s.generateBaseSlug(coupleName)
    
    available, err := s.IsAvailable(ctx, baseSlug)
    if err != nil {
        return "", err
    }
    
    if available {
        return baseSlug, nil
    }
    
    for i := 1; i <= 100; i++ {
        candidate := fmt.Sprintf("%s-%d", baseSlug, i)
        available, err := s.IsAvailable(ctx, candidate)
        if err != nil {
            return "", err
        }
        if available {
            return candidate, nil
        }
    }
    
    return fmt.Sprintf("%s-%s", baseSlug, generateRandomString(6)), nil
}

func (s *slugService) IsAvailable(ctx context.Context, slug string) (bool, error) {
    exists, err := s.weddingRepo.ExistsBySlug(ctx, slug)
    if err != nil {
        return false, err
    }
    return !exists, nil
}

func (s *slugService) generateBaseSlug(name string) string {
    slug := strings.ToLower(name)
    re := regexp.MustCompile(`[^a-z0-9]+`)
    slug = re.ReplaceAllString(slug, "-")
    slug = strings.Trim(slug, "-")
    re = regexp.MustCompile(`-+`)
    slug = re.ReplaceAllString(slug, "-")
    
    if len(slug) > 50 {
        slug = slug[:50]
        slug = strings.TrimRight(slug, "-")
    }
    
    return slug
}

func generateRandomString(length int) string {
    const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
    b := make([]byte, length)
    for i := range b {
        b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
    }
    return string(b)
}
```

#### Task 2.2: RSVP Service

```go
// internal/service/rsvp.go
package service

import (
    "context"
    "errors"
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "wedding-invitation-backend/internal/domain/models"
    "wedding-invitation-backend/internal/domain/repository"
)

var (
    ErrRSVPNotFound      = errors.New("rsvp not found")
    ErrRSVPClosed        = errors.New("rsvp is closed for this wedding")
    ErrInvalidRSVPStatus = errors.New("invalid rsvp status")
    ErrDuplicateRSVP     = errors.New("rsvp already submitted for this email")
)

type RSVPService interface {
    Submit(ctx context.Context, weddingID primitive.ObjectID, req SubmitRSVPRequest) (*models.RSVP, error)
    GetByID(ctx context.Context, id primitive.ObjectID) (*models.RSVP, error)
    ListByWedding(ctx context.Context, weddingID primitive.ObjectID, userID primitive.ObjectID, page, pageSize int) ([]*models.RSVP, int64, error)
    Update(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, req UpdateRSVPRequest) (*models.RSVP, error)
    Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error
    GetStatistics(ctx context.Context, weddingID primitive.ObjectID, userID primitive.ObjectID) (*models.RSVPStatistics, error)
}

type rsvpService struct {
    rsvpRepo    repository.RSVPRepository
    weddingRepo repository.WeddingRepository
}

type SubmitRSVPRequest struct {
    FirstName           string
    LastName            string
    Email               string
    Phone               string
    Status              string
    AttendanceCount     int
    PlusOnes            []models.PlusOneInfo
    DietaryRestrictions string
    DietarySelected     []string
    AdditionalNotes     string
    CustomAnswers       []models.CustomAnswer
    IPAddress           string
    UserAgent           string
    Source              string
}

type UpdateRSVPRequest struct {
    Status              *string
    AttendanceCount     *int
    PlusOnes            *[]models.PlusOneInfo
    DietaryRestrictions *string
    DietarySelected     *[]string
    AdditionalNotes     *string
    CustomAnswers       *[]models.CustomAnswer
}

func (s *rsvpService) Submit(ctx context.Context, weddingID primitive.ObjectID, req SubmitRSVPRequest) (*models.RSVP, error) {
    wedding, err := s.weddingRepo.GetByID(ctx, weddingID)
    if err != nil {
        return nil, ErrWeddingNotFound
    }
    
    if !wedding.IsRSVPOpen() {
        return nil, ErrRSVPClosed
    }
    
    if req.Status != "attending" && req.Status != "not-attending" && req.Status != "maybe" {
        return nil, ErrInvalidRSVPStatus
    }
    
    if req.Email != "" {
        existing, _ := s.rsvpRepo.GetByEmail(ctx, weddingID, req.Email)
        if existing != nil {
            return nil, ErrDuplicateRSVP
        }
    }
    
    if len(req.PlusOnes) > wedding.RSVP.MaxPlusOnes {
        return nil, errors.New("too many plus ones")
    }
    
    rsvp := &models.RSVP{
        ID:                  primitive.NewObjectID(),
        WeddingID:           weddingID,
        FirstName:           req.FirstName,
        LastName:            req.LastName,
        Email:               req.Email,
        Phone:               req.Phone,
        Status:              req.Status,
        AttendanceCount:     req.AttendanceCount,
        PlusOnes:            req.PlusOnes,
        PlusOneCount:        len(req.PlusOnes),
        DietaryRestrictions: req.DietaryRestrictions,
        DietarySelected:     req.DietarySelected,
        AdditionalNotes:     req.AdditionalNotes,
        CustomAnswers:       req.CustomAnswers,
        SubmittedAt:         time.Now(),
        IPAddress:           req.IPAddress,
        UserAgent:           req.UserAgent,
        Source:              req.Source,
        ConfirmationSent:    false,
    }
    
    if err := s.rsvpRepo.Create(ctx, rsvp); err != nil {
        return nil, err
    }
    
    return rsvp, nil
}
```

### Definition of Done (Phase 2)

- [ ] Slug generation creates unique, URL-safe slugs
- [ ] Complete wedding CRUD with authorization
- [ ] Public API returns sanitized wedding data
- [ ] RSVP submission works with validation
- [ ] File upload accepts images up to 5MB
- [ ] All endpoints have Swagger documentation
- [ ] Integration tests for all new features
- [ ] API rate limiting configured

---

## Phase 3: Guest Management (Week 4)

### Goals
- Implement guest list CRUD operations
- Create bulk CSV import functionality
- Complete RSVP management interface
- Set up email notification system

### Deliverables Checklist

| File/Component | Path | Purpose | Priority |
|----------------|------|---------|----------|
| **Guest Domain** | | | |
| `guest.go` | `/internal/domain/models/guest.go` | Guest entity | Must Have |
| `guest_repository.go` | `/internal/repository/mongodb/guest.go` | Guest data access | Must Have |
| `guest_service.go` | `/internal/service/guest.go` | Guest business logic | Must Have |
| `guest_handler.go` | `/internal/handler/guest.go` | Guest HTTP handlers | Must Have |
| **CSV Import** | | | |
| `csv_importer.go` | `/internal/utils/csv.go` | CSV parsing utilities | Must Have |
| `import_handler.go` | `/internal/handler/import.go` | Bulk import endpoint | Must Have |
| **Email Service** | | | |
| `email_service.go` | `/internal/service/email.go` | Email service interface | Must Have |
| `sendgrid_email.go` | `/pkg/email/sendgrid.go` | SendGrid implementation | Must Have |
| **Templates** | | | |
| `rsvp_confirmation.html` | `/templates/emails/rsvp_confirmation.html` | RSVP confirmation | Must Have |
| `invitation.html` | `/templates/emails/invitation.html` | Guest invitation | Must Have |
| **DTOs** | | | |
| `guest_dto.go` | `/internal/dto/guest.go` | Guest request/response | Must Have |
| `import_dto.go` | `/internal/dto/import.go` | Import request/response | Must Have |

### Implementation Tasks

#### Task 3.1: Guest Service

```go
// internal/service/guest.go
package service

import (
    "context"
    "errors"
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "wedding-invitation-backend/internal/domain/models"
    "wedding-invitation-backend/internal/domain/repository"
)

var ErrGuestNotFound = errors.New("guest not found")

type GuestService interface {
    Create(ctx context.Context, weddingID, userID primitive.ObjectID, req CreateGuestRequest) (*models.Guest, error)
    CreateMany(ctx context.Context, weddingID, userID primitive.ObjectID, req []CreateGuestRequest) (*models.GuestImportResult, error)
    GetByID(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) (*models.Guest, error)
    ListByWedding(ctx context.Context, weddingID primitive.ObjectID, userID primitive.ObjectID, page, pageSize int, filters GuestFilters) ([]*models.Guest, int64, error)
    Update(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID, req UpdateGuestRequest) (*models.Guest, error)
    Delete(ctx context.Context, id primitive.ObjectID, userID primitive.ObjectID) error
    ImportFromCSV(ctx context.Context, weddingID, userID primitive.ObjectID, csvData []byte) (*models.GuestImportResult, error)
    SendInvitations(ctx context.Context, weddingID, userID primitive.ObjectID, guestIDs []primitive.ObjectID) error
    GetStatistics(ctx context.Context, weddingID primitive.ObjectID, userID primitive.ObjectID) (*GuestStatistics, error)
}

type guestService struct {
    guestRepo    repository.GuestRepository
    weddingRepo  repository.WeddingRepository
    emailService EmailService
}

type CreateGuestRequest struct {
    FirstName    string
    LastName     string
    Email        string
    Phone        string
    Address      *models.Address
    Relationship string
    Side         string
    AllowPlusOne bool
    MaxPlusOnes  int
    DietaryNotes string
    VIP          bool
    Notes        string
}

type GuestFilters struct {
    RSVPStatus   string
    Side         string
    Relationship string
    Search       string
}

type GuestStatistics struct {
    TotalGuests     int64 `json:"total_guests"`
    PendingRSVP     int64 `json:"pending_rsvp"`
    Attending       int64 `json:"attending"`
    NotAttending    int64 `json:"not_attending"`
    Maybe           int64 `json:"maybe"`
    InvitationsSent int64 `json:"invitations_sent"`
    BrideSide       int64 `json:"bride_side"`
    GroomSide       int64 `json:"groom_side"`
}

func (s *guestService) Create(ctx context.Context, weddingID, userID primitive.ObjectID, req CreateGuestRequest) (*models.Guest, error) {
    wedding, err := s.weddingRepo.GetByID(ctx, weddingID)
    if err != nil {
        return nil, ErrWeddingNotFound
    }
    
    if wedding.UserID != userID {
        return nil, ErrUnauthorized
    }
    
    guest := &models.Guest{
        ID:               primitive.NewObjectID(),
        WeddingID:        weddingID,
        FirstName:        req.FirstName,
        LastName:         req.LastName,
        Email:            req.Email,
        Phone:            req.Phone,
        Address:          req.Address,
        Relationship:     req.Relationship,
        Side:             req.Side,
        InvitedVia:       "digital",
        InvitationStatus: "pending",
        AllowPlusOne:     req.AllowPlusOne,
        MaxPlusOnes:      req.MaxPlusOnes,
        RSVPStatus:       "pending",
        DietaryNotes:     req.DietaryNotes,
        VIP:              req.VIP,
        Notes:            req.Notes,
        CreatedAt:        time.Now(),
        UpdatedAt:        time.Now(),
        CreatedBy:        userID,
    }
    
    if err := s.guestRepo.Create(ctx, guest); err != nil {
        return nil, err
    }
    
    return guest, nil
}

func (s *guestService) ImportFromCSV(ctx context.Context, weddingID, userID primitive.ObjectID, csvData []byte) (*models.GuestImportResult, error) {
    wedding, err := s.weddingRepo.GetByID(ctx, weddingID)
    if err != nil {
        return nil, ErrWeddingNotFound
    }
    
    if wedding.UserID != userID {
        return nil, ErrUnauthorized
    }
    
    guests, importErrors := parseCSV(csvData, weddingID, userID)
    
    batchID := primitive.NewObjectID().Hex()
    for _, g := range guests {
        g.ImportBatchID = batchID
    }
    
    if len(guests) > 0 {
        if err := s.guestRepo.CreateMany(ctx, guests); err != nil {
            return nil, err
        }
    }
    
    return &models.GuestImportResult{
        SuccessCount: len(guests),
        ErrorCount:   len(importErrors),
        Errors:       importErrors,
        BatchID:      batchID,
    }, nil
}
```

#### Task 3.2: Email Service

```go
// internal/service/email.go
package service

import (
    "context"
    "wedding-invitation-backend/internal/domain/models"
)

type EmailService interface {
    SendRSVPConfirmation(ctx context.Context, rsvp *models.RSVP, wedding *models.Wedding) error
    SendGuestInvitation(ctx context.Context, guest *models.Guest, wedding *models.Wedding) error
    SendPasswordReset(ctx context.Context, email, token string) error
    SendEmailVerification(ctx context.Context, email, token string) error
}
```

### Definition of Done (Phase 3)

- [ ] Guest CRUD operations complete with filtering
- [ ] CSV import handles up to 500 guests per file
- [ ] Email notifications sent for RSVPs
- [ ] Guest statistics endpoint returns accurate counts
- [ ] Email service configured (SendGrid or SMTP)
- [ ] Bulk operations work without timeouts
- [ ] Integration tests for email sending

---

## Phase 4: Advanced Features (Week 5)

### Goals
- Implement analytics tracking system
- Add rate limiting to all endpoints
- Set up email verification
- Implement password reset flow
- Security hardening

### Deliverables Checklist

| File/Component | Path | Purpose | Priority |
|----------------|------|---------|----------|
| **Analytics** | | | |
| `analytics.go` | `/internal/domain/models/analytics.go` | Analytics entities | Must Have |
| `analytics_repository.go` | `/internal/repository/mongodb/analytics.go` | Analytics data access | Must Have |
| `analytics_service.go` | `/internal/service/analytics.go` | Analytics aggregation | Must Have |
| `analytics_handler.go` | `/internal/handler/analytics.go` | Analytics endpoints | Must Have |
| `tracking_middleware.go` | `/internal/middleware/tracking.go` | Request tracking | Must Have |
| **Rate Limiting** | | | |
| `rate_limiter.go` | `/internal/middleware/rate_limiter.go` | Rate limiting middleware | Must Have |
| `redis_client.go` | `/pkg/cache/redis.go` | Redis client | Must Have |
| **Email Verification** | | | |
| `verification_service.go` | `/internal/service/verification.go` | Email verification logic | Must Have |
| `verification_handler.go` | `/internal/handler/verification.go` | Verification endpoints | Must Have |
| **Password Reset** | | | |
| `password_reset_service.go` | `/internal/service/password_reset.go` | Password reset logic | Must Have |
| `password_reset_handler.go` | `/internal/handler/password_reset.go` | Reset endpoints | Must Have |
| **Security** | | | |
| `security.go` | `/internal/middleware/security.go` | Security headers | Must Have |
| `cors.go` | `/internal/middleware/cors.go` | CORS configuration | Must Have |

### Implementation Tasks

#### Task 4.1: Rate Limiting Middleware

```go
// internal/middleware/rate_limiter.go
package middleware

import (
    "fmt"
    "net/http"
    "time"
    "github.com/gin-gonic/gin"
    "github.com/go-redis/redis/v8"
)

type RateLimiterConfig struct {
    RequestsPerMinute int
    BurstSize         int
    BlockDuration     time.Duration
}

var defaultLimits = map[string]RateLimiterConfig{
    "default": {RequestsPerMinute: 60, BurstSize: 10, BlockDuration: 5 * time.Minute},
    "auth":    {RequestsPerMinute: 5, BurstSize: 3, BlockDuration: 15 * time.Minute},
    "upload":  {RequestsPerMinute: 10, BurstSize: 5, BlockDuration: 30 * time.Minute},
    "rsvp":    {RequestsPerMinute: 20, BurstSize: 10, BlockDuration: 10 * time.Minute},
    "import":  {RequestsPerMinute: 5, BurstSize: 2, BlockDuration: 30 * time.Minute},
}

func RateLimiter(redisClient *redis.Client) gin.HandlerFunc {
    return func(c *gin.Context) {
        category := getCategory(c.Request.URL.Path, c.Request.Method)
        config := defaultLimits[category]
        
        clientIP := c.ClientIP()
        key := fmt.Sprintf("ratelimit:%s:%s", category, clientIP)
        
        ctx := c.Request.Context()
        
        pipe := redisClient.Pipeline()
        incr := pipe.Incr(ctx, key)
        pipe.Expire(ctx, key, time.Minute)
        _, err := pipe.Exec(ctx)
        
        if err != nil {
            c.Next()
            return
        }
        
        count := incr.Val()
        
        c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
        c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, int64(config.RequestsPerMinute)-count)))
        
        if count > int64(config.RequestsPerMinute) {
            c.Header("Retry-After", fmt.Sprintf("%d", int(config.BlockDuration.Seconds())))
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "rate limit exceeded",
                "retry_after": config.BlockDuration.Seconds(),
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

func getCategory(path, method string) string {
    if method == "POST" && (path == "/api/v1/auth/login" || path == "/api/v1/auth/register") {
        return "auth"
    }
    if method == "POST" && contains(path, "/upload") {
        return "upload"
    }
    if method == "POST" && contains(path, "/rsvp") {
        return "rsvp"
    }
    if method == "POST" && contains(path, "/import") {
        return "import"
    }
    return "default"
}

func contains(s, substr string) bool {
    return strings.Contains(s, substr)
}

func max(a, b int64) int64 {
    if a > b {
        return a
    }
    return b
}
```

#### Task 4.2: Security Middleware

```go
// internal/middleware/security.go
package middleware

import "github.com/gin-gonic/gin"

func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self' https:")
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
        c.Next()
    }
}

func CORS(allowedOrigins []string) gin.HandlerFunc {
    return func(c *gin.Context) {
        origin := c.Request.Header.Get("Origin")
        
        allowed := false
        for _, o := range allowedOrigins {
            if o == "*" || o == origin {
                allowed = true
                break
            }
        }
        
        if allowed {
            c.Header("Access-Control-Allow-Origin", origin)
            c.Header("Access-Control-Allow-Credentials", "true")
            c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
            c.Header("Access-Control-Allow-Methods", "POST, HEAD, PATCH, OPTIONS, GET, PUT, DELETE")
        }
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    }
}
```

### Definition of Done (Phase 4)

- [ ] Rate limiting active on all endpoints
- [ ] Analytics events tracked for page views
- [ ] Email verification flow working
- [ ] Password reset flow working
- [ ] Security headers on all responses
- [ ] CORS configured for allowed origins
- [ ] Redis caching for rate limits
- [ ] Security audit complete (OWASP Top 10)

---

## Phase 5: Deployment (Week 6)

### Goals
- Create production Docker configuration
- Set up CI/CD pipeline
- Implement monitoring and alerting
- Establish backup strategy

### Deliverables Checklist

| File/Component | Path | Purpose | Priority |
|----------------|------|---------|----------|
| **Docker Production** | | | |
| `Dockerfile.prod` | `/Dockerfile.prod` | Production Dockerfile | Must Have |
| `docker-compose.prod.yml` | `/docker-compose.prod.yml` | Production compose | Must Have |
| `.dockerignore` | `/.dockerignore` | Docker ignore rules | Must Have |
| **CI/CD** | | | |
| `ci.yml` | `/.github/workflows/ci.yml` | CI pipeline | Must Have |
| `cd.yml` | `/.github/workflows/cd.yml` | CD pipeline | Must Have |
| **Monitoring** | | | |
| `prometheus.yml` | `/monitoring/prometheus/prometheus.yml` | Prometheus config | Must Have |
| `grafana-dashboard.json` | `/monitoring/grafana/dashboard.json` | Grafana dashboard | Must Have |
| `health_handler.go` | `/internal/handler/health.go` | Health check endpoint | Must Have |
| **Scripts** | | | |
| `backup.sh` | `/scripts/backup.sh` | Backup script | Must Have |
| `deploy.sh` | `/scripts/deploy.sh` | Deployment script | Must Have |
| **Documentation** | | | |
| `DEPLOYMENT.md` | `/docs/DEPLOYMENT.md` | Deployment guide | Must Have |
| `MONITORING.md` | `/docs/MONITORING.md` | Monitoring guide | Should Have |

### Implementation Tasks

#### Task 5.1: Production Docker Compose

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile.prod
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - PORT=8080
      - MONGODB_URI=${MONGODB_URI}
      - MONGODB_DATABASE=${MONGODB_DATABASE}
      - JWT_SECRET=${JWT_SECRET}
      - JWT_REFRESH_SECRET=${JWT_REFRESH_SECRET}
      - REDIS_URL=${REDIS_URL}
      - ALLOWED_ORIGINS=${ALLOWED_ORIGINS}
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    deploy:
      replicas: 2
      resources:
        limits:
          cpus: '1'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 128M
    networks:
      - wedding-network

  nginx:
    image: nginx:alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./certbot/conf:/etc/letsencrypt:ro
    depends_on:
      - api
    networks:
      - wedding-network

  prometheus:
    image: prom/prometheus:latest
    restart: unless-stopped
    ports:
      - "9090:9090"
    volumes:
      - ./monitoring/prometheus:/etc/prometheus:ro
      - prometheus_data:/prometheus
    networks:
      - wedding-network

  grafana:
    image: grafana/grafana:latest
    restart: unless-stopped
    ports:
      - "3000:3000"
    volumes:
      - grafana_data:/var/lib/grafana
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD}
    networks:
      - wedding-network

networks:
  wedding-network:
    driver: bridge

volumes:
  prometheus_data:
  grafana_data:
```

#### Task 5.2: CI/CD Pipeline

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      mongodb:
        image: mongo:6.0
        ports:
          - 27017:27017
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Download dependencies
      run: go mod download
    
    - name: Run tests
      run: go test -v -race -coverprofile=coverage.out ./...
      env:
        MONGODB_URI: mongodb://localhost:27017
        MONGODB_DATABASE: wedding_test
        JWT_SECRET: test-secret
        JWT_REFRESH_SECRET: test-refresh-secret
    
    - name: Run linter
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
    
    - name: Build
      run: go build -o main cmd/api/main.go
```

```yaml
# .github/workflows/cd.yml
name: CD

on:
  push:
    branches: [ main ]
    tags:
      - 'v*'

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2
    
    - name: Login to Docker Hub
      uses: docker/login-action@v2
      with:
        username: ${{ secrets.DOCKER_USERNAME }}
        password: ${{ secrets.DOCKER_PASSWORD }}
    
    - name: Build and push
      uses: docker/build-push-action@v4
      with:
        context: .
        file: ./Dockerfile.prod
        push: true
        tags: ${{ secrets.DOCKER_USERNAME }}/wedding-api:latest
        cache-from: type=gha
        cache-to: type=gha,mode=max

  deploy:
    needs: build-and-push
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
    - name: Deploy to production
      uses: appleboy/ssh-action@v0.1.10
      with:
        host: ${{ secrets.SSH_HOST }}
        username: ${{ secrets.SSH_USER }}
        key: ${{ secrets.SSH_KEY }}
        script: |
          cd /opt/wedding-app
          docker-compose -f docker-compose.prod.yml pull
          docker-compose -f docker-compose.prod.yml up -d
          docker system prune -f
```

#### Task 5.3: Backup Script

```bash
#!/bin/bash
# scripts/backup.sh

set -e

BACKUP_DIR="/var/backups/wedding-app"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30
MONGODB_URI="${MONGODB_URI}"
S3_BUCKET="${BACKUP_S3_BUCKET}"

mkdir -p "${BACKUP_DIR}"

echo "Creating MongoDB backup..."
mongodump --uri="${MONGODB_URI}" --out="${BACKUP_DIR}/mongodb_${DATE}"

echo "Compressing backup..."
tar -czf "${BACKUP_DIR}/mongodb_${DATE}.tar.gz" -C "${BACKUP_DIR}" "mongodb_${DATE}"
rm -rf "${BACKUP_DIR}/mongodb_${DATE}"

if [ -n "${S3_BUCKET}" ]; then
    echo "Uploading to S3..."
    aws s3 cp "${BACKUP_DIR}/mongodb_${DATE}.tar.gz" "s3://${S3_BUCKET}/backups/"
    rm "${BACKUP_DIR}/mongodb_${DATE}.tar.gz"
fi

echo "Cleaning up old backups..."
find "${BACKUP_DIR}" -name "mongodb_*.tar.gz" -mtime +${RETENTION_DAYS} -delete

echo "Backup completed: mongodb_${DATE}.tar.gz"
```

### Definition of Done (Phase 5)

- [ ] Production Docker Compose runs successfully
- [ ] CI pipeline passes all tests and builds
- [ ] CD pipeline deploys to production on merge
- [ ] Health endpoint returns 200 with all checks green
- [ ] Prometheus collecting metrics
- [ ] Grafana dashboard displays key metrics
- [ ] Backup script runs and uploads to S3
- [ ] SSL certificates auto-renew via certbot
- [ ] Monitoring alerts configured for critical errors
- [ ] Documentation complete (DEPLOYMENT.md, API docs)

---

## Dependencies Between Phases

### Critical Path

| Phase | Depends On | Used By |
|-------|-----------|---------|
| Phase 1 | - | All Phases |
| Phase 2 | Phase 1 | Phases 3-5 |
| Phase 3 | Phase 1, Phase 2 | Phases 4-5 |
| Phase 4 | Phase 1, Phase 2, Phase 3 | Phase 5 |
| Phase 5 | Phase 1-4 | - |

### Cross-Phase Dependencies

| Feature | Depends On | Used By |
|---------|-----------|---------|
| MongoDB Connection | - | All Phases |
| User Auth | Phase 1 | Phases 2-5 |
| Wedding CRUD | Phase 1 | Phases 3-5 |
| Email Service | Phase 1 | Phases 3-5 |
| File Upload | Phase 2 | Phases 3-5 |
| RSVP System | Phase 2 | Phases 3-5 |
| Analytics | Phase 2 | Phase 5 |

---

## Risk Mitigation Strategies

### Technical Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| MongoDB Performance Issues | High | Medium | Add indexes early, implement pagination |
| File Upload Failures | Medium | Medium | Implement retries, validate before upload |
| Email Delivery Issues | Medium | High | Use multiple providers, queue emails |
| Rate Limiting Bypass | High | Low | Use Redis, implement distributed rate limiting |
| Memory Leaks | High | Low | Use pprof, set resource limits |
| Data Loss | Critical | Low | Automated backups, transaction logs |
| Security Breach | Critical | Low | Regular audits, dependency scanning |

### Schedule Risks

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Scope Creep | High | Medium | Strict change control, MVP focus |
| Integration Delays | Medium | Medium | Daily standups, early integration testing |
| Resource Availability | Medium | Low | Cross-training, documentation |
| Third-party Service Issues | Medium | Medium | Fallback options, local development |

---

## Review Checkpoints

### Weekly Reviews

| Week | Checkpoint | Success Criteria |
|------|------------|------------------|
| Week 1 | Architecture review | Project structure created, Docker running |
| Week 2 | Auth system review | Login/register working, JWT implemented |
| Week 3 | Core features review | Weddings CRUD, RSVP submission working |
| Week 4 | Guest management review | CSV import, email notifications working |
| Week 5 | Security review | Rate limiting, analytics, security headers active |
| Week 6 | Production readiness | CI/CD pipeline, monitoring, backups complete |

### Phase Gates

| Phase | Gate Criteria | Reviewers |
|-------|---------------|-----------|
| Phase 1 | All week 1-2 deliverables complete | Tech Lead |
| Phase 2 | Core features functional, API documented | Product Owner |
| Phase 3 | Guest management complete, emails sending | QA Team |
| Phase 4 | Security audit passed, performance tested | Security Team |
| Phase 5 | Production deployed, monitoring active | DevOps Team |

---

## Quick Start Commands

### Development

```bash
# Clone and setup
git clone <repo>
cd wedding-invitation-backend
cp .env.example .env

# Start services
docker-compose up -d

# Run migrations
go run cmd/migrate/main.go

# Start development server
go run cmd/api/main.go

# Run tests
go test ./...
```

### Production Deployment

```bash
# Set environment variables
export MONGODB_URI="mongodb://..."
export JWT_SECRET="your-secret"

# Deploy
docker-compose -f docker-compose.prod.yml up -d

# Verify health
curl http://localhost:8080/health

# View logs
docker-compose -f docker-compose.prod.yml logs -f api
```

---

## Appendix: Environment Variables

### Required

| Variable | Description | Example |
|----------|-------------|---------|
| `MONGODB_URI` | MongoDB connection string | `mongodb://localhost:27017` |
| `JWT_SECRET` | JWT signing secret | `your-secret-key` |
| `JWT_REFRESH_SECRET` | JWT refresh secret | `your-refresh-secret` |
| `APP_ENV` | Application environment | `development` or `production` |

### Optional

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `REDIS_URL` | Redis connection URL | `localhost:6379` |
| `SENDGRID_API_KEY` | SendGrid API key | - |
| `AWS_ACCESS_KEY_ID` | AWS credentials | - |
| `S3_BUCKET_NAME` | S3 bucket for uploads | - |
| `ALLOWED_ORIGINS` | CORS allowed origins | `*` |

---

**Document Version:** 1.0  
**Last Updated:** 2026-02-03  
**Next Review:** End of Phase 2