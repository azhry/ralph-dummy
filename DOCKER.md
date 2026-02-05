# Wedding Invitation Backend - Docker Development Guide

## Overview

This guide explains how to set up and run the Wedding Invitation Backend using Docker for development and production environments.

## Prerequisites

- Docker and Docker Compose installed on your system
- At least 4GB of available RAM
- At least 10GB of free disk space

## Quick Start

### Development Environment

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd wedding-invitation-backend
   ```

2. **Start development environment:**
   ```bash
   # Start only databases (MongoDB + Redis)
   docker-compose -f docker-compose.dev.yml up -d
   
   # Or start with the application
   docker-compose -f docker-compose.yml up --build
   ```

3. **Verify setup:**
   - MongoDB: `mongodb://admin:password123@localhost:27017`
   - Redis: `redis://localhost:6379`
   - API: `http://localhost:8080`
   - Health Check: `http://localhost:8080/health`

### Production Environment

1. **Start production environment:**
   ```bash
   docker-compose --profile production up -d --build
   ```

2. **Access the application:**
   - HTTP: `http://localhost`
   - HTTPS: `https://localhost` (if SSL certificates are configured)

## Environment Variables

### Required Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
# Database
DATABASE_URI=mongodb://admin:password123@mongodb:27017/wedding_invitations?authSource=admin
DATABASE_NAME=wedding_invitations

# Authentication
AUTH_JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
AUTH_JWT_REFRESH_SECRET=your-super-secret-refresh-key-change-this-in-production

# Redis
REDIS_URL=redis://redis:6379
```

### Optional Environment Variables

See `docker-compose.yml` for a complete list of configurable options.

## Services

### MongoDB
- **Container:** `wedding-mongodb`
- **Port:** `27017`
- **Default Credentials:** `admin:password123`
- **Database:** `wedding_invitations`
- **Data Volume:** `mongodb_data`

### Redis
- **Container:** `wedding-redis`
- **Port:** `6379`
- **Data Volume:** `redis_data`
- **Purpose:** Token blacklist, caching, session storage

### Application
- **Container:** `wedding-api`
- **Port:** `8080`
- **Health Check:** `/health`
- **Mounts:** `./uploads:/root/uploads`

### Nginx (Production Only)
- **Container:** `wedding-nginx`
- **Ports:** `80:80`, `443:443`
- **Config:** `./nginx/nginx.conf`

## Development Workflow

### Running the Application Locally

1. **Start databases only:**
   ```bash
   docker-compose -f docker-compose.dev.yml up -d
   ```

2. **Run the Go application locally:**
   ```bash
   export DATABASE_URI="mongodb://admin:password123@localhost:27017/wedding_invitations?authSource=admin"
   export REDIS_URL="redis://localhost:6379"
   go run cmd/api/main.go
   ```

### Running Tests

```bash
# Run unit tests
go test ./...

# Run integration tests with Docker
docker-compose -f docker-compose.yml up -d mongodb redis
go test -tags=integration ./...
```

### Database Management

```bash
# Connect to MongoDB
docker exec -it wedding-mongodb mongosh -u admin -p password123 --authenticationDatabase admin

# Backup database
docker exec wedding-mongodb mongodump --uri="mongodb://admin:password123@localhost:27017/wedding_invitations?authSource=admin" -o /backup

# Restore database
docker exec wedding-mongodb mongorestore --uri="mongodb://admin:password123@localhost:27017/wedding_invitations?authSource=admin" /backup/wedding_invitations
```

## Production Deployment

### Environment Setup

1. **Create production environment file:**
   ```bash
   cp .env.example .env.production
   # Edit .env.production with production values
   ```

2. **Configure SSL certificates:**
   ```bash
   mkdir -p nginx/ssl
   # Copy your SSL certificates to nginx/ssl/
   ```

3. **Deploy with production profile:**
   ```bash
   docker-compose --env-file .env.production --profile production up -d --build
   ```

### Monitoring and Logging

```bash
# View logs
docker-compose logs -f app
docker-compose logs -f mongodb
docker-compose logs -f redis

# Monitor container status
docker-compose ps

# Check resource usage
docker stats
```

### Scaling

```bash
# Scale application containers
docker-compose up -d --scale app=3

# Scale with load balancer
docker-compose --profile production up -d --scale app=3
```

## Troubleshooting

### Common Issues

1. **Port conflicts:**
   - Ensure ports 8080, 27017, and 6379 are available
   - Modify ports in `docker-compose.yml` if needed

2. **Permission issues:**
   - Ensure proper file permissions for mounted volumes
   - Run with appropriate user permissions

3. **Database connection issues:**
   - Verify MongoDB is healthy: `docker-compose ps mongodb`
   - Check connection string in environment variables
   - Ensure proper authentication credentials

4. **Memory issues:**
   - Increase Docker memory allocation
   - Monitor resource usage with `docker stats`

### Health Checks

```bash
# Check application health
curl http://localhost:8080/health

# Check container health
docker-compose ps
docker inspect wedding-api | grep -A 10 Health

# Check database connectivity
docker exec wedding-mongodb mongosh --eval "db.adminCommand('ping')"
```

### Logs and Debugging

```bash
# View application logs
docker-compose logs -f app

# View MongoDB logs
docker-compose logs -f mongodb

# Debug with interactive shell
docker exec -it wedding-api sh
```

## Security Considerations

1. **Change default passwords** before deploying to production
2. **Use environment variables** for sensitive configuration
3. **Enable HTTPS** in production
4. **Configure firewall rules** appropriately
5. **Regularly update** base images and dependencies
6. **Use secrets management** for production secrets

## Performance Optimization

1. **Enable resource limits** in production
2. **Use Redis caching** for frequently accessed data
3. **Configure MongoDB indexes** properly
4. **Monitor and tune** database queries
5. **Use CDN** for static assets
6. **Enable gzip compression** in Nginx

## Backup and Recovery

### Automated Backups

```bash
# Create backup script
cat > backup.sh << 'EOF'
#!/bin/bash
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/backups/$DATE"
mkdir -p $BACKUP_DIR

# Backup MongoDB
docker exec wedding-mongodb mongodump --uri="mongodb://admin:password123@localhost:27017/wedding_invitations?authSource=admin" -o $BACKUP_DIR/mongodb

# Compress backup
tar -czf "$BACKUP_DIR.tar.gz" -C /backups "$DATE"
rm -rf $BACKUP_DIR

echo "Backup completed: $BACKUP_DIR.tar.gz"
EOF

chmod +x backup.sh
```

### Recovery

```bash
# Restore from backup
tar -xzf backup_20231201_120000.tar.gz
docker exec wedding-mongodb mongorestore --uri="mongodb://admin:password123@localhost:27017/wedding_invitations?authSource=admin" ./20231201_120000/mongodb/wedding_invitations
```

## Support

For issues and questions:
1. Check logs: `docker-compose logs -f`
2. Verify configuration: `docker-compose config`
3. Test connectivity: `docker-compose exec app wget -qO- http://localhost:8080/health`
4. Review documentation and known issues