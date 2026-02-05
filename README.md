# Wedding Invitation Backend

A production-ready backend for the Wedding Invitation system built with Go and MongoDB.

## Features

- JWT-based authentication with refresh tokens
- Wedding management with custom slugs
- Guest management with CSV import
- RSVP collection and tracking
- File upload with cloud storage integration
- Email notifications
- Analytics and insights
- Rate limiting and security

## Technology Stack

- **Language**: Go 1.21+
- **Framework**: Gin
- **Database**: MongoDB
- **Authentication**: JWT + bcrypt
- **Cache**: Redis
- **File Storage**: AWS S3 / Cloudflare R2
- **Email**: SendGrid
- **Testing**: Testify

## Development Setup

```bash
# Install dependencies
go mod tidy

# Start development environment
docker-compose up -d

# Run the server
go run cmd/api/main.go

# Run tests
go test ./...
```

## API Documentation

Start the server and visit `http://localhost:8080/swagger/index.html` for interactive API documentation.

## Project Structure

```
cmd/api/                 # Application entry point
internal/
  config/               # Configuration management
  domain/
    models/             # Domain entities
    repository/         # Repository interfaces
  handler/              # HTTP handlers
  middleware/           # Middleware (auth, logging, etc.)
  service/              # Business logic
  utils/                # Utilities
  dto/                  # Data transfer objects
pkg/                    # Public packages
  database/             # Database connections
  storage/              # File storage implementations
  email/                # Email service implementations
  cache/                # Cache implementations
tests/
  unit/                 # Unit tests
  integration/          # Integration tests
  e2e/                  # End-to-end tests
```

## Environment Variables

Copy `.env.example` to `.env` and configure the required variables.

```bash
cp .env.example .env
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Ensure all tests pass
6. Submit a pull request