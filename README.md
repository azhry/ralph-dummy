# Wedding Invitation Backend

A production-ready backend for the Wedding Invitation system built with Go and MongoDB. This comprehensive backend provides complete functionality for managing wedding invitations, guest lists, RSVPs, and analytics.

## 📚 Documentation

| Document | Description |
|----------|-------------|
| **[QUICK_START.md](./QUICK_START.md)** | Step-by-step setup guide for new users |
| **[VISUAL_GUIDE.md](./VISUAL_GUIDE.md)** | System architecture and workflow diagrams |
| **[USER_GUIDE.md](./USER_GUIDE.md)** | Comprehensive API usage examples |
| **[docs/API_DOCUMENTATION.md](./docs/API_DOCUMENTATION.md)** | Complete REST API reference |
| **[AGENTS.md](./AGENTS.md)** | Development documentation for contributors |

## 🚀 Quick Start

**New to the project?** Start with [QUICK_START.md](./QUICK_START.md) for a step-by-step setup guide.

### Prerequisites
- Go 1.21+
- MongoDB 6.0+
- Redis (optional, for rate limiting)

### One-Command Setup

```bash
# Quick setup (if you have Docker and Go)
git clone <repository-url> && cd wedding-invitation-backend && \
cp .env.example .env && go mod tidy && \
docker run -d --name wedding-mongo -p 27017:27017 mongo:6.0 && \
go run cmd/api/main.go
```

### Detailed Setup

For detailed setup instructions, troubleshooting, and first API calls, see [QUICK_START.md](./QUICK_START.md).

### Run Tests

```bash
# Run all tests
go test ./... -v

# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 📚 API Documentation

### Base URL
```
Development: http://localhost:8080/api/v1
Production: https://api.yourdomain.com/api/v1
```

### Authentication Headers
```bash
Authorization: Bearer <access_token>
Content-Type: application/json
```

## 🎯 Core API Endpoints

### Authentication
```bash
# Register new user
POST /api/v1/auth/register
{
  "email": "user@example.com",
  "password": "SecurePass123!",
  "first_name": "John",
  "last_name": "Doe"
}

# Login
POST /api/v1/auth/login
{
  "email": "user@example.com",
  "password": "SecurePass123!"
}

# Refresh token
POST /api/v1/auth/refresh
{
  "refresh_token": "<refresh_token>"
}
```

### Wedding Management
```bash
# Create wedding
POST /api/v1/weddings
{
  "title": "John & Jane's Wedding",
  "groom_name": "John Doe",
  "bride_name": "Jane Smith",
  "wedding_date": "2024-06-15T14:00:00Z",
  "venue": "Beautiful Garden",
  "description": "Join us for our special day"
}

# Get wedding by slug (public)
GET /api/v1/public/weddings/john-jane-wedding
```

### Guest Management
```bash
# Add guest
POST /api/v1/weddings/{wedding_id}/guests
{
  "first_name": "Bob",
  "last_name": "Wilson",
  "email": "bob@example.com",
  "relationship": "friend",
  "side": "groom"
}

# Import guests from CSV
POST /api/v1/weddings/{wedding_id}/guests/import
Content-Type: multipart/form-data
file: guests.csv
```

### RSVP Management
```bash
# Submit RSVP (public)
POST /api/v1/public/weddings/{slug}/rsvp
{
  "first_name": "Alice",
  "last_name": "Johnson",
  "email": "alice@example.com",
  "status": "attending",
  "attendance_count": 2,
  "dietary_restrictions": "Vegetarian"
}

# Get RSVP statistics
GET /api/v1/weddings/{wedding_id}/rsvps/statistics
```

### File Upload
```bash
# Upload wedding photo
POST /api/v1/upload
Content-Type: multipart/form-data
file: wedding_photo.jpg
wedding_id: <wedding_id>
type: "wedding_photo"
```

### Analytics
```bash
# Track page view (client-side)
POST /api/v1/analytics/track/page-view
{
  "wedding_id": "<wedding_id>",
  "page": "invitation",
  "referrer": "https://facebook.com"
}

