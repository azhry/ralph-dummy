# AGENTS.md - AI Agent Development Documentation

This file contains documentation for AI agents working on the Wedding Invitation Backend project. It includes development history, patterns, and guidelines for future AI agent work.

## 🤖 AI Agent Development History

### Initial Development Phase (Previous Iterations)

The Wedding Invitation Backend was developed through multiple AI agent iterations, with comprehensive implementation of all specification requirements.

## 🏗️ Architecture Patterns

### Project Structure
```
wedding-invitation-backend/
├── cmd/api/                    # Application entry point
├── internal/
│   ├── config/                 # Environment configuration
│   ├── domain/
│   │   ├── models/             # Domain entities
│   │   └── repository/         # Repository interfaces
│   ├── handlers/               # HTTP handlers (controllers)
│   ├── middleware/             # Gin middleware
│   ├── services/              # Business logic
│   ├── utils/                 # Utilities
│   └── dto/                   # Data transfer objects
├── pkg/                       # Reusable packages
│   ├── database/              # MongoDB connection
│   ├── storage/               # File storage
│   ├── email/                 # Email services
│   └── cache/                 # Redis caching
└── tests/                     # Test suites
```

### Design Patterns Used

1. **Repository Pattern**: Data access abstraction
2. **Service Layer**: Business logic separation
3. **Handler Pattern**: HTTP request/response handling
4. **Middleware Chain**: Cross-cutting concerns
5. **DTO Pattern**: Data transfer for API layers
6. **Factory Pattern**: Service creation
7. **Strategy Pattern**: Storage provider abstraction

### Dependency Injection
Services are injected into handlers, repositories into services, following clean architecture principles.

## 📋 Code Standards and Conventions

### Go Conventions
- **Naming**: CamelCase for exported, camelCase for unexported
- **Interfaces**: Defined in domain layer, implemented in infrastructure
- **Errors**: Explicit error handling with context
- **Testing**: Table-driven tests where appropriate

### File Organization
- **One type per file** for models
- **Interface separation** in domain/repository
- **Handler grouping** by resource type
- **Middleware separation** by concern

### Database Conventions
- **Collection names**: Plural (users, weddings, guests)
- **Field names**: snake_case in MongoDB, CamelCase in Go structs
- **Indexes**: Unique on email (users), slug (weddings)
- **Soft deletes**: Implemented with `deleted_at` timestamp

## 🔧 Technical Decisions (DEC-001 to DEC-050)

### Key Architectural Decisions

**DEC-001**: MongoDB as primary database
- **Reasoning**: Flexible schema for wedding data, good for JSON-heavy workloads
- **Confidence**: 85%
- **Alternatives**: PostgreSQL (considered but chose MongoDB for flexibility)

**DEC-002**: Gin web framework
- **Reasoning**: Lightweight, good performance, extensive middleware ecosystem
- **Confidence**: 90%
- **Alternatives**: Echo, Fiber (Gin chosen for maturity)

**DEC-003**: JWT authentication with refresh tokens
- **Reasoning**: Stateless, scalable, secure with proper rotation
- **Confidence**: 95%
- **Alternatives**: Session-based (JWT chosen for API-first design)

**DEC-004**: Local file storage with cloud abstraction
- **Reasoning**: Easy development, cloud-ready for production
- **Confidence**: 80%
- **Alternatives**: Direct S3 integration (abstraction provides flexibility)

## 🧪 Testing Strategy

### Test Coverage Requirements
- **Models**: 90%+ coverage (business logic)
- **Services**: 85%+ coverage (core functionality)
- **Handlers**: 80%+ coverage (HTTP layer)
- **Middleware**: 75%+ coverage (cross-cutting concerns)

### Test Categories
1. **Unit Tests**: Individual functions and methods
2. **Integration Tests**: Database operations, external services
3. **Handler Tests**: HTTP request/response cycles
4. **End-to-End Tests**: Complete user workflows

### Mock Strategy
- **Repository Interfaces**: Mocked for service layer tests
- **External Services**: Mocked for email, storage, caching
- **HTTP Clients**: Mocked for third-party API calls

## 🚀 Deployment Patterns

### Environment Configuration
- **Development**: Local MongoDB, Redis, file storage
- **Staging**: External services, production-like configuration
- **Production**: Managed services (MongoDB Atlas, ElastiCache, S3)

### Docker Strategy
- **Multi-stage builds**: Optimized production images
- **Separate services**: API, MongoDB, Redis, Nginx
- **Health checks**: Comprehensive endpoint monitoring

### CI/CD Pipeline
- **Automated testing**: All test suites, linting, security scanning
- **Automated deployment**: Tag-based releases to production
- **Rollback capability**: Previous version retention

## 🔍 Debugging and Troubleshooting

### Common Issues and Solutions

**MongoDB Connection Issues**
```bash
# Check connection string
MONGODB_URI=mongodb://localhost:27017/wedding_invitations

# Verify MongoDB is running
docker ps | grep mongo

# Test connection
mongosh $MONGODB_URI
```

**JWT Token Issues**
```bash
# Verify token contents
echo $TOKEN | cut -d'.' -f2 | base64 -d | jq .

# Check expiration
date -d @$(echo $TOKEN | cut -d'.' -f2 | base64 -d | jq '.exp')
```

**Rate Limiting Issues**
```bash
# Check Redis connection
redis-cli ping

# View rate limit keys
redis-cli --scan --pattern "ratelimit:*"

# Clear rate limits (dev only)
redis-cli flushall
```

## 📊 Performance Considerations

### Database Optimization
- **Indexes**: Email (users), slug (weddings), compound queries
- **Pagination**: Cursor-based for large datasets
- **Query Optimization**: Projection, limiting fields

