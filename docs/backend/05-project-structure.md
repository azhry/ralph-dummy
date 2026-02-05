# Project Structure

This document describes the Go project structure following the Standard Go Project Layout and Clean Architecture principles.

## Table of Contents

1. [Overview](#overview)
2. [Directory Structure](#directory-structure)
3. [Component Details](#component-details)
4. [Clean Architecture](#clean-architecture)
5. [File Naming Conventions](#file-naming-conventions)
6. [Package Organization](#package-organization)
7. [Import Organization](#import-organization)
8. [Request Flow Example](#request-flow-example)
9. [Testing Structure](#testing-structure)
10. [Build & Development Workflow](#build--development-workflow)

---

## Overview

This project follows the **Standard Go Project Layout** and implements **Clean Architecture** (also known as Hexagonal Architecture or Ports and Adapters). This ensures:

- **Separation of Concerns**: Each layer has a single responsibility
- **Testability**: Easy to mock dependencies for unit testing
- **Maintainability**: Changes in one layer don't cascade to others
- **Dependency Direction**: Inner layers don't know about outer layers

---

## Directory Structure

```
.
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/                       # Private application code
│   ├── config/
│   │   └── config.go               # Configuration management
│   ├── domain/                     # Domain layer (business entities)
│   │   ├── models/
│   │   │   ├── user.go
│   │   │   └── invitation.go
│   │   └── repository/
│   │       ├── user_repository.go
│   │       └── invitation_repository.go
│   ├── service/                    # Business logic layer
│   │   ├── user_service.go
│   │   └── invitation_service.go
│   ├── repository/
│   │   └── mongodb/                # Database implementations
│   │       ├── user_repository.go
│   │       └── invitation_repository.go
│   ├── handler/                    # HTTP handlers (controllers)
│   │   ├── user_handler.go
│   │   └── invitation_handler.go
│   ├── middleware/                 # HTTP middleware
│   │   ├── auth.go
│   │   ├── cors.go
│   │   └── logging.go
│   ├── dto/                        # Data Transfer Objects
│   │   ├── user_dto.go
│   │   └── invitation_dto.go
│   └── utils/                      # Internal utilities
│       ├── validator.go
│       └── response.go
├── pkg/                            # Public packages (can be imported by other projects)
│   ├── logger/
│   │   └── logger.go
│   └── errors/
│       └── errors.go
├── api/                            # API documentation
│   ├── swagger/
│   │   └── swagger.yaml
│   └── openapi/
│       └── openapi.yaml
├── tests/                          # Test files
│   ├── integration/
│   │   └── api_test.go
│   └── e2e/
│       └── wedding_flow_test.go
├── scripts/                        # Build and deployment scripts
│   ├── build.sh
│   └── migrate.sh
├── docs/                           # Documentation
├── Makefile
├── go.mod
├── go.sum
└── README.md
```

---

## Component Details

### 1. `cmd/api/` - Application Entry Point

**Purpose**: Contains the main application entry point. Follows Go convention where `main` packages are in `cmd/` directory.

**Key Files**:
- `main.go` - Application bootstrap and initialization

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    
    "wedding-invitation/internal/config"
    "wedding-invitation/internal/handler"
    "wedding-invitation/internal/middleware"
    "wedding-invitation/internal/repository/mongodb"
    "wedding-invitation/internal/service"
    "wedding-invitation/pkg/logger"
)

func main() {
    // Initialize logger
    log := logger.NewLogger()
    
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("failed to load config", err)
    }
    
    // Initialize database connection
    db, err := mongodb.NewConnection(cfg.MongoDB.URI)
    if err != nil {
        log.Fatal("failed to connect to database", err)
    }
    defer db.Close()
    
    // Initialize repositories
    userRepo := mongodb.NewUserRepository(db)
    invitationRepo := mongodb.NewInvitationRepository(db)
    
    // Initialize services
    userService := service.NewUserService(userRepo, log)
    invitationService := service.NewInvitationService(invitationRepo, userRepo, log)
    
    // Initialize handlers
    userHandler := handler.NewUserHandler(userService)
    invitationHandler := handler.NewInvitationHandler(invitationService)
    
    // Setup router
    r := gin.New()
    r.Use(middleware.Logger(log))
    r.Use(middleware.Recovery())
    r.Use(middleware.CORS())
    
    // Register routes
    api := r.Group("/api/v1")
    {
        users := api.Group("/users")
        {
            users.POST("/", userHandler.Create)
            users.GET("/:id", userHandler.GetByID)
            users.PUT("/:id", userHandler.Update)
            users.DELETE("/:id", userHandler.Delete)
        }
        
        invitations := api.Group("/invitations")
        {
            invitations.POST("/", invitationHandler.Create)
            invitations.GET("/:id", invitationHandler.GetByID)
            invitations.GET("/user/:userId", invitationHandler.GetByUserID)
        }
    }
    
    // Start server
    srv := &http.Server{
        Addr:    ":" + cfg.Server.Port,
        Handler: r,
    }
    
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal("server failed to start", err)
        }
    }()
    
    log.Info("server started", "port", cfg.Server.Port)
    
    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Info("shutting down server...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("server forced to shutdown", err)
    }
    
    log.Info("server exited")
}
```

---

### 2. `internal/config/` - Configuration Management

**Purpose**: Centralized configuration loading and validation.

**Key Files**:
- `config.go` - Configuration structure and loading logic

```go
package config

import (
    "fmt"
    "os"
    "strconv"
    "time"
)

type Config struct {
    Server   ServerConfig
    MongoDB  MongoDBConfig
    JWT      JWTConfig
    Log      LogConfig
}

type ServerConfig struct {
    Port         string
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
}

type MongoDBConfig struct {
    URI      string
    Database string
}

type JWTConfig struct {
    Secret     string
    Expiration time.Duration
}

type LogConfig struct {
    Level  string
    Format string
}

func Load() (*Config, error) {
    cfg := &Config{
        Server: ServerConfig{
            Port:         getEnv("SERVER_PORT", "8080"),
            ReadTimeout:  getDurationEnv("SERVER_READ_TIMEOUT", 10*time.Second),
            WriteTimeout: getDurationEnv("SERVER_WRITE_TIMEOUT", 10*time.Second),
        },
        MongoDB: MongoDBConfig{
            URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
            Database: getEnv("MONGODB_DATABASE", "wedding_invitation"),
        },
        JWT: JWTConfig{
            Secret:     getEnv("JWT_SECRET", "your-secret-key"),
            Expiration: getDurationEnv("JWT_EXPIRATION", 24*time.Hour),
        },
        Log: LogConfig{
            Level:  getEnv("LOG_LEVEL", "info"),
            Format: getEnv("LOG_FORMAT", "json"),
        },
    }
    
    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }
    
    return cfg, nil
}

func (c *Config) Validate() error {
    if c.MongoDB.URI == "" {
        return fmt.Errorf("mongodb uri is required")
    }
    if c.JWT.Secret == "" {
        return fmt.Errorf("jwt secret is required")
    }
    return nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}
```

---

### 3. `internal/domain/` - Domain Layer

**Purpose**: Contains business entities (models) and repository interfaces. This is the innermost layer - it has no dependencies on other layers.

**Key Files**:
- `models/` - Business entities
- `repository/` - Repository interfaces (ports)

```go
// internal/domain/models/user.go
package models

import (
    "time"
    
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
    ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    Email     string             `json:"email" bson:"email"`
    Password  string             `json:"-" bson:"password"`
    Name      string             `json:"name" bson:"name"`
    CreatedAt time.Time          `json:"created_at" bson:"created_at"`
    UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

func (u *User) BeforeCreate() {
    now := time.Now()
    u.CreatedAt = now
    u.UpdatedAt = now
}

func (u *User) BeforeUpdate() {
    u.UpdatedAt = time.Now()
}
```

```go
// internal/domain/models/invitation.go
package models

import (
    "time"
    
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type RSVPStatus string

const (
    RSVPStatusPending   RSVPStatus = "pending"
    RSVPStatusAttending RSVPStatus = "attending"
    RSVPStatusDeclined  RSVPStatus = "declined"
    RSVPStatusMaybe     RSVPStatus = "maybe"
)

type Invitation struct {
    ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    UserID      primitive.ObjectID `json:"user_id" bson:"user_id"`
    Title       string             `json:"title" bson:"title"`
    EventDate   time.Time          `json:"event_date" bson:"event_date"`
    Venue       string             `json:"venue" bson:"venue"`
    Message     string             `json:"message" bson:"message"`
    RSVPStatus  RSVPStatus         `json:"rsvp_status" bson:"rsvp_status"`
    GuestCount  int                `json:"guest_count" bson:"guest_count"`
    CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
    UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

func (i *Invitation) BeforeCreate() {
    now := time.Now()
    i.CreatedAt = now
    i.UpdatedAt = now
    if i.RSVPStatus == "" {
        i.RSVPStatus = RSVPStatusPending
    }
}
```

```go
// internal/domain/repository/user_repository.go
package repository

import (
    "context"
    
    "go.mongodb.org/mongo-driver/bson/primitive"
    "wedding-invitation/internal/domain/models"
)

// UserRepository defines the interface for user data access.
// This is a port in Clean Architecture - the domain defines what it needs,
// and infrastructure provides the implementation.
type UserRepository interface {
    Create(ctx context.Context, user *models.User) error
    GetByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
    GetByEmail(ctx context.Context, email string) (*models.User, error)
    Update(ctx context.Context, user *models.User) error
    Delete(ctx context.Context, id primitive.ObjectID) error
    Exists(ctx context.Context, email string) (bool, error)
}
```

```go
// internal/domain/repository/invitation_repository.go
package repository

import (
    "context"
    
    "go.mongodb.org/mongo-driver/bson/primitive"
    "wedding-invitation/internal/domain/models"
)

// InvitationRepository defines the interface for invitation data access.
type InvitationRepository interface {
    Create(ctx context.Context, invitation *models.Invitation) error
    GetByID(ctx context.Context, id primitive.ObjectID) (*models.Invitation, error)
    GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.Invitation, error)
    Update(ctx context.Context, invitation *models.Invitation) error
    Delete(ctx context.Context, id primitive.ObjectID) error
    UpdateRSVP(ctx context.Context, id primitive.ObjectID, status models.RSVPStatus, guestCount int) error
}
```

---

### 4. `internal/service/` - Business Logic Layer

**Purpose**: Contains business logic. Orchestrates between domain entities and repositories. Implements use cases.

**Key Files**:
- `user_service.go` - User business logic
- `invitation_service.go` - Invitation business logic

```go
// internal/service/user_service.go
package service

import (
    "context"
    "fmt"
    
    "go.mongodb.org/mongo-driver/bson/primitive"
    "golang.org/x/crypto/bcrypt"
    
    "wedding-invitation/internal/domain/models"
    "wedding-invitation/internal/domain/repository"
    "wedding-invitation/pkg/errors"
    "wedding-invitation/pkg/logger"
)

// UserService implements user-related business logic.
type UserService struct {
    repo   repository.UserRepository
    logger logger.Logger
}

// NewUserService creates a new UserService instance.
func NewUserService(repo repository.UserRepository, log logger.Logger) *UserService {
    return &UserService{
        repo:   repo,
        logger: log,
    }
}

// CreateUserInput contains the data needed to create a user.
type CreateUserInput struct {
    Email    string
    Password string
    Name     string
}

// Create creates a new user with validation and business rules.
func (s *UserService) Create(ctx context.Context, input CreateUserInput) (*models.User, error) {
    // Check if user already exists
    exists, err := s.repo.Exists(ctx, input.Email)
    if err != nil {
        s.logger.Error("failed to check user existence", "error", err)
        return nil, errors.NewInternalError("failed to create user")
    }
    if exists {
        return nil, errors.NewConflictError("user with this email already exists")
    }
    
    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        s.logger.Error("failed to hash password", "error", err)
        return nil, errors.NewInternalError("failed to create user")
    }
    
    // Create user entity
    user := &models.User{
        Email:    input.Email,
        Password: string(hashedPassword),
        Name:     input.Name,
    }
    user.BeforeCreate()
    
    // Persist user
    if err := s.repo.Create(ctx, user); err != nil {
        s.logger.Error("failed to create user", "error", err)
        return nil, errors.NewInternalError("failed to create user")
    }
    
    s.logger.Info("user created", "user_id", user.ID.Hex())
    return user, nil
}

// GetByID retrieves a user by ID.
func (s *UserService) GetByID(ctx context.Context, id string) (*models.User, error) {
    objectID, err := primitive.ObjectIDFromHex(id)
    if err != nil {
        return nil, errors.NewValidationError("invalid user id")
    }
    
    user, err := s.repo.GetByID(ctx, objectID)
    if err != nil {
        if errors.IsNotFound(err) {
            return nil, errors.NewNotFoundError("user not found")
        }
        s.logger.Error("failed to get user", "error", err)
        return nil, errors.NewInternalError("failed to get user")
    }
    
    return user, nil
}

// Authenticate validates user credentials.
func (s *UserService) Authenticate(ctx context.Context, email, password string) (*models.User, error) {
    user, err := s.repo.GetByEmail(ctx, email)
    if err != nil {
        if errors.IsNotFound(err) {
            return nil, errors.NewUnauthorizedError("invalid credentials")
        }
        s.logger.Error("failed to get user by email", "error", err)
        return nil, errors.NewInternalError("authentication failed")
    }
    
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
        return nil, errors.NewUnauthorizedError("invalid credentials")
    }
    
    return user, nil
}
```

```go
// internal/service/invitation_service.go
package service

import (
    "context"
    "fmt"
    "time"
    
    "go.mongodb.org/mongo-driver/bson/primitive"
    
    "wedding-invitation/internal/domain/models"
    "wedding-invitation/internal/domain/repository"
    "wedding-invitation/pkg/errors"
    "wedding-invitation/pkg/logger"
)

// InvitationService implements invitation-related business logic.
type InvitationService struct {
    invitationRepo repository.InvitationRepository
    userRepo       repository.UserRepository
    logger         logger.Logger
}

// NewInvitationService creates a new InvitationService instance.
func NewInvitationService(
    invitationRepo repository.InvitationRepository,
    userRepo repository.UserRepository,
    log logger.Logger,
) *InvitationService {
    return &InvitationService{
        invitationRepo: invitationRepo,
        userRepo:       userRepo,
        logger:         log,
    }
}

// CreateInvitationInput contains the data needed to create an invitation.
type CreateInvitationInput struct {
    UserID    string
    Title     string
    EventDate time.Time
    Venue     string
    Message   string
}

// Create creates a new invitation with validation.
func (s *InvitationService) Create(ctx context.Context, input CreateInvitationInput) (*models.Invitation, error) {
    // Validate user exists
    userID, err := primitive.ObjectIDFromHex(input.UserID)
    if err != nil {
        return nil, errors.NewValidationError("invalid user id")
    }
    
    _, err = s.userRepo.GetByID(ctx, userID)
    if err != nil {
        if errors.IsNotFound(err) {
            return nil, errors.NewNotFoundError("user not found")
        }
        s.logger.Error("failed to get user", "error", err)
        return nil, errors.NewInternalError("failed to create invitation")
    }
    
    // Business rule: Event date must be in the future
    if input.EventDate.Before(time.Now()) {
        return nil, errors.NewValidationError("event date must be in the future")
    }
    
    // Create invitation entity
    invitation := &models.Invitation{
        UserID:    userID,
        Title:     input.Title,
        EventDate: input.EventDate,
        Venue:     input.Venue,
        Message:   input.Message,
    }
    invitation.BeforeCreate()
    
    // Persist invitation
    if err := s.invitationRepo.Create(ctx, invitation); err != nil {
        s.logger.Error("failed to create invitation", "error", err)
        return nil, errors.NewInternalError("failed to create invitation")
    }
    
    s.logger.Info("invitation created", 
        "invitation_id", invitation.ID.Hex(),
        "user_id", userID.Hex())
    
    return invitation, nil
}

// UpdateRSVP updates the RSVP status for an invitation.
func (s *InvitationService) UpdateRSVP(
    ctx context.Context,
    invitationID string,
    status models.RSVPStatus,
    guestCount int,
) error {
    objectID, err := primitive.ObjectIDFromHex(invitationID)
    if err != nil {
        return errors.NewValidationError("invalid invitation id")
    }
    
    // Validate status
    if !isValidRSVPStatus(status) {
        return errors.NewValidationError("invalid rsvp status")
    }
    
    // Business rule: If attending, guest count must be positive
    if status == models.RSVPStatusAttending && guestCount <= 0 {
        return errors.NewValidationError("guest count must be positive when attending")
    }
    
    if err := s.invitationRepo.UpdateRSVP(ctx, objectID, status, guestCount); err != nil {
        if errors.IsNotFound(err) {
            return errors.NewNotFoundError("invitation not found")
        }
        s.logger.Error("failed to update rsvp", "error", err)
        return errors.NewInternalError("failed to update rsvp")
    }
    
    s.logger.Info("rsvp updated", 
        "invitation_id", invitationID,
        "status", status,
        "guest_count", guestCount)
    
    return nil
}

func isValidRSVPStatus(status models.RSVPStatus) bool {
    switch status {
    case models.RSVPStatusPending, 
         models.RSVPStatusAttending, 
         models.RSVPStatusDeclined, 
         models.RSVPStatusMaybe:
        return true
    }
    return false
}
```

---

### 5. `internal/repository/mongodb/` - Data Access Layer

**Purpose**: Implements repository interfaces using MongoDB. This is an adapter in Clean Architecture.

**Key Files**:
- `user_repository.go` - MongoDB user repository implementation
- `invitation_repository.go` - MongoDB invitation repository implementation
- `connection.go` - Database connection management

```go
// internal/repository/mongodb/connection.go
package mongodb

import (
    "context"
    "fmt"
    "time"
    
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type Connection struct {
    client   *mongo.Client
    database *mongo.Database
}

func NewConnection(uri string) (*Connection, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
    if err != nil {
        return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
    }
    
    // Verify connection
    if err := client.Ping(ctx, nil); err != nil {
        return nil, fmt.Errorf("failed to ping mongodb: %w", err)
    }
    
    return &Connection{
        client:   client,
        database: client.Database("wedding_invitation"),
    }, nil
}

func (c *Connection) Close() error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    return c.client.Disconnect(ctx)
}

func (c *Connection) Database() *mongo.Database {
    return c.database
}
```

```go
// internal/repository/mongodb/user_repository.go
package mongodb

import (
    "context"
    "errors"
    
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    
    "wedding-invitation/internal/domain/models"
    customerrors "wedding-invitation/pkg/errors"
)

// UserRepository implements repository.UserRepository using MongoDB.
type UserRepository struct {
    collection *mongo.Collection
}

// NewUserRepository creates a new UserRepository instance.
func NewUserRepository(conn *Connection) *UserRepository {
    return &UserRepository{
        collection: conn.Database().Collection("users"),
    }
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
    result, err := r.collection.InsertOne(ctx, user)
    if err != nil {
        return fmt.Errorf("failed to insert user: %w", err)
    }
    
    user.ID = result.InsertedID.(primitive.ObjectID)
    return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
    var user models.User
    err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
    if err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            return nil, customerrors.ErrNotFound
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
    var user models.User
    err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
    if err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            return nil, customerrors.ErrNotFound
        }
        return nil, fmt.Errorf("failed to get user by email: %w", err)
    }
    return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
    user.BeforeUpdate()
    
    result, err := r.collection.UpdateOne(
        ctx,
        bson.M{"_id": user.ID},
        bson.M{"$set": user},
    )
    if err != nil {
        return fmt.Errorf("failed to update user: %w", err)
    }
    
    if result.MatchedCount == 0 {
        return customerrors.ErrNotFound
    }
    
    return nil
}

func (r *UserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
    result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
    if err != nil {
        return fmt.Errorf("failed to delete user: %w", err)
    }
    
    if result.DeletedCount == 0 {
        return customerrors.ErrNotFound
    }
    
    return nil
}

func (r *UserRepository) Exists(ctx context.Context, email string) (bool, error) {
    count, err := r.collection.CountDocuments(ctx, bson.M{"email": email})
    if err != nil {
        return false, fmt.Errorf("failed to check user existence: %w", err)
    }
    return count > 0, nil
}
```

```go
// internal/repository/mongodb/invitation_repository.go
package mongodb

import (
    "context"
    "errors"
    "fmt"
    
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    
    "wedding-invitation/internal/domain/models"
    customerrors "wedding-invitation/pkg/errors"
)

// InvitationRepository implements repository.InvitationRepository using MongoDB.
type InvitationRepository struct {
    collection *mongo.Collection
}

// NewInvitationRepository creates a new InvitationRepository instance.
func NewInvitationRepository(conn *Connection) *InvitationRepository {
    return &InvitationRepository{
        collection: conn.Database().Collection("invitations"),
    }
}

func (r *InvitationRepository) Create(ctx context.Context, invitation *models.Invitation) error {
    result, err := r.collection.InsertOne(ctx, invitation)
    if err != nil {
        return fmt.Errorf("failed to insert invitation: %w", err)
    }
    
    invitation.ID = result.InsertedID.(primitive.ObjectID)
    return nil
}

func (r *InvitationRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Invitation, error) {
    var invitation models.Invitation
    err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&invitation)
    if err != nil {
        if errors.Is(err, mongo.ErrNoDocuments) {
            return nil, customerrors.ErrNotFound
        }
        return nil, fmt.Errorf("failed to get invitation: %w", err)
    }
    return &invitation, nil
}

func (r *InvitationRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID) ([]*models.Invitation, error) {
    cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, options.Find().SetSort(bson.M{"created_at": -1}))
    if err != nil {
        return nil, fmt.Errorf("failed to find invitations: %w", err)
    }
    defer cursor.Close(ctx)
    
    var invitations []*models.Invitation
    if err := cursor.All(ctx, &invitations); err != nil {
        return nil, fmt.Errorf("failed to decode invitations: %w", err)
    }
    
    return invitations, nil
}

func (r *InvitationRepository) Update(ctx context.Context, invitation *models.Invitation) error {
    invitation.BeforeUpdate()
    
    result, err := r.collection.UpdateOne(
        ctx,
        bson.M{"_id": invitation.ID},
        bson.M{"$set": invitation},
    )
    if err != nil {
        return fmt.Errorf("failed to update invitation: %w", err)
    }
    
    if result.MatchedCount == 0 {
        return customerrors.ErrNotFound
    }
    
    return nil
}

func (r *InvitationRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
    result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
    if err != nil {
        return fmt.Errorf("failed to delete invitation: %w", err)
    }
    
    if result.DeletedCount == 0 {
        return customerrors.ErrNotFound
    }
    
    return nil
}

func (r *InvitationRepository) UpdateRSVP(ctx context.Context, id primitive.ObjectID, status models.RSVPStatus, guestCount int) error {
    result, err := r.collection.UpdateOne(
        ctx,
        bson.M{"_id": id},
        bson.M{
            "$set": bson.M{
                "rsvp_status": status,
                "guest_count": guestCount,
                "updated_at":  time.Now(),
            },
        },
    )
    if err != nil {
        return fmt.Errorf("failed to update rsvp: %w", err)
    }
    
    if result.MatchedCount == 0 {
        return customerrors.ErrNotFound
    }
    
    return nil
}
```

---

### 6. `internal/handler/` - HTTP Handlers

**Purpose**: Handles HTTP requests and responses. Converts HTTP input to service input and service output to HTTP responses.

**Key Files**:
- `user_handler.go` - User HTTP handlers
- `invitation_handler.go` - Invitation HTTP handlers

```go
// internal/handler/user_handler.go
package handler

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    
    "wedding-invitation/internal/dto"
    "wedding-invitation/internal/service"
    "wedding-invitation/internal/utils"
    "wedding-invitation/pkg/errors"
)

