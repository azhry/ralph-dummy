# Deployment Guide

## Table of Contents

1. [Deployment Overview](#deployment-overview)
2. [Docker Configuration](#docker-configuration)
3. [Environment-Specific Configurations](#environment-specific-configurations)
4. [CI/CD Pipeline](#cicd-pipeline)
5. [Cloud Deployment Options](#cloud-deployment-options)
6. [Database Deployment](#database-deployment)
7. [File Storage Setup](#file-storage-setup)
8. [Load Balancing and Scaling](#load-balancing-and-scaling)
9. [Monitoring and Logging](#monitoring-and-logging)
10. [SSL/TLS and Domain Setup](#ssltls-and-domain-setup)
11. [Database Migrations](#database-migrations)
12. [Rollback Strategies](#rollback-strategies)
13. [Complete Configuration Examples](#complete-configuration-examples)

---

## Deployment Overview

### Deployment Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT LAYER                                    │
│                         (React Web Application)                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      │ HTTPS
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                            CDN / EDGE LAYER                                  │
│                      (Cloudflare / AWS CloudFront)                          │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │  Static Assets (JS, CSS, Images)                                       │ │
│  │  DDoS Protection | SSL Termination | Caching                          │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           LOAD BALANCER LAYER                                │
│                      (AWS ALB / Nginx / Traefik)                            │
│  ┌────────────────────────────────────────────────────────────────────────┐ │
│  │  Health Checks | SSL Passthrough | Sticky Sessions | Rate Limiting    │ │
│  └────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           APPLICATION LAYER                                  │
│                    (Docker Containers / Kubernetes)                         │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐   │
│  │  API Server  │  │  API Server  │  │  API Server  │  │  Worker      │   │
│  │   (Go/Gin)   │  │   (Go/Gin)   │  │   (Go/Gin)   │  │  (Email)     │   │
│  │   :8080      │  │   :8080      │  │   :8080      │  │              │   │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                           DATA LAYER                                         │
│  ┌────────────────────────┐  ┌────────────────────────┐  ┌───────────────┐   │
│  │    MongoDB Cluster     │  │    Redis Cluster       │  │   AWS S3      │   │
│  │  (Primary + 2 Replicas)│  │  (Cache + Sessions)    │  │  (File Store) │   │
│  └────────────────────────┘  └────────────────────────┘  └───────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Deployment Checklist

**Pre-Deployment:**
- [ ] Environment variables configured
- [ ] Database migrations ready
- [ ] SSL certificates obtained
- [ ] Domain DNS configured
- [ ] Secrets encrypted (AWS KMS, HashiCorp Vault)
- [ ] Health check endpoints tested
- [ ] Backup strategy verified
- [ ] Rollback plan documented

**Deployment Day:**
- [ ] Maintenance mode activated
- [ ] Database backup created
- [ ] New version deployed
- [ ] Health checks passing
- [ ] Smoke tests executed
- [ ] Maintenance mode disabled
- [ ] Monitoring dashboards verified

**Post-Deployment:**
- [ ] Error rates monitored
- [ ] Response times checked
- [ ] User feedback collected
- [ ] Rollback window monitored

---

## Docker Configuration

### Development Dockerfile

```dockerfile
# Dockerfile.dev
FROM golang:1.21-alpine AS base

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set timezone
ENV TZ=UTC

# Set working directory
WORKDIR /app

# Install air for live reloading
RUN go install github.com/cosmtrek/air@latest

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Expose port
EXPOSE 8080

# Run with air for hot reload
CMD ["air", "-c", ".air.toml"]
```

### Production Dockerfile (Multi-Stage)

```dockerfile
# Dockerfile.prod
# Stage 1: Build
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=$(git describe --tags --always) \
    -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -o /build/wedding-api \
    ./cmd/api

# Stage 2: Runtime
FROM alpine:3.19 AS runtime

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

# Set timezone
ENV TZ=UTC

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/wedding-api .

# Copy configuration files
COPY --from=builder /build/config ./config

# Copy migration files
COPY --from=builder /build/migrations ./migrations

# Change ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Run the binary
ENTRYPOINT ["./wedding-api"]
```

### .dockerignore

```
# Git
.git
.gitignore
.gitattributes

# CI/CD
.github
.gitlab-ci.yml

# Documentation
*.md
docs/

# Development
.env
.env.local
.env.*.local
*.log
logs/
tmp/
temp/
.air.toml

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# Dependencies (downloaded in container)
vendor/

# Test files
*_test.go
coverage.out
coverage.html

# Build artifacts
dist/
build/
bin/
wedding-api

# Docker
Dockerfile*
docker-compose*.yml
.docker/

# Misc
.DS_Store
Thumbs.db
```

### docker-compose.yml (Local Development)

```yaml
version: '3.8'

services:
  # Application API
  api:
    build:
      context: .
      dockerfile: Dockerfile.dev
    container_name: wedding-api-dev
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=development
      - APP_PORT=8080
      - LOG_LEVEL=debug
      - MONGODB_URI=mongodb://mongo:27017/wedding_dev
      - MONGODB_DATABASE=wedding_dev
      - REDIS_URL=redis://redis:6379/0
      - JWT_SECRET=dev-secret-key-change-in-production
      - JWT_EXPIRATION=24h
      - AWS_REGION=us-east-1
      - AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
      - AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}
      - S3_BUCKET=wedding-dev-bucket
      - S3_ENDPOINT=http://localstack:4566
      - EMAIL_PROVIDER=console
      - ENABLE_SWAGGER=true
    volumes:
      - .:/app
      - go-cache:/go/pkg/mod
    depends_on:
      mongo:
        condition: service_healthy
      redis:
        condition: service_healthy
      localstack:
        condition: service_healthy
    networks:
      - wedding-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  # MongoDB Database
  mongo:
    image: mongo:7.0
    container_name: wedding-mongo-dev
    ports:
      - "27017:27017"
    environment:
      - MONGO_INITDB_ROOT_USERNAME=admin
      - MONGO_INITDB_ROOT_PASSWORD=password
      - MONGO_INITDB_DATABASE=wedding_dev
    volumes:
      - mongo-data:/data/db
      - ./init-mongo.js:/docker-entrypoint-initdb.d/init-mongo.js:ro
    networks:
      - wedding-network
    healthcheck:
      test: ["CMD", "mongosh", "--eval", "db.adminCommand('ping')"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s

  # MongoDB Express (Database UI)
  mongo-express:
    image: mongo-express:1.0.0
    container_name: wedding-mongo-express
    ports:
      - "8081:8081"
    environment:
      - ME_CONFIG_MONGODB_ADMINUSERNAME=admin
      - ME_CONFIG_MONGODB_ADMINPASSWORD=password
      - ME_CONFIG_MONGODB_URL=mongodb://admin:password@mongo:27017/
      - ME_CONFIG_BASICAUTH_USERNAME=admin
      - ME_CONFIG_BASICAUTH_PASSWORD=admin
    depends_on:
      - mongo
    networks:
      - wedding-network

  # Redis Cache
  redis:
    image: redis:7-alpine
    container_name: wedding-redis-dev
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    command: redis-server --appendonly yes
    networks:
      - wedding-network
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Redis Insight (Redis UI)
  redis-insight:
    image: redis/redisinsight:2.40
    container_name: wedding-redis-insight
    ports:
      - "5540:5540"
    volumes:
      - redis-insight-data:/data
    depends_on:
      - redis
    networks:
      - wedding-network

  # LocalStack (AWS Services Emulator)
  localstack:
    image: localstack/localstack:3.0
    container_name: wedding-localstack
    ports:
      - "4566:4566"
      - "4510-4559:4510-4559"
    environment:
      - SERVICES=s3,ses,sns
      - DEFAULT_REGION=us-east-1
      - AWS_DEFAULT_REGION=us-east-1
      - AWS_ACCESS_KEY_ID=test
      - AWS_SECRET_ACCESS_KEY=test
    volumes:
      - localstack-data:/var/lib/localstack
      - ./init-aws.sh:/etc/localstack/init/ready.d/init-aws.sh:ro
    networks:
      - wedding-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:4566/_localstack/health"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s

  # MailHog (Email Testing)
  mailhog:
    image: mailhog/mailhog:v1.0.1
    container_name: wedding-mailhog
    ports:
      - "1025:1025"
      - "8025:8025"
    networks:
      - wedding-network

volumes:
  mongo-data:
  redis-data:
  redis-insight-data:
  localstack-data:
  go-cache:

networks:
  wedding-network:
    driver: bridge
```

### docker-compose.prod.yml (Production)

```yaml
version: '3.8'

services:
  # Application API
  api:
    build:
      context: .
      dockerfile: Dockerfile.prod
    image: wedding-api:${VERSION:-latest}
    container_name: wedding-api-${HOSTNAME}
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - APP_ENV=production
      - APP_PORT=8080
      - LOG_LEVEL=info
      - MONGODB_URI=${MONGODB_URI}
      - MONGODB_DATABASE=${MONGODB_DATABASE}
      - REDIS_URL=${REDIS_URL}
      - JWT_SECRET=${JWT_SECRET}
      - JWT_EXPIRATION=${JWT_EXPIRATION:-24h}
      - AWS_REGION=${AWS_REGION}
      - AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}
      - AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}
      - S3_BUCKET=${S3_BUCKET}
      - S3_CDN_DOMAIN=${S3_CDN_DOMAIN}
      - EMAIL_PROVIDER=${EMAIL_PROVIDER}
      - EMAIL_FROM_ADDRESS=${EMAIL_FROM_ADDRESS}
      - SENDGRID_API_KEY=${SENDGRID_API_KEY}
      - RATE_LIMIT_REQUESTS=${RATE_LIMIT_REQUESTS:-100}
      - RATE_LIMIT_WINDOW=${RATE_LIMIT_WINDOW:-60}
      - ENABLE_SWAGGER=false
      - OTEL_EXPORTER_OTLP_ENDPOINT=${OTEL_EXPORTER_OTLP_ENDPOINT}
      - OTEL_SERVICE_NAME=wedding-api
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
        reservations:
          cpus: '0.25'
          memory: 128M
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
        labels: "service,environment"
        env: "OS_VERSION"
    networks:
      - wedding-network

  # Nginx Reverse Proxy
  nginx:
    image: nginx:1.25-alpine
    container_name: wedding-nginx
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
      - nginx-cache:/var/cache/nginx
    depends_on:
      - api
    networks:
      - wedding-network
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost/health"]
      interval: 30s
      timeout: 10s
      retries: 3

volumes:
  nginx-cache:

networks:
  wedding-network:
    driver: bridge
```

---

## Environment-Specific Configurations

### Development Environment

**Purpose:** Local development with hot reloading and debugging tools

| Service | Endpoint | Purpose |
|---------|----------|---------|
| API | http://localhost:8080 | Main application API |
| MongoDB | localhost:27017 | Primary database |
| Mongo Express | http://localhost:8081 | Database management UI |
| Redis | localhost:6379 | Cache & sessions |
| Redis Insight | http://localhost:5540 | Redis management UI |
| LocalStack | http://localhost:4566 | AWS service emulator |
| MailHog | http://localhost:8025 | Email testing UI |

### Staging Environment

**Purpose:** Pre-production testing with production-like data

**Configuration:**
```yaml
# .env.staging
APP_ENV=staging
APP_PORT=8080
LOG_LEVEL=debug

# MongoDB - Shared cluster with production but isolated database
MONGODB_URI=mongodb+srv://staging-user:xxx@cluster.mongodb.net/wedding_staging
MONGODB_DATABASE=wedding_staging

# Redis - Shared staging instance
REDIS_URL=redis://staging-redis.cache.amazonaws.com:6379/1

# AWS - Separate staging bucket
AWS_REGION=us-east-1
S3_BUCKET=wedding-staging-uploads
S3_CDN_DOMAIN=cdn-staging.wedding.app

# Email - Test provider (Mailtrap)
EMAIL_PROVIDER=mailtrap
EMAIL_FROM_ADDRESS=noreply@staging.wedding.app
MAILTRAP_API_KEY=xxx

# Rate Limiting - Relaxed for testing
RATE_LIMIT_REQUESTS=200
RATE_LIMIT_WINDOW=60

# Feature Flags
ENABLE_SWAGGER=true
ENABLE_DEBUG_ENDPOINTS=true
```

### Production Environment

**Configuration:**
```yaml
# .env.production
APP_ENV=production
APP_PORT=8080
LOG_LEVEL=warn

# MongoDB - Production cluster with replicas
MONGODB_URI=mongodb+srv://prod-user:xxx@cluster.mongodb.net/wedding_prod?retryWrites=true&w=majority
MONGODB_DATABASE=wedding_prod

# Redis - Production cluster
REDIS_URL=redis://prod-redis.cache.amazonaws.com:6379/0

# AWS - Production bucket with CloudFront
AWS_REGION=us-east-1
S3_BUCKET=wedding-production-uploads
S3_CDN_DOMAIN=cdn.wedding.app

# Email - Production provider (SendGrid)
EMAIL_PROVIDER=sendgrid
EMAIL_FROM_ADDRESS=noreply@wedding.app
SENDGRID_API_KEY=xxx

# Rate Limiting - Strict
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60

# Security
JWT_SECRET=${JWT_SECRET_VAULT}
JWT_EXPIRATION=24h

# Monitoring
OTEL_EXPORTER_OTLP_ENDPOINT=https://collector.monitoring.io:4317
OTEL_SERVICE_NAME=wedding-api-prod
SENTRY_DSN=${SENTRY_DSN}

# Feature Flags
ENABLE_SWAGGER=false
ENABLE_DEBUG_ENDPOINTS=false
```

---

## CI/CD Pipeline

### GitHub Actions Workflow - Build

```yaml
# .github/workflows/build.yml
name: Build

on:
  push:
    branches: [main, develop]
    tags: ['v*']
  pull_request:
    branches: [main, develop]

env:
  GO_VERSION: '1.21'
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}/api

jobs:
  lint:
    name: Lint
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
      
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --timeout=5m

  test:
    name: Test
    runs-on: ubuntu-latest
    services:
      mongo:
        image: mongo:7.0
        ports:
          - 27017:27017
      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ env.GO_VERSION }}
      
      - name: Cache Go modules
        uses: actions/cache@v3
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-
      
      - name: Download dependencies
        run: go mod download
      
      - name: Run tests
        env:
          MONGODB_URI: mongodb://localhost:27017/wedding_test
          REDIS_URL: redis://localhost:6379/0
        run: |
          go test -v -race -coverprofile=coverage.out ./...
          go tool cover -html=coverage.out -o coverage.html
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          fail_ci_if_error: true

  security-scan:
    name: Security Scan
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Run Gosec Security Scanner
        uses: securego/gosec@master
        with:
          args: '-fmt sarif -out results.sarif ./...'
      
      - name: Upload SARIF file
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: results.sarif

  build:
    name: Build Docker Image
    runs-on: ubuntu-latest
    needs: [lint, test, security-scan]
    permissions:
      contents: read
      packages: write
    outputs:
      image_tag: ${{ steps.meta.outputs.tags }}
      image_digest: ${{ steps.build.outputs.digest }}
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3
      
      - name: Log in to Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha,prefix=,suffix=,format=short
      
      - name: Build and push
        id: build
        uses: docker/build-push-action@v5
        with:
          context: .
          file: ./Dockerfile.prod
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          platforms: linux/amd64,linux/arm64
      
      - name: Generate SBOM
        uses: anchore/sbom-action@v0
        with:
          image: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}@${{ steps.build.outputs.digest }}
          format: spdx-json
          output-file: sbom.spdx.json
      
      - name: Upload SBOM
        uses: actions/upload-artifact@v3
        with:
          name: sbom
          path: sbom.spdx.json
```

### GitHub Actions Workflow - Deploy to Staging

```yaml
# .github/workflows/deploy-staging.yml
name: Deploy to Staging

on:
  push:
    branches: [develop]
  workflow_dispatch:

env:
  AWS_REGION: us-east-1
  ECR_REPOSITORY: wedding-api
  ECS_CLUSTER: wedding-staging
  ECS_SERVICE: wedding-api-staging
  ECS_TASK_DEFINITION: .aws/task-definition-staging.json

jobs:
  deploy:
    name: Deploy to Staging
    runs-on: ubuntu-latest
    environment: staging
    steps:
      - uses: actions/checkout@v4
      
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: ${{ env.AWS_REGION }}
      
      - name: Login to Amazon ECR
        id: login-ecr
        uses: aws-actions/amazon-ecr-login@v2
      
      - name: Build, tag, and push image
        env:
          ECR_REGISTRY: ${{ steps.login-ecr.outputs.registry }}
          IMAGE_TAG: ${{ github.sha }}
        run: |
          docker build -t $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG -f Dockerfile.prod .
          docker push $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG
          docker tag $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG $ECR_REGISTRY/$ECR_REPOSITORY:staging
          docker push $ECR_REGISTRY/$ECR_REPOSITORY:staging
          echo "image=$ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG" >> $GITHUB_OUTPUT
      
      - name: Fill in image ID in task definition
        id: task-def
        uses: aws-actions/amazon-ecs-render-task-definition@v1
        with:
          task-definition: ${{ env.ECS_TASK_DEFINITION }}
          container-name: wedding-api
          image: ${{ steps.build-image.outputs.image }}
      
      - name: Deploy to ECS
        uses: aws-actions/amazon-ecs-deploy-task-definition@v1
        with:
          task-definition: ${{ steps.task-def.outputs.task-definition }}
          service: ${{ env.ECS_SERVICE }}
          cluster: ${{ env.ECS_CLUSTER }}
          wait-for-service-stability: true
      
      - name: Run smoke tests
        run: |
          sleep 30
          curl -sf https://api-staging.wedding.app/health || exit 1
          echo "Smoke tests passed!"
```

### GitHub Actions Workflow - Deploy to Production

```yaml
# .github/workflows/deploy-production.yml
name: Deploy to Production

on:
  release:
    types: [published]
  workflow_dispatch:
    inputs:
      version:
        description: 'Version to deploy (tag)'
        required: true
        type: string

env:
  AWS_REGION: us-east-1
  ECR_REPOSITORY: wedding-api
  ECS_CLUSTER: wedding-production
  ECS_SERVICE: wedding-api-production
  ECS_TASK_DEFINITION: .aws/task-definition-production.json

jobs:
  deploy:
    name: Deploy to Production
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.inputs.version || github.event.release.tag_name }}
      
      - name: Verify release exists
        run: |
          git fetch --tags
          git checkout ${{ github.event.inputs.version || github.event.release.tag_name }}
      
      - name: Run integration tests
        run: |
          echo "Running integration tests against staging..."
          go test -v -tags=integration ./tests/integration/...
      
      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: ${{ env.AWS_REGION }}
      
      - name: Login to Amazon ECR
        id: login-ecr
        uses: aws-actions/amazon-ecr-login@v2
      
      - name: Tag and push production image
        env:
          ECR_REGISTRY: ${{ steps.login-ecr.outputs.registry }}
          VERSION: ${{ github.event.inputs.version || github.event.release.tag_name }}
        run: |
          # Pull staging image, retag for production
          docker pull $ECR_REGISTRY/$ECR_REPOSITORY:staging
          docker tag $ECR_REGISTRY/$ECR_REPOSITORY:staging $ECR_REGISTRY/$ECR_REPOSITORY:$VERSION
          docker tag $ECR_REGISTRY/$ECR_REPOSITORY:staging $ECR_REGISTRY/$ECR_REPOSITORY:latest
          docker push $ECR_REGISTRY/$ECR_REPOSITORY:$VERSION
          docker push $ECR_REGISTRY/$ECR_REPOSITORY:latest
          echo "image=$ECR_REGISTRY/$ECR_REPOSITORY:$VERSION" >> $GITHUB_OUTPUT
      
      - name: Create backup
        run: |
          echo "Creating database backup..."
          aws mongodb create-backup \
            --cluster-id wedding-prod-cluster \
            --backup-name pre-deploy-$(date +%Y%m%d-%H%M%S)
      
      - name: Deploy blue-green
        run: |
          # Deploy to blue environment first
          echo "Deploying to blue environment..."
          aws ecs update-service \
            --cluster $ECS_CLUSTER \
            --service wedding-api-blue \
            --task-definition wedding-api-blue:$VERSION \
            --force-new-deployment
          
          # Wait for blue to be healthy
          echo "Waiting for blue deployment to stabilize..."
          aws ecs wait services-stable \
            --cluster $ECS_CLUSTER \
            --services wedding-api-blue
          
          # Run smoke tests on blue
          echo "Running smoke tests on blue..."
          curl -sf https://api-blue.wedding.app/health || exit 1
          
          # Switch traffic to blue
          echo "Switching traffic to blue..."
          aws route53 change-resource-record-sets \
            --hosted-zone-id ${{ secrets.HOSTED_ZONE_ID }} \
            --change-batch file://switch-to-blue.json
          
          # Wait and verify
          sleep 60
          curl -sf https://api.wedding.app/health || exit 1
      
      - name: Notify on success
        if: success()
        uses: slackapi/slack-github-action@v1
        with:
          payload: |
            {
              "text": "Production deployment successful: ${{ github.event.inputs.version || github.event.release.tag_name }}"
            }
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
      
      - name: Rollback on failure
        if: failure()
        run: |
          echo "Deployment failed! Rolling back..."
          aws route53 change-resource-record-sets \
            --hosted-zone-id ${{ secrets.HOSTED_ZONE_ID }} \
            --change-batch file://rollback-to-green.json
          
          curl -X POST ${{ secrets.SLACK_WEBHOOK_URL }} \
            -H 'Content-type: application/json' \
            --data '{"text":"Production deployment FAILED and was rolled back"}'
```

---

## Cloud Deployment Options

### AWS ECS with Fargate

**Architecture:**
```
┌─────────────────────────────────────────────────────────┐
│                    Route 53 (DNS)                      │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│              Application Load Balancer                  │
│        (SSL Termination | Health Checks | WAF)         │
└─────────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────┐
│                  ECS Cluster (Fargate)                  │
│  ┌─────────────────────────────────────────────────┐  │
│  │            ECS Service: wedding-api             │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐      │  │
│  │  │ Task 1   │  │ Task 2   │  │ Task 3   │      │  │
│  │  │ (v1.2.0) │  │ (v1.2.0) │  │ (v1.2.0) │      │  │
│  │  │ 256MB    │  │ 256MB    │  │ 256MB    │      │  │
│  │  │ 0.25 vCPU│  │ 0.25 vCPU│  │ 0.25 vCPU│      │  │
│  │  └──────────┘  └──────────┘  └──────────┘      │  │
│  │                                                 │  │
│  │  Min: 2 tasks | Max: 10 tasks | Desired: 3    │  │
│  │  Auto-scaling: CPU > 70% for 2 minutes          │  │
│  └─────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────┘
```

**ECS Task Definition:**
```json
{
  "family": "wedding-api",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "executionRoleArn": "arn:aws:iam::123456789012:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::123456789012:role/ecsTaskRole",
  "containerDefinitions": [
    {
      "name": "wedding-api",
      "image": "123456789012.dkr.ecr.us-east-1.amazonaws.com/wedding-api:latest",
      "essential": true,
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "APP_ENV",
          "value": "production"
        },
        {
          "name": "LOG_LEVEL",
          "value": "info"
        }
      ],
      "secrets": [
        {
          "name": "MONGODB_URI",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:123456789012:secret:wedding/mongodb-uri:AWSCURRENT"
        },
        {
          "name": "JWT_SECRET",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:123456789012:secret:wedding/jwt-secret:AWSCURRENT"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/wedding-api",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "healthCheck": {
        "command": ["CMD-SHELL", "curl -f http://localhost:8080/health || exit 1"],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      },
      "ulimits": [
        {
          "name": "nofile",
          "softLimit": 65536,
          "hardLimit": 65536
        }
      ]
    }
  ]
}
```

**Terraform Configuration:**
```hcl
# terraform/ecs.tf
provider "aws" {
  region = var.aws_region
}

# VPC and Networking
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "5.0.0"

  name = "wedding-vpc"
  cidr = "10.0.0.0/16"

  azs             = ["us-east-1a", "us-east-1b", "us-east-1c"]
  private_subnets = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  public_subnets  = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]

  enable_nat_gateway = true
  enable_vpn_gateway = false
  enable_dns_hostnames = true
  enable_dns_support = true

  tags = {
    Environment = var.environment
    Project     = "wedding-invitation"
  }
}

# ECS Cluster
resource "aws_ecs_cluster" "wedding" {
  name = "wedding-${var.environment}"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  configuration {
    execute_command_configuration {
      logging = "OVERRIDE"
      log_configuration {
        cloud_watch_encryption_enabled = true
        cloud_watch_log_group_name     = aws_cloudwatch_log_group.ecs_exec.name
      }
    }
  }
}

# ECS Service
resource "aws_ecs_service" "api" {
  name            = "wedding-api-${var.environment}"
  cluster         = aws_ecs_cluster.wedding.id
  task_definition = aws_ecs_task_definition.api.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = module.vpc.private_subnets
    security_groups  = [aws_security_group.ecs_tasks.id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.api.arn
    container_name   = "wedding-api"
    container_port   = 8080
  }

  deployment_configuration {
    maximum_percent         = 200
    minimum_healthy_percent = 100
    deployment_circuit_breaker {
      enable   = true
      rollback = true
    }
  }

  capacity_provider_strategy {
    base              = 1
    weight            = 100
    capacity_provider = "FARGATE"
  }

  depends_on = [aws_lb_listener.https]

  tags = {
    Environment = var.environment
  }
}

# Auto Scaling
resource "aws_appautoscaling_target" "api" {
  max_capacity       = var.max_count
  min_capacity       = var.min_count
  resource_id        = "service/${aws_ecs_cluster.wedding.name}/${aws_ecs_service.api.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "api_cpu" {
  name               = "wedding-api-cpu-autoscaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.api.resource_id
  scalable_dimension = aws_appautoscaling_target.api.scalable_dimension
  service_namespace  = aws_appautoscaling_target.api.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value       = 70.0
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}
```

### Google Cloud Platform - Cloud Run

**Deployment Configuration:**
```yaml
# cloudbuild.yaml
steps:
  # Build container image
  - name: 'gcr.io/cloud-builders/docker'
    args:
      - 'build'
      - '-t'
      - 'gcr.io/$PROJECT_ID/wedding-api:$COMMIT_SHA'
      - '-f'
      - 'Dockerfile.prod'
      - '.'
  
  # Push container image
  - name: 'gcr.io/cloud-builders/docker'
    args:
      - 'push'
      - 'gcr.io/$PROJECT_ID/wedding-api:$COMMIT_SHA'
  
  # Deploy to Cloud Run
  - name: 'gcr.io/google.com/cloudsdktool/cloud-sdk'
    entrypoint: gcloud
    args:
      - 'run'
      - 'deploy'
      - 'wedding-api'
      - '--image'
      - 'gcr.io/$PROJECT_ID/wedding-api:$COMMIT_SHA'
      - '--region'
      - 'us-central1'
      - '--platform'
      - 'managed'
      - '--allow-unauthenticated'
      - '--memory'
      - '512Mi'
      - '--cpu'
      - '1'
      - '--concurrency'
      - '100'
      - '--max-instances'
      - '10'
      - '--min-instances'
      - '1'
      - '--timeout'
      - '30s'
      - '--service-account'
      - 'wedding-api@$PROJECT_ID.iam.gserviceaccount.com'
      - '--set-secrets'
      - 'MONGODB_URI=mongodb-uri:latest,JWT_SECRET=jwt-secret:latest'
      - '--set-env-vars'
      - 'APP_ENV=production,LOG_LEVEL=info'
      - '--port'
      - '8080'

images:
  - 'gcr.io/$PROJECT_ID/wedding-api:$COMMIT_SHA'

options:
  logging: CLOUD_LOGGING_ONLY
```

### Railway (Platform as a Service)

**railway.yml:**
```yaml
build:
  builder: DOCKERFILE
  dockerfilePath: Dockerfile.prod

deploy:
  startCommand: ./wedding-api
  healthcheckPath: /health
  healthcheckTimeout: 100
  restartPolicyType: ON_FAILURE
  restartPolicyMaxRetries: 10
  
  resources:
    cpu: 1
    memory: 512
    
  numReplicas: 2
  
  multiRegionConfig:
    regions:
      us-west1:
        numReplicas: 2
      eu-west1:
        numReplicas: 1
```

### Render

**render.yaml:**
```yaml
services:
  - type: web
    name: wedding-api
    runtime: docker
    dockerfilePath: ./Dockerfile.prod
    plan: standard
    envVars:
      - key: APP_ENV
        value: production
      - key: PORT
        value: 8080
      - key: MONGODB_URI
        fromDatabase:
          name: wedding-mongodb
          property: connectionString
      - key: JWT_SECRET
        generateValue: true
      - key: AWS_ACCESS_KEY_ID
        sync: false
      - key: AWS_SECRET_ACCESS_KEY
        sync: false
    healthCheckPath: /health
    autoDeploy: true
    scaling:
      minInstances: 2
      maxInstances: 10
      targetCPUPercent: 70
      targetMemoryPercent: 80

databases:
  - name: wedding-mongodb
    databaseName: wedding_prod
    user: wedding_user
    plan: standard
```

---

## Database Deployment

### MongoDB Atlas Setup

**Cluster Configuration:**
```javascript
// Cluster Tier: M10 (Production)
// - 2GB RAM
// - 10GB Storage
// - 3-node replica set
// - Auto-scaling enabled

// Connection String
mongodb+srv://<username>:<password>@wedding-cluster.mongodb.net/wedding_prod?retryWrites=true&w=majority&maxPoolSize=100
```

**Atlas CLI Deployment:**
```bash
#!/bin/bash
# scripts/setup-atlas.sh

# Login to Atlas
atlas auth login

# Create cluster
atlas clusters create wedding-cluster \
  --provider AWS \
  --region US_EAST_1 \
  --tier M10 \
  --type REPLICASET \
  --members 3

# Configure maintenance window
atlas maintenanceWindows update \
  --dayOfWeek 6 \
  --hourOfDay 2 \
  --clusterName wedding-cluster

# Enable backups
atlas backups enable \
  --clusterName wedding-cluster

# Create database user
atlas dbusers create \
  --username wedding_app \
  --password $DB_PASSWORD \
  --role readWrite@wedding_prod \
  --role readWrite@wedding_staging

# Configure IP access list
atlas accessLists create \
  --type ipAddress \
  --ipAddress 0.0.0.0/0 \
  --comment "Production VPC"

# Wait for cluster
atlas clusters watch wedding-cluster

echo "MongoDB Atlas cluster ready!"
```

**Backup Configuration:**
```bash
#!/bin/bash
# scripts/backup-mongodb.sh

BACKUP_NAME="wedding-$(date +%Y%m%d-%H%M%S)"
S3_BUCKET="wedding-backups"

# Create backup using mongodump
docker run --rm \
  -v $(pwd)/backups:/backup \
  mongo:7.0 \
  mongodump \
  --uri="$MONGODB_URI" \
  --out=/backup/$BACKUP_NAME \
  --gzip

# Upload to S3
aws s3 sync \
  ./backups/$BACKUP_NAME \
  s3://$S3_BUCKET/mongodb/$BACKUP_NAME/ \
  --storage-class STANDARD_IA

# Clean up local backup
rm -rf ./backups/$BACKUP_NAME

# Keep only last 30 days of backups in S3
aws s3api list-objects-v2 \
  --bucket $S3_BUCKET \
  --prefix mongodb/ \
  --query "Contents[?LastModified<='$(date -d '30 days ago' +%Y-%m-%d)'].Key" \
  --output text | xargs -I {} aws s3 rm s3://$S3_BUCKET/{}

echo "Backup completed: $BACKUP_NAME"
```

### Self-Hosted MongoDB

**Docker Compose for Production:**
```yaml
version: '3.8'

services:
  mongo-primary:
    image: mongo:7.0
    container_name: mongo-primary
    command: mongod --replSet rs0 --bind_ip_all
    ports:
      - "27017:27017"
    volumes:
      - mongo-primary-data:/data/db
      - ./mongo-keyfile:/data/keyfile:ro
    environment:
      - MONGO_INITDB_ROOT_USERNAME=admin
      - MONGO_INITDB_ROOT_PASSWORD=${MONGO_ROOT_PASSWORD}
    networks:
      - mongo-network

  mongo-secondary-1:
    image: mongo:7.0
    container_name: mongo-secondary-1
    command: mongod --replSet rs0 --bind_ip_all
    volumes:
      - mongo-secondary-1-data:/data/db
      - ./mongo-keyfile:/data/keyfile:ro
    networks:
      - mongo-network

  mongo-secondary-2:
    image: mongo:7.0
    container_name: mongo-secondary-2
    command: mongod --replSet rs0 --bind_ip_all
    volumes:
      - mongo-secondary-2-data:/data/db
      - ./mongo-keyfile:/data/keyfile:ro
    networks:
      - mongo-network

  mongo-init:
    image: mongo:7.0
    depends_on:
      - mongo-primary
      - mongo-secondary-1
      - mongo-secondary-2
    entrypoint: |
      bash -c '
        sleep 10
        mongosh --host mongo-primary:27017 -u admin -p $MONGO_ROOT_PASSWORD --authenticationDatabase admin --eval "
          rs.initiate({
            _id: \"rs0\",
            members: [
              { _id: 0, host: \"mongo-primary:27017\", priority: 2 },
              { _id: 1, host: \"mongo-secondary-1:27017\", priority: 1 },
              { _id: 2, host: \"mongo-secondary-2:27017\", priority: 1 }
            ]
          })
        "
      '
    networks:
      - mongo-network

volumes:
  mongo-primary-data:
  mongo-secondary-1-data:
  mongo-secondary-2-data:

networks:
  mongo-network:
    driver: bridge
```

---

## File Storage Setup

### AWS S3 Configuration

**Bucket Setup:**
```bash
#!/bin/bash
# scripts/setup-s3.sh

BUCKET_NAME="wedding-production-uploads"
REGION="us-east-1"
CDN_DOMAIN="cdn.wedding.app"

# Create bucket
aws s3api create-bucket \
  --bucket $BUCKET_NAME \
  --region $REGION

# Enable versioning
aws s3api put-bucket-versioning \
  --bucket $BUCKET_NAME \
  --versioning-configuration Status=Enabled

# Block public access
aws s3api put-public-access-block \
  --bucket $BUCKET_NAME \
  --public-access-block-configuration \
  "BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true"

# Enable server-side encryption
aws s3api put-bucket-encryption \
  --bucket $BUCKET_NAME \
  --server-side-encryption-configuration '{
    "Rules": [
      {
        "ApplyServerSideEncryptionByDefault": {
          "SSEAlgorithm": "AES256"
        },
        "BucketKeyEnabled": true
      }
    ]
  }'

# Set lifecycle policy for old versions
aws s3api put-bucket-lifecycle-configuration \
  --bucket $BUCKET_NAME \
  --lifecycle-configuration file://lifecycle-policy.json

# Set CORS configuration
aws s3api put-bucket-cors \
  --bucket $BUCKET_NAME \
  --cors-configuration file://cors-config.json

# Create CloudFront distribution
DISTRIBUTION_ID=$(aws cloudfront create-distribution \
  --origin-domain-name ${BUCKET_NAME}.s3.amazonaws.com \
  --default-root-object index.html \
  --query 'Distribution.Id' \
  --output text)

echo "S3 bucket and CloudFront distribution created!"
echo "CDN Domain: $CDN_DOMAIN"
```

**IAM Policy for S3 Access:**
```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "WeddingS3Access",
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:PutObjectAcl",
        "s3:GetObject",
        "s3:DeleteObject",
        "s3:ListBucket",
        "s3:GetObjectVersion"
      ],
      "Resource": [
        "arn:aws:s3:::wedding-production-uploads",
        "arn:aws:s3:::wedding-production-uploads/*",
        "arn:aws:s3:::wedding-staging-uploads",
        "arn:aws:s3:::wedding-staging-uploads/*"
      ],
      "Condition": {
        "StringEquals": {
          "s3:x-amz-server-side-encryption": "AES256"
        }
      }
    },
    {
      "Sid": "WeddingS3Presigned",
      "Effect": "Allow",
      "Action": [
        "s3:GetObject",
        "s3:PutObject"
      ],
      "Resource": "arn:aws:s3:::wedding-production-uploads/weddings/*",
      "Condition": {
        "StringEquals": {
          "s3:signatureVersion": "AWS4-HMAC-SHA256"
        },
        "NumericLessThanEquals": {
          "s3:signatureAge": 900
        }
      }
    }
  ]
}
```

**CORS Configuration:**
```json
{
  "CORSRules": [
    {
      "AllowedOrigins": [
        "https://wedding.app",
        "https://*.wedding.app",
        "https://localhost:3000"
      ],
      "AllowedMethods": ["GET", "PUT", "POST", "DELETE"],
      "AllowedHeaders": ["*"],
      "MaxAgeSeconds": 3000,
      "ExposeHeaders": ["ETag", "x-amz-server-side-encryption"]
    }
  ]
}
```

### Cloudflare R2 Setup

```bash
#!/bin/bash
# scripts/setup-r2.sh

# Create R2 bucket using Wrangler
wrangler r2 bucket create wedding-uploads

# Configure bucket bindings in wrangler.toml
cat >> wrangler.toml << EOF
[[r2_buckets]]
binding = "WEDDING_UPLOADS"
bucket_name = "wedding-uploads"
EOF

# Set up custom domain
wrangler r2 bucket domain add wedding-uploads --domain cdn.wedding.app

echo "R2 bucket configured!"
```

---

## Load Balancing and Scaling

### Nginx Load Balancer Configuration

```nginx
# nginx.conf
user nginx;
worker_processes auto;
error_log /var/log/nginx/error.log warn;
pid /var/run/nginx.pid;

events {
    worker_connections 4096;
    use epoll;
    multi_accept on;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    # Logging
    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                    '$status $body_bytes_sent "$http_referer" '
                    '"$http_user_agent" "$http_x_forwarded_for" '
                    'rt=$request_time uct="$upstream_connect_time" '
                    'uht="$upstream_header_time" urt="$upstream_response_time"';

    access_log /var/log/nginx/access.log main;

    # Performance
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;
    client_max_body_size 50M;

    # Gzip
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types text/plain text/css text/xml application/json application/javascript application/rss+xml application/atom+xml image/svg+xml;

    # Rate limiting zones
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=auth:10m rate=5r/m;
    limit_req_zone $binary_remote_addr zone=upload:10m rate=2r/m;

    # Connection limiting
    limit_conn_zone $binary_remote_addr zone=addr:10m;

    # Upstream configuration
    upstream api_servers {
        least_conn;
        
        server api-1:8080 max_fails=3 fail_timeout=30s;
        server api-2:8080 max_fails=3 fail_timeout=30s;
        server api-3:8080 max_fails=3 fail_timeout=30s;
        
        keepalive 32;
    }

    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:50m;
    ssl_session_timeout 1d;
    ssl_session_tickets off;

    # Main server block - HTTP redirect
    server {
        listen 80;
        server_name api.wedding.app;
        
        location /.well-known/acme-challenge/ {
            root /var/www/certbot;
        }
        
        location / {
            return 301 https://$server_name$request_uri;
        }
    }

    # HTTPS server
    server {
        listen 443 ssl http2;
        server_name api.wedding.app;

        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;

        # Security headers
        add_header X-Frame-Options "SAMEORIGIN" always;
        add_header X-Content-Type-Options "nosniff" always;
        add_header X-XSS-Protection "1; mode=block" always;
        add_header Referrer-Policy "strict-origin-when-cross-origin" always;
        add_header Permissions-Policy "geolocation=(), microphone=(), camera=()" always;
        add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;

        # Health check endpoint (no rate limiting)
        location /health {
            proxy_pass http://api_servers;
            proxy_http_version 1.1;
            proxy_set_header Connection "";
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            
            proxy_connect_timeout 5s;
            proxy_send_timeout 60s;
            proxy_read_timeout 60s;
        }

        # API endpoints with rate limiting
        location /api/ {
            limit_req zone=api burst=20 nodelay;
            limit_conn addr 10;
            
            proxy_pass http://api_servers;
            proxy_http_version 1.1;
            proxy_set_header Connection "";
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_set_header X-Request-ID $request_id;
            
            proxy_connect_timeout 5s;
            proxy_send_timeout 60s;
            proxy_read_timeout 60s;
            
            # Buffer settings
            proxy_buffering on;
            proxy_buffer_size 4k;
            proxy_buffers 8 4k;
            proxy_busy_buffers_size 8k;
        }

        # Authentication endpoints (stricter rate limiting)
        location /api/auth/ {
            limit_req zone=auth burst=10 nodelay;
            limit_conn addr 5;
            
            proxy_pass http://api_servers;
            proxy_http_version 1.1;
            proxy_set_header Connection "";
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            
            proxy_connect_timeout 5s;
            proxy_send_timeout 30s;
            proxy_read_timeout 30s;
        }

        # File upload endpoints (strictest rate limiting)
        location /api/upload/ {
            limit_req zone=upload burst=5 nodelay;
            limit_conn addr 3;
            client_max_body_size 50M;
            
            proxy_pass http://api_servers;
            proxy_http_version 1.1;
            proxy_set_header Connection "";
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            
            proxy_connect_timeout 30s;
            proxy_send_timeout 300s;
            proxy_read_timeout 300s;
            proxy_request_buffering off;
        }

        # Static assets (if serving from nginx)
        location /static/ {
            alias /var/www/static/;
            expires 1y;
            add_header Cache-Control "public, immutable";
            access_log off;
        }
    }
}
```

### Horizontal Pod Autoscaling (Kubernetes)

```yaml
# k8s/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: wedding-api-hpa
  namespace: wedding
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: wedding-api
  minReplicas: 3
  maxReplicas: 20
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
    - type: Pods
      pods:
        metric:
          name: http_requests_per_second
        target:
          type: AverageValue
          averageValue: "100"
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
        - type: Percent
          value: 100
          periodSeconds: 60
        - type: Pods
          value: 4
          periodSeconds: 60
      selectPolicy: Max
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Percent
          value: 10
          periodSeconds: 60
        - type: Pods
          value: 2
          periodSeconds: 60
      selectPolicy: Min
```

---

## Monitoring and Logging

### Prometheus Metrics

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  external_labels:
    cluster: wedding-production
    replica: '{{.ExternalURL}}'

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093

rule_files:
  - /etc/prometheus/rules/*.yml

scrape_configs:
  # Prometheus self-monitoring
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  # API servers
  - job_name: 'wedding-api'
    static_configs:
      - targets:
        - api-1:8080
        - api-2:8080
        - api-3:8080
    metrics_path: /metrics
    scrape_interval: 15s
    scrape_timeout: 10s

  # MongoDB
  - job_name: 'mongodb'
    static_configs:
      - targets: ['mongodb-exporter:9216']

  # Redis
  - job_name: 'redis'
    static_configs:
      - targets: ['redis-exporter:9121']

  # Node exporter
  - job_name: 'node'
    static_configs:
      - targets: ['node-exporter:9100']

  # Docker containers
  - job_name: 'docker'
    static_configs:
      - targets: ['cadvisor:8080']
```

**Prometheus Rules:**
```yaml
# rules/api-alerts.yml
groups:
  - name: wedding_api
    rules:
      - alert: HighErrorRate
        expr: |
          (
            sum(rate(http_requests_total{job="wedding-api",status=~"5.."}[5m]))
            /
            sum(rate(http_requests_total{job="wedding-api"}[5m]))
          ) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate detected"
          description: "Error rate is {{ $value | humanizePercentage }} for the last 5 minutes"

      - alert: HighLatency
        expr: |
          histogram_quantile(0.95, 
            sum(rate(http_request_duration_seconds_bucket{job="wedding-api"}[5m])) by (le)
          ) > 0.5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High latency detected"
          description: "95th percentile latency is {{ $value }}s"

      - alert: LowAvailability
        expr: |
          (
            sum(up{job="wedding-api"}) 
            / 
            count(up{job="wedding-api"})
          ) < 0.99
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Service availability low"
          description: "Only {{ $value | humanizePercentage }} of instances are up"

      - alert: MongoDBHighConnections
        expr: mongodb_connections{state="current"} > 80
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "MongoDB connection count high"
          description: "MongoDB has {{ $value }} connections (threshold: 80)"
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Wedding API - Production",
    "tags": ["wedding", "api", "production"],
    "timezone": "UTC",
    "refresh": "30s",
    "panels": [
      {
        "id": 1,
        "title": "Request Rate",
        "type": "stat",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{job=\"wedding-api\"}[5m]))",
            "legendFormat": "req/s"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "reqps",
            "thresholds": {
              "steps": [
                {"color": "green", "value": null},
                {"color": "yellow", "value": 1000},
                {"color": "red", "value": 5000}
              ]
            }
          }
        },
        "gridPos": {"h": 4, "w": 6, "x": 0, "y": 0}
      },
      {
        "id": 2,
        "title": "Error Rate",
        "type": "stat",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{job=\"wedding-api\",status=~\"5..\"}[5m])) / sum(rate(http_requests_total{job=\"wedding-api\"}[5m]))",
            "legendFormat": "error rate"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "percentunit",
            "thresholds": {
              "steps": [
                {"color": "green", "value": null},
                {"color": "yellow", "value": 0.01},
                {"color": "red", "value": 0.05}
              ]
            }
          }
        },
        "gridPos": {"h": 4, "w": 6, "x": 6, "y": 0}
      },
      {
        "id": 3,
        "title": "P95 Latency",
        "type": "stat",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket{job=\"wedding-api\"}[5m])) by (le))",
            "legendFormat": "p95"
          }
        ],
        "fieldConfig": {
          "defaults": {
            "unit": "s",
            "thresholds": {
              "steps": [
                {"color": "green", "value": null},
                {"color": "yellow", "value": 0.5},
                {"color": "red", "value": 1.0}
              ]
            }
          }
        },
        "gridPos": {"h": 4, "w": 6, "x": 12, "y": 0}
      },
      {
        "id": 4,
        "title": "Active Instances",
        "type": "stat",
        "targets": [
          {
            "expr": "count(up{job=\"wedding-api\"} == 1)",
            "legendFormat": "instances"
          }
        ],
        "gridPos": {"h": 4, "w": 6, "x": 18, "y": 0}
      },
      {
        "id": 5,
        "title": "Request Latency Distribution",
        "type": "heatmap",
        "targets": [
          {
            "expr": "sum(rate(http_request_duration_seconds_bucket{job=\"wedding-api\"}[5m])) by (le)",
            "format": "heatmap",
            "legendFormat": "{{le}}"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 0, "y": 4}
      },
      {
        "id": 6,
        "title": "Error Rate by Endpoint",
        "type": "timeseries",
        "targets": [
          {
            "expr": "sum(rate(http_requests_total{job=\"wedding-api\",status=~\"5..\"}[5m])) by (handler)",
            "legendFormat": "{{handler}}"
          }
        ],
        "gridPos": {"h": 8, "w": 12, "x": 12, "y": 4}
      }
    ]
  }
}
```

### Health Check Endpoint

```go
// internal/handler/health.go
package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"github.com/go-redis/redis/v8"
)

type HealthHandler struct {
	mongoClient *mongo.Client
	redisClient *redis.Client
	version     string
	startTime   time.Time
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Uptime    string            `json:"uptime"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]Check  `json:"checks"`
}

type Check struct {
	Status  string `json:"status"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

func NewHealthHandler(mongoClient *mongo.Client, redisClient *redis.Client, version string) *HealthHandler {
	return &HealthHandler{
		mongoClient: mongoClient,
		redisClient: redisClient,
		version:     version,
		startTime:   time.Now(),
	}
}

func (h *HealthHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	checks := make(map[string]Check)
	overallStatus := "healthy"

	// Check MongoDB
	mongoStart := time.Now()
	if err := h.mongoClient.Ping(ctx, nil); err != nil {
		checks["mongodb"] = Check{
			Status: "unhealthy",
			Error:  err.Error(),
		}
		overallStatus = "unhealthy"
	} else {
		checks["mongodb"] = Check{
			Status:  "healthy",
			Latency: time.Since(mongoStart).String(),
		}
	}

	// Check Redis
	redisStart := time.Now()
	if err := h.redisClient.Ping(ctx).Err(); err != nil {
		checks["redis"] = Check{
			Status: "unhealthy",
			Error:  err.Error(),
		}
		overallStatus = "unhealthy"
	} else {
		checks["redis"] = Check{
			Status:  "healthy",
			Latency: time.Since(redisStart).String(),
		}
	}

	// Check disk space (if applicable)
	checks["disk"] = Check{Status: "healthy"}

	response := HealthResponse{
		Status:    overallStatus,
		Version:   h.version,
		Uptime:    time.Since(h.startTime).String(),
		Timestamp: time.Now(),
		Checks:    checks,
	}

	statusCode := http.StatusOK
	if overallStatus != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// Readiness probe - can accept traffic
func (h *HealthHandler) Ready(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	// Check if MongoDB is ready
	if err := h.mongoClient.Ping(ctx, nil); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}

// Liveness probe - is the process running
func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "alive"})
}
```

---

## SSL/TLS and Domain Setup

### Let's Encrypt Certificate Generation

```bash
#!/bin/bash
# scripts/setup-ssl.sh

DOMAIN="api.wedding.app"
EMAIL="admin@wedding.app"

# Install certbot if not present
if ! command -v certbot &> /dev/null; then
    apt-get update
    apt-get install -y certbot
fi

# Generate certificate
certbot certonly \
    --standalone \
    --preferred-challenges http \
    --agree-tos \
    --non-interactive \
    --email $EMAIL \
    -d $DOMAIN \
    -d www.$DOMAIN

# Copy certificates to nginx directory
mkdir -p /etc/nginx/ssl
cp /etc/letsencrypt/live/$DOMAIN/fullchain.pem /etc/nginx/ssl/cert.pem
cp /etc/letsencrypt/live/$DOMAIN/privkey.pem /etc/nginx/ssl/key.pem

# Set up auto-renewal
echo "0 0,12 * * * root certbot renew --quiet --deploy-hook 'systemctl reload nginx'" | tee -a /etc/crontab

# Reload nginx
systemctl reload nginx

echo "SSL certificates installed for $DOMAIN"
```

### DNS Configuration

```yaml
# DNS Records (Route 53 / Cloudflare)

# A Record - Point domain to load balancer
api.wedding.app:
  type: A
  value: 1.2.3.4  # Load balancer IP
  ttl: 300

# CNAME - WWW redirect
www.api.wedding.app:
  type: CNAME
  value: api.wedding.app
  ttl: 300

# MX Records (if using domain email)
api.wedding.app:
  - type: MX
    priority: 10
    value: aspmx.l.google.com
  - type: MX
    priority: 20
    value: alt1.aspmx.l.google.com

# TXT Records for SPF/DKIM
api.wedding.app:
  - type: TXT
    value: "v=spf1 include:_spf.google.com ~all"
```

---

## Database Migrations

### MongoDB Migration Tool

```go
// cmd/migrate/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Migration struct {
	Version int
	Name    string
	Up      func(*mongo.Database) error
	Down    func(*mongo.Database) error
}

var migrations = []Migration{
	{
		Version: 1,
		Name:    "create_weddings_collection",
		Up: func(db *mongo.Database) error {
			opts := options.CreateCollection()
			return db.CreateCollection(context.Background(), "weddings", opts)
		},
		Down: func(db *mongo.Database) error {
			return db.Collection("weddings").Drop(context.Background())
		},
	},
	{
		Version: 2,
		Name:    "add_slug_index_to_weddings",
		Up: func(db *mongo.Database) error {
			_, err := db.Collection("weddings").Indexes().CreateOne(
				context.Background(),
				mongo.IndexModel{
					Keys:    bson.D{{Key: "slug", Value: 1}},
					Options: options.Index().SetUnique(true),
				},
			)
			return err
		},
		Down: func(db *mongo.Database) error {
			_, err := db.Collection("weddings").Indexes().DropOne(context.Background(), "slug_1")
			return err
		},
	},
	{
		Version: 3,
		Name:    "create_guests_collection",
		Up: func(db *mongo.Database) error {
			opts := options.CreateCollection()
			return db.CreateCollection(context.Background(), "guests", opts)
		},
		Down: func(db *mongo.Database) error {
			return db.Collection("guests").Drop(context.Background())
		},
	},
}

func main() {
	ctx := context.Background()

	// Connect to MongoDB
	uri := os.Getenv("MONGODB_URI")
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database(os.Getenv("MONGODB_DATABASE"))

	// Get current version
	var versionDoc struct {
		Version int `bson:"version"`
	}
	err = db.Collection("migrations").FindOne(ctx, bson.M{}).Decode(&versionDoc)
	currentVersion := 0
	if err == nil {
		currentVersion = versionDoc.Version
	}

	fmt.Printf("Current database version: %d\n", currentVersion)

	// Parse command
	if len(os.Args) < 2 {
		fmt.Println("Usage: migrate [up|down|version|status]")
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "up":
		targetVersion := len(migrations)
		if len(os.Args) > 2 {
			targetVersion = atoi(os.Args[2])
		}

		for _, migration := range migrations {
			if migration.Version > currentVersion && migration.Version <= targetVersion {
				fmt.Printf("Running migration %d: %s\n", migration.Version, migration.Name)
				if err := migration.Up(db); err != nil {
					log.Fatal(err)
				}
				currentVersion = migration.Version
				_, _ = db.Collection("migrations").UpdateOne(
					ctx,
					bson.M{},
					bson.M{"$set": bson.M{"version": currentVersion}},
					options.Update().SetUpsert(true),
				)
			}
		}
		fmt.Println("Migrations completed successfully!")

	case "down":
		targetVersion := currentVersion - 1
		if len(os.Args) > 2 {
			targetVersion = atoi(os.Args[2])
		}

		for i := len(migrations) - 1; i >= 0; i-- {
			migration := migrations[i]
			if migration.Version <= currentVersion && migration.Version > targetVersion {
				fmt.Printf("Rolling back migration %d: %s\n", migration.Version, migration.Name)
				if err := migration.Down(db); err != nil {
					log.Fatal(err)
				}
				currentVersion = migration.Version - 1
				_, _ = db.Collection("migrations").UpdateOne(
					ctx,
					bson.M{},
					bson.M{"$set": bson.M{"version": currentVersion}},
					options.Update().SetUpsert(true),
				)
			}
		}
		fmt.Println("Rollback completed successfully!")

	case "version":
		fmt.Printf("Current version: %d\n", currentVersion)

	case "status":
		fmt.Println("Migration Status:")
		fmt.Println("================")
		for _, migration := range migrations {
			status := "pending"
			if migration.Version <= currentVersion {
				status = "applied"
			}
			fmt.Printf("[%s] %d: %s\n", status, migration.Version, migration.Name)
		}

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

func atoi(s string) int {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return 0
	}
	return n
}
```

**Makefile targets:**
```makefile
.PHONY: migrate-up migrate-down migrate-status

migrate-up:
	go run cmd/migrate/main.go up

migrate-down:
	go run cmd/migrate/main.go down

migrate-status:
	go run cmd/migrate/main.go status

migrate-create:
	@read -p "Migration name: " name; \
	echo "Creating migration $$name..."; \
	touch migrations/$$(date +%Y%m%d%H%M%S)_$$name.go
```

---

## Rollback Strategies

### Blue-Green Deployment

```bash
#!/bin/bash
# scripts/blue-green-deploy.sh

ENVIRONMENT=$1
NEW_VERSION=$2
CURRENT_VERSION=$(aws ecs describe-services \
    --cluster wedding-$ENVIRONMENT \
    --services wedding-api \
    --query 'services[0].deployments[0].taskDefinition' \
    --output text | cut -d':' -f2)

echo "Current version: $CURRENT_VERSION"
echo "Deploying version: $NEW_VERSION"

# Deploy to blue environment
aws ecs update-service \
    --cluster wedding-$ENVIRONMENT \
    --service wedding-api-blue \
    --task-definition wedding-api:$NEW_VERSION \
    --force-new-deployment

# Wait for blue to be healthy
echo "Waiting for blue deployment to stabilize..."
aws ecs wait services-stable \
    --cluster wedding-$ENVIRONMENT \
    --services wedding-api-blue

# Run smoke tests on blue
if ! curl -sf https://api-blue.wedding.$ENVIRONMENT/health; then
    echo "Blue deployment failed health check!"
    exit 1
fi

# Switch traffic to blue
aws route53 change-resource-record-sets \
    --hosted-zone-id $HOSTED_ZONE_ID \
    --change-batch file://switch-to-blue.json

echo "Traffic switched to blue environment"

# Keep green running for 10 minutes as backup
sleep 600

# Update green to new version (for next deployment)
aws ecs update-service \
    --cluster wedding-$ENVIRONMENT \
    --service wedding-api-green \
    --task-definition wedding-api:$NEW_VERSION \
    --force-new-deployment

echo "Blue-green deployment completed!"
```

### Quick Rollback Script

```bash
#!/bin/bash
# scripts/rollback.sh

ENVIRONMENT=$1
PREVIOUS_VERSION=$2

echo "Rolling back $ENVIRONMENT to version $PREVIOUS_VERSION..."

# Immediate rollback - update service
aws ecs update-service \
    --cluster wedding-$ENVIRONMENT \
    --service wedding-api \
    --task-definition wedding-api:$PREVIOUS_VERSION \
    --force-new-deployment

# Wait for rollback to complete
echo "Waiting for rollback to complete..."
aws ecs wait services-stable \
    --cluster wedding-$ENVIRONMENT \
    --services wedding-api

# Verify rollback
if curl -sf https://api.wedding.$ENVIRONMENT/health; then
    echo "Rollback successful!"
    
    # Send notification
    curl -X POST $SLACK_WEBHOOK_URL \
        -H 'Content-type: application/json' \
        --data "{\"text\":\"Rollback to version $PREVIOUS_VERSION completed successfully\"}"
else
    echo "Rollback verification failed!"
    exit 1
fi
```

---

## Complete Configuration Examples

### Systemd Service

```ini
# /etc/systemd/system/wedding-api.service
[Unit]
Description=Wedding API Server
After=network.target
Wants=network.target

[Service]
Type=simple
User=appuser
Group=appgroup

WorkingDirectory=/opt/wedding-api
ExecStart=/opt/wedding-api/wedding-api

# Restart configuration
Restart=always
RestartSec=5
StartLimitInterval=60s
StartLimitBurst=3

# Resource limits
LimitAS=1G
LimitRSS=512M
LimitNOFILE=65536
LimitNPROC=4096

# Security
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/wedding-api/logs

# Environment
Environment="APP_ENV=production"
Environment="APP_PORT=8080"
Environment="LOG_LEVEL=info"
Environment="MONGODB_URI=mongodb://localhost:27017/wedding_prod"
EnvironmentFile=-/opt/wedding-api/.env

# Health check
ExecStartPost=/bin/sh -c 'sleep 5 && curl -sf http://localhost:8080/health || exit 1'

[Install]
WantedBy=multi-user.target
```

**Systemd commands:**
```bash
# Enable service
sudo systemctl enable wedding-api

# Start service
sudo systemctl start wedding-api

# Check status
sudo systemctl status wedding-api

# View logs
sudo journalctl -u wedding-api -f

# Restart
sudo systemctl restart wedding-api

# Reload configuration
sudo systemctl daemon-reload
```

### Kubernetes Manifests

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: wedding
  labels:
    name: wedding
    environment: production
---
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: wedding-api
  namespace: wedding
  labels:
    app: wedding-api
    version: v1.2.0
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app: wedding-api
  template:
    metadata:
      labels:
        app: wedding-api
        version: v1.2.0
    spec:
      serviceAccountName: wedding-api
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      containers:
        - name: api
          image: wedding-api:v1.2.0
          imagePullPolicy: Always
          ports:
            - name: http
              containerPort: 8080
              protocol: TCP
          env:
            - name: APP_ENV
              value: "production"
            - name: APP_PORT
              value: "8080"
            - name: LOG_LEVEL
              value: "info"
            - name: MONGODB_URI
              valueFrom:
                secretKeyRef:
                  name: wedding-secrets
                  key: mongodb-uri
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: wedding-secrets
                  key: jwt-secret
          resources:
            requests:
              cpu: 250m
              memory: 256Mi
            limits:
              cpu: 500m
              memory: 512Mi
          livenessProbe:
            httpGet:
              path: /health/live
              port: http
            initialDelaySeconds: 30
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 3
          readinessProbe:
            httpGet:
              path: /health/ready
              port: http
            initialDelaySeconds: 5
            periodSeconds: 5
            timeoutSeconds: 3
            failureThreshold: 3
          startupProbe:
            httpGet:
              path: /health
              port: http
            initialDelaySeconds: 10
            periodSeconds: 5
            failureThreshold: 30
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop:
                - ALL
          volumeMounts:
            - name: tmp
              mountPath: /tmp
      volumes:
        - name: tmp
          emptyDir: {}
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 100
              podAffinityTerm:
                labelSelector:
                  matchExpressions:
                    - key: app
                      operator: In
                      values:
                        - wedding-api
                topologyKey: kubernetes.io/hostname
---
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: wedding-api
  namespace: wedding
  labels:
    app: wedding-api
spec:
  type: ClusterIP
  ports:
    - port: 80
      targetPort: http
      protocol: TCP
      name: http
  selector:
    app: wedding-api
---
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: wedding-api
  namespace: wedding
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-body-size: "50m"
    nginx.ingress.kubernetes.io/rate-limit: "100"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/cors-allow-origin: "https://wedding.app"
    nginx.ingress.kubernetes.io/cors-allow-methods: "GET, POST, PUT, DELETE, OPTIONS"
    nginx.ingress.kubernetes.io/cors-allow-headers: "Authorization, Content-Type, X-Request-ID"
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - api.wedding.app
      secretName: wedding-api-tls
  rules:
    - host: api.wedding.app
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: wedding-api
                port:
                  number: 80
---
# k8s/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: wedding-secrets
  namespace: wedding
type: Opaque
stringData:
  mongodb-uri: "mongodb+srv://..."
  jwt-secret: "..."
  aws-access-key: "..."
  aws-secret-key: "..."
---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: wedding-config
  namespace: wedding
data:
  APP_ENV: "production"
  LOG_LEVEL: "info"
  RATE_LIMIT_REQUESTS: "100"
  RATE_LIMIT_WINDOW: "60"
  S3_BUCKET: "wedding-production-uploads"
  S3_REGION: "us-east-1"
  ENABLE_SWAGGER: "false"
```

**Deployment Script:**
```bash
#!/bin/bash
# scripts/k8s-deploy.sh

NAMESPACE="wedding"
VERSION=${1:-latest}

echo "Deploying version $VERSION to Kubernetes..."

# Update image tag
sed -i "s|image: wedding-api:.*|image: wedding-api:$VERSION|" k8s/deployment.yaml

# Apply manifests
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/secrets.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml
kubectl apply -f k8s/hpa.yaml

# Wait for rollout
echo "Waiting for deployment to complete..."
kubectl rollout status deployment/wedding-api -n $NAMESPACE

# Verify deployment
kubectl get pods -n $NAMESPACE
kubectl get svc -n $NAMESPACE
kubectl get ingress -n $NAMESPACE

echo "Deployment completed!"
```

---

## Deployment Checklist Summary

### Pre-Deployment
- [ ] All tests passing
- [ ] Security scan completed
- [ ] Database migrations reviewed
- [ ] Backup created
- [ ] Rollback plan documented
- [ ] Environment variables configured
- [ ] SSL certificates valid
- [ ] Monitoring dashboards ready
- [ ] On-call team notified

### During Deployment
- [ ] Maintenance mode enabled (if applicable)
- [ ] Database migrations executed
- [ ] New version deployed
- [ ] Health checks passing
- [ ] Smoke tests executed
- [ ] Logs monitored for errors
- [ ] Performance metrics checked

### Post-Deployment
- [ ] Maintenance mode disabled
- [ ] All endpoints responding correctly
- [ ] Error rates within acceptable limits
- [ ] Response times normal
- [ ] Database connections healthy
- [ ] File uploads working
- [ ] Email notifications functioning
- [ ] Rollback window passed successfully
- [ ] Team notified of success

### Emergency Rollback
- [ ] Identify issue and severity
- [ ] Alert team members
- [ ] Execute rollback procedure
- [ ] Verify rollback success
- [ ] Create incident report
- [ ] Plan fix for next deployment