### Caching Strategy
- **Application Layer**: Redis for frequent data
- **HTTP Caching**: ETags for static resources
- **CDN**: For uploaded media files

### File Upload Optimization
- **Streaming**: Direct to storage, not memory
- **Thumbnails**: Generated asynchronously
- **Compression**: WebP format when possible

## 🔄 Maintenance Patterns

### Database Maintenance
```bash
# Backup MongoDB
mongodump --uri="$MONGODB_URI" --out="/backup/$(date +%Y%m%d)"

# Restore MongoDB
mongorestore --uri="$MONGODB_URI" "/backup/20240115"

# Compact collections (maintenance)
db.runCommand({compact: "weddings"})
```

### Log Management
- **Structured Logging**: JSON format with zap
- **Log Levels**: Debug, Info, Warn, Error
- **Log Rotation**: Daily with size limits

### Monitoring
- **Health Endpoints**: `/health`, `/health/detailed`
- **Metrics Collection**: Prometheus-compatible
- **Alerting**: Error rates, response times, resource usage

## 🚨 Security Best Practices

### Input Validation
- **All Input**: Validated at handler layer
- **File Uploads**: Type, size, content validation
- **SQL Injection**: Parameterized queries (MongoDB driver)
- **XSS Prevention**: Content-Type headers, CSP

### Authentication Security
- **Password Hashing**: bcrypt with cost 12
- **JWT Secrets**: Environment-specific, rotated regularly
- **Token Expiration**: Short access, longer refresh
- **Rate Limiting**: Endpoint-specific limits

### Infrastructure Security
- **HTTPS Required**: Production environments
- **CORS Configuration**: Specific allowed origins
- **Security Headers**: HSTS, CSP, X-Frame-Options
- **Dependency Scanning**: Regular vulnerability checks

## 📚 Documentation Standards

### Code Documentation
- **Godoc Comments**: All exported functions/types
- **Example Usage**: Complex business logic
- **API Documentation**: Swagger/OpenAPI specifications
- **Architecture Decisions**: ADR format for major decisions

### API Documentation
- **Endpoint Descriptions**: Clear purpose and usage
- **Request/Response Examples**: JSON payloads
- **Error Codes**: Comprehensive error handling
- **Authentication**: Bearer token requirements

### User Documentation
- **Getting Started**: Quick setup guide
- **API Examples**: Real-world usage scenarios
- **Common Workflows**: End-to-end processes
- **Troubleshooting**: Common issues and solutions

## 🎯 Future Enhancement Roadmap

### Phase 6: Advanced Features (Future)
- **Real-time Notifications**: WebSocket support
- **Advanced Analytics**: Machine learning insights
- **Multi-language Support**: Internationalization
- **Mobile API**: Optimized for mobile clients

### Phase 7: Enterprise Features (Future)
- **Multi-tenancy**: Wedding planner accounts
- **White-labeling**: Custom branding options
- **Advanced Integrations**: Third-party service connections
- **Compliance**: GDPR, CCPA data handling

## 🤝 AI Agent Guidelines

### For Future AI Agents Working on This Project

1. **Understand the Architecture**: Review project structure and patterns
2. **Follow Conventions**: Use existing code style and patterns
3. **Test Thoroughly**: Maintain high test coverage
4. **Document Decisions**: Use ADR format for major changes
5. **Security First**: Validate all inputs and outputs
6. **Performance Matters**: Consider database and caching implications
7. **Error Handling**: Provide meaningful error messages
8. **Backward Compatibility**: Consider API versioning

### Common Tasks for AI Agents

1. **Feature Implementation**: Follow service → handler → test pattern
2. **Bug Fixes**: Identify root cause, add tests, fix systematically
3. **Performance Optimization**: Profile before optimizing
4. **Security Updates**: Regular dependency and security reviews
5. **Documentation Updates**: Keep docs in sync with code

### Development Workflow for AI Agents

1. **Analyze Requirements**: Understand what needs to be built
2. **Design Solution**: Plan architecture and implementation
3. **Write Tests**: Define expected behavior first
4. **Implement Code**: Write production-ready code
5. **Verify Functionality**: Test manually and automatically
6. **Update Documentation**: Document changes and decisions
7. **Commit Changes**: Use descriptive commit messages


---

**Document Version**: 1.3  
**Last Updated**: 2026-02-07  
**Agent**: opencode (big-pickle)  
**Status**: Production Ready ✅

## 🔍 Integration Testing (Latest)

### Integration Test Suite Created
- **Location**: `tests/integration/api_integration_test.go`
- **Coverage**: 13 comprehensive tests covering all API endpoints
- **Tests Passing**: 13/13 (100% success rate)
- **Focus**: Testing against running server with real HTTP requests

### Test Categories
1. **Basic Endpoints**: Health check, API info
2. **Wedding Management**: CRUD operations
3. **Guest Management**: Creation and retrieval
4. **RSVP Management**: Status tracking
5. **Error Handling**: Invalid requests, authentication
6. **Security**: Input validation, method restrictions
7. **Performance**: Concurrent request handling

### Running Integration Tests
```bash
# Start the backend server
go run cmd/api/main.go

# In another terminal, run integration tests
go test ./tests/integration -v
```

### Key Results
- ✅ All 13 integration tests pass
- ✅ Server starts correctly on port 8080
- ✅ Health endpoints respond properly
- ✅ API documentation accessible at `/swagger/index.html`
- ✅ Error handling works as expected
- ✅ Input validation prevents invalid data
- ✅ Concurrent requests handled gracefully

This document serves as a guide for future AI agents working on the Wedding Invitation Backend project. It contains architectural decisions, patterns, and best practices established during development.