// UserHandler handles HTTP requests for user operations.
type UserHandler struct {
    service *service.UserService
}

// NewUserHandler creates a new UserHandler instance.
func NewUserHandler(service *service.UserService) *UserHandler {
    return &UserHandler{service: service}
}

// Create handles POST /users
func (h *UserHandler) Create(c *gin.Context) {
    var req dto.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.RespondError(c, errors.NewValidationError(err.Error()))
        return
    }
    
    input := service.CreateUserInput{
        Email:    req.Email,
        Password: req.Password,
        Name:     req.Name,
    }
    
    user, err := h.service.Create(c.Request.Context(), input)
    if err != nil {
        utils.RespondError(c, err)
        return
    }
    
    utils.RespondJSON(c, http.StatusCreated, dto.UserToResponse(user))
}

// GetByID handles GET /users/:id
func (h *UserHandler) GetByID(c *gin.Context) {
    id := c.Param("id")
    
    user, err := h.service.GetByID(c.Request.Context(), id)
    if err != nil {
        utils.RespondError(c, err)
        return
    }
    
    utils.RespondJSON(c, http.StatusOK, dto.UserToResponse(user))
}
```

```go
// internal/handler/invitation_handler.go
package handler

import (
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    
    "wedding-invitation/internal/domain/models"
    "wedding-invitation/internal/dto"
    "wedding-invitation/internal/service"
    "wedding-invitation/internal/utils"
    "wedding-invitation/pkg/errors"
)

