# 12. Cost Analysis

## Table of Contents

1. [Cost Overview and Factors](#1-cost-overview-and-factors)
2. [Infrastructure Components](#2-infrastructure-components)
3. [Cost Scenarios](#3-cost-scenarios)
4. [Cost Breakdown by Service](#4-cost-breakdown-by-service)
5. [Cost Optimization Strategies](#5-cost-optimization-strategies)
6. [Free Tier Options](#6-free-tier-options)
7. [Hidden Costs](#7-hidden-costs)
8. [Monitoring Costs](#8-monitoring-costs)
9. [Scaling Cost Projections](#9-scaling-cost-projections)
10. [ROI Calculation](#10-roi-calculation)
11. [Budget Recommendations](#11-budget-recommendations)
12. [Cost Monitoring Tools](#12-cost-monitoring-tools)
13. [Complete Pricing Tables](#13-complete-pricing-tables)

---

## 1. Cost Overview and Factors

### Key Cost Drivers

The total cost of running a wedding invitation backend depends on several factors:

1. **Number of Active Weddings**: Each wedding generates database records, file uploads, and API traffic
2. **Guest Count**: Higher guest counts = more invitations sent = more email delivery
3. **File Upload Volume**: Photos, videos, and media assets consume storage and bandwidth
4. **Traffic Patterns**: RSVP page visits, API calls, and concurrent users
5. **Geographic Distribution**: CDN costs vary based on user locations
6. **Data Retention**: How long you keep wedding data after the event

### Cost Scaling Factors

| Factor | Small Impact | Medium Impact | Large Impact |
|--------|-------------|---------------|--------------|
| Weddings | 1-50 | 50-500 | 500+ |
| Guests per Wedding | 20-50 | 50-150 | 150+ |
| Media Uploads | 10MB/wedding | 50MB/wedding | 200MB+ |
| Concurrent Users | < 10 | 10-50 | 50+ |
| API Requests/Month | 10K | 100K | 1M+ |

---

## 2. Infrastructure Components

### 2.1 Compute (Server/VM Costs)

**Options and Pricing:**

| Provider | Instance Type | Specs | Monthly Cost |
|----------|--------------|-------|--------------|
| AWS EC2 | t3.micro | 2 vCPU, 1GB RAM | $8.47 |
| AWS EC2 | t3.small | 2 vCPU, 2GB RAM | $16.94 |
| AWS EC2 | t3.medium | 2 vCPU, 4GB RAM | $33.87 |
| AWS EC2 | t3.large | 2 vCPU, 8GB RAM | $67.74 |
| DigitalOcean | Basic | 1 vCPU, 1GB RAM | $6.00 |
| DigitalOcean | Basic | 2 vCPU, 4GB RAM | $24.00 |
| DigitalOcean | General | 4 vCPU, 8GB RAM | $48.00 |
| Heroku | Hobby | 1 dyno | $7.00 |
| Heroku | Standard | 1 dyno | $25.00 |
| Heroku | Performance M | 1 dyno | $250.00 |
| Railway | Starter | Shared | $5.00 |
| Railway | Pro | 2 vCPU, 2GB RAM | $20.00 |
| Render | Starter | 512MB RAM | $7.00 |
| Render | Standard | 2GB RAM | $25.00 |
| Vercel | Pro (Serverless) | 100GB bandwidth | $20.00 |

**Recommended Configurations:**
- Small Scale: 1 vCPU, 2GB RAM ($16-25/month)
- Medium Scale: 2 vCPU, 4GB RAM ($33-48/month)
- Large Scale: 4 vCPU, 8GB RAM ($67-100/month)

### 2.2 Database (MongoDB Hosting)

**MongoDB Atlas Tiers:**

| Tier | Specs | Storage | Monthly Cost |
|------|-------|---------|--------------|
| M0 (Free) | Shared RAM | 512MB | $0 |
| M2 | 2GB RAM, 10GB storage | 10GB | $9 |
| M5 | 2GB RAM, 20GB storage | 20GB | $25 |
| M10 | 2GB RAM, 10GB storage | 10GB | $57 |
| M20 | 4GB RAM, 20GB storage | 20GB | $115 |
| M30 | 8GB RAM, 40GB storage | 40GB | $230 |
| M40 | 16GB RAM, 80GB storage | 80GB | $460 |

**Alternative Database Options:**

| Provider | Plan | Storage | Monthly Cost |
|----------|------|---------|--------------|
| AWS DocumentDB | db.t3.medium | 100GB | ~$140 |
| AWS RDS PostgreSQL | db.t3.micro | 20GB | ~$13 |
| DigitalOcean Managed MongoDB | 2GB RAM | 25GB | $15 |
| MongoDB Community (self-hosted) | Your server | Unlimited | $0 + server cost |

### 2.3 Storage (File Uploads, Backups)

**Cloud Storage Pricing:**

| Provider | Service | Storage Cost/GB | Egress Cost/GB |
|----------|---------|-----------------|----------------|
| AWS S3 | Standard | $0.023 | $0.09 |
| AWS S3 | Standard-IA | $0.0125 | $0.09 |
| AWS S3 | Glacier | $0.004 | $0.09 |
| Cloudflare R2 | Standard | $0.015 | $0.00 (free) |
| Google Cloud Storage | Standard | $0.020 | $0.12 |
| Google Cloud Storage | Nearline | $0.010 | $0.12 |
| Azure Blob Storage | Hot | $0.0184 | $0.087 |
| Azure Blob Storage | Cool | $0.0100 | $0.087 |
| DigitalOcean Spaces | - | $0.020 | $0.00 (free) |
| Backblaze B2 | - | $0.005 | $0.00 (first 1GB/day free) |

**Estimated Storage Needs:**
- Small: 5-20GB ($0.10-0.40/month)
- Medium: 20-100GB ($0.30-2.00/month)
- Large: 100-500GB ($1.50-10.00/month)

**Backup Storage:**
- AWS S3 Glacier: $0.004/GB/month
- Daily backups of 10GB = $1.20/month
- Weekly backups of 50GB = $0.86/month

### 2.4 CDN (Content Delivery)

**CDN Provider Pricing:**

| Provider | Bandwidth Cost | Requests | Additional Features |
|----------|----------------|----------|---------------------|
| Cloudflare | $0.00 | Unlimited | Free tier includes CDN |
| Cloudflare Pro | $20/month | Unlimited | $20/month flat |
| AWS CloudFront | $0.085/GB | $0.0075/10K | First 10TB tier |
| AWS CloudFront | $0.080/GB | $0.0075/10K | Next 40TB tier |
| Fastly | $0.12/GB | Included | Minimum $50/month |
| KeyCDN | $0.04/GB | Included | $4 minimum/month |
| BunnyCDN | $0.01/GB | Included | Volume pricing available |
| DigitalOcean CDN | $0.02/GB | Included | Spaces integration |

**Bandwidth Estimates:**
- Small Scale (100 weddings): 50-100GB/month = $0-5/month
- Medium Scale (500 weddings): 200-500GB/month = $0-20/month
- Large Scale (2000 weddings): 1-2TB/month = $20-100/month

### 2.5 Email Service

**Transactional Email Providers:**

| Provider | Free Tier | Paid Tier 1 | Cost/1K Emails |
|----------|-----------|-------------|----------------|
| SendGrid | 100/day | $19.95/month | $0.001 |
| Mailgun | 5K/month (3mo) | $35/month | $0.0008 |
| AWS SES | 62K free (from EC2) | $0.10/1K | $0.0001 |
| Postmark | 100/month | $10/month | $0.00125 |
| Mailjet | 200/day | $15/month | $0.001 |
| Sendinblue | 300/day | $25/month | $0.001 |
| Resend | 3K/month | $20/month | $0.0012 |

**Email Volume Estimates:**
- Invitations: 50-200 per wedding
- RSVPs: 50-150 per wedding
- Reminders: 2-5 per guest
- Thank you notes: 50-200 per wedding

### 2.6 Domain and SSL

| Service | Provider | Annual Cost | Monthly Equivalent |
|---------|----------|-------------|-------------------|
| Domain (.com) | Namecheap | $9-15/year | $0.75-1.25 |
| Domain (.com) | Cloudflare | $9.15/year | $0.76 |
| Domain (.wedding) | Various | $30-50/year | $2.50-4.17 |
| SSL Certificate | Let's Encrypt | Free | $0.00 |
| SSL Certificate (Wildcard) | Cloudflare | $0 (included) | $0.00 |
| SSL Certificate (EV) | DigiCert | $300+/year | $25.00+ |

### 2.7 Monitoring and Logging

| Service | Plan | Cost | Features |
|---------|------|------|----------|
| Datadog | Free | $0 | 5 hosts, 1-day retention |
| Datadog | Pro | $15/host/month | Full features |
| New Relic | Free | $0 | 100GB/month data |
| New Relic | Pro | $49/month | 250GB/month data |
| AWS CloudWatch | Free Tier | $0 | 10 metrics, alarms |
| AWS CloudWatch | Paid | $0.30/metric | Custom metrics |
| LogDNA (Mezmo) | Free | $0 | 10GB/day |
| LogDNA (Mezmo) | Pro | $3/GB | 30-day retention |
| Sentry | Developer | $0 | 5K errors/month |
| Sentry | Team | $26/month | 50K errors/month |
| PagerDuty | Free | $0 | 5 users |
| PagerDuty | Professional | $29/user | Full features |

---

## 3. Cost Scenarios

### 3.1 Small Scale (1-100 Weddings)

**Profile:**
- 1-100 active weddings
- 20-100 guests per wedding
- 1,000-10,000 total guests
- Low to moderate traffic
- Minimal media uploads

**Monthly Cost Breakdown:**

| Component | Service | Cost Range |
|-------------|---------|------------|
| Server | AWS EC2 t3.small / DigitalOcean Basic | $20-50 |
| Database | MongoDB Atlas M5 / M10 | $60-80 |
| Storage | Cloudflare R2 (20GB) / S3 | $5-20 |
| CDN | Cloudflare Free | $0-20 |
| Email | SendGrid Free / SES | $0-10 |
| Domain + SSL | Cloudflare / Let's Encrypt | $1-5 |
| Monitoring | Free tiers only | $0-15 |
| **Total** | | **$85-200/month** |

**Yearly Cost:** $1,020-2,400

### 3.2 Medium Scale (100-1,000 Weddings)

**Profile:**
- 100-1,000 active weddings
- 50-150 guests per wedding
- 5,000-150,000 total guests
- Moderate to high traffic
- Regular media uploads

**Monthly Cost Breakdown:**

| Component | Service | Cost Range |
|-------------|---------|------------|
| Server | AWS EC2 t3.medium / 2x t3.small | $50-100 |
| Database | MongoDB Atlas M20 / M30 | $80-150 |
| Storage | Cloudflare R2 / S3 (100GB) | $20-50 |
| CDN | Cloudflare Pro / CloudFront | $20-50 |
| Email | SendGrid Pro / SES | $10-30 |
| Domain + SSL | Multiple domains | $5-20 |
| Monitoring | New Relic / Datadog | $20-50 |
| **Total** | | **$180-380/month** |

**Yearly Cost:** $2,160-4,560

### 3.3 Large Scale (1,000+ Weddings)

**Profile:**
- 1,000+ active weddings
- 50-300 guests per wedding
- 50,000+ total guests
- High traffic with spikes
- Heavy media uploads
- Multiple regions

**Monthly Cost Breakdown:**

| Component | Service | Cost Range |
|-------------|---------|------------|
| Server | AWS EC2 t3.large / Auto Scaling | $100-300 |
| Database | MongoDB Atlas M40 / M50+ | $200-500 |
| Storage | R2 + S3 (500GB+) | $50-200 |
| CDN | Cloudflare Enterprise / Multi-CDN | $20-100 |
| Email | SendGrid Pro / Dedicated IP | $20-100 |
| Domain + SSL | Multiple + Wildcards | $20-50 |
| Monitoring | Enterprise plans | $50-150 |
| **Total** | | **$390-1,200/month** |

**Yearly Cost:** $4,680-14,400

---

## 4. Cost Breakdown by Service

### 4.1 AWS Pricing Details

**Compute (EC2):**

| Instance | vCPU | Memory | On-Demand | Spot |
|----------|------|--------|-----------|------|
| t3.micro | 2 | 1 GiB | $0.0104/hr | ~$0.003/hr |
| t3.small | 2 | 2 GiB | $0.0208/hr | ~$0.006/hr |
| t3.medium | 2 | 4 GiB | $0.0416/hr | ~$0.012/hr |
| t3.large | 2 | 8 GiB | $0.0832/hr | ~$0.025/hr |
| t3.xlarge | 4 | 16 GiB | $0.1664/hr | ~$0.050/hr |
| m6g.medium | 1 | 4 GiB | $0.0385/hr | ~$0.011/hr |
| c6g.large | 2 | 4 GiB | $0.068/hr | ~$0.020/hr |

**Additional AWS Costs:**
- EBS Storage: $0.10/GB/month (gp2)
- Data Transfer Out: $0.09/GB
- Elastic IP: $0.005/hr (when unattached)
- Load Balancer: $0.0225/hr + $0.008/LCU-hour
- NAT Gateway: $0.045/hr + $0.045/GB processed

**AWS Savings Plans (1-year, No Upfront):**
- Compute Savings Plan: 20-30% savings
- EC2 Instance Savings: 25-40% savings for specific families
- Reserved Instances: 30-60% savings for 1-3 year commitments

### 4.2 MongoDB Atlas Tiers

**Shared Clusters (Free - M5):**

| Tier | RAM | Storage | Connections | Price |
|------|-----|---------|-------------|-------|
| M0 | Shared | 512MB | 500 | Free |
| M2 | Shared | 2GB | 500 | $9/month |
| M5 | Shared | 5GB | 500 | $25/month |

**Dedicated Clusters (M10+):**

| Tier | RAM | vCPU | Storage | Connections | Price |
|------|-----|------|---------|-------------|-------|
| M10 | 2GB | 2 | 10GB | 350 | $57/month |
| M20 | 4GB | 2 | 20GB | 500 | $115/month |
| M30 | 8GB | 2 | 40GB | 650 | $230/month |
| M40 | 16GB | 4 | 80GB | 850 | $460/month |
| M50 | 32GB | 8 | 160GB | 1,300 | $920/month |
| M60 | 64GB | 16 | 320GB | 2,000 | $1,840/month |

**Atlas Add-ons:**
- Backup (M10+): $0.20/GB/month
- Private Endpoint: $0.01/hr
- Advanced Security: $0.50/1M requests
- Data Transfer Out: $0.10/GB

### 4.3 Cloudflare R2 vs S3 Costs

**Storage Cost Comparison:**

| Provider | Storage/GB | Class A Operations | Class B Operations | Egress |
|----------|------------|-------------------|-------------------|--------|
| Cloudflare R2 | $0.015 | $0.0045/1K | $0.00036/1K | Free |
| AWS S3 Standard | $0.023 | $0.005/1K | $0.0004/1K | $0.09/GB |
| AWS S3 Standard-IA | $0.0125 | $0.01/1K | $0.001/1K | $0.09/GB |
| AWS S3 Glacier | $0.004 | $0.10/1K | $0.01/1K | $0.09/GB |

**Cost Example: 100GB Storage, 1M requests, 500GB egress**

| Provider | Storage | Requests | Egress | Total |
|----------|---------|----------|--------|-------|
| Cloudflare R2 | $1.50 | $4.50 | $0.00 | **$6.00** |
| AWS S3 Standard | $2.30 | $4.90 | $45.00 | **$52.20** |
| AWS S3 Standard-IA | $1.25 | $10.00 | $45.00 | **$56.25** |

### 4.4 SendGrid Pricing Tiers

**Email API Plans:**

| Plan | Monthly Emails | Price | Features |
|------|----------------|-------|----------|
| Free | 100/day (3,000/mo) | $0 | Basic features |
| Essentials 50K | 50,000 | $19.95 | Dedicated IP, 1 user |
| Essentials 100K | 100,000 | $34.95 | Dedicated IP, 1 user |
| Essentials 200K | 200,000 | $54.95 | Dedicated IP, 1 user |
| Pro 100K | 100,000 | $89.95 | Subusers, 10 users |
| Pro 300K | 300,000 | $249.95 | Subusers, 15 users |
| Pro 700K | 700,000 | $449.95 | Subusers, 20 users |
| Premier | Custom | Custom | Dedicated support |

**Overages:**
- Essentials: $0.00125/email
- Pro: $0.001/email

---

## 5. Cost Optimization Strategies

### 5.1 Right-Sizing Instances

**Best Practices:**
1. Monitor actual CPU/Memory usage for 2-4 weeks
2. Start with smallest viable instance
3. Scale up only when consistently hitting 70%+ utilization
4. Use vertical scaling (bigger instance) before horizontal

**Instance Selection Guide:**
- <10 concurrent users: t3.micro (1GB RAM)
- 10-50 concurrent users: t3.small (2GB RAM)
- 50-200 concurrent users: t3.medium (4GB RAM)
- 200+ concurrent users: t3.large (8GB RAM) or auto-scaling

### 5.2 Reserved Instances

**AWS Reserved Instances (1-year commitment):**

| Instance Type | On-Demand | 1-Year No Upfront | Savings |
|---------------|-----------|-------------------|---------|
| t3.micro | $8.47 | $5.55 | 35% |
| t3.small | $16.94 | $11.11 | 34% |
| t3.medium | $33.87 | $22.25 | 34% |
| t3.large | $67.74 | $44.49 | 34% |

**3-Year Commitment Savings: 50-60%**

**MongoDB Atlas Reserved Capacity:**
- 1-year prepay: 10-15% discount
- 3-year prepay: 20-25% discount

### 5.3 CDN Optimization

**Cost-Saving CDN Strategies:**
1. Use Cloudflare R2 (free egress) instead of S3
2. Enable aggressive browser caching (1 year for static assets)
3. Compress images to WebP format (30-50% smaller)
4. Use Cloudflare Polish for automatic optimization
5. Implement lazy loading for images

**Bandwidth Savings Example:**
- Original bandwidth: 500GB/month
- With optimization: 250GB/month
- S3 egress cost saved: $22.50/month

### 5.4 Image Compression

**Compression Tools and Savings:**

| Format | Original Size | Compressed | Savings |
|--------|---------------|------------|---------|
| JPEG | 100KB | 60KB (quality 80) | 40% |
| PNG | 200KB | 80KB (WebP) | 60% |
| WebP | 100KB | 100KB | 0% (already optimal) |

**Implementation:**
```javascript
// Sharp.js example
const sharp = require('sharp');
await sharp('input.jpg')
  .resize(1920, null, { withoutEnlargement: true })
  .jpeg({ quality: 80, progressive: true })
  .toFile('output.jpg');
```

**Storage Impact:**
- 1000 images @ 5MB each = 5GB
- Compressed to 2MB each = 2GB
- 3GB saved = $0.045/month (R2) or $0.069/month (S3)

### 5.5 Database Indexing

**Indexing for Cost Efficiency:**
- Reduces query time = less CPU usage
- Reduces read operations = lower costs (at scale)
- Prevents unnecessary full collection scans

**Critical Indexes:**
```javascript
// Wedding collection
db.weddings.createIndex({ userId: 1, createdAt: -1 });
db.weddings.createIndex({ slug: 1 }, { unique: true });

// Guest collection
db.guests.createIndex({ weddingId: 1, rsvpStatus: 1 });
db.guests.createIndex({ email: 1 });

// RSVP collection
db.rsvps.createIndex({ guestId: 1, createdAt: -1 });
```

### 5.6 Caching Strategies

**Multi-Layer Caching:**

| Layer | Technology | Cost Impact |
|-------|------------|-------------|
| Browser | Cache-Control headers | Free, reduces requests |
| CDN | Cloudflare cache | Free tier, reduces origin hits |
| Application | Redis/Memcached | $15-30/month, reduces DB calls |
| Database | MongoDB query cache | Built-in, reduces disk reads |

**Redis Cloud Pricing:**
- Free: 30MB, 1 database
- 250MB: $7/month
- 1GB: $22/month
- 5GB: $80/month

**Cache Configuration Example:**
```javascript
// API response caching
res.set('Cache-Control', 'public, max-age=300'); // 5 minutes

// Database query caching
const cacheKey = `wedding:${weddingId}`;
let wedding = await redis.get(cacheKey);
if (!wedding) {
  wedding = await Wedding.findById(weddingId);
  await redis.setex(cacheKey, 300, JSON.stringify(wedding));
}
```

---

## 6. Free Tier Options

### 6.1 MongoDB Atlas Free Tier

**M0 Cluster (Free Forever):**
- 512MB storage
- Shared RAM
- 3 regions available
- 500 max connections
- Monitoring included
- Community support

**Limitations:**
- No backup (must export manually)
- 10GB data transfer limit/month
- No VPC peering
- No advanced security features
- Auto-expires after 6 months of inactivity

**Migration Path:**
M0 → M2 ($9) → M5 ($25) → M10 ($57) → M20+ (dedicated)

### 6.2 AWS Free Tier

**12-Month Free Tier:**

| Service | Free Tier | Monthly Value |
|---------|-----------|---------------|
| EC2 | 750 hrs t2.micro/t3.micro | $8.47 |
| S3 | 5GB standard | $0.12 |
| RDS | 750 hrs db.t2.micro | $13 |
| CloudFront | 50GB data transfer | $4.25 |
| Data Transfer | 15GB out | $1.35 |
| EBS | 30GB | $3.00 |

**Always Free:**
- Lambda: 1M requests + 400K GB-seconds
- DynamoDB: 25GB + 200M reads/writes
- SNS: 1M publishes
- CloudWatch: 10 alarms + 10 metrics

**Wedding App Free Tier Strategy:**
1. Run on EC2 t3.micro (12 months free)
2. Use MongoDB Atlas M0 (free)
3. Store files on S3 5GB (12 months free)
4. Use Cloudflare CDN (always free)
5. SendGrid 100/day (always free)

**Total First Year Cost: ~$10/month** (domain only)

### 6.3 SendGrid Free Tier

**Free Plan Features:**
- 100 emails/day (3,000/month)
- 1 teammate
- API access
- Real-time analytics
- Deliverability features
- 2-day email history

**Rate Limits:**
- 10 requests/second (API)
- No dedicated IP
- Shared reputation

**When to Upgrade:**
- >100 emails/day consistently
- Need dedicated IP (high volume)
- Need subuser accounts
- Need phone support

### 6.4 Cloudflare Free Tier

**Free Plan Includes:**
- Unlimited bandwidth
- Unlimited requests
- Global CDN (200+ locations)
- DDoS protection (unmetered)
- SSL certificate
- DNS (unlimited queries)
- Page rules (3)
- Security features
- Analytics

**Limitations:**
- No custom cache rules (beyond 3 page rules)
- No advanced WAF
- No image optimization (Polish)
- No load balancing
- Support: Community only

---

## 7. Hidden Costs

### 7.1 Data Transfer

**AWS Data Transfer Pricing:**

| Transfer Type | Cost |
|---------------|------|
| Inbound | Free |
| Outbound to internet | $0.09/GB |
| Between regions | $0.02/GB |
| Within AZ | Free |
| Between AZs | $0.01/GB |

**Example Scenarios:**

| Scenario | Data Volume | Monthly Cost |
|----------|-------------|--------------|
| 500 weddings, 100 photos each, 1 view/photo | 500GB | $45 |
| 1000 weddings, video uploads (10MB avg) | 1TB | $90 |
| Backup transfer to different region | 100GB | $9 |

**Mitigation Strategies:**
1. Use Cloudflare R2 (free egress)
2. Keep backups in same region
3. Compress before transfer
4. Use CDN for all static assets

### 7.2 API Requests

**AWS API Gateway:**
- REST API: $3.50/million requests
- HTTP API: $1.00/million requests
- WebSocket: $1.00/million messages + $0.25/million connection minutes

**S3 API Costs:**
- PUT, COPY, POST, LIST: $0.005/1K requests
- GET, SELECT: $0.0004/1K requests

**Example:**
- 1M API calls/month: $3.50 (REST) or $1.00 (HTTP)
- 10M GET requests to S3: $4.00
- 1M PUT requests to S3: $5.00

### 7.3 Backup Storage

**MongoDB Atlas Backup Costs:**
- M10+: $0.20/GB/month
- M0-M5: No automated backup (manual export only)

**Calculation:**
- Database size: 50GB
- 7 daily backups: 350GB
- 4 weekly backups: 200GB
- 12 monthly backups: 600GB
- Total backup storage: ~1TB
- Cost: $200/month

**Cost-Effective Backup Strategy:**
- Daily incremental: $0.20 × 50GB = $10
- Weekly full: $0.20 × 50GB × 4 = $40
- Monthly snapshot: $0.20 × 50GB × 3 = $30
- **Total: $80/month** (vs $200 for all daily)

### 7.4 Support Plans

**AWS Support Plans:**

| Plan | Cost | Features |
|------|------|----------|
| Basic | Included | 24/7 customer service, documentation |
| Developer | $29/month | Business hours email, <12hr response |
| Business | $100/month or 10% monthly bill | 24/7 phone, <4hr response, architecture guidance |
| Enterprise | $15,000/month or 10% monthly bill | <15min response, dedicated TAM, Well-Architected |

**When to Pay:**
- Developer: Testing/production workloads
- Business: >$1,000/month spend, need 24/7 support
- Enterprise: Mission-critical, need dedicated support

---

## 8. Monitoring Costs

### 8.1 CloudWatch Costs

| Metric Type | Cost |
|-------------|------|
| Custom metrics | $0.30/metric/month |
| Dashboard | $3.00/dashboard/month |
| Alarm | $0.10/alarm/month |
| Logs ingestion | $0.50/GB |
| Logs storage | $0.03/GB/month |
| Logs Insights queries | $0.005/GB scanned |

**Typical Wedding App Setup:**
- 20 custom metrics: $6/month
- 2 dashboards: $6/month
- 10 alarms: $1/month
- 10GB logs: $5/month
- **Total: ~$18/month**

### 8.2 Third-Party Monitoring

| Service | Plan | Cost | Best For |
|---------|------|------|----------|
| Datadog | Pro | $15/host/month | Full-stack observability |
| New Relic | Pro | $49/month | Application performance |
| Sentry | Team | $26/month | Error tracking |
| LogDNA | Pro | $3/GB | Log management |
| Pingdom | Standard | $15/month | Uptime monitoring |
| StatusCake | Superior | $25/month | Uptime + performance |

**Recommended Stack (Medium Scale):**
- CloudWatch (AWS): Included + $18/month
- Sentry: $26/month
- Pingdom: $15/month
- **Total: ~$60/month**

### 8.3 Log Management

**Volume Estimates:**
- Small: 1GB/day = 30GB/month
- Medium: 5GB/day = 150GB/month
- Large: 20GB/day = 600GB/month

**Cost Comparison:**

| Provider | Small | Medium | Large |
|----------|-------|--------|-------|
| CloudWatch | $15 | $75 | $300 |
| Datadog | $15 | $75 | $300 |
| LogDNA | $90 | $450 | $1,800 |
| Mezmo | $90 | $450 | $1,800 |
| Self-hosted (ELK) | ~$50 | ~$150 | ~$400 |

---

## 9. Scaling Cost Projections

### 9.1 Linear Growth Model

**Assumptions:**
- Average wedding has 100 guests
- Each guest generates 10 API requests/month
- Each wedding stores 50MB of media

| Weddings | Guests | API Calls | Storage | Monthly Cost |
|----------|--------|-----------|---------|--------------|
| 10 | 1,000 | 10K | 500MB | $45 |
| 50 | 5,000 | 50K | 2.5GB | $75 |
| 100 | 10,000 | 100K | 5GB | $120 |
| 250 | 25,000 | 250K | 12.5GB | $180 |
| 500 | 50,000 | 500K | 25GB | $250 |
| 1,000 | 100,000 | 1M | 50GB | $380 |
| 2,500 | 250,000 | 2.5M | 125GB | $600 |
| 5,000 | 500,000 | 5M | 250GB | $900 |

### 9.2 Bursty Traffic Scenarios

**Wedding Season Impact:**
- Peak months (May-October): 3x normal traffic
- Invitation sends: 100 emails in 1 hour
- RSVP deadline day: 50 concurrent users
- Photo upload after event: 1000 uploads in 1 day

**Cost Impact:**
- Auto-scaling adds 2x servers for 6 months: +$100/month average
- CDN costs increase 3x during peak: +$60/month average
- Database may need temporary upgrade: +$100/month

### 9.3 Multi-Region Costs

**Adding Regions:**

| Component | Single Region | Two Regions | Three Regions |
|-----------|---------------|-------------|---------------|
| Compute | $50 | $100 | $150 |
| Database (replica) | $80 | $160 | $240 |
| Storage (replicated) | $20 | $40 | $60 |
| CDN | $20 | $20 | $20 |
| Data Transfer | $10 | $30 | $50 |
| **Total** | $180 | $350 | $520 |

**When Multi-Region Makes Sense:**
- >20% users in different continent
- Regulatory requirements (GDPR, data residency)
- Uptime SLA >99.9%
- 10,000+ concurrent users

---

## 10. ROI Calculation

### 10.1 Revenue Models

**Freemium Model:**
| Tier | Price | Features |
|------|-------|----------|
| Free | $0 | 1 wedding, 50 guests, basic templates |
| Pro | $19.99 | 1 wedding, 200 guests, all templates |
| Premium | $49.99 | 1 wedding, unlimited guests, custom domain |
| Plus | $99.99 | 2 weddings, unlimited guests, analytics |

**Subscription Model:**
| Plan | Monthly | Annual | Features |
|------|---------|--------|----------|
| Basic | $9.99 | $99.99 | 1 active wedding |
| Family | $19.99 | $199.99 | 3 active weddings |
| Event Planner | $49.99 | $499.99 | 10 active weddings |
| Enterprise | Custom | Custom | Unlimited |

### 10.2 Break-Even Analysis

**Scenario: Freemium with 2% Conversion**

| Weddings | Free Users | Paid Users (2%) | Revenue | Costs | Profit |
|----------|------------|-----------------|---------|-------|--------|
| 100 | 98 | 2 × $35 avg | $70 | $120 | -$50 |
| 500 | 490 | 10 × $35 avg | $350 | $250 | $100 |
| 1,000 | 980 | 20 × $35 avg | $700 | $380 | $320 |
| 5,000 | 4,900 | 100 × $35 avg | $3,500 | $900 | $2,600 |
| 10,000 | 9,800 | 200 × $35 avg | $7,000 | $1,500 | $5,500 |

**Break-even Point:** ~400 weddings (assuming $35 ARPU)

### 10.3 Cost Per Wedding

| Scale | Total Monthly | Weddings | Cost/Wedding |
|-------|---------------|----------|--------------|
| Small | $150 | 50 | $3.00 |
| Medium | $300 | 300 | $1.00 |
| Large | $800 | 1,000 | $0.80 |
| Enterprise | $2,000 | 5,000 | $0.40 |

**Gross Margin Calculation (Freemium):**
- Revenue per paid wedding: $35 (one-time)
- Infrastructure cost per wedding: $1-3/month
- Payment processing (Stripe): 2.9% + $0.30 = $1.32
- Gross profit per wedding: $30-33

---

## 11. Budget Recommendations

### 11.1 Startup Budget (Year 1)

**Phase 1: MVP (Months 1-3)**

| Category | Amount | Notes |
|----------|--------|-------|
| Infrastructure | $30/month | Free tier + t3.micro |
| Tools/Services | $50/month | Monitoring, email |
| Domain | $20/year | .com + .wedding |
| Development | $0 | Self-built |
| Marketing | $100/month | Ads, content |
| **Total (3 months)** | **$740** | |

**Phase 2: Launch (Months 4-6)**

| Category | Amount | Notes |
|----------|--------|-------|
| Infrastructure | $100/month | Small scale setup |
| Tools/Services | $100/month | Full monitoring stack |
| Marketing | $300/month | Growth phase |
| Support | $50/month | Part-time help |
| **Total (3 months)** | **$1,650** | |

**Phase 3: Growth (Months 7-12)**

| Category | Amount | Notes |
|----------|--------|-------|
| Infrastructure | $200/month | Medium scale |
| Tools/Services | $150/month | Advanced features |
| Marketing | $500/month | Scale marketing |
| Support | $200/month | Customer service |
| **Total (6 months)** | **$6,300** | |

**Year 1 Total Budget: $8,690**

### 11.2 Sustainable Business Budget

**Monthly Operating Costs (1000 weddings):**

| Category | Amount | % of Total |
|----------|--------|------------|
| Infrastructure | $400 | 33% |
| Payment Processing | $200 | 17% |
| Customer Support | $300 | 25% |
| Marketing | $200 | 17% |
| Tools/Software | $100 | 8% |
| **Total Monthly** | **$1,200** | |
| **Total Annual** | **$14,400** | |

**Required Revenue to Break Even:**
- At $35 ARPU: 411 paid weddings/year
- At $20 ARPU: 720 paid weddings/year
- At $50 ARPU: 288 paid weddings/year

### 11.3 Enterprise Budget

**For 10,000+ weddings with 99.99% uptime SLA:**

| Category | Monthly | Annual |
|----------|---------|--------|
| Multi-region infrastructure | $3,000 | $36,000 |
| Enterprise database cluster | $2,000 | $24,000 |
| CDN (enterprise tier) | $500 | $6,000 |
| Security & compliance | $1,000 | $12,000 |
| Monitoring & logging | $500 | $6,000 |
| Support team (3 people) | $15,000 | $180,000 |
| DevOps engineer | $10,000 | $120,000 |
| **Infrastructure Total** | $32,000 | $384,000 |

---

## 12. Cost Monitoring Tools

### 12.1 AWS Cost Management

**Free Tools:**
1. **AWS Cost Explorer**
   - Visualize costs by service
   - Forecast future costs
   - Identify trends

2. **AWS Budgets**
   - Set budget thresholds
   - Email alerts at 80%, 100%, forecasted
   - Free: 2 budgets
   - Paid: $0.10/budget/day beyond 2

3. **AWS Cost Anomaly Detection**
   - ML-powered anomaly detection
   - Free for AWS usage
   - Alerts on unexpected costs

4. **AWS Cost and Usage Report (CUR)**
   - Detailed CSV reports
   - Delivered to S3
   - Free (S3 storage costs apply)

### 12.2 Third-Party Cost Management

| Tool | Price | Features |
|------|-------|----------|
| CloudHealth | $500+/month | Multi-cloud, optimization |
| CloudCheckr | $0.007/resource/day | AWS, Azure, GCP |
| Kubecost | $449/cluster/month | Kubernetes cost analysis |
| Vantage | Free tier available | AWS cost tracking |
| CloudForecast | $99/month | AWS cost forecasting |

### 12.3 Open Source Tools

1. **Cloud Custodian**
   - Policy-based resource management
   - Automated cost optimization
   - Free (self-hosted)

2. **Infracost**
   - Terraform cost estimation
   - CI/CD integration
   - Free tier: 1,000 runs/month

3. **OpenCost**
   - Kubernetes cost monitoring
   - Prometheus integration
   - Free (open source)

### 12.4 Cost Alerts Setup

**AWS Budget Alert Example:**
```json
{
  "BudgetName": "WeddingApp-Monthly",
  "BudgetLimit": {
    "Amount": "500",
    "Unit": "USD"
  },
  "TimeUnit": "MONTHLY",
  "BudgetType": "COST",
  "NotificationsWithSubscribers": [
    {
      "Notification": {
        "NotificationType": "ACTUAL",
        "ComparisonOperator": "GREATER_THAN",
        "Threshold": 80
      },
      "Subscribers": [
        {
          "SubscriptionType": "EMAIL",
          "Address": "admin@example.com"
        }
      ]
    }
  ]
}
```

---

## 13. Complete Pricing Tables

### 13.1 Compute Pricing Comparison

| Provider | Service | Instance | vCPU | RAM | Price/Month | Notes |
|----------|---------|----------|------|-----|-------------|-------|
| AWS | EC2 | t3.micro | 2 | 1GB | $8.47 | Burstable |
| AWS | EC2 | t3.small | 2 | 2GB | $16.94 | Burstable |
| AWS | EC2 | t3.medium | 2 | 4GB | $33.87 | Burstable |
| AWS | EC2 | t3.large | 2 | 8GB | $67.74 | Burstable |
| AWS | ECS Fargate | 1 vCPU + 2GB | 1 | 2GB | ~$35 | Serverless |
| AWS | Lambda | - | - | - | $0.20/1M requests | Pay per use |
| Azure | B2s | 2 | 4GB | $28.38 | Dev/Test |
| Azure | D2s v3 | 2 | 8GB | $70.08 | General |
| GCP | e2-micro | 2 | 1GB | $6.11 | Free tier eligible |
| GCP | e2-small | 2 | 2GB | $12.23 | Burstable |
| DigitalOcean | Basic | 1 | 1GB | $6.00 | Fixed price |
| DigitalOcean | Basic | 2 | 4GB | $24.00 | Fixed price |
| Linode | Nanode | 1 | 1GB | $5.00 | Fixed price |
| Linode | Linode 2GB | 1 | 2GB | $10.00 | Fixed price |
| Vultr | Cloud Compute | 1 | 1GB | $5.00 | Fixed price |
| Vultr | Cloud Compute | 2 | 4GB | $20.00 | Fixed price |
| Heroku | Hobby | 1 | 512MB | $7.00 | Dyno |
| Heroku | Standard 1X | 1 | 512MB | $25.00 | Dyno |
| Railway | Starter | Shared | Shared | $5.00 | Managed |
| Railway | Pro | 2 | 2GB | $20.00 | Managed |
| Render | Starter | 0.5 | 512MB | $7.00 | Container |
| Render | Standard | 2 | 2GB | $25.00 | Container |

### 13.2 Database Pricing Comparison

| Provider | Service | Plan | RAM | Storage | Price/Month |
|----------|---------|------|-----|---------|-------------|
| MongoDB Atlas | M0 | Free | Shared | 512MB | $0 |
| MongoDB Atlas | M2 | Shared | Shared | 2GB | $9 |
| MongoDB Atlas | M5 | Shared | Shared | 5GB | $25 |
| MongoDB Atlas | M10 | Dedicated | 2GB | 10GB | $57 |
| MongoDB Atlas | M20 | Dedicated | 4GB | 20GB | $115 |
| MongoDB Atlas | M30 | Dedicated | 8GB | 40GB | $230 |
| AWS | DocumentDB | db.t3.medium | 4GB | 100GB | ~$140 |
| AWS | RDS MySQL | db.t3.micro | 1GB | 20GB | ~$13 |
| AWS | RDS PostgreSQL | db.t3.micro | 1GB | 20GB | ~$13 |
| AWS | RDS PostgreSQL | db.t3.small | 2GB | 20GB | ~$26 |
| DigitalOcean | Managed MongoDB | Basic | 2GB | 25GB | $15 |
| DigitalOcean | Managed MySQL | Basic | 1GB | 10GB | $15 |
| ScaleGrid | MongoDB | Dedicated | 2GB | 10GB | $27 |
| Aiven | PostgreSQL | Hobbyist | Shared | 5GB | $0 |
| Aiven | PostgreSQL | Startup-4 | 4GB | 32GB | $167 |
| PlanetScale | MySQL | Free | Shared | 5GB | $0 |
| PlanetScale | MySQL | Scaler | 1GB | 10GB | $29 |
| Supabase | PostgreSQL | Free | Shared | 500MB | $0 |
| Supabase | PostgreSQL | Pro | Shared | 8GB | $25 |

### 13.3 Storage Pricing Comparison

| Provider | Service | Storage/GB | GET/1K | PUT/1K | Egress/GB | Notes |
|----------|---------|------------|--------|--------|-----------|-------|
| Cloudflare R2 | Object Storage | $0.015 | $0.00036 | $0.0045 | $0.00 | No egress fees |
| AWS S3 | Standard | $0.023 | $0.0004 | $0.005 | $0.09 | Most popular |
| AWS S3 | Standard-IA | $0.0125 | $0.001 | $0.01 | $0.09 | Infrequent access |
| AWS S3 | One Zone-IA | $0.01 | $0.001 | $0.01 | $0.09 | Single AZ |
| AWS S3 | Glacier | $0.004 | $0.01 | $0.1 | $0.09 | Archive |
| AWS S3 | Deep Archive | $0.00099 | $0.025 | $0.18 | $0.09 | Long-term |
| Google Cloud | Standard | $0.020 | $0.0004 | $0.005 | $0.12 | Multi-regional |
| Google Cloud | Nearline | $0.010 | $0.001 | $0.01 | $0.12 | Infrequent |
| Google Cloud | Coldline | $0.004 | $0.005 | $0.02 | $0.12 | Rare access |
| Google Cloud | Archive | $0.0012 | $0.05 | $0.05 | $0.12 | Backup only |
| Azure Blob | Hot | $0.0184 | $0.0005 | $0.005 | $0.087 | Frequent access |
| Azure Blob | Cool | $0.0100 | $0.001 | $0.01 | $0.087 | 30+ days |
| Azure Blob | Archive | $0.00099 | $0.022 | $0.10 | $0.087 | 180+ days |
| DigitalOcean | Spaces | $0.020 | Included | Included | $0.00 | First 250GB |
| Backblaze B2 | Object Storage | $0.005 | $0.0004 | Included | $0.00 | First 1GB/day free |
| Wasabi | Hot Storage | $0.0059 | Included | Included | $0.00 | 90-day minimum |

### 13.4 CDN Pricing Comparison

| Provider | Bandwidth/GB | Requests/10K | Minimum | Features |
|----------|--------------|--------------|---------|----------|
| Cloudflare | $0.00 | $0.00 | $0.00 | Unlimited free tier |
| Cloudflare Pro | $20.00 flat | Unlimited | $20.00 | Advanced features |
| Cloudflare Business | $200.00 flat | Unlimited | $200.00 | Enterprise features |
| AWS CloudFront | $0.085-0.020 | $0.0075 | $0.00 | AWS integration |
| Google Cloud CDN | $0.08-0.02 | $0.0075 | $0.00 | GCP integration |
| Azure CDN | $0.087-0.037 | Included | $0.00 | Azure integration |
| Fastly | $0.12 | Included | $50.00 | Real-time config |
| KeyCDN | $0.04 | Included | $4.00 | Pay-as-you-go |
| BunnyCDN | $0.01-0.005 | Included | $0.00 | Volume discounts |
| StackPath | $0.06 | Included | $10.00 | Security focus |
| CDN77 | $0.049 | Included | $0.00 | No minimums |
| BelugaCDN | $0.02 | Included | $5.00 | Transparent pricing |

### 13.5 Email Service Pricing Comparison

| Provider | Free Tier | 10K/Month | 50K/Month | 100K/Month | 500K/Month | Notes |
|----------|-----------|-----------|-----------|------------|------------|-------|
| AWS SES | 62K (from EC2) | $1.00 | $5.00 | $10.00 | $50.00 | Cheapest option |
| SendGrid | 3K/month | $19.95 | $19.95 | $34.95 | $199.95 | Good deliverability |
| Mailgun | 5K (3mo trial) | $35.00 | $35.00 | $80.00 | $350.00 | Developer friendly |
| Postmark | 100/month | $10.00 | $50.00 | $100.00 | $400.00 | Fast delivery |
| Mailjet | 6K/month | $15.00 | $15.00 | $25.00 | $95.00 | Good for EU |
| Sendinblue | 300/day | $25.00 | $25.00 | $39.00 | $159.00 | Marketing + trans |
| Mandrill (Mailchimp) | None | $20.00 | $20.00 | $40.00 | $200.00 | Mailchimp integration |
| Elastic Email | 100/day | $10.00 | $10.00 | $20.00 | $90.00 | Budget option |
| SparkPost | 15K/month | $20.00 | $20.00 | $30.00 | $75.00 | High volume |
| MessageBird | 500/day | $0.00 | $50.00 | $50.00 | $300.00 | Global reach |
| Resend | 3K/month | $20.00 | $20.00 | $20.00 | $90.00 | Developer focused |

### 13.6 Monitoring Tools Pricing

| Provider | Free Tier | Starter | Pro | Enterprise | Notes |
|----------|-----------|---------|-----|------------|-------|
| Datadog | 5 hosts | - | $15/host | Custom | Full observability |
| New Relic | 100GB data | - | $49/month | Custom | APM leader |
| Dynatrace | 15 days | - | $0.08/host/hr | Custom | AI-powered |
| AppDynamics | Trial | - | - | Custom | Enterprise focus |
| Instana | 14 days | - | $75/host | Custom | Auto-discovery |
| Splunk | 500MB/day | - | Custom | Custom | Log analysis |
| LogDNA | 10GB/day | - | $3/GB | Custom | Log management |
| Sentry | 5K errors | - | $26/month | Custom | Error tracking |
| Rollbar | 5K errors | - | $49/month | Custom | Error tracking |
| Bugsnag | 1.5K events | - | $59/month | Custom | Error tracking |
| PagerDuty | 5 users | - | $29/user | Custom | Incident management |
| Opsgenie | 5 users | - | $9/user | Custom | Atlassian |
| Pingdom | - | $15/month | $45/month | Custom | Uptime |
| UptimeRobot | 50 monitors | - | $8/month | $34/month | Simple |
| StatusCake | - | - | $25/month | $150/month | Uptime + pages |

### 13.7 Complete Cost Summary Table

**Small Scale (1-100 weddings):**

| Component | Cheapest Option | Mid-Range | Premium | Notes |
|-----------|-----------------|-----------|---------|-------|
| Compute | $8 (EC2 t3.micro) | $25 (DO) | $50 (Heroku) | Start small |
| Database | $0 (Atlas M0) | $60 (Atlas M10) | $80 (Atlas M20) | Free tier works |
| Storage | $0 (R2) | $5 (R2) | $20 (S3) | R2 recommended |
| CDN | $0 (Cloudflare) | $0 (Cloudflare) | $20 (Cloudflare Pro) | Free is sufficient |
| Email | $0 (SES) | $0 (SendGrid Free) | $20 (SendGrid) | Free tier covers |
| Domain/SSL | $10/year | $15/year | $50/year | Shop around |
| Monitoring | $0 | $15 | $30 | Free tier fine |
| **Total/Month** | **$9** | **$120** | **$190** | |
| **Total/Year** | **$118** | **$1,440** | **$2,290** | |

**Medium Scale (100-1,000 weddings):**

| Component | Cheapest Option | Mid-Range | Premium | Notes |
|-----------|-----------------|-----------|---------|-------|
| Compute | $25 (EC2 t3.small) | $50 (EC2 t3.medium) | $100 (2x t3.medium) | Monitor CPU |
| Database | $60 (Atlas M10) | $115 (Atlas M20) | $150 (Atlas M30) | Plan for growth |
| Storage | $5 (R2) | $20 (R2) | $50 (R2) | R2 wins |
| CDN | $0 (Cloudflare) | $20 (Cloudflare Pro) | $50 (Cloudflare) | Free if <200GB |
| Email | $0 (SES) | $20 (SendGrid) | $50 (SendGrid) | Track volume |
| Domain/SSL | $20/year | $50/year | $100/year | Multiple domains |
| Monitoring | $15 | $30 | $50 | Worth investing |
| **Total/Month** | **$110** | **$255** | **$370** | |
| **Total/Year** | **$1,340** | **$3,110** | **$4,540** | |

**Large Scale (1,000+ weddings):**

| Component | Cheapest Option | Mid-Range | Premium | Notes |
|-----------|-----------------|-----------|---------|-------|
| Compute | $50 (EC2 t3.medium) | $150 (t3.large + auto-scaling) | $300 (multi-AZ) | Auto-scale |
| Database | $115 (Atlas M20) | $230 (Atlas M30) | $500 (Atlas M50) | Replica sets |
| Storage | $20 (R2) | $100 (R2) | $200 (R2+S3) | Multi-region |
| CDN | $20 (Cloudflare) | $50 (Cloudflare) | $100 (Enterprise) | Essential |
| Email | $20 (SendGrid) | $50 (SendGrid) | $100 (dedicated IP) | Reputation |
| Domain/SSL | $50/year | $100/year | $200/year | Wildcards |
| Monitoring | $30 | $75 | $150 | Don't skimp |
| **Total/Month** | **$275** | **$655** | **$1,200** | |
| **Total/Year** | **$3,350** | **$7,960** | **$14,600** | |

---

## Quick Reference: Cost-Saving Tips

1. **Start with free tiers** - AWS free tier, MongoDB M0, Cloudflare Free
2. **Use Cloudflare R2** - No egress fees saves significant money
3. **Compress images** - WebP format, 60% smaller than JPEG
4. **Right-size instances** - Monitor and adjust based on actual usage
5. **Use reserved instances** - 30-60% savings for 1-3 year commitments
6. **Enable caching everywhere** - Browser, CDN, application layer
7. **Monitor costs weekly** - Set up alerts at 50%, 80%, 100%
8. **Review unused resources** - Terminate idle instances monthly
9. **Negotiate at scale** - AWS Enterprise, Cloudflare Enterprise
10. **Consider alternatives** - DigitalOcean, Linode for predictable pricing

## Cost Checklist by Stage

### Pre-Launch (Months 1-3)
- [ ] Set up AWS free tier account
- [ ] Create MongoDB Atlas M0 cluster
- [ ] Configure Cloudflare free CDN
- [ ] Use SendGrid free tier (100/day)
- [ ] Register domain ($10-15/year)
- [ ] Set up CloudWatch basic monitoring
- [ ] **Target Budget: $30/month**

### Launch (Months 4-6)
- [ ] Upgrade to t3.small ($17/month)
- [ ] Move to Atlas M5 ($25/month)
- [ ] Add S3 for file storage ($5/month)
- [ ] Upgrade SendGrid if needed ($20/month)
- [ ] Add Sentry for error tracking ($26/month)
- [ ] Set up AWS Budget alerts
- [ ] **Target Budget: $100/month**

### Growth (Months 7-12)
- [ ] Consider auto-scaling setup
- [ ] Upgrade to t3.medium ($34/month)
- [ ] Move to Atlas M10 ($57/month)
- [ ] Implement Cloudflare Pro ($20/month)
- [ ] Add advanced monitoring ($30/month)
- [ ] Review and optimize storage classes
- [ ] **Target Budget: $250/month**

### Scale (Year 2+)
- [ ] Multi-region deployment
- [ ] Reserved instance purchases
- [ ] Enterprise support plans
- [ ] Dedicated email IPs
- [ ] Custom monitoring stack
- [ ] Quarterly cost reviews
- [ ] **Target Budget: $500+/month**

---

**Document Version:** 1.0  
**Last Updated:** 2026-02-03  
**Maintained By:** DevOps Team  
**Next Review:** Quarterly
