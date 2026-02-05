# Backend Testing Guide

Comprehensive testing strategy for the Wedding Invitation backend API built with Go, Gin, and MongoDB.

## Table of Contents

1. [Testing Strategy Overview](#testing-strategy-overview)
2. [Test Types](#test-types)
3. [Testing Tools](#testing-tools)
4. [Writing Unit Tests](#writing-unit-tests)
5. [Writing Integration Tests](#writing-integration-tests)
6. [Repository Testing](#repository-testing)
7. [Handler Testing](#handler-testing)
8. [Test Organization](#test-organization)
9. [Running Tests](#running-tests)
10. [CI/CD Integration](#cicd-integration)
11. [Coverage Targets](#coverage-targets)
12. [Complete Examples](#complete-examples)

---

## Testing Strategy Overview

Our testing strategy follows the **Testing Pyramid** approach:

```
       /\
      /  \    E2E Tests (5%)
     /----\
    /      \  Integration Tests (20%)
   /--------\
  /          \ Unit Tests (75%)
 /____________\
```

### Principles

1. **Fast Feedback**: Unit tests run in milliseconds
2. **Isolation**: Tests don't depend on external services
3. **Deterministic**: Same input always produces same output
4. **Maintainable**: Tests are easy to understand and update
5. **Comprehensive**: 70% minimum code coverage

### Test Categories

| Category | Scope | Speed | Dependencies |
|----------|-------|-------|--------------|
| Unit | Single function/component | Fast (<10ms) | None (mocked) |
| Integration | Multiple components | Medium (<1s) | Test DB/HTTP |
| E2E | Full application flow | Slow (<10s) | Full stack |

---

## Test Types

### 1. Unit Tests

Test individual functions and methods in isolation.

**Characteristics:**
- No external dependencies (database, HTTP, file system)
- All dependencies mocked
- Fast execution
- High volume

**Coverage Targets:**
- Business logic: 90%+
- Utility functions: 80%+
- Configuration: 50%+

### 2. Integration Tests

Test component interactions with real dependencies.

**Characteristics:**
- Real database connections (test containers)
- HTTP server with real routing
- Test data seeding and cleanup
- Slower but more realistic

**Coverage Targets:**
- API endpoints: 80%+
- Database operations: 85%+
- External integrations: 60%+

### 3. Repository Tests

Test database layer with actual MongoDB instances.

**Characteristics:**
- Use testcontainers-go for MongoDB
- CRUD operation validation
- Index and constraint testing
- Transaction testing

### 4. Handler Tests

Test HTTP handlers with mocked Gin contexts.

**Characteristics:**
- Mocked Gin context
- Request/response validation
- Middleware chain testing
- Error handling verification

---

## Testing Tools

### Testify Suite

The primary testing framework providing assertions and mocking.

```go
import (
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
    "github.com/stretchr/testify/require"
)
```

**Key Features:**
- `assert`: Non-fatal assertions
- `require`: Fatal assertions (stop test immediately)
- `mock`: Mocking framework
- `suite`: Test suite organization

### HTTP Testing

Standard library `httptest` for handler testing.

```go
import "net/http/httptest"
```

**Components:**
- `httptest.NewRecorder()`: Records HTTP responses
- `httptest.NewRequest()`: Creates test requests
- Response validation and inspection

### MongoDB Test Containers

Spin up real MongoDB instances for integration tests.

```go
import (
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/mongodb"
)
```

**Benefits:**
- Real database behavior
- Automatic cleanup
- Parallel test execution
- No local MongoDB required

### Mocking with Testify

Create mock implementations of interfaces.

```go
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) FindByID(ctx context.Context, id string) (*Model, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*Model), args.Error(1)
}
```

---

## Writing Unit Tests

### Service Layer Testing

Test business logic without database dependencies.

```go
// services/user_service_test.go
package services

import (
    "context"
    "errors"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
    
    "wedding-invitation/backend/internal/domain"
    "wedding-invitation/backend/internal/mocks"
)

type UserServiceTestSuite struct {
    suite.Suite
    mockRepo *mocks.MockUserRepository
    service  *UserService
    ctx      context.Context
}

func (s *UserServiceTestSuite) SetupTest() {
    s.mockRepo = new(mocks.MockUserRepository)
    s.service = NewUserService(s.mockRepo)
    s.ctx = context.Background()
}

func (s *UserServiceTestSuite) TearDownTest() {
    s.mockRepo.AssertExpectations(s.T())
}

func TestUserServiceSuite(t *testing.T) {
    suite.Run(t, new(UserServiceTestSuite))
}

func (s *UserServiceTestSuite) TestCreateUser_Success() {
    // Arrange
    input := domain.CreateUserInput{
        Email:     "test@example.com",
        Password:  "securePassword123",
        FirstName: "John",
        LastName:  "Doe",
    }
    
    s.mockRepo.On("FindByEmail", s.ctx, input.Email).
        Return(nil, nil). // User doesn't exist
        Once()
    
    s.mockRepo.On("Create", s.ctx, mock.AnythingOfType("*domain.User")).
        Return(nil).
        Once()
    
    // Act
    user, err := s.service.CreateUser(s.ctx, input)
    
    // Assert
    s.NoError(err)
    s.NotNil(user)
    s.Equal(input.Email, user.Email)
    s.Equal(input.FirstName, user.FirstName)
    s.NotEmpty(user.ID)
    s.NotEmpty(user.PasswordHash) // Password should be hashed
}

func (s *UserServiceTestSuite) TestCreateUser_DuplicateEmail() {
    // Arrange
    input := domain.CreateUserInput{
        Email:     "existing@example.com",
        Password:  "password123",
        FirstName: "Jane",
        LastName:  "Doe",
    }
    
    existingUser := &domain.User{
        ID:    "existing-id",
        Email: input.Email,
    }
    
    s.mockRepo.On("FindByEmail", s.ctx, input.Email).
        Return(existingUser, nil).
        Once()
    
    // Act
    user, err := s.service.CreateUser(s.ctx, input)
    
    // Assert
    s.Error(err)
    s.Nil(user)
    s.True(errors.Is(err, domain.ErrUserAlreadyExists))
}

func (s *UserServiceTestSuite) TestCreateUser_InvalidInput() {
    tests := []struct {
        name    string
        input   domain.CreateUserInput
        wantErr string
    }{
        {
            name: "empty email",
            input: domain.CreateUserInput{
                Email:     "",
                Password:  "password123",
                FirstName: "John",
            },
            wantErr: "email is required",
        },
        {
            name: "invalid email format",
            input: domain.CreateUserInput{
                Email:     "invalid-email",
                Password:  "password123",
                FirstName: "John",
            },
            wantErr: "invalid email format",
        },
        {
            name: "short password",
            input: domain.CreateUserInput{
                Email:     "test@example.com",
                Password:  "123",
                FirstName: "John",
            },
            wantErr: "password must be at least 8 characters",
        },
        {
            name: "missing first name",
            input: domain.CreateUserInput{
                Email:     "test@example.com",
                Password:  "password123",
                FirstName: "",
            },
            wantErr: "first name is required",
        },
    }
    
    for _, tt := range tests {
        s.Run(tt.name, func() {
            // Act
            user, err := s.service.CreateUser(s.ctx, tt.input)
            
            // Assert
            s.Error(err)
            s.Nil(user)
            s.Contains(err.Error(), tt.wantErr)
        })
    }
}
```

### Table-Driven Tests

Efficiently test multiple scenarios with shared setup.

```go
// services/wedding_service_test.go
package services

import (
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
)

func TestWeddingService_ValidateWeddingDate(t *testing.T) {
    now := time.Now()
    
    tests := []struct {
        name      string
        date      time.Time
        wantError bool
        errorMsg  string
    }{
        {
            name:      "future date",
            date:      now.AddDate(0, 1, 0),
            wantError: false,
        },
        {
            name:      "today",
            date:      now,
            wantError: true,
            errorMsg:  "wedding date must be in the future",
        },
        {
            name:      "past date",
            date:      now.AddDate(0, -1, 0),
            wantError: true,
            errorMsg:  "wedding date must be in the future",
        },
        {
            name:      "exactly 1 year from now",
            date:      now.AddDate(1, 0, 0),
            wantError: false,
        },
        {
            name:      "more than 2 years in future",
            date:      now.AddDate(2, 0, 1),
            wantError: true,
            errorMsg:  "wedding date must be within 2 years",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateWeddingDate(tt.date)
            
            if tt.wantError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Mocking Repositories

Create mock implementations for isolated testing.

```go
// mocks/user_repository_mock.go
package mocks

import (
    "context"
    
    "github.com/stretchr/testify/mock"
    
    "wedding-invitation/backend/internal/domain"
)

type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
    args := m.Called(ctx, email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
    args := m.Called(ctx, id)
    return args.Error(0)
}
```

---

## Writing Integration Tests

### Full API Endpoint Testing

Test complete request/response cycles.

```go
// tests/integration/wedding_api_test.go
package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    
    "wedding-invitation/backend/internal/api/handlers"
    "wedding-invitation/backend/internal/api/middleware"
    "wedding-invitation/backend/internal/config"
    "wedding-invitation/backend/internal/domain"
    "wedding-invitation/backend/internal/repository"
    "wedding-invitation/backend/internal/services"
    "wedding-invitation/backend/tests/testutil"
)

type WeddingAPIIntegrationTestSuite struct {
    suite.Suite
    router   *gin.Engine
    db       *testutil.TestDatabase
    ctx      *gin.Context
    recorder *httptest.ResponseRecorder
}

func (s *WeddingAPIIntegrationTestSuite) SetupSuite() {
    // Set Gin to test mode
    gin.SetMode(gin.TestMode)
    
    // Start test database
    s.db = testutil.NewTestDatabase(s.T())
    s.db.Start()
    
    // Setup router with real dependencies
    s.setupRouter()
}

func (s *WeddingAPIIntegrationTestSuite) TearDownSuite() {
    s.db.Stop()
}

func (s *WeddingAPIIntegrationTestSuite) SetupTest() {
    s.recorder = httptest.NewRecorder()
    s.db.Cleanup()
}

func TestWeddingAPIIntegrationSuite(t *testing.T) {
    suite.Run(t, new(WeddingAPIIntegrationTestSuite))
}

func (s *WeddingAPIIntegrationTestSuite) setupRouter() {
    s.router = gin.New()
    
    // Initialize real dependencies
    cfg := config.Load()
    weddingRepo := repository.NewWeddingRepository(s.db.Client())
    weddingService := services.NewWeddingService(weddingRepo)
    weddingHandler := handlers.NewWeddingHandler(weddingService)
    
    // Setup routes
    api := s.router.Group("/api/v1")
    {
        weddings := api.Group("/weddings")
        weddings.Use(middleware.AuthMiddleware(cfg))
        {
            weddings.POST("", weddingHandler.Create)
            weddings.GET("/:id", weddingHandler.Get)
            weddings.PUT("/:id", weddingHandler.Update)
            weddings.DELETE("/:id", weddingHandler.Delete)
        }
    }
}

func (s *WeddingAPIIntegrationTestSuite) TestCreateWedding_Success() {
    // Arrange
    input := map[string]interface{}{
        "title":       "Our Beautiful Wedding",
        "description": "Join us for our special day",
        "date":        "2024-12-25T15:00:00Z",
        "venue": map[string]interface{}{
            "name":    "Grand Ballroom",
            "address": "123 Wedding Lane",
            "city":    "New York",
        },
        "theme": "classic-elegant",
    }
    
    jsonData, _ := json.Marshal(input)
    req := httptest.NewRequest(http.MethodPost, "/api/v1/weddings", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer test-token")
    
    // Act
    s.router.ServeHTTP(s.recorder, req)
    
    // Assert
    assert.Equal(s.T(), http.StatusCreated, s.recorder.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(s.recorder.Body.Bytes(), &response)
    s.NoError(err)
    
    s.NotNil(response["id"])
    s.Equal("Our Beautiful Wedding", response["title"])
    s.Equal("classic-elegant", response["theme"])
    s.NotNil(response["created_at"])
}

func (s *WeddingAPIIntegrationTestSuite) TestGetWedding_NotFound() {
    // Arrange
    req := httptest.NewRequest(http.MethodGet, "/api/v1/weddings/non-existent-id", nil)
    req.Header.Set("Authorization", "Bearer test-token")
    
    // Act
    s.router.ServeHTTP(s.recorder, req)
    
    // Assert
    assert.Equal(s.T(), http.StatusNotFound, s.recorder.Code)
    
    var response map[string]interface{}
    json.Unmarshal(s.recorder.Body.Bytes(), &response)
    
    s.Equal("wedding not found", response["error"])
}

func (s *WeddingAPIIntegrationTestSuite) TestUpdateWedding_Success() {
    // Arrange - Create a wedding first
    wedding := &domain.Wedding{
        Title:       "Original Title",
        Description: "Original description",
        UserID:      "test-user-id",
    }
    err := s.db.SeedWedding(wedding)
    s.NoError(err)
    
    update := map[string]interface{}{
        "title":       "Updated Title",
        "description": "Updated description",
    }
    
    jsonData, _ := json.Marshal(update)
    url := "/api/v1/weddings/" + wedding.ID
    req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer test-token")
    
    // Act
    s.router.ServeHTTP(s.recorder, req)
    
    // Assert
    assert.Equal(s.T(), http.StatusOK, s.recorder.Code)
    
    var response map[string]interface{}
    json.Unmarshal(s.recorder.Body.Bytes(), &response)
    
    s.Equal("Updated Title", response["title"])
    s.Equal("Updated description", response["description"])
}
```

### Database Setup and Teardown

Manage test database lifecycle.

```go
// tests/testutil/database.go
package testutil

import (
    "context"
    "fmt"
    "testing"
    "time"
    
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/mongodb"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type TestDatabase struct {
    t         *testing.T
    container *mongodb.MongoDBContainer
    client    *mongo.Client
    dbName    string
}

func NewTestDatabase(t *testing.T) *TestDatabase {
    return &TestDatabase{
        t:      t,
        dbName: fmt.Sprintf("test_%d", time.Now().UnixNano()),
    }
}

func (td *TestDatabase) Start() {
    ctx := context.Background()
    
    // Start MongoDB container
    container, err := mongodb.Run(ctx, "mongo:6.0")
    if err != nil {
        td.t.Fatalf("Failed to start MongoDB container: %v", err)
    }
    
    td.container = container
    
    // Get connection string
    connStr, err := container.ConnectionString(ctx)
    if err != nil {
        td.t.Fatalf("Failed to get connection string: %v", err)
    }
    
    // Connect to MongoDB
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(connStr))
    if err != nil {
        td.t.Fatalf("Failed to connect to MongoDB: %v", err)
    }
    
    // Ping to verify connection
    err = client.Ping(ctx, nil)
    if err != nil {
        td.t.Fatalf("Failed to ping MongoDB: %v", err)
    }
    
    td.client = client
}

func (td *TestDatabase) Stop() {
    ctx := context.Background()
    
    if td.client != nil {
        td.client.Disconnect(ctx)
    }
    
    if td.container != nil {
        td.container.Terminate(ctx)
    }
}

func (td *TestDatabase) Cleanup() {
    ctx := context.Background()
    
    // Drop all collections
    db := td.client.Database(td.dbName)
    collections, err := db.ListCollectionNames(ctx, nil)
    if err != nil {
        td.t.Logf("Failed to list collections: %v", err)
        return
    }
    
    for _, coll := range collections {
        err := db.Collection(coll).Drop(ctx)
        if err != nil {
            td.t.Logf("Failed to drop collection %s: %v", coll, err)
        }
    }
}

func (td *TestDatabase) Client() *mongo.Client {
    return td.client
}

func (td *TestDatabase) DB() *mongo.Database {
    return td.client.Database(td.dbName)
}

// Seed helpers
func (td *TestDatabase) SeedWedding(wedding *domain.Wedding) error {
    ctx := context.Background()
    _, err := td.DB().Collection("weddings").InsertOne(ctx, wedding)
    return err
}

func (td *TestDatabase) SeedUser(user *domain.User) error {
    ctx := context.Background()
    _, err := td.DB().Collection("users").InsertOne(ctx, user)
    return err
}

func (td *TestDatabase) SeedRSVP(rsvp *domain.RSVP) error {
    ctx := context.Background()
    _, err := td.DB().Collection("rsvps").InsertOne(ctx, rsvp)
    return err
}
```

### Test Data Seeding

Populate database with test data.

```go
// tests/testutil/fixtures.go
package testutil

import (
    "time"
    
    "wedding-invitation/backend/internal/domain"
)

// Fixture builders for test data

type WeddingFixture struct {
    ID          string
    Title       string
    Description string
    UserID      string
    Date        time.Time
    Venue       domain.Venue
    Theme       string
}

func NewWeddingFixture() *WeddingFixture {
    return &WeddingFixture{
        ID:          generateTestID(),
        Title:       "Test Wedding",
        Description: "A beautiful test wedding",
        UserID:      generateTestID(),
        Date:        time.Now().AddDate(0, 6, 0),
        Venue: domain.Venue{
            Name:    "Test Venue",
            Address: "123 Test St",
            City:    "Test City",
        },
        Theme: "classic",
    }
}

func (f *WeddingFixture) Build() *domain.Wedding {
    return &domain.Wedding{
        ID:          f.ID,
        Title:       f.Title,
        Description: f.Description,
        UserID:      f.UserID,
        Date:        f.Date,
        Venue:       f.Venue,
        Theme:       f.Theme,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
}

func (f *WeddingFixture) WithTitle(title string) *WeddingFixture {
    f.Title = title
    return f
}

func (f *WeddingFixture) WithUserID(userID string) *WeddingFixture {
    f.UserID = userID
    return f
}

func (f *WeddingFixture) WithDate(date time.Time) *WeddingFixture {
    f.Date = date
    return f
}

// User fixture
type UserFixture struct {
    ID        string
    Email     string
    FirstName string
    LastName  string
}

func NewUserFixture() *UserFixture {
    return &UserFixture{
        ID:        generateTestID(),
        Email:     "test@example.com",
        FirstName: "Test",
        LastName:  "User",
    }
}

func (f *UserFixture) Build() *domain.User {
    return &domain.User{
        ID:        f.ID,
        Email:     f.Email,
        FirstName: f.FirstName,
        LastName:    f.LastName,
        CreatedAt: time.Now(),
    }
}

func generateTestID() string {
    // Simple ID generation for tests
    return fmt.Sprintf("test_%d", time.Now().UnixNano())
}
```

---

## Repository Testing

### MongoDB Test Containers

Test real database operations.

```go
// repository/rsvp_repository_test.go
package repository

import (
    "context"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    
    "wedding-invitation/backend/internal/domain"
    "wedding-invitation/backend/tests/testutil"
)

type RSVPRepositoryTestSuite struct {
    suite.Suite
    repo *RSVPRepository
    db   *testutil.TestDatabase
    ctx  context.Context
}

func (s *RSIPRepositoryTestSuite) SetupSuite() {
    s.db = testutil.NewTestDatabase(s.T())
    s.db.Start()
    
    s.repo = NewRSVPRepository(s.db.Client(), s.db.DB().Name())
    s.ctx = context.Background()
}

func (s *RSVPRepositoryTestSuite) TearDownSuite() {
    s.db.Stop()
}

func (s *RSVPRepositoryTestSuite) SetupTest() {
    s.db.Cleanup()
}

func TestRSVPRepositorySuite(t *testing.T) {
    suite.Run(t, new(RSVPRepositoryTestSuite))
}

func (s *RSVPRepositoryTestSuite) TestCreateRSVP_Success() {
    // Arrange
    rsvp := &domain.RSVP{
        ID:         generateTestID(),
        WeddingID:  generateTestID(),
        GuestName:  "John Doe",
        GuestEmail: "john@example.com",
        Status:     domain.RSVPStatusAttending,
        GuestCount: 2,
        DietaryRestrictions: "Vegetarian",
        Message:    "Looking forward to it!",
        CreatedAt:  time.Now(),
    }
    
    // Act
    err := s.repo.Create(s.ctx, rsvp)
    
    // Assert
    s.NoError(err)
    
    // Verify it was saved
    found, err := s.repo.FindByID(s.ctx, rsvp.ID)
    s.NoError(err)
    s.NotNil(found)
    s.Equal(rsvp.GuestName, found.GuestName)
    s.Equal(rsvp.Status, found.Status)
}

func (s *RSVPRepositoryTestSuite) TestFindByWeddingID_Success() {
    // Arrange
    weddingID := generateTestID()
    
    rsvp1 := &domain.RSVP{
        ID:        generateTestID(),
        WeddingID: weddingID,
        GuestName: "Guest One",
        Status:    domain.RSVPStatusAttending,
        CreatedAt: time.Now(),
    }
    
    rsvp2 := &domain.RSVP{
        ID:        generateTestID(),
        WeddingID: weddingID,
        GuestName: "Guest Two",
        Status:    domain.RSVPStatusDeclined,
        CreatedAt: time.Now(),
    }
    
    rsvp3 := &domain.RSVP{
        ID:        generateTestID(),
        WeddingID: generateTestID(), // Different wedding
        GuestName: "Guest Three",
        Status:    domain.RSVPStatusAttending,
        CreatedAt: time.Now(),
    }
    
    s.db.SeedRSVP(rsvp1)
    s.db.SeedRSVP(rsvp2)
    s.db.SeedRSVP(rsvp3)
    
    // Act
    results, err := s.repo.FindByWeddingID(s.ctx, weddingID)
    
    // Assert
    s.NoError(err)
    s.Len(results, 2)
    
    guestNames := make([]string, len(results))
    for i, r := range results {
        guestNames[i] = r.GuestName
    }
    s.Contains(guestNames, "Guest One")
    s.Contains(guestNames, "Guest Two")
}

func (s *RSVPRepositoryTestSuite) TestUpdateRSVP_Success() {
    // Arrange
    rsvp := &domain.RSVP{
        ID:         generateTestID(),
        WeddingID:  generateTestID(),
        GuestName:  "Jane Doe",
        Status:     domain.RSVPStatusPending,
        GuestCount: 1,
        CreatedAt:  time.Now(),
    }
    s.db.SeedRSVP(rsvp)
    
    // Update
    rsvp.Status = domain.RSVPStatusAttending
    rsvp.GuestCount = 3
    rsvp.Message = "Updated message"
    
    // Act
    err := s.repo.Update(s.ctx, rsvp)
    
    // Assert
    s.NoError(err)
    
    // Verify update
    updated, err := s.repo.FindByID(s.ctx, rsvp.ID)
    s.NoError(err)
    s.Equal(domain.RSVPStatusAttending, updated.Status)
    s.Equal(3, updated.GuestCount)
    s.Equal("Updated message", updated.Message)
}

func (s *RSVPRepositoryTestSuite) TestDeleteRSVP_Success() {
    // Arrange
    rsvp := &domain.RSVP{
        ID:        generateTestID(),
        WeddingID: generateTestID(),
        GuestName: "To Delete",
        Status:    domain.RSVPStatusPending,
        CreatedAt: time.Now(),
    }
    s.db.SeedRSVP(rsvp)
    
    // Verify exists
    found, err := s.repo.FindByID(s.ctx, rsvp.ID)
    s.NoError(err)
    s.NotNil(found)
    
    // Act
    err = s.repo.Delete(s.ctx, rsvp.ID)
    
    // Assert
    s.NoError(err)
    
    // Verify deleted
    deleted, err := s.repo.FindByID(s.ctx, rsvp.ID)
    s.Error(err)
    s.Nil(deleted)
}

func (s *RSVPRepositoryTestSuite) TestGetRSVPStats_Success() {
    // Arrange
    weddingID := generateTestID()
    
    rsvps := []*domain.RSVP{
        {Status: domain.RSVPStatusAttending, GuestCount: 2},
        {Status: domain.RSVPStatusAttending, GuestCount: 3},
        {Status: domain.RSVPStatusDeclined, GuestCount: 0},
        {Status: domain.RSVPStatusPending, GuestCount: 1},
    }
    
    for i, r := range rsvps {
        r.ID = generateTestID()
        r.WeddingID = weddingID
        r.GuestName = fmt.Sprintf("Guest %d", i+1)
        r.CreatedAt = time.Now()
        s.db.SeedRSVP(r)
    }
    
    // Act
    stats, err := s.repo.GetStatsByWeddingID(s.ctx, weddingID)
    
    // Assert
    s.NoError(err)
    s.Equal(2, stats.Attending)
    s.Equal(1, stats.Declined)
    s.Equal(1, stats.Pending)
    s.Equal(5, stats.TotalGuests) // 2 + 3
}

func (s *RSVPRepositoryTestSuite) TestIndexVerification() {
    // Verify that unique indexes are working
    weddingID := generateTestID()
    email := "duplicate@example.com"
    
    rsvp1 := &domain.RSVP{
        ID:         generateTestID(),
        WeddingID:  weddingID,
        GuestEmail: email,
        GuestName:  "First",
        CreatedAt:  time.Now(),
    }
    
    rsvp2 := &domain.RSVP{
        ID:         generateTestID(),
        WeddingID:  weddingID,
        GuestEmail: email, // Same email, should fail
        GuestName:  "Second",
        CreatedAt:  time.Now(),
    }
    
    // First insert should succeed
    err := s.repo.Create(s.ctx, rsvp1)
    s.NoError(err)
    
    // Second insert should fail due to unique index
    err = s.repo.Create(s.ctx, rsvp2)
    s.Error(err)
    s.Contains(err.Error(), "duplicate")
}
```

---

## Handler Testing

### Gin Context Mocking

Test HTTP handlers in isolation.

```go
// handlers/wedding_handler_test.go
package handlers

import (
    "bytes"
    "encoding/json"
    "errors"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
    
    "wedding-invitation/backend/internal/domain"
    "wedding-invitation/backend/internal/mocks"
)

type WeddingHandlerTestSuite struct {
    suite.Suite
    mockService *mocks.MockWeddingService
    handler     *WeddingHandler
    router      *gin.Engine
}

func (s *WeddingHandlerTestSuite) SetupTest() {
    gin.SetMode(gin.TestMode)
    
    s.mockService = new(mocks.MockWeddingService)
    s.handler = NewWeddingHandler(s.mockService)
    
    // Setup router
    s.router = gin.New()
    s.router.POST("/weddings", s.handler.Create)
    s.router.GET("/weddings/:id", s.handler.Get)
    s.router.PUT("/weddings/:id", s.handler.Update)
}

func (s *WeddingHandlerTestSuite) TearDownTest() {
    s.mockService.AssertExpectations(s.T())
}

func TestWeddingHandlerSuite(t *testing.T) {
    suite.Run(t, new(WeddingHandlerTestSuite))
}

func (s *WeddingHandlerTestSuite) TestCreateWedding_Success() {
    // Arrange
    input := domain.CreateWeddingInput{
        Title:       "Test Wedding",
        Description: "Test description",
        Date:        "2024-12-25",
        Venue: domain.VenueInput{
            Name:    "Test Venue",
            Address: "123 Test St",
            City:    "Test City",
        },
        Theme: "classic",
    }
    
    expectedWedding := &domain.Wedding{
        ID:          "wedding-123",
        Title:       input.Title,
        Description: input.Description,
        UserID:      "user-123",
        Theme:       input.Theme,
    }
    
    s.mockService.On("Create", mock.Anything, input, "user-123").
        Return(expectedWedding, nil).
        Once()
    
    jsonData, _ := json.Marshal(input)
    req := httptest.NewRequest(http.MethodPost, "/weddings", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    
    // Add user to context (simulating auth middleware)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Set("userID", "user-123")
    
    // Act
    s.handler.Create(c)
    
    // Assert
    assert.Equal(s.T(), http.StatusCreated, w.Code)
    
    var response domain.Wedding
    err := json.Unmarshal(w.Body.Bytes(), &response)
    s.NoError(err)
    s.Equal(expectedWedding.ID, response.ID)
    s.Equal(expectedWedding.Title, response.Title)
}

func (s *WeddingHandlerTestSuite) TestCreateWedding_InvalidInput() {
    // Arrange - invalid JSON
    req := httptest.NewRequest(http.MethodPost, "/weddings", bytes.NewBufferString("invalid json"))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Set("userID", "user-123")
    
    // Act
    s.handler.Create(c)
    
    // Assert
    assert.Equal(s.T(), http.StatusBadRequest, w.Code)
    
    var response map[string]string
    json.Unmarshal(w.Body.Bytes(), &response)
    s.Contains(response["error"], "invalid request")
}

func (s *WeddingHandlerTestSuite) TestCreateWedding_ServiceError() {
    // Arrange
    input := domain.CreateWeddingInput{
        Title: "Test Wedding",
        Date:  "2024-12-25",
    }
    
    s.mockService.On("Create", mock.Anything, input, "user-123").
        Return(nil, errors.New("database error")).
        Once()
    
    jsonData, _ := json.Marshal(input)
    req := httptest.NewRequest(http.MethodPost, "/weddings", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Set("userID", "user-123")
    
    // Act
    s.handler.Create(c)
    
    // Assert
    assert.Equal(s.T(), http.StatusInternalServerError, w.Code)
}

func (s *WeddingHandlerTestSuite) TestGetWedding_Success() {
    // Arrange
    weddingID := "wedding-123"
    expectedWedding := &domain.Wedding{
        ID:     weddingID,
        Title:  "Test Wedding",
        UserID: "user-123",
    }
    
    s.mockService.On("GetByID", mock.Anything, weddingID).
        Return(expectedWedding, nil).
        Once()
    
    req := httptest.NewRequest(http.MethodGet, "/weddings/"+weddingID, nil)
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Params = gin.Params{{Key: "id", Value: weddingID}}
    
    // Act
    s.handler.Get(c)
    
    // Assert
    assert.Equal(s.T(), http.StatusOK, w.Code)
    
    var response domain.Wedding
    json.Unmarshal(w.Body.Bytes(), &response)
    s.Equal(expectedWedding.ID, response.ID)
}

func (s *WeddingHandlerTestSuite) TestGetWedding_NotFound() {
    // Arrange
    weddingID := "non-existent"
    
    s.mockService.On("GetByID", mock.Anything, weddingID).
        Return(nil, domain.ErrWeddingNotFound).
        Once()
    
    req := httptest.NewRequest(http.MethodGet, "/weddings/"+weddingID, nil)
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Params = gin.Params{{Key: "id", Value: weddingID}}
    
    // Act
    s.handler.Get(c)
    
    // Assert
    assert.Equal(s.T(), http.StatusNotFound, w.Code)
}

func (s *WeddingHandlerTestSuite) TestGetWedding_Unauthorized() {
    // Arrange
    weddingID := "wedding-123"
    wedding := &domain.Wedding{
        ID:     weddingID,
        Title:  "Test Wedding",
        UserID: "different-user", // Not the requesting user
    }
    
    s.mockService.On("GetByID", mock.Anything, weddingID).
        Return(wedding, nil).
        Once()
    
    req := httptest.NewRequest(http.MethodGet, "/weddings/"+weddingID, nil)
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Params = gin.Params{{Key: "id", Value: weddingID}}
    c.Set("userID", "user-123")
    
    // Act
    s.handler.Get(c)
    
    // Assert
    assert.Equal(s.T(), http.StatusForbidden, w.Code)
}
```

### Middleware Testing

Test authentication and other middleware.

```go
// middleware/auth_test.go
package middleware

import (
    "errors"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    
    "wedding-invitation/backend/internal/config"
    "wedding-invitation/backend/internal/mocks"
)

func TestAuthMiddleware_ValidToken(t *testing.T) {
    // Arrange
    gin.SetMode(gin.TestMode)
    
    mockAuth := new(mocks.MockAuthService)
    cfg := &config.Config{
        JWTSecret: "test-secret",
    }
    
    mockAuth.On("ValidateToken", "valid-token").
        Return("user-123", nil).
        Once()
    
    router := gin.New()
    router.Use(AuthMiddleware(cfg, mockAuth))
    router.GET("/protected", func(c *gin.Context) {
        userID, _ := c.Get("userID")
        c.JSON(200, gin.H{"user_id": userID})
    })
    
    req := httptest.NewRequest(http.MethodGet, "/protected", nil)
    req.Header.Set("Authorization", "Bearer valid-token")
    
    w := httptest.NewRecorder()
    
    // Act
    router.ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Contains(t, w.Body.String(), "user-123")
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
    // Arrange
    gin.SetMode(gin.TestMode)
    
    cfg := &config.Config{
        JWTSecret: "test-secret",
    }
    
    router := gin.New()
    router.Use(AuthMiddleware(cfg, nil))
    router.GET("/protected", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    req := httptest.NewRequest(http.MethodGet, "/protected", nil)
    w := httptest.NewRecorder()
    
    // Act
    router.ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
    // Arrange
    gin.SetMode(gin.TestMode)
    
    mockAuth := new(mocks.MockAuthService)
    cfg := &config.Config{
        JWTSecret: "test-secret",
    }
    
    mockAuth.On("ValidateToken", "invalid-token").
        Return("", errors.New("invalid token")).
        Once()
    
    router := gin.New()
    router.Use(AuthMiddleware(cfg, mockAuth))
    router.GET("/protected", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    req := httptest.NewRequest(http.MethodGet, "/protected", nil)
    req.Header.Set("Authorization", "Bearer invalid-token")
    
    w := httptest.NewRecorder()
    
    // Act
    router.ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_MalformedHeader(t *testing.T) {
    // Arrange
    gin.SetMode(gin.TestMode)
    
    cfg := &config.Config{
        JWTSecret: "test-secret",
    }
    
    router := gin.New()
    router.Use(AuthMiddleware(cfg, nil))
    router.GET("/protected", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    req := httptest.NewRequest(http.MethodGet, "/protected", nil)
    req.Header.Set("Authorization", "Basic invalid") // Wrong scheme
    
    w := httptest.NewRecorder()
    
    // Act
    router.ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, http.StatusUnauthorized, w.Code)
}
```

---

## Test Organization

### Directory Structure

```
backend/
├── internal/
│   ├── domain/
│   │   └── domain_test.go         # Domain logic tests
│   ├── services/
│   │   ├── user_service.go
│   │   ├── user_service_test.go   # Unit tests
│   │   ├── wedding_service.go
│   │   └── wedding_service_test.go
│   ├── repository/
│   │   ├── user_repository.go
│   │   ├── user_repository_test.go # Repository tests
│   │   ├── rsvp_repository.go
│   │   └── rsvp_repository_test.go
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── user_handler.go
│   │   │   ├── user_handler_test.go # Handler tests
│   │   │   ├── wedding_handler.go
│   │   │   └── wedding_handler_test.go
│   │   └── middleware/
│   │       ├── auth.go
│   │       └── auth_test.go
│   └── mocks/                     # Generated mocks
│       ├── user_repository_mock.go
│       ├── wedding_repository_mock.go
│       └── mockery_generated/
├── tests/
│   ├── integration/               # Integration tests
│   │   ├── wedding_api_test.go
│   │   ├── rsvp_api_test.go
│   │   └── suite_test.go
│   ├── testutil/                  # Test utilities
│   │   ├── database.go
│   │   ├── fixtures.go
│   │   └── http.go
│   └── fixtures/                  # Test data files
│       ├── users.json
│       └── weddings.json
└── go.mod
```

### File Naming Conventions

- Unit tests: `*_test.go` (same package)
- Integration tests: `*_integration_test.go` (separate `integration` package)
- Test utilities: `tests/testutil/*.go`
- Fixtures: `tests/fixtures/*.json`

### Package Organization

```go
// Internal tests (white box)
package services // same package as code

// External tests (black box)
package services_test // separate test package

// Integration tests
package integration // dedicated package
```

---

## Running Tests

### Individual Packages

```bash
# Run all tests
go test ./...

# Run tests in specific package
go test ./internal/services/...

# Run specific test
go test ./internal/services -run TestUserService

# Run specific test suite
go test ./internal/services -run TestUserServiceSuite

# Verbose output
go test -v ./internal/services/...
```

### With Coverage

```bash
# Basic coverage
go test -cover ./...

# Coverage with detailed report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Coverage for specific package
go test -coverprofile=coverage.out ./internal/services
go tool cover -func=coverage.out

# Exclude mocks and generated code
go test -coverprofile=coverage.out $(go list ./... | grep -v /mocks | grep -v /testutil)
```

### Race Detection

```bash
# Run with race detector
go test -race ./...

# Race detection on specific package
go test -race ./internal/services/...

# Race detection with short timeout
go test -race -timeout=5m ./...
```

### Performance Testing

```bash
# Run benchmarks
go test -bench=. ./...

# Run benchmarks with memory allocation stats
go test -bench=. -benchmem ./...

# Run specific benchmark
go test -bench=BenchmarkCreateWedding ./internal/services
```

### Parallel Execution

```bash
# Run tests in parallel
go test -parallel=4 ./...

# Set parallel count per CPU
go test -parallel=$(nproc) ./...
```

### Integration Test Flags

```bash
# Run only unit tests (skip integration)
go test -short ./...

# Run only integration tests
go test ./tests/integration/...

# Run with custom tag
go test -tags=integration ./...
```

---

## CI/CD Integration

### GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Tests

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      mongodb:
        image: mongo:6.0
        ports:
          - 27017:27017
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
          cache: true
      
      - name: Install dependencies
        run: go mod download
      
      - name: Run unit tests
        run: go test -v -race -short ./...
        env:
          CI: true
      
      - name: Run integration tests
        run: go test -v ./tests/integration/...
        env:
          MONGODB_URI: mongodb://localhost:27017/test
          CI: true
      
      - name: Generate coverage report
        run: |
          go test -coverprofile=coverage.out ./...
          go tool cover -func=coverage.out
      
      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          file: ./coverage.out
          fail_ci_if_error: true

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --timeout=5m
```

### Makefile Targets

```makefile
# Makefile
.PHONY: test test-unit test-integration test-race coverage lint

test: test-unit test-integration

test-unit:
	go test -v -race -short ./...

test-integration:
	go test -v ./tests/integration/...

test-race:
	go test -race -timeout=5m ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

coverage-report:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

lint:
	golangci-lint run ./...

benchmark:
	go test -bench=. -benchmem ./...
```

---

## Coverage Targets

### Minimum Requirements

| Component | Minimum Coverage | Target Coverage |
|-----------|------------------|-----------------|
| Domain Layer | 70% | 90% |
| Services | 70% | 85% |
| Repositories | 70% | 80% |
| Handlers | 70% | 80% |
| Middleware | 70% | 75% |
| Utilities | 50% | 70% |

### Overall Project: 70% Minimum

### Exclusions

```bash
# Exclude from coverage
//go:build ignore

# Or use pattern in coverage command
go test -coverprofile=coverage.out $(go list ./... | \
    grep -v /mocks | \
    grep -v /testutil | \
    grep -v /generated | \
    grep -v /cmd | \
    grep -v /docs)
```

### Coverage Badge

Add to README.md:

```markdown
[![Coverage](https://codecov.io/gh/yourorg/wedding-invitation/branch/main/graph/badge.svg)](https://codecov.io/gh/yourorg/wedding-invitation)
```

---

## Complete Examples

### User Service Test (Complete)

```go
// internal/services/user_service_test.go
package services

import (
    "context"
    "errors"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
    "golang.org/x/crypto/bcrypt"
    
    "wedding-invitation/backend/internal/domain"
    "wedding-invitation/backend/internal/mocks"
)

type UserServiceTestSuite struct {
    suite.Suite
    mockRepo *mocks.MockUserRepository
    service  *UserService
    ctx      context.Context
}

func (s *UserServiceTestSuite) SetupTest() {
    s.mockRepo = new(mocks.MockUserRepository)
    s.service = NewUserService(s.mockRepo)
    s.ctx = context.Background()
}

func (s *UserServiceTestSuite) TearDownTest() {
    s.mockRepo.AssertExpectations(s.T())
}

func TestUserServiceSuite(t *testing.T) {
    suite.Run(t, new(UserServiceTestSuite))
}

func (s *UserServiceTestSuite) TestCreateUser_Success() {
    // Arrange
    input := domain.CreateUserInput{
        Email:     "test@example.com",
        Password:  "securePassword123!",
        FirstName: "John",
        LastName:  "Doe",
    }
    
    s.mockRepo.On("FindByEmail", s.ctx, input.Email).
        Return(nil, nil).
        Once()
    
    s.mockRepo.On("Create", s.ctx, mock.AnythingOfType("*domain.User")).
        Return(nil).
        Once()
    
    // Act
    user, err := s.service.CreateUser(s.ctx, input)
    
    // Assert
    s.NoError(err)
    s.NotNil(user)
    s.Equal(input.Email, user.Email)
    s.Equal(input.FirstName, user.FirstName)
    s.Equal(input.LastName, user.LastName)
    s.NotEmpty(user.ID)
    s.NotEmpty(user.PasswordHash)
    s.NotEqual(input.Password, user.PasswordHash) // Password should be hashed
    
    // Verify password hash works
    err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
    s.NoError(err)
}

func (s *UserServiceTestSuite) TestCreateUser_DuplicateEmail() {
    input := domain.CreateUserInput{
        Email:     "existing@example.com",
        Password:  "password123",
        FirstName: "Jane",
        LastName:  "Doe",
    }
    
    existingUser := &domain.User{
        ID:    "existing-id",
        Email: input.Email,
    }
    
    s.mockRepo.On("FindByEmail", s.ctx, input.Email).
        Return(existingUser, nil).
        Once()
    
    user, err := s.service.CreateUser(s.ctx, input)
    
    s.Error(err)
    s.Nil(user)
    s.True(errors.Is(err, domain.ErrUserAlreadyExists))
}

func (s *UserServiceTestSuite) TestCreateUser_InvalidInput() {
    tests := []struct {
        name    string
        input   domain.CreateUserInput
        wantErr string
    }{
        {
            name: "empty email",
            input: domain.CreateUserInput{
                Email:     "",
                Password:  "password123!",
                FirstName: "John",
            },
            wantErr: "email is required",
        },
        {
            name: "invalid email format",
            input: domain.CreateUserInput{
                Email:     "invalid-email",
                Password:  "password123!",
                FirstName: "John",
            },
            wantErr: "invalid email format",
        },
        {
            name: "short password",
            input: domain.CreateUserInput{
                Email:     "test@example.com",
                Password:  "123",
                FirstName: "John",
            },
            wantErr: "password must be at least 8 characters",
        },
        {
            name: "password without uppercase",
            input: domain.CreateUserInput{
                Email:     "test@example.com",
                Password:  "password123!",
                FirstName: "John",
            },
            wantErr: "password must contain at least one uppercase letter",
        },
        {
            name: "missing first name",
            input: domain.CreateUserInput{
                Email:     "test@example.com",
                Password:  "SecurePass123!",
                FirstName: "",
            },
            wantErr: "first name is required",
        },
    }
    
    for _, tt := range tests {
        s.Run(tt.name, func() {
            user, err := s.service.CreateUser(s.ctx, tt.input)
            
            s.Error(err)
            s.Nil(user)
            s.Contains(err.Error(), tt.wantErr)
        })
    }
}

func (s *UserServiceTestSuite) TestAuthenticate_Success() {
    // Arrange
    email := "test@example.com"
    password := "securePassword123!"
    
    // Create password hash
    hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    
    user := &domain.User{
        ID:           "user-123",
        Email:        email,
        PasswordHash: string(hash),
        FirstName:    "John",
        LastName:     "Doe",
    }
    
    s.mockRepo.On("FindByEmail", s.ctx, email).
        Return(user, nil).
        Once()
    
    // Act
    authenticatedUser, err := s.service.Authenticate(s.ctx, email, password)
    
    // Assert
    s.NoError(err)
    s.NotNil(authenticatedUser)
    s.Equal(user.ID, authenticatedUser.ID)
    s.Equal(user.Email, authenticatedUser.Email)
}

func (s *UserServiceTestSuite) TestAuthenticate_InvalidPassword() {
    email := "test@example.com"
    correctPassword := "securePassword123!"
    wrongPassword := "wrongpassword"
    
    hash, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)
    
    user := &domain.User{
        ID:           "user-123",
        Email:        email,
        PasswordHash: string(hash),
    }
    
    s.mockRepo.On("FindByEmail", s.ctx, email).
        Return(user, nil).
        Once()
    
    authenticatedUser, err := s.service.Authenticate(s.ctx, email, wrongPassword)
    
    s.Error(err)
    s.Nil(authenticatedUser)
    s.True(errors.Is(err, domain.ErrInvalidCredentials))
}

func (s *UserServiceTestSuite) TestGetUserByID_Success() {
    userID := "user-123"
    expectedUser := &domain.User{
        ID:        userID,
        Email:     "test@example.com",
        FirstName: "John",
        LastName:  "Doe",
    }
    
    s.mockRepo.On("FindByID", s.ctx, userID).
        Return(expectedUser, nil).
        Once()
    
    user, err := s.service.GetUserByID(s.ctx, userID)
    
    s.NoError(err)
    s.NotNil(user)
    s.Equal(expectedUser.ID, user.ID)
    s.Equal(expectedUser.Email, user.Email)
}

func (s *UserServiceTestSuite) TestGetUserByID_NotFound() {
    userID := "non-existent"
    
    s.mockRepo.On("FindByID", s.ctx, userID).
        Return(nil, domain.ErrUserNotFound).
        Once()
    
    user, err := s.service.GetUserByID(s.ctx, userID)
    
    s.Error(err)
    s.Nil(user)
    s.True(errors.Is(err, domain.ErrUserNotFound))
}

func (s *UserServiceTestSuite) TestUpdateUser_Success() {
    userID := "user-123"
    input := domain.UpdateUserInput{
        FirstName: "Updated",
        LastName:  "Name",
    }
    
    existingUser := &domain.User{
        ID:        userID,
        Email:     "test@example.com",
        FirstName: "John",
        LastName:  "Doe",
    }
    
    s.mockRepo.On("FindByID", s.ctx, userID).
        Return(existingUser, nil).
        Once()
    
    s.mockRepo.On("Update", s.ctx, mock.AnythingOfType("*domain.User")).
        Return(nil).
        Once()
    
    updatedUser, err := s.service.UpdateUser(s.ctx, userID, input)
    
    s.NoError(err)
    s.NotNil(updatedUser)
    s.Equal(input.FirstName, updatedUser.FirstName)
    s.Equal(input.LastName, updatedUser.LastName)
    s.Equal(existingUser.Email, updatedUser.Email) // Email unchanged
}

func (s *UserServiceTestSuite) TestDeleteUser_Success() {
    userID := "user-123"
    
    s.mockRepo.On("Delete", s.ctx, userID).
        Return(nil).
        Once()
    
    err := s.service.DeleteUser(s.ctx, userID)
    
    s.NoError(err)
}
```

### Wedding Handler Test (Complete)

```go
// internal/api/handlers/wedding_handler_test.go
package handlers

import (
    "bytes"
    "encoding/json"
    "errors"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/suite"
    
    "wedding-invitation/backend/internal/domain"
    "wedding-invitation/backend/internal/mocks"
)

type WeddingHandlerTestSuite struct {
    suite.Suite
    mockService *mocks.MockWeddingService
    handler     *WeddingHandler
    router      *gin.Engine
}

func (s *WeddingHandlerTestSuite) SetupTest() {
    gin.SetMode(gin.TestMode)
    
    s.mockService = new(mocks.MockWeddingService)
    s.handler = NewWeddingHandler(s.mockService)
    
    // Setup router
    s.router = gin.New()
    s.router.POST("/weddings", s.handler.Create)
    s.router.GET("/weddings/:id", s.handler.Get)
    s.router.PUT("/weddings/:id", s.handler.Update)
    s.router.DELETE("/weddings/:id", s.handler.Delete)
    s.router.GET("/weddings/:id/stats", s.handler.GetStats)
    s.router.POST("/weddings/:id/publish", s.handler.Publish)
}

func (s *WeddingHandlerTestSuite) TearDownTest() {
    s.mockService.AssertExpectations(s.T())
}

func TestWeddingHandlerSuite(t *testing.T) {
    suite.Run(t, new(WeddingHandlerTestSuite))
}

func (s *WeddingHandlerTestSuite) TestCreateWedding_Success() {
    input := domain.CreateWeddingInput{
        Title:       "Test Wedding",
        Description: "Test description",
        Date:        "2024-12-25T15:00:00Z",
        Venue: domain.VenueInput{
            Name:    "Test Venue",
            Address: "123 Test St",
            City:    "Test City",
        },
        Theme: "classic-elegant",
    }
    
    expectedWedding := &domain.Wedding{
        ID:          "wedding-123",
        Title:       input.Title,
        Description: input.Description,
        UserID:      "user-123",
        Theme:       input.Theme,
        Date:        time.Date(2024, 12, 25, 15, 0, 0, 0, time.UTC),
        Venue: domain.Venue{
            Name:    input.Venue.Name,
            Address: input.Venue.Address,
            City:    input.Venue.City,
        },
        Status:    domain.WeddingStatusDraft,
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
    
    s.mockService.On("Create", mock.Anything, input, "user-123").
        Return(expectedWedding, nil).
        Once()
    
    jsonData, _ := json.Marshal(input)
    req := httptest.NewRequest(http.MethodPost, "/weddings", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Set("userID", "user-123")
    
    s.handler.Create(c)
    
    assert.Equal(s.T(), http.StatusCreated, w.Code)
    
    var response domain.Wedding
    err := json.Unmarshal(w.Body.Bytes(), &response)
    s.NoError(err)
    s.Equal(expectedWedding.ID, response.ID)
    s.Equal(expectedWedding.Title, response.Title)
    s.Equal(expectedWedding.Theme, response.Theme)
}

func (s *WeddingHandlerTestSuite) TestCreateWedding_InvalidJSON() {
    req := httptest.NewRequest(http.MethodPost, "/weddings", bytes.NewBufferString("invalid json"))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Set("userID", "user-123")
    
    s.handler.Create(c)
    
    assert.Equal(s.T(), http.StatusBadRequest, w.Code)
    
    var response map[string]string
    json.Unmarshal(w.Body.Bytes(), &response)
    s.Contains(response["error"], "invalid request")
}

func (s *WeddingHandlerTestSuite) TestCreateWedding_ValidationErrors() {
    input := map[string]interface{}{
        "title": "", // Missing required fields
    }
    
    jsonData, _ := json.Marshal(input)
    req := httptest.NewRequest(http.MethodPost, "/weddings", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Set("userID", "user-123")
    
    s.handler.Create(c)
    
    assert.Equal(s.T(), http.StatusBadRequest, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    s.NotNil(response["errors"])
}

func (s *WeddingHandlerTestSuite) TestCreateWedding_ServiceError() {
    input := domain.CreateWeddingInput{
        Title: "Test Wedding",
        Date:  "2024-12-25T15:00:00Z",
        Venue: domain.VenueInput{Name: "Test Venue"},
    }
    
    s.mockService.On("Create", mock.Anything, input, "user-123").
        Return(nil, errors.New("database error")).
        Once()
    
    jsonData, _ := json.Marshal(input)
    req := httptest.NewRequest(http.MethodPost, "/weddings", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Set("userID", "user-123")
    
    s.handler.Create(c)
    
    assert.Equal(s.T(), http.StatusInternalServerError, w.Code)
    
    var response map[string]string
    json.Unmarshal(w.Body.Bytes(), &response)
    s.Equal("internal server error", response["error"])
}

func (s *WeddingHandlerTestSuite) TestGetWedding_Success() {
    weddingID := "wedding-123"
    expectedWedding := &domain.Wedding{
        ID:     weddingID,
        Title:  "Test Wedding",
        UserID: "user-123",
        Status: domain.WeddingStatusPublished,
    }
    
    s.mockService.On("GetByID", mock.Anything, weddingID).
        Return(expectedWedding, nil).
        Once()
    
    req := httptest.NewRequest(http.MethodGet, "/weddings/"+weddingID, nil)
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Params = gin.Params{{Key: "id", Value: weddingID}}
    c.Set("userID", "user-123")
    
    s.handler.Get(c)
    
    assert.Equal(s.T(), http.StatusOK, w.Code)
    
    var response domain.Wedding
    json.Unmarshal(w.Body.Bytes(), &response)
    s.Equal(expectedWedding.ID, response.ID)
    s.Equal(expectedWedding.Title, response.Title)
}

func (s *WeddingHandlerTestSuite) TestGetWedding_NotFound() {
    weddingID := "non-existent"
    
    s.mockService.On("GetByID", mock.Anything, weddingID).
        Return(nil, domain.ErrWeddingNotFound).
        Once()
    
    req := httptest.NewRequest(http.MethodGet, "/weddings/"+weddingID, nil)
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Params = gin.Params{{Key: "id", Value: weddingID}}
    
    s.handler.Get(c)
    
    assert.Equal(s.T(), http.StatusNotFound, w.Code)
    
    var response map[string]string
    json.Unmarshal(w.Body.Bytes(), &response)
    s.Equal("wedding not found", response["error"])
}

func (s *WeddingHandlerTestSuite) TestGetWedding_Unauthorized() {
    weddingID := "wedding-123"
    wedding := &domain.Wedding{
        ID:     weddingID,
        Title:  "Test Wedding",
        UserID: "different-user",
    }
    
    s.mockService.On("GetByID", mock.Anything, weddingID).
        Return(wedding, nil).
        Once()
    
    req := httptest.NewRequest(http.MethodGet, "/weddings/"+weddingID, nil)
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Params = gin.Params{{Key: "id", Value: weddingID}}
    c.Set("userID", "user-123")
    
    s.handler.Get(c)
    
    assert.Equal(s.T(), http.StatusForbidden, w.Code)
    
    var response map[string]string
    json.Unmarshal(w.Body.Bytes(), &response)
    s.Equal("access denied", response["error"])
}

func (s *WeddingHandlerTestSuite) TestUpdateWedding_Success() {
    weddingID := "wedding-123"
    input := domain.UpdateWeddingInput{
        Title:       "Updated Title",
        Description: "Updated description",
    }
    
    existingWedding := &domain.Wedding{
        ID:          weddingID,
        Title:       "Original Title",
        Description: "Original description",
        UserID:      "user-123",
    }
    
    updatedWedding := &domain.Wedding{
        ID:          weddingID,
        Title:       input.Title,
        Description: input.Description,
        UserID:      "user-123",
        UpdatedAt:   time.Now(),
    }
    
    s.mockService.On("GetByID", mock.Anything, weddingID).
        Return(existingWedding, nil).
        Once()
    
    s.mockService.On("Update", mock.Anything, weddingID, input).
        Return(updatedWedding, nil).
        Once()
    
    jsonData, _ := json.Marshal(input)
    req := httptest.NewRequest(http.MethodPut, "/weddings/"+weddingID, bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Params = gin.Params{{Key: "id", Value: weddingID}}
    c.Set("userID", "user-123")
    
    s.handler.Update(c)
    
    assert.Equal(s.T(), http.StatusOK, w.Code)
    
    var response domain.Wedding
    json.Unmarshal(w.Body.Bytes(), &response)
    s.Equal(input.Title, response.Title)
    s.Equal(input.Description, response.Description)
}

func (s *WeddingHandlerTestSuite) TestDeleteWedding_Success() {
    weddingID := "wedding-123"
    wedding := &domain.Wedding{
        ID:     weddingID,
        Title:  "Test Wedding",
        UserID: "user-123",
    }
    
    s.mockService.On("GetByID", mock.Anything, weddingID).
        Return(wedding, nil).
        Once()
    
    s.mockService.On("Delete", mock.Anything, weddingID).
        Return(nil).
        Once()
    
    req := httptest.NewRequest(http.MethodDelete, "/weddings/"+weddingID, nil)
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Params = gin.Params{{Key: "id", Value: weddingID}}
    c.Set("userID", "user-123")
    
    s.handler.Delete(c)
    
    assert.Equal(s.T(), http.StatusNoContent, w.Code)
    assert.Empty(s.T(), w.Body.String())
}

func (s *WeddingHandlerTestSuite) TestGetStats_Success() {
    weddingID := "wedding-123"
    wedding := &domain.Wedding{
        ID:     weddingID,
        UserID: "user-123",
    }
    
    stats := &domain.WeddingStats{
        TotalGuests:   100,
        Attending:     75,
        Declined:      15,
        Pending:       10,
        DietaryCounts: map[string]int{"Vegetarian": 10, "Vegan": 5},
    }
    
    s.mockService.On("GetByID", mock.Anything, weddingID).
        Return(wedding, nil).
        Once()
    
    s.mockService.On("GetStats", mock.Anything, weddingID).
        Return(stats, nil).
        Once()
    
    req := httptest.NewRequest(http.MethodGet, "/weddings/"+weddingID+"/stats", nil)
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Params = gin.Params{{Key: "id", Value: weddingID}}
    c.Set("userID", "user-123")
    
    s.handler.GetStats(c)
    
    assert.Equal(s.T(), http.StatusOK, w.Code)
    
    var response domain.WeddingStats
    json.Unmarshal(w.Body.Bytes(), &response)
    s.Equal(stats.TotalGuests, response.TotalGuests)
    s.Equal(stats.Attending, response.Attending)
}

func (s *WeddingHandlerTestSuite) TestPublishWedding_Success() {
    weddingID := "wedding-123"
    wedding := &domain.Wedding{
        ID:     weddingID,
        Title:  "Test Wedding",
        UserID: "user-123",
        Status: domain.WeddingStatusDraft,
    }
    
    publishedWedding := &domain.Wedding{
        ID:        weddingID,
        Title:     wedding.Title,
        UserID:    "user-123",
        Status:    domain.WeddingStatusPublished,
        PublicURL: "https://weddings.example.com/w/test-wedding-abc123",
    }
    
    s.mockService.On("GetByID", mock.Anything, weddingID).
        Return(wedding, nil).
        Once()
    
    s.mockService.On("Publish", mock.Anything, weddingID).
        Return(publishedWedding, nil).
        Once()
    
    req := httptest.NewRequest(http.MethodPost, "/weddings/"+weddingID+"/publish", nil)
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Params = gin.Params{{Key: "id", Value: weddingID}}
    c.Set("userID", "user-123")
    
    s.handler.Publish(c)
    
    assert.Equal(s.T(), http.StatusOK, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    s.Equal("published", response["status"])
    s.NotNil(response["public_url"])
}

func (s *WeddingHandlerTestSuite) TestPublishWedding_AlreadyPublished() {
    weddingID := "wedding-123"
    wedding := &domain.Wedding{
        ID:     weddingID,
        Title:  "Test Wedding",
        UserID: "user-123",
        Status: domain.WeddingStatusPublished,
    }
    
    s.mockService.On("GetByID", mock.Anything, weddingID).
        Return(wedding, nil).
        Once()
    
    s.mockService.On("Publish", mock.Anything, weddingID).
        Return(nil, domain.ErrWeddingAlreadyPublished).
        Once()
    
    req := httptest.NewRequest(http.MethodPost, "/weddings/"+weddingID+"/publish", nil)
    
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = req
    c.Params = gin.Params{{Key: "id", Value: weddingID}}
    c.Set("userID", "user-123")
    
    s.handler.Publish(c)
    
    assert.Equal(s.T(), http.StatusBadRequest, w.Code)
    
    var response map[string]string
    json.Unmarshal(w.Body.Bytes(), &response)
    s.Equal("wedding is already published", response["error"])
}
```

### RSVP Repository Test (Complete with Testcontainers)

```go
// repository/rsvp_repository_test.go
package repository

import (
    "context"
    "fmt"
    "testing"
    "time"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"
    "github.com/testcontainers/testcontainers-go"
    "github.com/testcontainers/testcontainers-go/modules/mongodb"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    
    "wedding-invitation/backend/internal/domain"
)

type RSVPRepositoryTestSuite struct {
    suite.Suite
    container *mongodb.MongoDBContainer
    client    *mongo.Client
    repo      *RSVPRepository
    ctx       context.Context
    dbName    string
}

func (s *RSVPRepositoryTestSuite) SetupSuite() {
    s.ctx = context.Background()
    s.dbName = "test_rsvp_db"
    
    // Start MongoDB container
    container, err := mongodb.Run(s.ctx, "mongo:6.0")
    s.Require().NoError(err)
    s.container = container
    
    // Get connection string
    connStr, err := container.ConnectionString(s.ctx)
    s.Require().NoError(err)
    
    // Connect to MongoDB
    ctx, cancel := context.WithTimeout(s.ctx, 10*time.Second)
    defer cancel()
    
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(connStr))
    s.Require().NoError(err)
    s.client = client
    
    // Verify connection
    err = client.Ping(ctx, nil)
    s.Require().NoError(err)
    
    // Create repository
    s.repo = NewRSVPRepository(client, s.dbName)
    
    // Create indexes
    s.createIndexes()
}

func (s *RSVPRepositoryTestSuite) TearDownSuite() {
    ctx := context.Background()
    
    if s.client != nil {
        s.client.Disconnect(ctx)
    }
    
    if s.container != nil {
        s.container.Terminate(ctx)
    }
}

func (s *RSVPRepositoryTestSuite) SetupTest() {
    // Clean up before each test
    db := s.client.Database(s.dbName)
    collections := []string{"rsvps", "weddings"}
    
    for _, coll := range collections {
        err := db.Collection(coll).Drop(s.ctx)
        if err != nil {
            s.T().Logf("Failed to drop collection %s: %v", coll, err)
        }
    }
    
    // Recreate indexes
    s.createIndexes()
}

func TestRSVPRepositorySuite(t *testing.T) {
    suite.Run(t, new(RSVPRepositoryTestSuite))
}

func (s *RSVPRepositoryTestSuite) createIndexes() {
    db := s.client.Database(s.dbName)
    collection := db.Collection("rsvps")
    
    // Create unique index on wedding_id + guest_email
    indexModel := mongo.IndexModel{
        Keys: bson.D{
            {Key: "wedding_id", Value: 1},
            {Key: "guest_email", Value: 1},
        },
        Options: options.Index().SetUnique(true),
    }
    
    _, err := collection.Indexes().CreateOne(s.ctx, indexModel)
    s.Require().NoError(err)
}

func generateTestID() string {
    return fmt.Sprintf("test_%d", time.Now().UnixNano())
}

func (s *RSVPRepositoryTestSuite) TestCreate_Success() {
    // Arrange
    rsvp := &domain.RSVP{
        ID:                  generateTestID(),
        WeddingID:           generateTestID(),
        GuestName:           "John Doe",
        GuestEmail:          "john@example.com",
        Status:              domain.RSVPStatusAttending,
        GuestCount:          2,
        DietaryRestrictions: "Vegetarian",
        Message:             "Can't wait!",
        CreatedAt:           time.Now(),
        UpdatedAt:           time.Now(),
    }
    
    // Act
    err := s.repo.Create(s.ctx, rsvp)
    
    // Assert
    s.NoError(err)
    
    // Verify in database
    found, err := s.repo.FindByID(s.ctx, rsvp.ID)
    s.NoError(err)
    s.NotNil(found)
    s.Equal(rsvp.GuestName, found.GuestName)
    s.Equal(rsvp.GuestEmail, found.GuestEmail)
    s.Equal(rsvp.Status, found.Status)
}

func (s *RSVPRepositoryTestSuite) TestCreate_DuplicateEmail() {
    // Arrange
    weddingID := generateTestID()
    email := "duplicate@example.com"
    
    rsvp1 := &domain.RSVP{
        ID:         generateTestID(),
        WeddingID:  weddingID,
        GuestName:  "First Guest",
        GuestEmail: email,
        Status:     domain.RSVPStatusAttending,
        CreatedAt:  time.Now(),
    }
    
    rsvp2 := &domain.RSVP{
        ID:         generateTestID(),
        WeddingID:  weddingID,
        GuestName:  "Second Guest",
        GuestEmail: email, // Same email
        Status:     domain.RSVPStatusPending,
        CreatedAt:  time.Now(),
    }
    
    // First insert
    err := s.repo.Create(s.ctx, rsvp1)
    s.NoError(err)
    
    // Second insert should fail
    err = s.repo.Create(s.ctx, rsvp2)
    s.Error(err)
    s.Contains(err.Error(), "duplicate")
}

func (s *RSVPRepositoryTestSuite) TestFindByID_Success() {
    // Arrange
    rsvp := &domain.RSVP{
        ID:         generateTestID(),
        WeddingID:  generateTestID(),
        GuestName:  "Jane Smith",
        GuestEmail: "jane@example.com",
        Status:     domain.RSVPStatusAttending,
        GuestCount: 3,
        CreatedAt:  time.Now(),
    }
    
    err := s.repo.Create(s.ctx, rsvp)
    s.NoError(err)
    
    // Act
    found, err := s.repo.FindByID(s.ctx, rsvp.ID)
    
    // Assert
    s.NoError(err)
    s.NotNil(found)
    s.Equal(rsvp.ID, found.ID)
    s.Equal(rsvp.GuestName, found.GuestName)
    s.Equal(rsvp.GuestCount, found.GuestCount)
}

func (s *RSVPRepositoryTestSuite) TestFindByID_NotFound() {
    // Act
    found, err := s.repo.FindByID(s.ctx, "non-existent-id")
    
    // Assert
    s.Error(err)
    s.Nil(found)
    s.Equal(domain.ErrRSVPNotFound, err)
}

func (s *RSVPRepositoryTestSuite) TestFindByWeddingID_Success() {
    // Arrange
    weddingID := generateTestID()
    otherWeddingID := generateTestID()
    
    rsvps := []*domain.RSVP{
        {
            ID:         generateTestID(),
            WeddingID:  weddingID,
            GuestName:  "Guest 1",
            GuestEmail: "guest1@example.com",
            Status:     domain.RSVPStatusAttending,
            GuestCount: 2,
            CreatedAt:  time.Now(),
        },
        {
            ID:         generateTestID(),
            WeddingID:  weddingID,
            GuestName:  "Guest 2",
            GuestEmail: "guest2@example.com",
            Status:     domain.RSVPStatusDeclined,
            GuestCount: 0,
            CreatedAt:  time.Now(),
        },
        {
            ID:         generateTestID(),
            WeddingID:  weddingID,
            GuestName:  "Guest 3",
            GuestEmail: "guest3@example.com",
            Status:     domain.RSVPStatusPending,
            GuestCount: 1,
            CreatedAt:  time.Now(),
        },
        {
            ID:         generateTestID(),
            WeddingID:  otherWeddingID, // Different wedding
            GuestName:  "Other Guest",
            GuestEmail: "other@example.com",
            Status:     domain.RSVPStatusAttending,
            CreatedAt:  time.Now(),
        },
    }
    
    for _, r := range rsvps {
        err := s.repo.Create(s.ctx, r)
        s.NoError(err)
    }
    
    // Act
    results, err := s.repo.FindByWeddingID(s.ctx, weddingID)
    
    // Assert
    s.NoError(err)
    s.Len(results, 3)
    
    // Verify order (by created_at desc)
    s.Equal("Guest 3", results[0].GuestName)
    s.Equal("Guest 2", results[1].GuestName)
    s.Equal("Guest 1", results[2].GuestName)
}

func (s *RSVPRepositoryTestSuite) TestFindByWeddingID_Empty() {
    weddingID := generateTestID()
    
    results, err := s.repo.FindByWeddingID(s.ctx, weddingID)
    
    s.NoError(err)
    s.Empty(results)
    s.NotNil(results)
}

func (s *RSVPRepositoryTestSuite) TestUpdate_Success() {
    // Arrange
    rsvp := &domain.RSVP{
        ID:         generateTestID(),
        WeddingID:  generateTestID(),
        GuestName:  "Original Name",
        GuestEmail: "original@example.com",
        Status:     domain.RSVPStatusPending,
        GuestCount: 1,
        CreatedAt:  time.Now(),
    }
    
    err := s.repo.Create(s.ctx, rsvp)
    s.NoError(err)
    
    // Update fields
    rsvp.GuestName = "Updated Name"
    rsvp.Status = domain.RSVPStatusAttending
    rsvp.GuestCount = 3
    rsvp.Message = "New message"
    rsvp.UpdatedAt = time.Now()
    
    // Act
    err = s.repo.Update(s.ctx, rsvp)
    
    // Assert
    s.NoError(err)
    
    // Verify update
    found, err := s.repo.FindByID(s.ctx, rsvp.ID)
    s.NoError(err)
    s.Equal("Updated Name", found.GuestName)
    s.Equal(domain.RSVPStatusAttending, found.Status)
    s.Equal(3, found.GuestCount)
    s.Equal("New message", found.Message)
}

func (s *RSVPRepositoryTestSuite) TestUpdate_NotFound() {
    rsvp := &domain.RSVP{
        ID:        generateTestID(),
        WeddingID: generateTestID(),
        GuestName: "Ghost",
        Status:    domain.RSVPStatusPending,
        UpdatedAt: time.Now(),
    }
    
    err := s.repo.Update(s.ctx, rsvp)
    
    s.Error(err)
    s.Equal(domain.ErrRSVPNotFound, err)
}

func (s *RSVPRepositoryTestSuite) TestDelete_Success() {
    // Arrange
    rsvp := &domain.RSVP{
        ID:         generateTestID(),
        WeddingID:  generateTestID(),
        GuestName:  "To Delete",
        GuestEmail: "delete@example.com",
        Status:     domain.RSVPStatusPending,
        CreatedAt:  time.Now(),
    }
    
    err := s.repo.Create(s.ctx, rsvp)
    s.NoError(err)
    
    // Verify exists
    found, err := s.repo.FindByID(s.ctx, rsvp.ID)
    s.NoError(err)
    s.NotNil(found)
    
    // Act
    err = s.repo.Delete(s.ctx, rsvp.ID)
    
    // Assert
    s.NoError(err)
    
    // Verify deletion
    deleted, err := s.repo.FindByID(s.ctx, rsvp.ID)
    s.Error(err)
    s.Nil(deleted)
}

func (s *RSVPRepositoryTestSuite) TestDelete_NotFound() {
    err := s.repo.Delete(s.ctx, "non-existent-id")
    
    s.Error(err)
    s.Equal(domain.ErrRSVPNotFound, err)
}

func (s *RSVPRepositoryTestSuite) TestGetStatsByWeddingID_Success() {
    // Arrange
    weddingID := generateTestID()
    
    rsvps := []*domain.RSVP{
        {
            ID:                  generateTestID(),
            WeddingID:           weddingID,
            GuestName:           "Attending 1",
            GuestEmail:          "a1@example.com",
            Status:              domain.RSVPStatusAttending,
            GuestCount:          2,
            DietaryRestrictions: "Vegetarian",
            CreatedAt:           time.Now(),
        },
        {
            ID:                  generateTestID(),
            WeddingID:           weddingID,
            GuestName:           "Attending 2",
            GuestEmail:          "a2@example.com",
            Status:              domain.RSVPStatusAttending,
            GuestCount:          3,
            DietaryRestrictions: "Vegan",
            CreatedAt:           time.Now(),
        },
        {
            ID:         generateTestID(),
            WeddingID:  weddingID,
            GuestName:  "Declined 1",
            GuestEmail: "d1@example.com",
            Status:     domain.RSVPStatusDeclined,
            GuestCount: 0,
            CreatedAt:  time.Now(),
        },
        {
            ID:         generateTestID(),
            WeddingID:  weddingID,
            GuestName:  "Pending 1",
            GuestEmail: "p1@example.com",
            Status:     domain.RSVPStatusPending,
            GuestCount: 1,
            CreatedAt:  time.Now(),
        },
    }
    
    for _, r := range rsvps {
        err := s.repo.Create(s.ctx, r)
        s.NoError(err)
    }
    
    // Act
    stats, err := s.repo.GetStatsByWeddingID(s.ctx, weddingID)
    
    // Assert
    s.NoError(err)
    s.NotNil(stats)
    s.Equal(2, stats.Attending)
    s.Equal(1, stats.Declined)
    s.Equal(1, stats.Pending)
    s.Equal(5, stats.TotalGuests) // 2 + 3 + 0 + 0 (pending doesn't count until confirmed)
    
    // Verify dietary restrictions
    s.Equal(1, stats.DietaryCounts["Vegetarian"])
    s.Equal(1, stats.DietaryCounts["Vegan"])
}

func (s *RSVPRepositoryTestSuite) TestGetStatsByWeddingID_Empty() {
    weddingID := generateTestID()
    
    stats, err := s.repo.GetStatsByWeddingID(s.ctx, weddingID)
    
    s.NoError(err)
    s.NotNil(stats)
    s.Equal(0, stats.Attending)
    s.Equal(0, stats.Declined)
    s.Equal(0, stats.Pending)
    s.Equal(0, stats.TotalGuests)
}

func (s *RSVPRepositoryTestSuite) TestFindByWeddingIDAndEmail_Success() {
    // Arrange
    weddingID := generateTestID()
    email := "find@example.com"
    
    rsvp := &domain.RSVP{
        ID:         generateTestID(),
        WeddingID:  weddingID,
        GuestName:  "Find Me",
        GuestEmail: email,
        Status:     domain.RSVPStatusAttending,
        CreatedAt:  time.Now(),
    }
    
    err := s.repo.Create(s.ctx, rsvp)
    s.NoError(err)
    
    // Act
    found, err := s.repo.FindByWeddingIDAndEmail(s.ctx, weddingID, email)
    
    // Assert
    s.NoError(err)
    s.NotNil(found)
    s.Equal(rsvp.ID, found.ID)
    s.Equal(rsvp.GuestName, found.GuestName)
}

func (s *RSVPRepositoryTestSuite) TestFindByWeddingIDAndEmail_NotFound() {
    found, err := s.repo.FindByWeddingIDAndEmail(s.ctx, "wedding-id", "email@example.com")
    
    s.Error(err)
    s.Nil(found)
    s.Equal(domain.ErrRSVPNotFound, err)
}

func (s *RSVPRepositoryTestSuite) TestCountByWeddingID_Success() {
    // Arrange
    weddingID := generateTestID()
    
    for i := 0; i < 5; i++ {
        rsvp := &domain.RSVP{
            ID:         generateTestID(),
            WeddingID:  weddingID,
            GuestName:  fmt.Sprintf("Guest %d", i),
            GuestEmail: fmt.Sprintf("guest%d@example.com", i),
            Status:     domain.RSVPStatusAttending,
            CreatedAt:  time.Now(),
        }
        err := s.repo.Create(s.ctx, rsvp)
        s.NoError(err)
    }
    
    // Act
    count, err := s.repo.CountByWeddingID(s.ctx, weddingID)
    
    // Assert
    s.NoError(err)
    s.Equal(int64(5), count)
}

func (s *RSVPRepositoryTestSuite) TestBulkCreate_Success() {
    // Arrange
    weddingID := generateTestID()
    
    rsvps := []*domain.RSVP{
        {
            ID:         generateTestID(),
            WeddingID:  weddingID,
            GuestName:  "Bulk 1",
            GuestEmail: "bulk1@example.com",
            Status:     domain.RSVPStatusPending,
            CreatedAt:  time.Now(),
        },
        {
            ID:         generateTestID(),
            WeddingID:  weddingID,
            GuestName:  "Bulk 2",
            GuestEmail: "bulk2@example.com",
            Status:     domain.RSVPStatusPending,
            CreatedAt:  time.Now(),
        },
        {
            ID:         generateTestID(),
            WeddingID:  weddingID,
            GuestName:  "Bulk 3",
            GuestEmail: "bulk3@example.com",
            Status:     domain.RSVPStatusPending,
            CreatedAt:  time.Now(),
        },
    }
    
    // Act
    err := s.repo.BulkCreate(s.ctx, rsvps)
    
    // Assert
    s.NoError(err)
    
    // Verify all created
    for _, r := range rsvps {
        found, err := s.repo.FindByID(s.ctx, r.ID)
        s.NoError(err)
        s.NotNil(found)
    }
}

func (s *RSVPRepositoryTestSuite) TestContextCancellation() {
    // Create a cancelled context
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    
    rsvp := &domain.RSVP{
        ID:         generateTestID(),
        WeddingID:  generateTestID(),
        GuestName:  "Test",
        GuestEmail: "test@example.com",
        Status:     domain.RSVPStatusPending,
        CreatedAt:  time.Now(),
    }
    
    err := s.repo.Create(ctx, rsvp)
    s.Error(err)
    s.Contains(err.Error(), "context")
}
```

---

## Quick Reference

### Running Specific Tests

```bash
# Run single test
go test -run TestUserService_CreateUser ./internal/services

# Run test suite
go test -run TestUserServiceSuite ./internal/services

# Run specific sub-test
go test -run "TestUserServiceSuite/TestCreateUser_Success" ./internal/services

# Run with verbose output
go test -v ./internal/services
```

### Test Shortcuts

```bash
# Run only quick tests
go test -short ./...

# Run with timeout
go test -timeout=30s ./...

# Run with race detection
go test -race ./...

# Fail fast (stop on first failure)
go test -failfast ./...
```

### Coverage Commands

```bash
# View coverage in browser
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out

# View coverage by function
go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out | grep -v "mocks\|testutil"

# Coverage for specific package
go test -coverprofile=services.out ./internal/services/... && go tool cover -func=services.out
```

### Mock Generation

```bash
# Using mockery
go install github.com/vektra/mockery/v2@latest
mockery --all --output ./internal/mocks

# Using mockgen
go install github.com/golang/mock/mockgen@latest
go generate ./...
```

---

## Best Practices

### 1. Test Independence
- Each test should be independent
- Don't rely on test execution order
- Clean up resources in TearDown

### 2. Table-Driven Tests
- Use for multiple similar test cases
- Clear test case names
- Easy to add new cases

### 3. Mock Verification
- Always verify mock expectations
- Use `Once()` for one-time calls
- Use `mock.Anything` for irrelevant parameters

### 4. Test Naming
- `Test<FunctionName>_<Scenario>` for unit tests
- `Test<SuiteName>_<MethodName>_<Scenario>` for suites
- Clear, descriptive names

### 5. Assertions
- Use `assert` for non-fatal checks
- Use `require` for fatal checks (stop test immediately)
- Check both success and error cases

### 6. Integration Test Containers
- Always clean up containers
- Use unique database names
- Handle container startup failures gracefully

### 7. Test Data
- Use fixtures for complex data
- Generate unique IDs to avoid conflicts
- Clean up test data after each test

### 8. Parallel Execution
- Use `t.Parallel()` where safe
- Avoid shared state between tests
- Use `sync/atomic` for shared counters

---

This testing guide provides comprehensive coverage of testing strategies for the Wedding Invitation backend. Follow these patterns to maintain high code quality, catch bugs early, and ensure reliable deployments.