// InvitationHandler handles HTTP requests for invitation operations.
type InvitationHandler struct {
    service *service.InvitationService
}

// NewInvitationHandler creates a new InvitationHandler instance.
func NewInvitationHandler(service *service.InvitationService) *InvitationHandler {
    return &InvitationHandler{service: service}
}

// Create handles POST /invitations
func (h *InvitationHandler) Create(c *gin.Context) {
    var req dto.CreateInvitationRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.RespondError(c, errors.NewValidationError(err.Error()))
        return
    }
    
    eventDate, err := time.Parse(time.RFC3339, req.EventDate)
    if err != nil {
        utils.RespondError(c, errors.NewValidationError("invalid event date format"))
        return
    }
    
    input := service.CreateInvitationInput{
        UserID:    req.UserID,
        Title:     req.Title,
        EventDate: eventDate,
        Venue:     req.Venue,
        Message:   req.Message,
    }
    
    invitation, err := h.service.Create(c.Request.Context(), input)
    if err != nil {
        utils.RespondError(c, err)
        return
    }
    
    utils.RespondJSON(c, http.StatusCreated, dto.InvitationToResponse(invitation))
}

// GetByID handles GET /invitations/:id
func (h *InvitationHandler) GetByID(c *gin.Context) {
    id := c.Param("id")
    
    invitation, err := h.service.GetByID(c.Request.Context(), id)
    if err != nil {
        utils.RespondError(c, err)
        return
    }
    
    utils.RespondJSON(c, http.StatusOK, dto.InvitationToResponse(invitation))
}