# Get wedding analytics
GET /api/v1/weddings/{wedding_id}/analytics
```

## 🔧 Configuration

### Required Environment Variables
```bash
# Database
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=wedding_invitations

# Authentication
JWT_SECRET=your-super-secret-jwt-key
JWT_REFRESH_SECRET=your-super-secret-refresh-key

# Server
PORT=8080
APP_ENV=development
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
```

### Optional Configuration
```bash
# File Storage
STORAGE_PROVIDER=local  # or "aws_s3", "cloudflare_r2"
UPLOAD_LOCAL_PATH=./uploads

# Email
EMAIL_PROVIDER=sendgrid
SENDGRID_API_KEY=your-api-key

# Rate Limiting
REDIS_URL=redis://localhost:6379
```

## 🏗️ Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend API   │    │   Database      │
│   (Web/Mobile)  │───▶│   (Go + Gin)    │───▶│   (MongoDB)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   External      │
                       │   Services      │
                       │ (S3, SendGrid) │
                       └─────────────────┘
```

## 📊 Features in Detail

### 🔐 Security
- JWT access tokens with refresh token rotation
- bcrypt password hashing (cost 12)
- Rate limiting on all endpoints
- CORS protection with configurable origins
- Security headers (CSP, HSTS, XSS protection)
- Input validation and sanitization

### 📈 Analytics & Insights
- Page view tracking with referrer analysis
- RSVP conversion funnel tracking
- Guest engagement metrics
- Real-time statistics dashboard
- CSV export functionality

### 📁 File Management
- Multi-format image support (JPEG, PNG, WebP)
- Automatic thumbnail generation
- Cloud storage integration (S3/R2)
- Presigned URL support for direct uploads
- Media metadata tracking

### 👥 Guest Management
- Individual and bulk guest creation
- CSV import with error handling
- Guest categorization (side, relationship, VIP)
- RSVP status tracking
- Email notification system

## 🧪 Testing

### Test Coverage
- **Unit Tests**: Models, services, utilities (80%+ coverage)
- **Integration Tests**: Repository layer, API endpoints
- **Handler Tests**: HTTP request/response validation

### Running Tests
```bash
# Run specific test suites
go test ./internal/domain/models -v
go test ./internal/services -v
go test ./internal/handlers -v

# Run with race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## 📦 Deployment

### Production Deployment

```bash
# Build for production
go build -o wedding-api cmd/api/main.go

# Using Docker
docker build -t wedding-api .
docker run -p 8080:8080 wedding-api

# With Docker Compose
docker-compose -f docker-compose.prod.yml up -d
```

### Environment Setup
1. **Development**: Use `docker-compose.yml` with local MongoDB
2. **Staging**: Deploy with external MongoDB and Redis
3. **Production**: Use managed services (MongoDB Atlas, ElastiCache)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass (`go test ./...`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

## 📝 API Response Format

### Success Response
```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "title": "John & Jane's Wedding"
  },
  "meta": {
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "req-123"
  }
}
```

### Error Response
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
  }
}
```

