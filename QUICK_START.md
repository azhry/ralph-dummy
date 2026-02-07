# Wedding Invitation Backend - Quick Start Guide

This guide provides step-by-step instructions for getting the Wedding Invitation Backend up and running quickly.

## 🚀 Quick Start

### Prerequisites
- Go 1.21+ installed
- MongoDB 6.0+ running locally or accessible
- Git for cloning the repository

### Step 1: Clone and Setup

```bash
# Clone the repository
git clone <repository-url>
cd wedding-invitation-backend

# Install dependencies
go mod tidy

# Copy environment configuration
cp .env.example .env
```

### Step 2: Configure Environment

Edit the `.env` file with your settings:

```bash
# Basic configuration
PORT=8080
APP_ENV=development

# Database - make sure MongoDB is running!
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=wedding_invitations

# JWT secrets (change these in production!)
JWT_SECRET=your-super-secret-jwt-key
JWT_REFRESH_SECRET=your-super-secret-refresh-key
```

### Step 3: Start MongoDB (if not already running)

```bash
# Using Docker (recommended)
docker run -d --name wedding-mongo \
  -p 27017:27017 \
  -e MONGO_INITDB_DATABASE=wedding_invitations \
  mongo:6.0

# Or install MongoDB locally following official docs
```

### Step 4: Run the Application

```bash
# Run the server
go run cmd/api/main.go

# You should see output like:
# 2024-01-15T10:30:00Z INFO api/main.go:67 Server starting on port 8080
# 2024-01-15T10:30:00Z INFO api/main.go:72 Database connected successfully
```

### Step 5: Verify Installation

```bash
# Check health endpoint
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy","timestamp":"2024-01-15T10:30:00Z"}
```

## 📝 First API Calls

### 1. Register a User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!",
    "first_name": "Test",
    "last_name": "User"
  }'
```

### 2. Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "SecurePass123!"
  }'
```

Save the `access_token` from the response for the next steps.

### 3. Create Your First Wedding

```bash
curl -X POST http://localhost:8080/api/v1/weddings \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-access-token>" \
  -d '{
    "title": "My Wedding",
    "slug": "my-wedding-2024",
    "description": "We invite you to celebrate our special day",
    "couple_info": {
      "partner1_name": "John Doe",
      "partner2_name": "Jane Smith",
      "partner1_role": "groom",
      "partner2_role": "bride"
    },
    "event_details": {
      "date": "2024-12-31T17:00:00Z",
      "venue_name": "Beautiful Venue",
      "venue_address": "123 Main St, City, State"
    }
  }'
```

## 🛠 Common Issues & Solutions

### Issue: "Failed to connect to database"
**Solution**: Make sure MongoDB is running and accessible at the URI specified in `.env`

```bash
# Test MongoDB connection
mongosh mongodb://localhost:27017/wedding_invitations
```

### Issue: "Port already in use"
**Solution**: Change the PORT in `.env` or stop the service using the port

```bash
# Find what's using port 8080
lsof -i :8080

# Kill the process (replace PID)
kill -9 <PID>
```

### Issue: "JWT token invalid"
**Solution**: Make sure you're using the correct token and it hasn't expired

```bash
# Decode JWT to check contents (use jwt.io or similar)
echo "your-token-here" | cut -d'.' -f2 | base64 -d
```

## 📊 Next Steps

Once you have the basic setup working:

1. **Explore the API**: Check the full API documentation in `docs/API_DOCUMENTATION.md`
2. **Test Features**: Try guest management, file uploads, and RSVP functionality
3. **Review Security**: Set up proper JWT secrets and enable security features
4. **Deploy**: Follow the deployment guide for production setup

## 🔧 Development Tips

### Running Tests
```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/handlers

# Run with verbose output
go test -v ./internal/services
```

### Building for Production
```bash
# Build binary
go build -o wedding-api ./cmd/api

# Run the binary
./wedding-api
```

### Environment Variables
Key environment variables to configure:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `APP_ENV` | Environment | `development` |
| `MONGODB_URI` | Database connection | `mongodb://localhost:27017` |
| `JWT_SECRET` | JWT signing secret | **Must be set** |
| `STORAGE_PROVIDER` | File storage | `local` |

## 📚 Additional Resources

- **Full API Documentation**: `docs/API_DOCUMENTATION.md`
- **User Guide**: `USER_GUIDE.md`
- **Development Guide**: `AGENTS.md`
- **Deployment Guide**: `docs/backend/09-deployment-guide.md`

## 🆘 Getting Help

If you encounter issues:

1. Check the logs for error messages
2. Verify your `.env` configuration
3. Ensure MongoDB is running and accessible
4. Review the troubleshooting section in the full documentation

---

**Need more help?** Check the comprehensive documentation in the `docs/` directory or open an issue on GitHub.