// UpdateRSVP handles PUT /invitations/:id/rsvp
func (h *InvitationHandler) UpdateRSVP(c *gin.Context) {
    id := c.Param("id")
    
    var req dto.UpdateRSVPRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.RespondError(c, errors.NewValidationError(err.Error()))
        return
    }
    
    if err := h.service.UpdateRSVP(c.Request.Context(), id, models.RSVPStatus(req.Status), req.GuestCount); err != nil {
        utils.RespondError(c, err)
        return
    }
    
    utils.RespondJSON(c, http.StatusOK, gin.H{"message": "rsvp updated successfully"})
}
```

---

### 7. `internal/middleware/` - HTTP Middleware

**Purpose**: Cross-cutting concerns like authentication, logging, CORS, recovery from panics.

**Key Files**:
- `auth.go` - Authentication middleware
- `cors.go` - CORS middleware
- `logging.go` - Request logging middleware

```go
// internal/middleware/auth.go
package middleware

import (
    "net/http"
    "strings"
    
    "github.com/gin-gonic/gin"
    
    "wedding-invitation/internal/utils"
    "wedding-invitation/pkg/errors"
)

// Auth creates an authentication middleware.
func Auth(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            utils.RespondError(c, errors.NewUnauthorizedError("authorization header required"))
            c.Abort()
            return
        }
        
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
            utils.RespondError(c, errors.NewUnauthorizedError("invalid authorization header format"))
            c.Abort()
            return
        }
        
        tokenString := parts[1]
        claims, err := utils.ValidateJWT(tokenString, jwtSecret)
        if err != nil {
            utils.RespondError(c, errors.NewUnauthorizedError("invalid token"))
            c.Abort()
            return
        }
        
        // Store user info in context
        c.Set("user_id", claims.UserID)
        c.Set("email", claims.Email)
        
        c.Next()
    }
}
```

```go
// internal/middleware/cors.go
package middleware

import (
    "time"
    
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

// CORS creates a CORS middleware with default configuration.
func CORS() gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins:     []string{"*"}, // Configure for production
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    })
}
```

```go
// internal/middleware/logging.go
package middleware

import (
    "time"
    
    "github.com/gin-gonic/gin"
    
    "wedding-invitation/pkg/logger"
)

// Logger creates a request logging middleware.
func Logger(log logger.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        raw := c.Request.URL.RawQuery
        
        if raw != "" {
            path = path + "?" + raw
        }
        
        c.Next()
        
        latency := time.Since(start)
        status := c.Writer.Status()
        
        log.Info("request",
            "method", c.Request.Method,
            "path", path,
            "status", status,
            "latency", latency,
            "client_ip", c.ClientIP(),
        )
    }
}

// Recovery recovers from panics and logs them.
func Recovery() gin.HandlerFunc {
    return gin.Recovery()
}
```

---

### 8. `internal/dto/` - Data Transfer Objects

**Purpose**: Defines request and response structures. Separates API contracts from domain models.

**Key Files**:
- `user_dto.go` - User-related DTOs
- `invitation_dto.go` - Invitation-related DTOs

```go
// internal/dto/user_dto.go
package dto

import (
    "wedding-invitation/internal/domain/models"
)

// Request DTOs

type CreateUserRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
    Name     string `json:"name" binding:"required"`
}