## 🛠 Health Check

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "ok",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "1.0.0"
}
```

## 📁 Project Structure

```
wedding-invitation-backend/
├── cmd/
│   └── api/                 # Application entry point
│       └── main.go          # Main server file
├── internal/
│   ├── config/              # Configuration management
│   ├── domain/              # Domain layer
│   │   ├── models/          # Domain entities (User, Wedding, RSVP, etc.)
│   │   └── repository/      # Repository interfaces
│   ├── handlers/            # HTTP handlers (controllers)
│   ├── middleware/          # Gin middleware (auth, security, logging)
│   ├── services/            # Business logic layer
│   ├── utils/               # Utility functions
│   └── dto/                 # Data transfer objects
├── pkg/                     # Public packages
│   ├── database/            # Database connections
│   ├── storage/             # File storage implementations
│   ├── email/               # Email service implementations
│   └── cache/               # Cache implementations
├── tests/                   # Test files
│   ├── unit/                # Unit tests
│   ├── integration/         # Integration tests
│   └── e2e/                 # End-to-end tests
├── docs/                    # Documentation
│   └── backend/             # Backend specification docs
├── scripts/                 # Utility scripts
├── docker-compose.yml       # Development environment
├── Dockerfile              # Production build
├── .env.example            # Environment template
├── go.mod                  # Go module definition
└── README.md               # This file
```

## 🔧 Environment Variables

### Setup
Copy `.env.example` to `.env` and configure the required variables:

```bash
cp .env.example .env
```

### Configuration Categories

#### Database Configuration
```bash
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=wedding_invitations
MONGODB_TIMEOUT_SECONDS=10
```

#### Authentication Configuration
```bash
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_REFRESH_SECRET=your-super-secret-refresh-key-change-in-production
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=168h
BCRYPT_COST=12
```

#### Server Configuration
```bash
PORT=8080
APP_ENV=development
ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8080
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s
```

#### Storage Configuration
```bash
STORAGE_PROVIDER=local
AWS_REGION=us-east-1
S3_BUCKET_NAME=your-wedding-app-bucket
AWS_ACCESS_KEY_ID=your-aws-access-key
AWS_SECRET_ACCESS_KEY=your-aws-secret-key
CDN_URL=
UPLOAD_LOCAL_PATH=./uploads
```

#### Email Configuration
```bash
EMAIL_PROVIDER=sendgrid
SENDGRID_API_KEY=your-sendgrid-api-key
EMAIL_FROM=noreply@yourdomain.com
```

## 🤝 Contributing

### Development Workflow

1. **Fork the Repository**
   ```bash
   git clone https://github.com/your-username/wedding-invitation-backend.git
   cd wedding-invitation-backend
   ```

2. **Set Up Development Environment**
   ```bash
   # Copy environment configuration
   cp .env.example .env
   # Install dependencies
   go mod tidy
   # Start development services
   docker-compose up -d mongodb
   ```

3. **Create a Feature Branch**
   ```bash
   git checkout -b feature/amazing-feature
   ```

4. **Make Your Changes**
   - Follow the existing code style
   - Add tests for new functionality
   - Update documentation as needed

5. **Run Tests and Quality Checks**
   ```bash
   # Run all tests
   go test ./... -v
   
   # Run with coverage
   go test -coverprofile=coverage.out ./...
   
   # Run linter (if configured)
   golangci-lint run
   
   # Build the application
   go build -o main cmd/api/main.go
   ```

6. **Commit Your Changes**
   ```bash
   git add .
   git commit -m "feat: add amazing feature"
   ```

7. **Push and Create Pull Request**
   ```bash
   git push origin feature/amazing-feature
   # Create PR on GitHub
   ```

### Code Style Guidelines

- **Naming**: Use Go conventions (CamelCase for exported, camelCase for unexported)
- **Comments**: Add godoc comments for exported functions and types
- **Error Handling**: Always handle errors explicitly
- **Testing**: Write tests for all public functions and HTTP handlers
- **Dependencies**: Keep dependencies minimal and well-documented

### Pull Request Process

1. **Description**: Clearly describe what your PR does
2. **Testing**: Ensure all tests pass in CI
3. **Documentation**: Update relevant documentation
4. **Review**: Address feedback from maintainers
5. **Approval**: Wait for approval before merging

## 📞 Support

### Documentation
- **Backend Specification**: See `docs/backend/` for detailed API documentation
- **Database Schema**: See `docs/backend/02-database-schema.md`
- **API Reference**: See `docs/backend/03-api-reference.md`

### Common Issues

**MongoDB Connection Issues**
```bash
# Check if MongoDB is running
docker ps | grep mongo

# Check connection string in .env
MONGODB_URI=mongodb://localhost:27017
```

**Port Already in Use**
```bash
# Check what's using port 8080
lsof -i :8080

# Kill the process or change port in .env
PORT=8081
```

**Environment Variables Not Loading**
```bash
# Ensure .env file exists and is readable
ls -la .env

# Check .env.example for required variables
cat .env.example
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Gin](https://gin-gonic.com/) web framework
- Database powered by [MongoDB](https://www.mongodb.com/)
- Authentication secured with [JWT](https://jwt.io/)
- Testing with [Testify](https://github.com/stretchr/testify)

---

**Version**: 1.0.0  
**Last Updated**: 2026-02-07  
**Status**: Production Ready ✅