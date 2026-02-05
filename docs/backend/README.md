# Wedding Invitation Backend Documentation

## 🎯 Overview

This documentation provides comprehensive guidance for building and deploying a production-ready backend for the Wedding Invitation system using **Golang** and **MongoDB**.

## 📚 Documentation Structure

### Getting Started
| Document | Description |
|----------|-------------|
| [README.md](./README.md) | This file - navigation and overview |
| [01-architecture-overview.md](./01-architecture-overview.md) | System architecture, design patterns, and technology choices |

### Core Documentation
| Document | Description |
|----------|-------------|
| [02-database-schema.md](./02-database-schema.md) | MongoDB collections, Go structs, indexes, and queries |
| [03-api-reference.md](./03-api-reference.md) | Complete REST API specifications with examples |
| [04-authentication-flow.md](./04-authentication-flow.md) | JWT implementation, security flows, and best practices |
| [05-project-structure.md](./05-project-structure.md) | Go project layout and code organization |

### Implementation Guides
| Document | Description |
|----------|-------------|
| [06-implementation-roadmap.md](./06-implementation-roadmap.md) | 5-phase development plan with deliverables |
| [07-file-upload-strategy.md](./07-file-upload-strategy.md) | Image upload architecture and S3/R2 integration |
| [08-testing-guide.md](./08-testing-guide.md) | Unit testing, integration testing, and mocking strategies |
| [13-frontend-integration.md](./13-frontend-integration.md) | Connecting the existing frontend to the backend |

### Operations & Deployment
| Document | Description |
|----------|-------------|
| [09-deployment-guide.md](./09-deployment-guide.md) | Docker, CI/CD, and production deployment |
| [10-environment-configuration.md](./10-environment-configuration.md) | Environment variables and secrets management |
| [11-security-hardening.md](./11-security-hardening.md) | Security best practices and hardening |
| [12-cost-analysis.md](./12-cost-analysis.md) | Infrastructure cost estimates and scaling |

## 🚀 Quick Start

### For Backend Developers
1. Start with [01-architecture-overview.md](./01-architecture-overview.md) to understand the system design
2. Review [02-database-schema.md](./02-database-schema.md) for data models
3. Check [05-project-structure.md](./05-project-structure.md) for code organization
4. Follow [06-implementation-roadmap.md](./06-implementation-roadmap.md) for development phases

### For Frontend Developers
1. Read [03-api-reference.md](./03-api-reference.md) for API endpoints
2. Check [13-frontend-integration.md](./13-frontend-integration.md) for JavaScript client examples
3. Review [04-authentication-flow.md](./04-authentication-flow.md) for auth implementation

### For DevOps/Deployment
1. Review [09-deployment-guide.md](./09-deployment-guide.md) for deployment strategies
2. Check [10-environment-configuration.md](./10-environment-configuration.md) for configuration
3. Read [12-cost-analysis.md](./12-cost-analysis.md) for infrastructure planning

## 📋 Key Features

### Authentication & Security
- JWT-based authentication with refresh tokens
- Password hashing with bcrypt
- Rate limiting and CORS protection
- Email verification and password reset

### Wedding Management
- Create and manage wedding invitations
- Custom URL slugs for sharing
- Public and private wedding modes
- Password protection option

### Guest & RSVP Management
- Guest list with CSV import
- RSVP collection and tracking
- Dietary restrictions and custom questions
- Plus-one management
- Email confirmations

### Media & Storage
- Image upload for couple photos and galleries
- Cloud storage integration (S3/R2)
- CDN delivery for fast loading
- Automatic image optimization

### Analytics
- Page view tracking
- RSVP statistics and insights
- Export to CSV/Excel
- Real-time dashboards

## 🛠 Technology Stack

| Layer | Technology |
|-------|-----------|
| **Language** | Go 1.21+ |
| **Web Framework** | Gin |
| **Database** | MongoDB 6.0+ |
| **Authentication** | JWT + bcrypt |
| **File Storage** | AWS S3 / Cloudflare R2 |
| **Email** | SendGrid / AWS SES |
| **Documentation** | Swagger/OpenAPI |
| **Testing** | Testify + httptest |
| **Deployment** | Docker + Docker Compose |

## 📊 Project Timeline

**Total Duration:** 6 weeks

- **Phase 1:** Foundation (Week 1-2) - Project setup, auth, basic CRUD
- **Phase 2:** Core Features (Week 3) - Wedding management, public API, file uploads
- **Phase 3:** Guest Management (Week 4) - RSVPs, guest lists, email notifications
- **Phase 4:** Advanced Features (Week 5) - Analytics, security hardening
- **Phase 5:** Deployment (Week 6) - Production setup, CI/CD, monitoring

See [06-implementation-roadmap.md](./06-implementation-roadmap.md) for detailed breakdown.

## 🔐 Security Checklist

- ✅ JWT with RS256 asymmetric keys
- ✅ bcrypt password hashing (cost 12)
- ✅ Rate limiting on all endpoints
- ✅ Input validation and sanitization
- ✅ CORS configuration
- ✅ Security headers (CSP, HSTS, etc.)
- ✅ File upload validation
- ✅ SQL/NoSQL injection prevention

See [11-security-hardening.md](./11-security-hardening.md) for complete security guide.

## 💰 Cost Estimates

### Small Scale (1-100 weddings)
- **Total:** $85-200/month
- Server: $20-50
- MongoDB: $60-80
- Storage: $5-20
- Email: $0 (free tier)

### Large Scale (1000+ weddings)
- **Total:** $390-1200/month
- Server: $100-300
- MongoDB: $200-500
- Storage: $50-200
- Email: $20-100

See [12-cost-analysis.md](./12-cost-analysis.md) for detailed breakdown.

## 🆘 Getting Help

### Common Issues
- **MongoDB Connection:** Check [02-database-schema.md](./02-database-schema.md) connection section
- **Authentication Errors:** Review [04-authentication-flow.md](./04-authentication-flow.md) troubleshooting
- **File Uploads:** See [07-file-upload-strategy.md](./07-file-upload-strategy.md) configuration
- **Deployment:** Check [09-deployment-guide.md](./09-deployment-guide.md) for platform-specific guides

### Next Steps
1. Read the architecture overview
2. Set up your development environment
3. Follow the implementation roadmap
4. Join the team for code reviews

---

**Version:** 1.0  
**Last Updated:** 2024-01-15  
**Status:** Ready for Implementation