type UpdateUserRequest struct {
    Name string `json:"name" binding:"omitempty"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

// Response DTOs

type UserResponse struct {
    ID        string `json:"id"`
    Email     string `json:"email"`
    Name      string `json:"name"`
    CreatedAt string `json:"created_at"`
}

// Conversion functions

func UserToResponse(user *models.User) UserResponse {
    return UserResponse{
        ID:        user.ID.Hex(),
        Email:     user.Email,
        Name:      user.Name,
        CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
    }
}

func UsersToResponse(users []*models.User) []UserResponse {
    responses := make([]UserResponse, len(users))
    for i, user := range users {
        responses[i] = UserToResponse(user)
    }
    return responses
}
```

```go
// internal/dto/invitation_dto.go
package dto

import (
    "wedding-invitation/internal/domain/models"
)

// Request DTOs

type CreateInvitationRequest struct {
    UserID    string `json:"user_id" binding:"required"`
    Title     string `json:"title" binding:"required"`
    EventDate string `json:"event_date" binding:"required"` // RFC3339 format
    Venue     string `json:"venue" binding:"required"`
    Message   string `json:"message" binding:"omitempty"`
}

type UpdateInvitationRequest struct {
    Title     string `json:"title" binding:"omitempty"`
    EventDate string `json:"event_date" binding:"omitempty"`
    Venue     string `json:"venue" binding:"omitempty"`
    Message   string `json:"message" binding:"omitempty"`
}

type UpdateRSVPRequest struct {
    Status     string `json:"status" binding:"required,oneof=pending attending declined maybe"`
    GuestCount int    `json:"guest_count" binding:"min=0"`
}

// Response DTOs

type InvitationResponse struct {
    ID         string `json:"id"`
    UserID     string `json:"user_id"`
    Title      string `json:"title"`
    EventDate  string `json:"event_date"`
    Venue      string `json:"venue"`
    Message    string `json:"message"`
    RSVPStatus string `json:"rsvp_status"`
    GuestCount int    `json:"guest_count"`
    CreatedAt  string `json:"created_at"`
}

// Conversion functions

func InvitationToResponse(invitation *models.Invitation) InvitationResponse {
    return InvitationResponse{
        ID:         invitation.ID.Hex(),
        UserID:     invitation.UserID.Hex(),
        Title:      invitation.Title,
        EventDate:  invitation.EventDate.Format("2006-01-02T15:04:05Z"),
        Venue:      invitation.Venue,
        Message:    invitation.Message,
        RSVPStatus: string(invitation.RSVPStatus),
        GuestCount: invitation.GuestCount,
        CreatedAt:  invitation.CreatedAt.Format("2006-01-02T15:04:05Z"),
    }
}

func InvitationsToResponse(invitations []*models.Invitation) []InvitationResponse {
    responses := make([]InvitationResponse, len(invitations))
    for i, invitation := range invitations {
        responses[i] = InvitationToResponse(invitation)
    }
    return responses
}
```

---

### 9. `internal/utils/` - Internal Utilities

**Purpose**: Helper functions and utilities used within the internal package.

**Key Files**:
- `response.go` - HTTP response helpers
- `validator.go` - Custom validators

```go
// internal/utils/response.go
package utils

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
    
    "wedding-invitation/pkg/errors"
)

// RespondJSON sends a JSON response with the given status code.
func RespondJSON(c *gin.Context, status int, data interface{}) {
    c.JSON(status, data)
}

// RespondError sends an error response.
func RespondError(c *gin.Context, err error) {
    var status int
    var message string
    
    switch {
    case errors.IsValidationError(err):
        status = http.StatusBadRequest
        message = err.Error()
    case errors.IsNotFound(err):
        status = http.StatusNotFound
        message = err.Error()
    case errors.IsUnauthorized(err):
        status = http.StatusUnauthorized
        message = err.Error()
    case errors.IsConflict(err):
        status = http.StatusConflict
        message = err.Error()
    default:
        status = http.StatusInternalServerError
        message = "internal server error"
    }
    
    c.JSON(status, gin.H{
        "error": message,
    })
}
```

---

### 10. `pkg/` - Public Packages

**Purpose**: Reusable packages that can be imported by other projects. These are domain-agnostic.

**Key Files**:
- `logger/logger.go` - Logging interface and implementations
- `errors/errors.go` - Custom error types

```go
// pkg/logger/logger.go
package logger

import (
    "log/slog"
    "os"
)

// Logger interface abstracts logging operations.
type Logger interface {
    Debug(msg string, args ...interface{})
    Info(msg string, args ...interface{})
    Warn(msg string, args ...interface{})
    Error(msg string, args ...interface{})
    Fatal(msg string, args ...interface{})
}

// SlogLogger implements Logger using slog.
type SlogLogger struct {
    logger *slog.Logger
}

// NewLogger creates a new Logger instance.
func NewLogger() Logger {
    opts := &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }
    handler := slog.NewJSONHandler(os.Stdout, opts)
    
    return &SlogLogger{
        logger: slog.New(handler),
    }
}

func (l *SlogLogger) Debug(msg string, args ...interface{}) {
    l.logger.Debug(msg, args...)
}

func (l *SlogLogger) Info(msg string, args ...interface{}) {
    l.logger.Info(msg, args...)
}

func (l *SlogLogger) Warn(msg string, args ...interface{}) {
    l.logger.Warn(msg, args...)
}

func (l *SlogLogger) Error(msg string, args ...interface{}) {
    l.logger.Error(msg, args...)
}

func (l *SlogLogger) Fatal(msg string, args ...interface{}) {
    l.logger.Error(msg, args...)
    os.Exit(1)
}
```

```go
// pkg/errors/errors.go
package errors

import "errors"

var (
    ErrNotFound = errors.New("not found")
)

// AppError represents an application error.
type AppError struct {
    Type    ErrorType
    Message string
    Err     error
}

type ErrorType string

const (
    ValidationError  ErrorType = "validation"
    NotFoundError    ErrorType = "not_found"
    UnauthorizedError ErrorType = "unauthorized"
    ConflictError    ErrorType = "conflict"
    InternalError    ErrorType = "internal"
)

func (e *AppError) Error() string {
    if e.Err != nil {
        return e.Message + ": " + e.Err.Error()
    }
    return e.Message
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// Error constructors

func NewValidationError(msg string) error {
    return &AppError{Type: ValidationError, Message: msg}
}

func NewNotFoundError(msg string) error {
    return &AppError{Type: NotFoundError, Message: msg}
}

func NewUnauthorizedError(msg string) error {
    return &AppError{Type: UnauthorizedError, Message: msg}
}

func NewConflictError(msg string) error {
    return &AppError{Type: ConflictError, Message: msg}
}

func NewInternalError(msg string) error {
    return &AppError{Type: InternalError, Message: msg}
}

// Error checkers

func IsValidationError(err error) bool {
    var appErr *AppError
    if errors.As(err, &appErr) {
        return appErr.Type == ValidationError
    }
    return false
}

func IsNotFound(err error) bool {
    if errors.Is(err, ErrNotFound) {
        return true
    }
    var appErr *AppError
    if errors.As(err, &appErr) {
        return appErr.Type == NotFoundError
    }
    return false
}

func IsUnauthorized(err error) bool {
    var appErr *AppError
    if errors.As(err, &appErr) {
        return appErr.Type == UnauthorizedError
    }
    return false
}

func IsConflict(err error) bool {
    var appErr *AppError
    if errors.As(err, &appErr) {
        return appErr.Type == ConflictError
    }
    return false
}
```

---

## Clean Architecture

### Dependency Direction

Clean Architecture enforces a strict dependency rule: **dependencies can only point inward**.

```
┌─────────────────────────────────────┐
│           cmd/api/                  │  (Outer Layer)
│        Infrastructure               │
│   - HTTP Server                     │
│   - Database Connection             │
│   - Framework Setup                 │
└─────────────┬───────────────────────┘
              │ depends on
              ▼
┌─────────────────────────────────────┐
│        internal/handler/            │
│      Interface Adapters             │
│   - HTTP Handlers                   │
│   - Middleware                      │
│   - DTOs                            │
└─────────────┬───────────────────────┘
              │ depends on
              ▼
┌─────────────────────────────────────┐
│        internal/service/            │
│     Application Business Rules      │
│   - Use Cases                       │
│   - Business Logic                  │
│   - Orchestration                   │
└─────────────┬───────────────────────┘
              │ depends on
              ▼
┌─────────────────────────────────────┐
│     internal/domain/models/         │
│      Enterprise Business Rules        │
│   - Domain Entities                 │
│   - Value Objects                   │
└─────────────┬───────────────────────┘
              │ implements
              ▼
┌─────────────────────────────────────┐
│  internal/domain/repository/        │
│       Repository Interfaces         │
│   - Ports (driven)                  │
└─────────────────────────────────────┘
              ▲
              │ implements
┌─────────────────────────────────────┐
│   internal/repository/mongodb/      │
│    Database Implementation          │
│   - Adapters (driven)               │
│   - MongoDB specific code           │
└─────────────────────────────────────┘
```

### Key Principles

1. **Independence of Frameworks**: The domain layer doesn't know about HTTP, MongoDB, or any framework
2. **Testability**: Business rules can be tested without external dependencies
3. **Independence of UI**: The UI can change without affecting business rules
4. **Independence of Database**: Database can be swapped without affecting business rules
5. **Independence of External Services**: External services are accessed through interfaces

---

## File Naming Conventions

### Standard Go Conventions

| Pattern | Description | Example |
|---------|-------------|---------|
| `snake_case.go` | Regular source files | `user_service.go` |
| `_test.go` | Test files | `user_service_test.go` |
| `_mock.go` | Mock implementations | `user_repository_mock.go` |
| `doc.go` | Package documentation | `doc.go` |
| `types.go` | Type definitions | `types.go` |
| `interfaces.go` | Interface definitions | `interfaces.go` |

### Naming Rules

```go
// Files: lowercase with underscores
user_repository.go          // good
UserRepository.go         // bad
userRepository.go         // bad
userrepo.go               // bad

// Packages: lowercase, no underscores (unless necessary)
package mongodb           // good
package mongo_db          // bad
package mongoDb           // bad

// Types: PascalCase
UserRepository            // good
userRepository            // bad
user_repository           // bad

// Interfaces: PascalCase with descriptive names
Repository                // good if clear from context
UserRepository            // good for clarity
IUserRepository           // bad (Go convention)

// Functions: camelCase
getUserByID               // good
GetUserByID               // good if exported
get_user_by_id            // bad

// Variables: camelCase
userID                    // good
userId                    // acceptable
user_id                   // bad
UserID                    // bad (only for exported)

// Constants: UPPER_SNAKE_CASE for exported, camelCase for unexported
MaxRetries                // good (exported)
maxRetries                // good (unexported)
MAX_RETRIES               // bad in Go

// Acronyms: All caps
HTTPClient                // good
HttpClient                // bad
userID                    // good
userId                    // acceptable
```

---

## Package Organization

### Best Practices

```go
// 1. One package per directory
// internal/service/       -> package service
// internal/handler/       -> package handler

// 2. Clear package naming
package service           // good - matches directory
package svc               // bad - abbreviated

// 3. Package documentation
// File: internal/service/doc.go
// Package service implements business logic and use cases for the wedding
// invitation application.
package service

// 4. Avoid package name collisions
// Use import aliases when necessary
import (
    domainRepo "wedding-invitation/internal/domain/repository"
    mongoRepo "wedding-invitation/internal/repository/mongodb"
)

// 5. Organize by functionality, not type
// Good:
// internal/
//   ├── user/
//   │   ├── handler.go
//   │   ├── service.go
//   │   └── repository.go
//   └── invitation/
//       ├── handler.go
//       ├── service.go
//       └── repository.go

// Better (our approach - separation of concerns):
// internal/
//   ├── handler/
//   │   ├── user_handler.go
//   │   └── invitation_handler.go
//   ├── service/
//   │   ├── user_service.go
//   │   └── invitation_service.go
//   └── repository/
//       ├── mongodb/
//       │   ├── user_repository.go
//       │   └── invitation_repository.go

// 6. Keep packages small and focused
// - Single responsibility per package
// - Clear public API (exported names)
// - Hide implementation details
```

---

## Import Organization

### Standard Format

```go
package main

import (
    // 1. Standard library packages
    "context"
    "fmt"
    "net/http"
    "time"
    
    // 2. Third-party packages (blank line separates)
    "github.com/gin-gonic/gin"
    "go.mongodb.org/mongo-driver/mongo"
    
    // 3. Project packages (blank line separates)
    "wedding-invitation/internal/config"
    "wedding-invitation/internal/handler"
    "wedding-invitation/pkg/logger"
)
```

### Import Aliases

```go
import (
    // Standard library - no aliases needed
    "time"
    
    // Third-party - use alias if name is unclear or conflicts
    "github.com/google/uuid"
    gonanoid "github.com/matoous/go-nanoid"  // alias for clarity
    
    // Internal - use aliases to avoid name collisions
    domainModels "wedding-invitation/internal/domain/models"
    domainRepo "wedding-invitation/internal/domain/repository"
    mongoRepo "wedding-invitation/internal/repository/mongodb"
    
    // pkg - usually no alias needed
    "wedding-invitation/pkg/errors"
    "wedding-invitation/pkg/logger"
)
```

### goimports Configuration

Use `goimports` to automatically format imports:

```bash
# Install
go install golang.org/x/tools/cmd/goimports@latest

# Run on file
goimports -w internal/service/user_service.go

# Run on all files
goimports -w .
```

### Tools Configuration

Add to your editor/IDE:

```json
// VS Code settings.json
{
    "gopls": {
        "formatting.gofumpt": true,
        "ui.diagnostic.annotations": {
            "bounds": true,
            "escape": true,
            "inline": true,
            "nil": true
        }
    },
    "go.formatTool": "goimports"
}
```

---

## Request Flow Example

Let's trace a complete request through all layers: **Creating a New Invitation**

### 1. HTTP Handler (Interface Adapter)

```go
// HTTP Request: POST /api/v1/invitations
// Body: {"user_id": "...", "title": "Wedding", "event_date": "2024-12-25T18:00:00Z", "venue": "Grand Hotel"}

// internal/handler/invitation_handler.go
func (h *InvitationHandler) Create(c *gin.Context) {
    // 1. Parse and validate HTTP request
    var req dto.CreateInvitationRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.RespondError(c, errors.NewValidationError(err.Error()))
        return
    }
    
    // 2. Transform DTO to service input
    eventDate, err := time.Parse(time.RFC3339, req.EventDate)
    if err != nil {
        utils.RespondError(c, errors.NewValidationError("invalid event date format"))
        return
    }
    
    input := service.CreateInvitationInput{
        UserID:    req.UserID,
        Title:     req.Title,
        EventDate: eventDate,
        Venue:     req.Venue,
        Message:   req.Message,
    }
    
    // 3. Call service layer
    invitation, err := h.service.Create(c.Request.Context(), input)
    if err != nil {
        utils.RespondError(c, err)
        return
    }
    
    // 4. Transform domain model to response DTO
    response := dto.InvitationToResponse(invitation)
    
    // 5. Send HTTP response
    utils.RespondJSON(c, http.StatusCreated, response)
}
```

### 2. Service Layer (Use Case)

```go
// internal/service/invitation_service.go
func (s *InvitationService) Create(ctx context.Context, input CreateInvitationInput) (*models.Invitation, error) {
    // 1. Validate business rules
    userID, err := primitive.ObjectIDFromHex(input.UserID)
    if err != nil {
        return nil, errors.NewValidationError("invalid user id")
    }
    
    // 2. Check dependencies (user must exist)
    _, err = s.userRepo.GetByID(ctx, userID)
    if err != nil {
        if errors.IsNotFound(err) {
            return nil, errors.NewNotFoundError("user not found")
        }
        return nil, errors.NewInternalError("failed to create invitation")
    }
    
    // 3. Apply business logic (event date in future)
    if input.EventDate.Before(time.Now()) {
        return nil, errors.NewValidationError("event date must be in the future")
    }
    
    // 4. Create domain entity
    invitation := &models.Invitation{
        UserID:    userID,
        Title:     input.Title,
        EventDate: input.EventDate,
        Venue:     input.Venue,
        Message:   input.Message,
    }
    invitation.BeforeCreate() // Set timestamps
    
    // 5. Persist through repository
    if err := s.invitationRepo.Create(ctx, invitation); err != nil {
        s.logger.Error("failed to create invitation", "error", err)
        return nil, errors.NewInternalError("failed to create invitation")
    }
    
    // 6. Return domain entity
    return invitation, nil
}
```

### 3. Repository Layer (Data Access)

```go
// internal/repository/mongodb/invitation_repository.go
func (r *InvitationRepository) Create(ctx context.Context, invitation *models.Invitation) error {
    // 1. Insert into MongoDB
    result, err := r.collection.InsertOne(ctx, invitation)
    if err != nil {
        return fmt.Errorf("failed to insert invitation: %w", err)
    }
    
    // 2. Set the generated ID on the entity
    invitation.ID = result.InsertedID.(primitive.ObjectID)
    
    return nil
}
```

### 4. Data Transformation Flow

```
HTTP Request (JSON)
       │
       ▼
┌──────────────────┐
│   DTO (Request)  │  dto.CreateInvitationRequest
│  {               │
│    "user_id":    │
│    "title":      │
│    "event_date": │
│    "venue":      │
│  }               │
└────────┬─────────┘
         │ Parse + Validation
         ▼
┌──────────────────┐
│  Service Input   │  service.CreateInvitationInput
│  (Parsed date)   │  time.Time for EventDate
└────────┬─────────┘
         │ Business Logic
         ▼
┌──────────────────┐
│  Domain Model    │  models.Invitation
│  (Entity)        │  ObjectID fields
│                  │  Validation hooks
└────────┬─────────┘
         │ Repository
         ▼
┌──────────────────┐
│  Database        │  MongoDB Document
│  (BSON)          │  _id: ObjectId
└────────┬─────────┘
         │ Response
         ▼
┌──────────────────┐
│  Domain Model    │  models.Invitation (with ID)
│  (Entity)        │
└────────┬─────────┘
         │ Transform
         ▼
┌──────────────────┐
│   DTO (Response) │  dto.InvitationResponse
│  {               │  string IDs (hex)
│    "id":         │  formatted dates
│    "user_id":    │
│    "title":      │
│  }               │
└────────┬─────────┘
         │ JSON Marshal
         ▼
HTTP Response (JSON)
```

### 5. Complete Flow Diagram

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Client     │────▶│   Handler    │────▶│   Service    │────▶│  Repository  │
│  (Browser)   │     │  (HTTP/Gin)  │     │ (Business)   │     │   (MongoDB)  │
└──────────────┘     └──────────────┘     └──────────────┘     └──────────────┘
                            │                    │                    │
                            ▼                    ▼                    ▼
                     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
                     │  Bind JSON   │     │   Validate   │     │   Execute    │
                     │   to DTO     │     │   Rules      │     │    Query     │
                     └──────────────┘     └──────────────┘     └──────────────┘
                            │                    │                    │
                            ▼                    ▼                    ▼
                     ┌──────────────┐     ┌──────────────┐     ┌──────────────┐
                     │  Transform   │     │  Create      │     │   Return     │
                     │  to Input    │     │  Entity      │     │    Entity    │
                     └──────────────┘     └──────────────┘     └──────────────┘
                                                 │
                                                 ▼
                                          ┌──────────────┐
                                          │   Call       │
                                          │ Repository   │
                                          └──────────────┘
```

---

## Testing Structure

### Test Organization

```
tests/
├── unit/                      # Unit tests (no external dependencies)
│   ├── service/
│   │   ├── user_service_test.go
│   │   └── invitation_service_test.go
│   └── domain/
│       └── models/
│           └── user_test.go
├── integration/               # Integration tests (with database)
│   ├── repository/
│   │   ├── user_repository_test.go
│   │   └── invitation_repository_test.go
│   └── api/
│       └── handlers_test.go
└── e2e/                      # End-to-end tests
    └── wedding_flow_test.go
```

### Unit Test Example (Service Layer)

```go
// tests/unit/service/user_service_test.go
package service_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    
    "wedding-invitation/internal/domain/models"
    "wedding-invitation/internal/domain/repository"
    "wedding-invitation/internal/service"
    mockRepo "wedding-invitation/tests/mocks/repository"
    mockLogger "wedding-invitation/tests/mocks/logger"
)

func TestUserService_Create(t *testing.T) {
    // Arrange
    ctx := context.Background()
    
    mockUserRepo := new(mockRepo.UserRepository)
    mockLog := new(mockLogger.Logger)
    
    svc := service.NewUserService(mockUserRepo, mockLog)
    
    input := service.CreateUserInput{
        Email:    "test@example.com",
        Password: "password123",
        Name:     "John Doe",
    }
    
    // Mock expectations
    mockUserRepo.On("Exists", ctx, input.Email).Return(false, nil)
    mockUserRepo.On("Create", ctx, mock.AnythingOfType("*models.User")).
        Run(func(args mock.Arguments) {
            user := args.Get(1).(*models.User)
            user.ID = primitive.NewObjectID()
        }).
        Return(nil)
    mockLog.On("Info", mock.Anything, mock.Anything, mock.Anything)
    
    // Act
    user, err := svc.Create(ctx, input)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, input.Email, user.Email)
    assert.Equal(t, input.Name, user.Name)
    assert.NotEmpty(t, user.ID)
    
    mockUserRepo.AssertExpectations(t)
}

func TestUserService_Create_DuplicateEmail(t *testing.T) {
    // Arrange
    ctx := context.Background()
    
    mockUserRepo := new(mockRepo.UserRepository)
    mockLog := new(mockLogger.Logger)
    
    svc := service.NewUserService(mockUserRepo, mockLog)
    
    input := service.CreateUserInput{
        Email:    "existing@example.com",
        Password: "password123",
        Name:     "John Doe",
    }
    
    mockUserRepo.On("Exists", ctx, input.Email).Return(true, nil)
    
    // Act
    user, err := svc.Create(ctx, input)
    
    // Assert
    assert.Error(t, err)
    assert.Nil(t, user)
    assert.True(t, errors.IsConflict(err))
    
    mockUserRepo.AssertExpectations(t)
}
```

### Mock Repository Implementation

```go
// tests/mocks/repository/user_repository_mock.go
package repository

import (
    "context"
    
    "github.com/stretchr/testify/mock"
    "go.mongodb.org/mongo-driver/bson/primitive"
    
    "wedding-invitation/internal/domain/models"
)

type UserRepository struct {
    mock.Mock
}

func (m *UserRepository) Create(ctx context.Context, user *models.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *UserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
    args := m.Called(ctx, email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*models.User), args.Error(1)
}

func (m *UserRepository) Update(ctx context.Context, user *models.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *UserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}

func (m *UserRepository) Exists(ctx context.Context, email string) (bool, error) {
    args := m.Called(ctx, email)
    return args.Bool(0), args.Error(1)
}
```

### Integration Test Example

```go
// tests/integration/repository/user_repository_test.go
package repository_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "go.mongodb.org/mongo-driver/bson/primitive"
    
    "wedding-invitation/internal/domain/models"
    "wedding-invitation/internal/repository/mongodb"
    "wedding-invitation/tests/testutil"
)

type UserRepositorySuite struct {
    suite.Suite
    repo *mongodb.UserRepository
    conn *mongodb.Connection
}

func (s *UserRepositorySuite) SetupSuite() {
    // Setup test database connection
    conn, err := testutil.SetupTestDatabase()
    if err != nil {
        s.T().Fatal(err)
    }
    s.conn = conn
    s.repo = mongodb.NewUserRepository(conn)
}

func (s *UserRepositorySuite) TearDownSuite() {
    s.conn.Close()
}

func (s *UserRepositorySuite) SetupTest() {
    // Clean database before each test
    testutil.CleanCollection(s.conn.Database(), "users")
}

func (s *UserRepositorySuite) TestCreate() {
    ctx := context.Background()
    
    user := &models.User{
        Email:    "test@example.com",
        Password: "hashedpassword",
        Name:     "Test User",
    }
    user.BeforeCreate()
    
    err := s.repo.Create(ctx, user)
    
    assert.NoError(s.T(), err)
    assert.NotEqual(s.T(), primitive.NilObjectID, user.ID)
}

func (s *UserRepositorySuite) TestGetByEmail() {
    ctx := context.Background()
    
    // Create a user first
    user := &models.User{
        Email:    "find@example.com",
        Password: "hashedpassword",
        Name:     "Find User",
    }
    user.BeforeCreate()
    s.repo.Create(ctx, user)
    
    // Retrieve by email
    found, err := s.repo.GetByEmail(ctx, "find@example.com")
    
    assert.NoError(s.T(), err)
    assert.NotNil(s.T(), found)
    assert.Equal(s.T(), user.Email, found.Email)
}

func TestUserRepositorySuite(t *testing.T) {
    suite.Run(t, new(UserRepositorySuite))
}
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific test
go test ./internal/service/... -run TestUserService_Create

# Run integration tests only
go test ./tests/integration/...

# Run with verbose output
go test -v ./...

# Run with race detection
go test -race ./...
```

---

## Build & Development Workflow

### Makefile

```makefile
# Makefile for wedding-invitation backend

.PHONY: all build test clean run dev lint fmt deps mock swagger

# Variables
BINARY_NAME=wedding-api
MAIN_PATH=cmd/api/main.go
BUILD_DIR=./build

# Default target
all: clean deps fmt lint test build

# Development build
dev:
	go run $(MAIN_PATH)

# Production build
build:
	@echo "Building..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	goimports -w .
	gofumpt -w .

# Lint code
lint:
	@echo "Linting..."
	golangci-lint run ./...

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -cover ./...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	go test -v ./tests/integration/...

# Generate mocks
mock:
	@echo "Generating mocks..."
	mockery --all --case=underscore --output=tests/mocks

# Generate swagger documentation
swagger:
	@echo "Generating swagger docs..."
	swag init -g $(MAIN_PATH) -o api/swagger

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean

# Docker commands
docker-build:
	docker build -t wedding-api:latest .

docker-run:
	docker run -p 8080:8080 --env-file .env wedding-api:latest

# Database migrations
migrate-up:
	@echo "Running migrations up..."
	# Add migration command here

migrate-down:
	@echo "Running migrations down..."
	# Add migration command here

# Hot reload for development (requires air)
watch:
	air

# Security scan
security:
	gosec ./...

# Check for outdated dependencies
outdated:
	go list -u -m all

# Update dependencies
update:
	go get -u ./...
	go mod tidy
```

### Development Tools Setup

```bash
# Install required tools
make deps

# Install additional development tools
go install golang.org/x/tools/cmd/goimports@latest
go install mvdan.cc/gofumpt@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/vektra/mockery/v2@latest
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/cosmtrek/air@latest  # For hot reload

# Setup pre-commit hooks
# .git/hooks/pre-commit
#!/bin/sh
make fmt
make lint
make test
```

### Environment Configuration

```bash
# .env.example
SERVER_PORT=8080
SERVER_READ_TIMEOUT=10s
SERVER_WRITE_TIMEOUT=10s

MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=wedding_invitation

JWT_SECRET=your-secret-key-here
JWT_EXPIRATION=24h

LOG_LEVEL=info
LOG_FORMAT=json

# Testing
TEST_MONGODB_URI=mongodb://localhost:27017
TEST_MONGODB_DATABASE=wedding_invitation_test
```

### Docker Support

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o wedding-api cmd/api/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/wedding-api .

# Expose port
EXPOSE 8080

# Run
CMD ["./wedding-api"]
```

```yaml
# docker-compose.yml
version: '3.8'

services:
  api:
    build: .
    ports:
      - "8080:8080"
    environment:
      - SERVER_PORT=8080
      - MONGODB_URI=mongodb://mongo:27017
      - MONGODB_DATABASE=wedding_invitation
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - mongo
    volumes:
      - ./:/app  # For development hot reload
    command: air  # Use air for hot reload in dev
    
  mongo:
    image: mongo:6
    ports:
      - "27017:27017"
    volumes:
      - mongo_data:/data/db
      
  mongo-express:
    image: mongo-express
    ports:
      - "8081:8081"
    environment:
      - ME_CONFIG_MONGODB_URL=mongodb://mongo:27017/
    depends_on:
      - mongo

volumes:
  mongo_data:
```

### CI/CD Pipeline (GitHub Actions)

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Install golangci-lint
      run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    
    - name: Run linter
      run: golangci-lint run ./...

  test:
    runs-on: ubuntu-latest
    services:
      mongodb:
        image: mongo:6
        ports:
          - 27017:27017
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Run tests
      run: go test -race -coverprofile=coverage.out ./...
      env:
        TEST_MONGODB_URI: mongodb://localhost:27017
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  build:
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Build
      run: go build -v ./...
```

### Development Workflow Summary

```
1. Setup
   ├── Install Go 1.21+
   ├── Clone repository
   ├── Copy .env.example to .env
   ├── Run `make deps` to install dependencies
   └── Start MongoDB (docker-compose up -d mongo)

2. Development
   ├── Make changes to code
   ├── Run `make fmt` to format
   ├── Run `make lint` to check
   ├── Run `make test` to verify
   └── Run `make dev` to start server

3. Testing
   ├── Write unit tests with mocks
   ├── Run `make test` for unit tests
   ├── Run `make test-integration` for integration tests
   └── Check coverage with `go test -cover`

4. Building
   ├── Run `make build` for production build
   ├── Test binary with `./build/wedding-api`
   └── Build Docker image with `make docker-build`

5. Deployment
   ├── Push to GitHub triggers CI/CD
   ├── CI runs lint, test, build
   ├── CD deploys to staging/production
   └── Monitor with logs and metrics
```

---

## Summary

This project structure follows:

1. **Standard Go Project Layout**: Industry-standard directory organization
2. **Clean Architecture**: Dependency rule, separation of concerns, testability
3. **Domain-Driven Design**: Domain models at the center, infrastructure at edges
4. **Interface-Based Design**: Dependencies are interfaces, enabling mocking
5. **Dependency Injection**: All dependencies injected, no global state
6. **Comprehensive Testing**: Unit, integration, and e2e tests
7. **Developer Experience**: Makefile, Docker, hot reload, CI/CD

Key benefits:
- **Maintainable**: Changes in one layer don't cascade
- **Testable**: Easy to mock dependencies
- **Scalable**: Clear separation allows team scaling
- **Flexible**: Can swap database or framework without touching business logic
- **Documented**: Clean structure is self-documenting

This architecture ensures the codebase remains clean, tested, and maintainable as the application grows.
