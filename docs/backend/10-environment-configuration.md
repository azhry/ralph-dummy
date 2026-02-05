# Environment Configuration Guide

## Table of Contents
1. [Configuration Overview](#1-configuration-overview)
2. [Environment Variables Reference](#2-environment-variables-reference)
3. [Configuration Management with Viper](#3-configuration-management-with-viper)
4. [Environment Files](#4-environment-files)
5. [Secrets Management](#5-secrets-management)
6. [Configuration Validation](#6-configuration-validation)
7. [Environment-Specific Behaviors](#7-environment-specific-behaviors)
8. [Code Examples](#8-code-examples)
9. [Security Best Practices](#9-security-best-practices)
10. [Troubleshooting](#10-troubleshooting)

---

## 1. Configuration Overview

### 12-Factor App Methodology

Our configuration system follows the 12-Factor App methodology for configuration management:

1. **Store config in environment variables** - Never hardcode credentials or settings
2. **Strict separation of config across deploys** - Same code, different configs
3. **No config in version control** - All secrets and environment-specific values excluded
4. **No group logic by env** - No "dev" or "production" conditionals in code

### Configuration Hierarchy

```
1. Environment Variables (highest priority)
2. .env file (local development only)
3. Viper defaults
4. Application defaults (lowest priority)
```

### Supported Environments

| Environment | Description | Typical Deploy Target |
|-------------|-------------|----------------------|
| `development` | Local development | Developer workstation |
| `test` | Automated testing | CI/CD pipelines |
| `staging` | Pre-production testing | Staging servers |
| `production` | Live production | Production servers |

---

## 2. Environment Variables Reference

### Server Settings

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `APP_ENV` | string | Yes | `development` | Application environment |
| `APP_NAME` | string | No | `wedding-invitation-api` | Application identifier |
| `PORT` | int | No | `8080` | HTTP server port |
| `HOST` | string | No | `0.0.0.0` | Server bind address |
| `DEBUG` | bool | No | `false` | Enable debug mode |
| `LOG_LEVEL` | string | No | `info` | Log level (debug, info, warn, error) |
| `REQUEST_TIMEOUT` | duration | No | `30s` | HTTP request timeout |
| `SHUTDOWN_TIMEOUT` | duration | No | `10s` | Graceful shutdown timeout |

### Database Configuration

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `MONGODB_URI` | string | Yes | - | MongoDB connection string |
| `MONGODB_DATABASE` | string | Yes | - | Database name |
| `MONGODB_MAX_POOL_SIZE` | int | No | `100` | Connection pool max size |
| `MONGODB_MIN_POOL_SIZE` | int | No | `10` | Connection pool min size |
| `MONGODB_CONNECT_TIMEOUT` | duration | No | `10s` | Connection timeout |
| `MONGODB_SOCKET_TIMEOUT` | duration | No | `0` | Socket timeout (0 = no timeout) |
| `MONGODB_RETRY_WRITES` | bool | No | `true` | Enable retryable writes |
| `MONGODB_MAX_COMMIT_TIME` | duration | No | `5s` | Max transaction commit time |

**Example MongoDB URI Formats:**
```bash
# Local development
MONGODB_URI=mongodb://localhost:27017

# Replica set
MONGODB_URI=mongodb://user:pass@host1:27017,host2:27017,host3:27017/database?replicaSet=rs0

# MongoDB Atlas
MONGODB_URI=mongodb+srv://user:pass@cluster.mongodb.net/database?retryWrites=true&w=majority
```

### Authentication Configuration

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `JWT_SECRET` | string | Yes* | - | JWT signing secret (min 32 chars) |
| `JWT_REFRESH_SECRET` | string | Yes* | - | Refresh token signing secret |
| `JWT_ACCESS_EXPIRY` | duration | No | `15m` | Access token expiration |
| `JWT_REFRESH_EXPIRY` | duration | No | `7d` | Refresh token expiration |
| `JWT_ISSUER` | string | No | `wedding-invitation-api` | JWT issuer claim |
| `BCRYPT_COST` | int | No | `12` | Password hashing cost |

**Note:** `JWT_SECRET` and `JWT_REFRESH_SECRET` are required in production.

### Storage Configuration

#### AWS S3

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `AWS_ACCESS_KEY_ID` | string | Yes* | - | AWS access key |
| `AWS_SECRET_ACCESS_KEY` | string | Yes* | - | AWS secret key |
| `AWS_REGION` | string | Yes* | - | AWS region (e.g., us-east-1) |
| `AWS_S3_BUCKET` | string | Yes* | - | S3 bucket name |
| `AWS_S3_ENDPOINT` | string | No | - | Custom endpoint (for MinIO, etc.) |
| `AWS_S3_FORCE_PATH_STYLE` | bool | No | `false` | Use path-style URLs |

#### Cloudflare R2

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `R2_ACCOUNT_ID` | string | Yes* | - | R2 account ID |
| `R2_ACCESS_KEY_ID` | string | Yes* | - | R2 access key |
| `R2_SECRET_ACCESS_KEY` | string | Yes* | - | R2 secret key |
| `R2_BUCKET` | string | Yes* | - | R2 bucket name |
| `R2_PUBLIC_URL` | string | No | - | Custom public URL for assets |

**Note:** S3 and R2 credentials are mutually exclusive - configure one or the other.

### Email Configuration

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `SENDGRID_API_KEY` | string | Yes* | - | SendGrid API key |
| `SMTP_HOST` | string | Yes* | - | SMTP server host |
| `SMTP_PORT` | int | Yes* | `587` | SMTP server port |
| `SMTP_USER` | string | Yes* | - | SMTP username |
| `SMTP_PASSWORD` | string | Yes* | - | SMTP password |
| `SMTP_TLS` | bool | No | `true` | Enable TLS/SSL |
| `EMAIL_FROM` | string | Yes | - | Default from address |
| `EMAIL_FROM_NAME` | string | No | `Wedding Invitation` | Default from name |

**Note:** Configure either SendGrid OR SMTP, not both.

### Security Configuration

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `CORS_ALLOWED_ORIGINS` | string | Yes | `*` | Comma-separated allowed origins |
| `CORS_ALLOWED_METHODS` | string | No | `GET,POST,PUT,DELETE,OPTIONS` | Allowed HTTP methods |
| `CORS_ALLOWED_HEADERS` | string | No | `*` | Allowed headers |
| `CORS_MAX_AGE` | int | No | `86400` | Preflight cache time (seconds) |
| `RATE_LIMIT_REQUESTS` | int | No | `100` | Requests per window |
| `RATE_LIMIT_WINDOW` | duration | No | `1m` | Rate limit window |
| `RATE_LIMIT_BURST` | int | No | `10` | Burst capacity |
| `TRUSTED_PROXIES` | string | No | - | Comma-separated trusted proxy IPs |
| `SECURE_HEADERS` | bool | No | `true` | Enable security headers |
| `HSTS_MAX_AGE` | int | No | `31536000` | HSTS max-age (seconds) |

### Monitoring & Observability

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `METRICS_ENABLED` | bool | No | `true` | Enable Prometheus metrics |
| `METRICS_PORT` | int | No | `9090` | Metrics server port |
| `METRICS_PATH` | string | No | `/metrics` | Metrics endpoint path |
| `TRACING_ENABLED` | bool | No | `false` | Enable distributed tracing |
| `TRACING_SERVICE_NAME` | string | No | `wedding-api` | Tracing service name |
| `TRACING_ENDPOINT` | string | No | - | Jaeger/OTLP endpoint |
| `LOG_FORMAT` | string | No | `json` | Log format (json, console) |
| `LOG_OUTPUT` | string | No | `stdout` | Log output (stdout, stderr, file) |
| `LOG_FILE_PATH` | string | No | - | Log file path (if LOG_OUTPUT=file) |

### Feature Flags

| Variable | Type | Required | Default | Description |
|----------|------|----------|---------|-------------|
| `FEATURE_REGISTRATION` | bool | No | `true` | Enable user registration |
| `FEATURE_INVITATION_EXPORT` | bool | No | `true` | Enable invitation exports |
| `FEATURE_ANALYTICS` | bool | No | `false` | Enable analytics collection |
| `FEATURE_BULK_IMPORT` | bool | No | `false` | Enable bulk import APIs |

---

## 3. Configuration Management with Viper

### Why Viper?

Viper is the premier configuration solution for Go applications because it provides:

- **Multiple sources**: Environment variables, config files, flags, remote systems
- **Dynamic reloading**: Watch for config changes without restart
- **Type safety**: Automatic type conversion with validation
- **Nesting**: Support for complex nested configurations
- **Environment mapping**: Automatic env var binding

### Installation

```bash
go get github.com/spf13/viper
```

### Complete Configuration Implementation

```go
// config/config.go
package config

import (
    "errors"
    "fmt"
    "os"
    "strings"
    "sync"
    "time"

    "github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
    sync.RWMutex
    
    // Server configuration
    Server ServerConfig `mapstructure:"server"`
    
    // Database configuration
    Database DatabaseConfig `mapstructure:"database"`
    
    // Authentication configuration
    Auth AuthConfig `mapstructure:"auth"`
    
    // Storage configuration
    Storage StorageConfig `mapstructure:"storage"`
    
    // Email configuration
    Email EmailConfig `mapstructure:"email"`
    
    // Security configuration
    Security SecurityConfig `mapstructure:"security"`
    
    // Monitoring configuration
    Monitoring MonitoringConfig `mapstructure:"monitoring"`
    
    // Feature flags
    Features FeatureConfig `mapstructure:"features"`
    
    // Raw values for direct access
    raw map[string]interface{}
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
    Environment     string        `mapstructure:"environment"`
    Name           string        `mapstructure:"name"`
    Port           int           `mapstructure:"port"`
    Host           string        `mapstructure:"host"`
    Debug          bool          `mapstructure:"debug"`
    LogLevel       string        `mapstructure:"log_level"`
    RequestTimeout time.Duration `mapstructure:"request_timeout"`
    ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// DatabaseConfig holds MongoDB settings
type DatabaseConfig struct {
    URI            string        `mapstructure:"uri"`
    Database       string        `mapstructure:"database"`
    MaxPoolSize    int           `mapstructure:"max_pool_size"`
    MinPoolSize    int           `mapstructure:"min_pool_size"`
    ConnectTimeout time.Duration `mapstructure:"connect_timeout"`
    SocketTimeout  time.Duration `mapstructure:"socket_timeout"`
    RetryWrites    bool          `mapstructure:"retry_writes"`
    MaxCommitTime  time.Duration `mapstructure:"max_commit_time"`
}

// AuthConfig holds authentication settings
type AuthConfig struct {
    JWTSecret        string        `mapstructure:"jwt_secret"`
    JWTRefreshSecret string        `mapstructure:"jwt_refresh_secret"`
    AccessExpiry     time.Duration `mapstructure:"access_expiry"`
    RefreshExpiry    time.Duration `mapstructure:"refresh_expiry"`
    Issuer           string        `mapstructure:"issuer"`
    BcryptCost       int           `mapstructure:"bcrypt_cost"`
}

// StorageConfig holds file storage settings
type StorageConfig struct {
    Provider    string `mapstructure:"provider"` // s3, r2, local
    S3          S3Config `mapstructure:"s3"`
    R2          R2Config `mapstructure:"r2"`
    Local       LocalStorageConfig `mapstructure:"local"`
}

type S3Config struct {
    AccessKeyID     string `mapstructure:"access_key_id"`
    SecretAccessKey string `mapstructure:"secret_access_key"`
    Region          string `mapstructure:"region"`
    Bucket          string `mapstructure:"bucket"`
    Endpoint        string `mapstructure:"endpoint"`
    ForcePathStyle  bool   `mapstructure:"force_path_style"`
}

type R2Config struct {
    AccountID       string `mapstructure:"account_id"`
    AccessKeyID     string `mapstructure:"access_key_id"`
    SecretAccessKey string `mapstructure:"secret_access_key"`
    Bucket          string `mapstructure:"bucket"`
    PublicURL       string `mapstructure:"public_url"`
}

type LocalStorageConfig struct {
    BasePath string `mapstructure:"base_path"`
    BaseURL  string `mapstructure:"base_url"`
}

// EmailConfig holds email settings
type EmailConfig struct {
    Provider string    `mapstructure:"provider"` // sendgrid, smtp
    Sendgrid SendgridConfig `mapstructure:"sendgrid"`
    SMTP     SMTPConfig     `mapstructure:"smtp"`
    From     string         `mapstructure:"from"`
    FromName string         `mapstructure:"from_name"`
}

type SendgridConfig struct {
    APIKey string `mapstructure:"api_key"`
}

type SMTPConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    Username string `mapstructure:"username"`
    Password string `mapstructure:"password"`
    TLS      bool   `mapstructure:"tls"`
}

// SecurityConfig holds security-related settings
type SecurityConfig struct {
    CorsAllowedOrigins string        `mapstructure:"cors_allowed_origins"`
    CorsAllowedMethods string        `mapstructure:"cors_allowed_methods"`
    CorsAllowedHeaders string        `mapstructure:"cors_allowed_headers"`
    CorsMaxAge         int           `mapstructure:"cors_max_age"`
    RateLimitRequests  int           `mapstructure:"rate_limit_requests"`
    RateLimitWindow    time.Duration `mapstructure:"rate_limit_window"`
    RateLimitBurst     int           `mapstructure:"rate_limit_burst"`
    TrustedProxies     []string      `mapstructure:"trusted_proxies"`
    SecureHeaders      bool          `mapstructure:"secure_headers"`
    HSTMaxAge          int           `mapstructure:"hsts_max_age"`
}

// MonitoringConfig holds observability settings
type MonitoringConfig struct {
    MetricsEnabled      bool   `mapstructure:"metrics_enabled"`
    MetricsPort         int    `mapstructure:"metrics_port"`
    MetricsPath         string `mapstructure:"metrics_path"`
    TracingEnabled      bool   `mapstructure:"tracing_enabled"`
    TracingServiceName  string `mapstructure:"tracing_service_name"`
    TracingEndpoint     string `mapstructure:"tracing_endpoint"`
    LogFormat           string `mapstructure:"log_format"`
    LogOutput           string `mapstructure:"log_output"`
    LogFilePath         string `mapstructure:"log_file_path"`
}

// FeatureConfig holds feature flags
type FeatureConfig struct {
    Registration     bool `mapstructure:"registration"`
    InvitationExport bool `mapstructure:"invitation_export"`
    Analytics        bool `mapstructure:"analytics"`
    BulkImport       bool `mapstructure:"bulk_import"`
}

var (
    // Global configuration instance
    globalConfig *Config
    configOnce   sync.Once
)

// Load initializes and loads configuration
func Load() (*Config, error) {
    var loadErr error
    configOnce.Do(func() {
        globalConfig = &Config{}
        loadErr = globalConfig.load()
    })
    
    if loadErr != nil {
        return nil, loadErr
    }
    
    return globalConfig, nil
}

// Get returns the current configuration
func Get() *Config {
    if globalConfig == nil {
        panic("configuration not loaded. Call Load() first")
    }
    return globalConfig
}

// Reload reloads configuration (useful for development)
func Reload() error {
    if globalConfig == nil {
        return errors.New("configuration not initialized")
    }
    return globalConfig.load()
}

// load performs the actual configuration loading
func (c *Config) load() error {
    c.Lock()
    defer c.Unlock()
    
    // Initialize Viper
    v := viper.New()
    
    // Set config name and paths
    v.SetConfigName("config")
    v.SetConfigType("yaml")
    v.AddConfigPath(".")
    v.AddConfigPath("./config")
    v.AddConfigPath("/etc/wedding-api/")
    
    // Set defaults
    c.setDefaults(v)
    
    // Try to read config file (optional)
    if err := v.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return fmt.Errorf("failed to read config file: %w", err)
        }
        // Config file not found is OK, we'll use env vars
    }
    
    // Enable environment variable support
    v.SetEnvPrefix("WEDDING") // Will look for WEDDING_* variables
    v.AutomaticEnv()
    
    // Set up environment variable mappings
    c.setupEnvMappings(v)
    
    // Load from .env file in development
    if v.GetString("app_env") == "development" || v.GetString("app_env") == "" {
        _ = v.ReadInConfig() // Already tried above
        v.SetEnvPrefix("") // Allow direct env vars in dev
    }
    
    // Unmarshal configuration
    if err := v.Unmarshal(c); err != nil {
        return fmt.Errorf("failed to unmarshal config: %w", err)
    }
    
    // Store raw values
    c.raw = v.AllSettings()
    
    // Validate configuration
    if err := c.Validate(); err != nil {
        return fmt.Errorf("configuration validation failed: %w", err)
    }
    
    return nil
}

// setDefaults sets default configuration values
func (c *Config) setDefaults(v *viper.Viper) {
    // Server defaults
    v.SetDefault("server.environment", "development")
    v.SetDefault("server.name", "wedding-invitation-api")
    v.SetDefault("server.port", 8080)
    v.SetDefault("server.host", "0.0.0.0")
    v.SetDefault("server.debug", false)
    v.SetDefault("server.log_level", "info")
    v.SetDefault("server.request_timeout", "30s")
    v.SetDefault("server.shutdown_timeout", "10s")
    
    // Database defaults
    v.SetDefault("database.max_pool_size", 100)
    v.SetDefault("database.min_pool_size", 10)
    v.SetDefault("database.connect_timeout", "10s")
    v.SetDefault("database.socket_timeout", "0s")
    v.SetDefault("database.retry_writes", true)
    v.SetDefault("database.max_commit_time", "5s")
    
    // Auth defaults
    v.SetDefault("auth.access_expiry", "15m")
    v.SetDefault("auth.refresh_expiry", "168h") // 7 days
    v.SetDefault("auth.issuer", "wedding-invitation-api")
    v.SetDefault("auth.bcrypt_cost", 12)
    
    // Storage defaults
    v.SetDefault("storage.provider", "local")
    
    // Email defaults
    v.SetDefault("email.from_name", "Wedding Invitation")
    
    // Security defaults
    v.SetDefault("security.cors_allowed_origins", "*")
    v.SetDefault("security.cors_allowed_methods", "GET,POST,PUT,DELETE,OPTIONS")
    v.SetDefault("security.cors_allowed_headers", "*")
    v.SetDefault("security.cors_max_age", 86400)
    v.SetDefault("security.rate_limit_requests", 100)
    v.SetDefault("security.rate_limit_window", "1m")
    v.SetDefault("security.rate_limit_burst", 10)
    v.SetDefault("security.secure_headers", true)
    v.SetDefault("security.hsts_max_age", 31536000)
    
    // Monitoring defaults
    v.SetDefault("monitoring.metrics_enabled", true)
    v.SetDefault("monitoring.metrics_port", 9090)
    v.SetDefault("monitoring.metrics_path", "/metrics")
    v.SetDefault("monitoring.tracing_enabled", false)
    v.SetDefault("monitoring.tracing_service_name", "wedding-api")
    v.SetDefault("monitoring.log_format", "json")
    v.SetDefault("monitoring.log_output", "stdout")
    
    // Feature defaults
    v.SetDefault("features.registration", true)
    v.SetDefault("features.invitation_export", true)
    v.SetDefault("features.analytics", false)
    v.SetDefault("features.bulk_import", false)
}

// setupEnvMappings configures environment variable bindings
func (c *Config) setupEnvMappings(v *viper.Viper) {
    // Server
    v.BindEnv("server.environment", "APP_ENV")
    v.BindEnv("server.name", "APP_NAME")
    v.BindEnv("server.port", "PORT")
    v.BindEnv("server.host", "HOST")
    v.BindEnv("server.debug", "DEBUG")
    v.BindEnv("server.log_level", "LOG_LEVEL")
    v.BindEnv("server.request_timeout", "REQUEST_TIMEOUT")
    v.BindEnv("server.shutdown_timeout", "SHUTDOWN_TIMEOUT")
    
    // Database
    v.BindEnv("database.uri", "MONGODB_URI")
    v.BindEnv("database.database", "MONGODB_DATABASE")
    v.BindEnv("database.max_pool_size", "MONGODB_MAX_POOL_SIZE")
    v.BindEnv("database.min_pool_size", "MONGODB_MIN_POOL_SIZE")
    v.BindEnv("database.connect_timeout", "MONGODB_CONNECT_TIMEOUT")
    v.BindEnv("database.socket_timeout", "MONGODB_SOCKET_TIMEOUT")
    v.BindEnv("database.retry_writes", "MONGODB_RETRY_WRITES")
    v.BindEnv("database.max_commit_time", "MONGODB_MAX_COMMIT_TIME")
    
    // Auth
    v.BindEnv("auth.jwt_secret", "JWT_SECRET")
    v.BindEnv("auth.jwt_refresh_secret", "JWT_REFRESH_SECRET")
    v.BindEnv("auth.access_expiry", "JWT_ACCESS_EXPIRY")
    v.BindEnv("auth.refresh_expiry", "JWT_REFRESH_EXPIRY")
    v.BindEnv("auth.issuer", "JWT_ISSUER")
    v.BindEnv("auth.bcrypt_cost", "BCRYPT_COST")
    
    // Storage - AWS S3
    v.BindEnv("storage.provider", "STORAGE_PROVIDER")
    v.BindEnv("storage.s3.access_key_id", "AWS_ACCESS_KEY_ID")
    v.BindEnv("storage.s3.secret_access_key", "AWS_SECRET_ACCESS_KEY")
    v.BindEnv("storage.s3.region", "AWS_REGION")
    v.BindEnv("storage.s3.bucket", "AWS_S3_BUCKET")
    v.BindEnv("storage.s3.endpoint", "AWS_S3_ENDPOINT")
    v.BindEnv("storage.s3.force_path_style", "AWS_S3_FORCE_PATH_STYLE")
    
    // Storage - R2
    v.BindEnv("storage.r2.account_id", "R2_ACCOUNT_ID")
    v.BindEnv("storage.r2.access_key_id", "R2_ACCESS_KEY_ID")
    v.BindEnv("storage.r2.secret_access_key", "R2_SECRET_ACCESS_KEY")
    v.BindEnv("storage.r2.bucket", "R2_BUCKET")
    v.BindEnv("storage.r2.public_url", "R2_PUBLIC_URL")
    
    // Storage - Local
    v.BindEnv("storage.local.base_path", "STORAGE_BASE_PATH")
    v.BindEnv("storage.local.base_url", "STORAGE_BASE_URL")
    
    // Email - Sendgrid
    v.BindEnv("email.provider", "EMAIL_PROVIDER")
    v.BindEnv("email.sendgrid.api_key", "SENDGRID_API_KEY")
    
    // Email - SMTP
    v.BindEnv("email.smtp.host", "SMTP_HOST")
    v.BindEnv("email.smtp.port", "SMTP_PORT")
    v.BindEnv("email.smtp.username", "SMTP_USER")
    v.BindEnv("email.smtp.password", "SMTP_PASSWORD")
    v.BindEnv("email.smtp.tls", "SMTP_TLS")
    v.BindEnv("email.from", "EMAIL_FROM")
    v.BindEnv("email.from_name", "EMAIL_FROM_NAME")
    
    // Security
    v.BindEnv("security.cors_allowed_origins", "CORS_ALLOWED_ORIGINS")
    v.BindEnv("security.cors_allowed_methods", "CORS_ALLOWED_METHODS")
    v.BindEnv("security.cors_allowed_headers", "CORS_ALLOWED_HEADERS")
    v.BindEnv("security.cors_max_age", "CORS_MAX_AGE")
    v.BindEnv("security.rate_limit_requests", "RATE_LIMIT_REQUESTS")
    v.BindEnv("security.rate_limit_window", "RATE_LIMIT_WINDOW")
    v.BindEnv("security.rate_limit_burst", "RATE_LIMIT_BURST")
    v.BindEnv("security.secure_headers", "SECURE_HEADERS")
    v.BindEnv("security.hsts_max_age", "HSTS_MAX_AGE")
    
    // Parse trusted proxies from comma-separated string
    if proxies := os.Getenv("TRUSTED_PROXIES"); proxies != "" {
        v.Set("security.trusted_proxies", strings.Split(proxies, ","))
    }
    
    // Monitoring
    v.BindEnv("monitoring.metrics_enabled", "METRICS_ENABLED")
    v.BindEnv("monitoring.metrics_port", "METRICS_PORT")
    v.BindEnv("monitoring.metrics_path", "METRICS_PATH")
    v.BindEnv("monitoring.tracing_enabled", "TRACING_ENABLED")
    v.BindEnv("monitoring.tracing_service_name", "TRACING_SERVICE_NAME")
    v.BindEnv("monitoring.tracing_endpoint", "TRACING_ENDPOINT")
    v.BindEnv("monitoring.log_format", "LOG_FORMAT")
    v.BindEnv("monitoring.log_output", "LOG_OUTPUT")
    v.BindEnv("monitoring.log_file_path", "LOG_FILE_PATH")
    
    // Features
    v.BindEnv("features.registration", "FEATURE_REGISTRATION")
    v.BindEnv("features.invitation_export", "FEATURE_INVITATION_EXPORT")
    v.BindEnv("features.analytics", "FEATURE_ANALYTICS")
    v.BindEnv("features.bulk_import", "FEATURE_BULK_IMPORT")
}

// Validate performs comprehensive configuration validation
func (c *Config) Validate() error {
    errors := make([]string, 0)
    
    // Validate Server
    if c.Server.Environment == "" {
        errors = append(errors, "server.environment is required")
    }
    validEnvs := []string{"development", "test", "staging", "production"}
    if !contains(validEnvs, c.Server.Environment) {
        errors = append(errors, fmt.Sprintf("server.environment must be one of: %v", validEnvs))
    }
    
    if c.Server.Port <= 0 || c.Server.Port > 65535 {
        errors = append(errors, "server.port must be between 1 and 65535")
    }
    
    // Validate Database
    if c.Database.URI == "" {
        errors = append(errors, "database.uri (MONGODB_URI) is required")
    }
    if c.Database.Database == "" {
        errors = append(errors, "database.database (MONGODB_DATABASE) is required")
    }
    
    // Validate Auth (only in production)
    if c.Server.Environment == "production" {
        if c.Auth.JWTSecret == "" {
            errors = append(errors, "auth.jwt_secret is required in production")
        }
        if len(c.Auth.JWTSecret) < 32 {
            errors = append(errors, "auth.jwt_secret must be at least 32 characters")
        }
        if c.Auth.JWTRefreshSecret == "" {
            errors = append(errors, "auth.jwt_refresh_secret is required in production")
        }
        if len(c.Auth.JWTRefreshSecret) < 32 {
            errors = append(errors, "auth.jwt_refresh_secret must be at least 32 characters")
        }
    }
    
    // Validate Storage
    if c.Storage.Provider == "s3" {
        if c.Storage.S3.AccessKeyID == "" || c.Storage.S3.SecretAccessKey == "" {
            errors = append(errors, "storage.s3 credentials are required when provider is s3")
        }
        if c.Storage.S3.Bucket == "" {
            errors = append(errors, "storage.s3.bucket is required when provider is s3")
        }
        if c.Storage.S3.Region == "" {
            errors = append(errors, "storage.s3.region is required when provider is s3")
        }
    }
    
    if c.Storage.Provider == "r2" {
        if c.Storage.R2.AccountID == "" {
            errors = append(errors, "storage.r2.account_id is required when provider is r2")
        }
        if c.Storage.R2.AccessKeyID == "" || c.Storage.R2.SecretAccessKey == "" {
            errors = append(errors, "storage.r2 credentials are required when provider is r2")
        }
        if c.Storage.R2.Bucket == "" {
            errors = append(errors, "storage.r2.bucket is required when provider is r2")
        }
    }
    
    // Validate Email
    if c.Email.Provider == "sendgrid" && c.Email.Sendgrid.APIKey == "" {
        errors = append(errors, "email.sendgrid.api_key is required when provider is sendgrid")
    }
    
    if c.Email.Provider == "smtp" {
        if c.Email.SMTP.Host == "" {
            errors = append(errors, "email.smtp.host is required when provider is smtp")
        }
        if c.Email.SMTP.Username == "" || c.Email.SMTP.Password == "" {
            errors = append(errors, "email.smtp credentials are required when provider is smtp")
        }
    }
    
    // Validate Monitoring
    if c.Monitoring.LogOutput == "file" && c.Monitoring.LogFilePath == "" {
        errors = append(errors, "monitoring.log_file_path is required when log_output is file")
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("configuration validation failed:\n- %s", strings.Join(errors, "\n- "))
    }
    
    return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
    c.RLock()
    defer c.RUnlock()
    return c.Server.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
    c.RLock()
    defer c.RUnlock()
    return c.Server.Environment == "production"
}

// GetString safely retrieves a string value from raw config
func (c *Config) GetString(key string) string {
    c.RLock()
    defer c.RUnlock()
    if val, ok := c.raw[key]; ok {
        if s, ok := val.(string); ok {
            return s
        }
    }
    return ""
}

// GetInt safely retrieves an int value from raw config
func (c *Config) GetInt(key string) int {
    c.RLock()
    defer c.RUnlock()
    if val, ok := c.raw[key]; ok {
        switch v := val.(type) {
        case int:
            return v
        case float64:
            return int(v)
        }
    }
    return 0
}

// GetBool safely retrieves a bool value from raw config
func (c *Config) GetBool(key string) bool {
    c.RLock()
    defer c.RUnlock()
    if val, ok := c.raw[key]; ok {
        if b, ok := val.(bool); ok {
            return b
        }
    }
    return false
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}
```

### Hot Reload Configuration (Development)

```go
// config/watcher.go
package config

import (
    "fmt"
    "log"
    "path/filepath"
    "sync"
    "time"

    "github.com/fsnotify/fsnotify"
    "github.com/spf13/viper"
)

// ConfigWatcher watches for configuration changes
type ConfigWatcher struct {
    watcher  *fsnotify.Watcher
    mu       sync.RWMutex
    onChange func(*Config)
    stopped  bool
}

// NewConfigWatcher creates a new configuration watcher
func NewConfigWatcher(onChange func(*Config)) (*ConfigWatcher, error) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, fmt.Errorf("failed to create watcher: %w", err)
    }
    
    return &ConfigWatcher{
        watcher:  watcher,
        onChange: onChange,
    }, nil
}

// Watch starts watching configuration files
func (cw *ConfigWatcher) Watch(paths ...string) error {
    for _, path := range paths {
        // Watch the directory containing the config file
        dir := filepath.Dir(path)
        if err := cw.watcher.Add(dir); err != nil {
            return fmt.Errorf("failed to watch %s: %w", dir, err)
        }
        log.Printf("Watching config directory: %s", dir)
    }
    
    go cw.run()
    return nil
}

// run processes file system events
func (cw *ConfigWatcher) run() {
    debounceTimer := time.NewTimer(0)
    <-debounceTimer.C
    
    for {
        select {
        case event, ok := <-cw.watcher.Events:
            if !ok {
                return
            }
            
            // Only react to config file changes
            if filepath.Ext(event.Name) == ".yaml" || 
               filepath.Ext(event.Name) == ".yml" ||
               filepath.Ext(event.Name) == ".env" {
                
                if event.Op&fsnotify.Write == fsnotify.Write ||
                   event.Op&fsnotify.Create == fsnotify.Create {
                    
                    // Debounce multiple rapid changes
                    debounceTimer.Reset(500 * time.Millisecond)
                    
                    go func() {
                        <-debounceTimer.C
                        cw.handleChange()
                    }()
                }
            }
            
        case err, ok := <-cw.watcher.Errors:
            if !ok {
                return
            }
            log.Printf("Config watcher error: %v", err)
            
        case <-cw.getStopChannel():
            return
        }
    }
}

// handleChange reloads configuration and notifies listeners
func (cw *ConfigWatcher) handleChange() {
    cw.mu.Lock()
    defer cw.mu.Unlock()
    
    if cw.stopped {
        return
    }
    
    log.Println("Configuration change detected, reloading...")
    
    if err := Reload(); err != nil {
        log.Printf("Failed to reload configuration: %v", err)
        return
    }
    
    if cw.onChange != nil {
        cw.onChange(Get())
    }
    
    log.Println("Configuration reloaded successfully")
}

// getStopChannel returns a channel that closes when stopped
func (cw *ConfigWatcher) getStopChannel() <-chan struct{} {
    cw.mu.RLock()
    defer cw.mu.RUnlock()
    
    if cw.stopped {
        ch := make(chan struct{})
        close(ch)
        return ch
    }
    
    // Use a ticker that we never send to as a "never" channel
    return make(chan struct{})
}

// Stop stops the configuration watcher
func (cw *ConfigWatcher) Stop() error {
    cw.mu.Lock()
    defer cw.mu.Unlock()
    
    if cw.stopped {
        return nil
    }
    
    cw.stopped = true
    return cw.watcher.Close()
}
```

---

## 4. Environment Files

### File Structure

```
.
├── .env                    # Local development (gitignored)
├── .env.example            # Template with all variables
├── .env.staging            # Staging environment config
├── .env.production         # Production environment config (encrypted)
├── .env.test               # Test environment config
└── config/
    └── config.yaml         # Optional YAML config file
```

### .env.example (Template)

```bash
# =============================================================================
# Environment Configuration Template
# Copy this file to .env and fill in your actual values
# DO NOT commit .env files with real values to version control
# =============================================================================

# =============================================================================
# Server Configuration
# =============================================================================
APP_ENV=development
APP_NAME=wedding-invitation-api
PORT=8080
HOST=0.0.0.0
DEBUG=false
LOG_LEVEL=info
REQUEST_TIMEOUT=30s
SHUTDOWN_TIMEOUT=10s

# =============================================================================
# Database Configuration
# =============================================================================
# MongoDB connection string (REQUIRED)
# Format: mongodb://[username:password@]host1[:port1][,host2[:port2],...][/database][?options]
MONGODB_URI=mongodb://localhost:27017

# Database name (REQUIRED)
MONGODB_DATABASE=wedding_invitation_dev

# Connection pool settings
MONGODB_MAX_POOL_SIZE=100
MONGODB_MIN_POOL_SIZE=10
MONGODB_CONNECT_TIMEOUT=10s
MONGODB_SOCKET_TIMEOUT=0s
MONGODB_RETRY_WRITES=true
MONGODB_MAX_COMMIT_TIME=5s

# =============================================================================
# Authentication Configuration
# =============================================================================
# JWT Secrets (REQUIRED in production - min 32 chars)
# Generate with: openssl rand -base64 64
JWT_SECRET=your-super-secret-jwt-key-change-in-production-at-least-32-chars
JWT_REFRESH_SECRET=your-super-secret-refresh-key-change-in-production-at-least-32-chars

# Token expiration times
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d
JWT_ISSUER=wedding-invitation-api

# Password hashing cost (10-14 recommended)
BCRYPT_COST=12

# =============================================================================
# Storage Configuration (Choose ONE provider)
# =============================================================================

# Option 1: AWS S3
# STORAGE_PROVIDER=s3
# AWS_ACCESS_KEY_ID=your-access-key
# AWS_SECRET_ACCESS_KEY=your-secret-key
# AWS_REGION=us-east-1
# AWS_S3_BUCKET=your-bucket-name
# AWS_S3_ENDPOINT=                      # For MinIO or other S3-compatible
# AWS_S3_FORCE_PATH_STYLE=false

# Option 2: Cloudflare R2
# STORAGE_PROVIDER=r2
# R2_ACCOUNT_ID=your-account-id
# R2_ACCESS_KEY_ID=your-access-key
# R2_SECRET_ACCESS_KEY=your-secret-key
# R2_BUCKET=your-bucket-name
# R2_PUBLIC_URL=https://your-custom-domain.com

# Option 3: Local Storage (Development only)
STORAGE_PROVIDER=local
STORAGE_BASE_PATH=./uploads
STORAGE_BASE_URL=http://localhost:8080/uploads

# =============================================================================
# Email Configuration (Choose ONE provider)
# =============================================================================

# Option 1: SendGrid
# EMAIL_PROVIDER=sendgrid
# SENDGRID_API_KEY=SG.your-api-key

# Option 2: SMTP
EMAIL_PROVIDER=smtp
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_TLS=true

# Email defaults
EMAIL_FROM=noreply@example.com
EMAIL_FROM_NAME=Wedding Invitation

# =============================================================================
# Security Configuration
# =============================================================================
# CORS settings (comma-separated)
# In production, specify exact origins:
# CORS_ALLOWED_ORIGINS=https://app.example.com,https://admin.example.com
CORS_ALLOWED_ORIGINS=*
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=Authorization,Content-Type,X-Requested-With
CORS_MAX_AGE=86400

# Rate limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
RATE_LIMIT_BURST=10

# Trusted proxy IPs (comma-separated)
# TRUSTED_PROXIES=10.0.0.1,10.0.0.2

# Security headers
SECURE_HEADERS=true
HSTS_MAX_AGE=31536000

# =============================================================================
# Monitoring & Observability
# =============================================================================
METRICS_ENABLED=true
METRICS_PORT=9090
METRICS_PATH=/metrics

# Distributed tracing
TRACING_ENABLED=false
TRACING_SERVICE_NAME=wedding-api
TRACING_ENDPOINT=http://jaeger:14268/api/traces

# Logging
LOG_FORMAT=json
LOG_OUTPUT=stdout
# LOG_FILE_PATH=/var/log/wedding-api/app.log

# =============================================================================
# Feature Flags
# =============================================================================
FEATURE_REGISTRATION=true
FEATURE_INVITATION_EXPORT=true
FEATURE_ANALYTICS=false
FEATURE_BULK_IMPORT=false
```

### .env.development

```bash
# =============================================================================
# Development Environment Configuration
# Optimized for local development with debug features enabled
# =============================================================================

# Server
APP_ENV=development
APP_NAME=wedding-invitation-api-dev
PORT=8080
HOST=localhost
DEBUG=true
LOG_LEVEL=debug
REQUEST_TIMEOUT=60s
SHUTDOWN_TIMEOUT=5s

# Database
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=wedding_invitation_dev
MONGODB_MAX_POOL_SIZE=10
MONGODB_MIN_POOL_SIZE=1

# Authentication (weak secrets OK for dev)
JWT_SECRET=dev-secret-not-for-production-32-chars
JWT_REFRESH_SECRET=dev-refresh-not-for-production-32-chars
JWT_ACCESS_EXPIRY=1h
JWT_REFRESH_EXPIRY=30d
BCRYPT_COST=4

# Storage
STORAGE_PROVIDER=local
STORAGE_BASE_PATH=./uploads
STORAGE_BASE_URL=http://localhost:8080/uploads

# Email (use Mailtrap or similar for testing)
EMAIL_PROVIDER=smtp
SMTP_HOST=sandbox.smtp.mailtrap.io
SMTP_PORT=587
SMTP_USER=your-mailtrap-user
SMTP_PASSWORD=your-mailtrap-password
SMTP_TLS=true
EMAIL_FROM=test@example.com
EMAIL_FROM_NAME=Wedding Invitation Test

# Security (permissive for development)
CORS_ALLOWED_ORIGINS=*
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS,PATCH
CORS_ALLOWED_HEADERS=*
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_WINDOW=1m
RATE_LIMIT_BURST=50
SECURE_HEADERS=false

# Monitoring
METRICS_ENABLED=true
METRICS_PORT=9090
LOG_FORMAT=console
LOG_OUTPUT=stdout

# Features
FEATURE_REGISTRATION=true
FEATURE_INVITATION_EXPORT=true
FEATURE_ANALYTICS=true
FEATURE_BULK_IMPORT=true
```

### .env.staging

```bash
# =============================================================================
# Staging Environment Configuration
# Pre-production testing with production-like settings
# =============================================================================

# Server
APP_ENV=staging
APP_NAME=wedding-invitation-api-staging
PORT=8080
HOST=0.0.0.0
DEBUG=false
LOG_LEVEL=info
REQUEST_TIMEOUT=30s
SHUTDOWN_TIMEOUT=10s

# Database
MONGODB_URI=mongodb+srv://user:pass@staging-cluster.mongodb.net/wedding_invitation_staging?retryWrites=true&w=majority
MONGODB_DATABASE=wedding_invitation_staging

# Authentication (use strong secrets)
JWT_SECRET=${STAGING_JWT_SECRET}
JWT_REFRESH_SECRET=${STAGING_JWT_REFRESH_SECRET}
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d
BCRYPT_COST=12

# Storage
STORAGE_PROVIDER=s3
AWS_ACCESS_KEY_ID=${STAGING_AWS_ACCESS_KEY_ID}
AWS_SECRET_ACCESS_KEY=${STAGING_AWS_SECRET_ACCESS_KEY}
AWS_REGION=us-east-1
AWS_S3_BUCKET=wedding-invitation-staging-uploads

# Email
EMAIL_PROVIDER=sendgrid
SENDGRID_API_KEY=${STAGING_SENDGRID_API_KEY}
EMAIL_FROM=staging@example.com
EMAIL_FROM_NAME=Wedding Invitation Staging

# Security
CORS_ALLOWED_ORIGINS=https://staging-app.example.com,https://staging-admin.example.com
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
RATE_LIMIT_REQUESTS=200
RATE_LIMIT_WINDOW=1m
RATE_LIMIT_BURST=20
SECURE_HEADERS=true
HSTS_MAX_AGE=31536000

# Monitoring
METRICS_ENABLED=true
METRICS_PORT=9090
LOG_FORMAT=json
LOG_OUTPUT=stdout
TRACING_ENABLED=true
TRACING_SERVICE_NAME=wedding-api-staging
TRACING_ENDPOINT=https://staging-jaeger.example.com/api/traces

# Features
FEATURE_REGISTRATION=true
FEATURE_INVITATION_EXPORT=true
FEATURE_ANALYTICS=true
FEATURE_BULK_IMPORT=true
```

### .env.production (Template)

```bash
# =============================================================================
# Production Environment Configuration
# All sensitive values should be injected via secrets management
# This file serves as documentation of required variables
# =============================================================================

# Server
APP_ENV=production
APP_NAME=wedding-invitation-api
PORT=8080
HOST=0.0.0.0
DEBUG=false
LOG_LEVEL=warn
REQUEST_TIMEOUT=30s
SHUTDOWN_TIMEOUT=15s

# Database
MONGODB_URI=${MONGODB_URI}              # Injected from AWS Secrets Manager
MONGODB_DATABASE=wedding_invitation_prod

# Authentication
JWT_SECRET=${JWT_SECRET}                # Injected from AWS Secrets Manager
JWT_REFRESH_SECRET=${JWT_REFRESH_SECRET} # Injected from AWS Secrets Manager
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d
BCRYPT_COST=12

# Storage
STORAGE_PROVIDER=s3
AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}      # Injected from AWS Secrets Manager
AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} # Injected from AWS Secrets Manager
AWS_REGION=us-east-1
AWS_S3_BUCKET=wedding-invitation-prod-uploads

# Email
EMAIL_PROVIDER=sendgrid
SENDGRID_API_KEY=${SENDGRID_API_KEY}        # Injected from AWS Secrets Manager
EMAIL_FROM=noreply@example.com
EMAIL_FROM_NAME=Wedding Invitation

# Security
CORS_ALLOWED_ORIGINS=${CORS_ORIGINS}        # Injected - comma-separated list
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1m
RATE_LIMIT_BURST=10
TRUSTED_PROXIES=${TRUSTED_PROXIES}          # Injected from infrastructure
SECURE_HEADERS=true
HSTS_MAX_AGE=31536000

# Monitoring
METRICS_ENABLED=true
METRICS_PORT=9090
LOG_FORMAT=json
LOG_OUTPUT=stdout
TRACING_ENABLED=true
TRACING_SERVICE_NAME=wedding-api-prod
TRACING_ENDPOINT=${JAEGER_ENDPOINT}         # Injected from infrastructure

# Features
FEATURE_REGISTRATION=true
FEATURE_INVITATION_EXPORT=true
FEATURE_ANALYTICS=true
FEATURE_BULK_IMPORT=false
```

### .env.test

```bash
# =============================================================================
# Test Environment Configuration
# Used by CI/CD and automated testing
# =============================================================================

APP_ENV=test
APP_NAME=wedding-invitation-api-test
PORT=8080
HOST=localhost
DEBUG=false
LOG_LEVEL=error
REQUEST_TIMEOUT=10s
SHUTDOWN_TIMEOUT=5s

# Use test database (dropped/recreated between test runs)
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=wedding_invitation_test
MONGODB_MAX_POOL_SIZE=5
MONGODB_MIN_POOL_SIZE=1

# Weak secrets OK for tests
JWT_SECRET=test-secret-not-for-production
JWT_REFRESH_SECRET=test-refresh-not-for-production
JWT_ACCESS_EXPIRY=5m
JWT_REFRESH_EXPIRY=1h
BCRYPT_COST=4

# Local storage for tests
STORAGE_PROVIDER=local
STORAGE_BASE_PATH=./test-uploads
STORAGE_BASE_URL=http://localhost:8080/uploads

# No email in tests
EMAIL_PROVIDER=mock

# Relaxed CORS for test client
CORS_ALLOWED_ORIGINS=*
RATE_LIMIT_REQUESTS=10000
SECURE_HEADERS=false

# Minimal monitoring
METRICS_ENABLED=false
LOG_FORMAT=console

# Enable all features for testing
FEATURE_REGISTRATION=true
FEATURE_INVITATION_EXPORT=true
FEATURE_ANALYTICS=true
FEATURE_BULK_IMPORT=true
```

---

## 5. Secrets Management

### Development: Local .env

For local development, use a `.env` file:

```bash
# Create from template
cp .env.example .env

# Edit with your values
nano .env

# Load in shell
export $(cat .env | xargs)
```

**Security considerations:**
- Never commit `.env` files
- Add to `.gitignore`
- Use different secrets for each developer
- Rotate secrets periodically

### Production: AWS Secrets Manager

```go
// secrets/aws.go
package secrets

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

// AWSSecretsManager retrieves secrets from AWS
type AWSSecretsManager struct {
    client *secretsmanager.Client
}

// NewAWSSecretsManager creates a new AWS secrets manager client
func NewAWSSecretsManager(ctx context.Context, region string) (*AWSSecretsManager, error) {
    cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
    if err != nil {
        return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
    }

    return &AWSSecretsManager{
        client: secretsmanager.NewFromConfig(cfg),
    }, nil
}

// GetSecret retrieves a secret by name
func (a *AWSSecretsManager) GetSecret(ctx context.Context, secretName string) (map[string]string, error) {
    input := &secretsmanager.GetSecretValueInput{
        SecretId: aws.String(secretName),
    }

    result, err := a.client.GetSecretValue(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("failed to get secret %s: %w", secretName, err)
    }

    var secrets map[string]string
    if err := json.Unmarshal([]byte(*result.SecretString), &secrets); err != nil {
        return nil, fmt.Errorf("failed to parse secret %s: %w", secretName, err)
    }

    return secrets, nil
}

// LoadIntoViper loads secrets from AWS into viper
func (a *AWSSecretsManager) LoadIntoViper(ctx context.Context, v *viper.Viper, secretName string) error {
    secrets, err := a.GetSecret(ctx, secretName)
    if err != nil {
        return err
    }

    for key, value := range secrets {
        v.Set(key, value)
    }

    return nil
}
```

**Usage:**

```go
// In your main.go or config initialization
if cfg.IsProduction() {
    secretsManager, err := secrets.NewAWSSecretsManager(ctx, "us-east-1")
    if err != nil {
        log.Fatal(err)
    }
    
    // Load application secrets
    if err := secretsManager.LoadIntoViper(ctx, v, "wedding-api/production"); err != nil {
        log.Fatal(err)
    }
}
```

### Production: HashiCorp Vault

```go
// secrets/vault.go
package secrets

import (
    "context"
    "fmt"

    "github.com/hashicorp/vault/api"
)

// VaultManager retrieves secrets from HashiCorp Vault
type VaultManager struct {
    client *api.Client
}

// NewVaultManager creates a new Vault client
func NewVaultManager(addr, token string) (*VaultManager, error) {
    config := api.DefaultConfig()
    config.Address = addr

    client, err := api.NewClient(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create vault client: %w", err)
    }

    client.SetToken(token)

    return &VaultManager{client: client}, nil
}

// GetSecret retrieves secrets from a path
func (v *VaultManager) GetSecret(ctx context.Context, path string) (map[string]interface{}, error) {
    secret, err := v.client.Logical().ReadWithContext(ctx, path)
    if err != nil {
        return nil, fmt.Errorf("failed to read secret at %s: %w", path, err)
    }

    if secret == nil {
        return nil, fmt.Errorf("no secret found at path: %s", path)
    }

    data, ok := secret.Data["data"].(map[string]interface{})
    if !ok {
        // Try direct data (KV v1)
        data = secret.Data
    }

    return data, nil
}

// LoadIntoViper loads vault secrets into viper
func (v *VaultManager) LoadIntoViper(ctx context.Context, vip *viper.Viper, path string) error {
    secrets, err := v.GetSecret(ctx, path)
    if err != nil {
        return err
    }

    for key, value := range secrets {
        vip.Set(key, value)
    }

    return nil
}
```

### Kubernetes Secrets

```yaml
# kubernetes/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: wedding-api-secrets
  namespace: production
type: Opaque
stringData:
  MONGODB_URI: "mongodb+srv://..."
  JWT_SECRET: "super-secret-key"
  JWT_REFRESH_SECRET: "super-secret-refresh-key"
  AWS_ACCESS_KEY_ID: "AKIA..."
  AWS_SECRET_ACCESS_KEY: "..."
  SENDGRID_API_KEY: "SG..."
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: wedding-api-config
  namespace: production
data:
  APP_ENV: "production"
  APP_NAME: "wedding-invitation-api"
  PORT: "8080"
  LOG_LEVEL: "info"
  STORAGE_PROVIDER: "s3"
  AWS_REGION: "us-east-1"
  AWS_S3_BUCKET: "wedding-invitation-prod"
```

**Deployment with Kubernetes:**

```yaml
# kubernetes/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: wedding-api
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: wedding-api
  template:
    metadata:
      labels:
        app: wedding-api
    spec:
      containers:
      - name: api
        image: wedding-api:latest
        envFrom:
        - configMapRef:
            name: wedding-api-config
        - secretRef:
            name: wedding-api-secrets
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
```

### Docker Secrets (Swarm Mode)

```yaml
# docker-compose.yml
version: '3.8'

services:
  api:
    image: wedding-api:latest
    environment:
      - APP_ENV=production
      - APP_NAME=wedding-invitation-api
    secrets:
      - mongodb_uri
      - jwt_secret
      - jwt_refresh_secret
      - aws_access_key_id
      - aws_secret_access_key
      - sendgrid_api_key
    deploy:
      replicas: 3
      update_config:
        parallelism: 1
        delay: 10s
      restart_policy:
        condition: on-failure

secrets:
  mongodb_uri:
    external: true
  jwt_secret:
    external: true
  jwt_refresh_secret:
    external: true
  aws_access_key_id:
    external: true
  aws_secret_access_key:
    external: true
  sendgrid_api_key:
    external: true
```

**Creating Docker secrets:**

```bash
# Create secrets
echo "mongodb+srv://..." | docker secret create mongodb_uri -
echo "super-secret-key" | docker secret create jwt_secret -
echo "super-secret-refresh-key" | docker secret create jwt_refresh_secret -

# Verify
docker secret ls

# Use in container - secrets mounted at /run/secrets/
# Read in application: cat /run/secrets/jwt_secret
```

### Secrets Rotation Strategy

```go
// secrets/rotation.go
package secrets

import (
    "context"
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "time"
)

// Rotator handles automatic secret rotation
type Rotator struct {
    manager SecretsManager
    interval time.Duration
}

// SecretsManager interface for different backends
type SecretsManager interface {
    GetSecret(ctx context.Context, name string) (map[string]string, error)
    SetSecret(ctx context.Context, name string, values map[string]string) error
}

// GenerateSecret creates a cryptographically secure secret
func GenerateSecret(length int) (string, error) {
    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        return "", fmt.Errorf("failed to generate random bytes: %w", err)
    }
    return base64.URLEncoding.EncodeToString(bytes), nil
}

// RotateJWTSecrets performs JWT secret rotation
func (r *Rotator) RotateJWTSecrets(ctx context.Context) error {
    // Generate new secrets
    newSecret, err := GenerateSecret(64)
    if err != nil {
        return err
    }
    
    newRefreshSecret, err := GenerateSecret(64)
    if err != nil {
        return err
    }
    
    // Store with versioning
    secrets := map[string]string{
        "JWT_SECRET":         newSecret,
        "JWT_SECRET_VERSION": time.Now().Format("20060102-150405"),
        "JWT_REFRESH_SECRET": newRefreshSecret,
        "JWT_REFRESH_VERSION": time.Now().Format("20060102-150405"),
    }
    
    // Update in secrets manager
    if err := r.manager.SetSecret(ctx, "wedding-api/jwt", secrets); err != nil {
        return fmt.Errorf("failed to update secrets: %w", err)
    }
    
    return nil
}

// StartRotation begins automatic rotation
func (r *Rotator) StartRotation(ctx context.Context) {
    ticker := time.NewTicker(r.interval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            if err := r.RotateJWTSecrets(ctx); err != nil {
                // Log error, alert on-call
                fmt.Printf("Secret rotation failed: %v\n", err)
            }
        case <-ctx.Done():
            return
        }
    }
}
```

---

## 6. Configuration Validation

### Startup Validation

```go
// config/validation.go
package config

import (
    "context"
    "fmt"
    "net/url"
    "regexp"
    "strings"
    "time"
)

// ValidationError represents a validation error
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validator performs configuration validation
type Validator struct {
    errors []ValidationError
}

// NewValidator creates a new validator
func NewValidator() *Validator {
    return &Validator{
        errors: make([]ValidationError, 0),
    }
}

// Validate performs all validation checks
func (v *Validator) Validate(cfg *Config) error {
    v.validateServer(cfg)
    v.validateDatabase(cfg)
    v.validateAuth(cfg)
    v.validateStorage(cfg)
    v.validateEmail(cfg)
    v.validateSecurity(cfg)
    
    if len(v.errors) > 0 {
        return v.buildError()
    }
    
    return nil
}

func (v *Validator) validateServer(cfg *Config) {
    // Environment validation
    validEnvs := []string{"development", "test", "staging", "production"}
    if !contains(validEnvs, cfg.Server.Environment) {
        v.addError("server.environment", 
            fmt.Sprintf("must be one of: %s", strings.Join(validEnvs, ", ")))
    }
    
    // Port validation
    if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
        v.addError("server.port", "must be between 1 and 65535")
    }
    
    // Log level validation
    validLevels := []string{"debug", "info", "warn", "error", "fatal"}
    if !contains(validLevels, cfg.Server.LogLevel) {
        v.addError("server.log_level", 
            fmt.Sprintf("must be one of: %s", strings.Join(validLevels, ", ")))
    }
    
    // Timeout validation
    if cfg.Server.RequestTimeout < time.Second {
        v.addError("server.request_timeout", "must be at least 1 second")
    }
}

func (v *Validator) validateDatabase(cfg *Config) {
    // URI validation
    if cfg.Database.URI == "" {
        v.addError("database.uri", "is required")
    } else {
        if _, err := url.Parse(cfg.Database.URI); err != nil {
            v.addError("database.uri", "must be a valid URL")
        }
        
        if !strings.HasPrefix(cfg.Database.URI, "mongodb://") &&
           !strings.HasPrefix(cfg.Database.URI, "mongodb+srv://") {
            v.addError("database.uri", "must start with mongodb:// or mongodb+srv://")
        }
    }
    
    // Database name validation
    if cfg.Database.Database == "" {
        v.addError("database.database", "is required")
    } else if matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, cfg.Database.Database); !matched {
        v.addError("database.database", "must contain only alphanumeric characters, hyphens, and underscores")
    }
    
    // Pool size validation
    if cfg.Database.MaxPoolSize <= 0 {
        v.addError("database.max_pool_size", "must be greater than 0")
    }
    if cfg.Database.MinPoolSize < 0 {
        v.addError("database.min_pool_size", "must be non-negative")
    }
    if cfg.Database.MinPoolSize > cfg.Database.MaxPoolSize {
        v.addError("database.min_pool_size", "must not exceed max_pool_size")
    }
}

func (v *Validator) validateAuth(cfg *Config) {
    // Production requires strong secrets
    if cfg.IsProduction() {
        if len(cfg.Auth.JWTSecret) < 32 {
            v.addError("auth.jwt_secret", "must be at least 32 characters in production")
        }
        if len(cfg.Auth.JWTRefreshSecret) < 32 {
            v.addError("auth.jwt_refresh_secret", "must be at least 32 characters in production")
        }
        
        // Ensure secrets are different
        if cfg.Auth.JWTSecret == cfg.Auth.JWTRefreshSecret {
            v.addError("auth.jwt_secret", "must be different from jwt_refresh_secret")
        }
    }
    
    // Bcrypt cost validation
    if cfg.Auth.BcryptCost < 4 || cfg.Auth.BcryptCost > 31 {
        v.addError("auth.bcrypt_cost", "must be between 4 and 31")
    }
    
    // Token expiry validation
    if cfg.Auth.AccessExpiry < time.Minute {
        v.addError("auth.access_expiry", "must be at least 1 minute")
    }
    if cfg.Auth.RefreshExpiry < time.Hour {
        v.addError("auth.refresh_expiry", "must be at least 1 hour")
    }
    if cfg.Auth.RefreshExpiry < cfg.Auth.AccessExpiry {
        v.addError("auth.refresh_expiry", "must be greater than access_expiry")
    }
}

func (v *Validator) validateStorage(cfg *Config) {
    validProviders := []string{"s3", "r2", "local"}
    if !contains(validProviders, cfg.Storage.Provider) {
        v.addError("storage.provider", 
            fmt.Sprintf("must be one of: %s", strings.Join(validProviders, ", ")))
    }
    
    switch cfg.Storage.Provider {
    case "s3":
        if cfg.Storage.S3.AccessKeyID == "" {
            v.addError("storage.s3.access_key_id", "is required when provider is s3")
        }
        if cfg.Storage.S3.SecretAccessKey == "" {
            v.addError("storage.s3.secret_access_key", "is required when provider is s3")
        }
        if cfg.Storage.S3.Region == "" {
            v.addError("storage.s3.region", "is required when provider is s3")
        }
        if cfg.Storage.S3.Bucket == "" {
            v.addError("storage.s3.bucket", "is required when provider is s3")
        }
        
    case "r2":
        if cfg.Storage.R2.AccountID == "" {
            v.addError("storage.r2.account_id", "is required when provider is r2")
        }
        if cfg.Storage.R2.AccessKeyID == "" {
            v.addError("storage.r2.access_key_id", "is required when provider is r2")
        }
        if cfg.Storage.R2.SecretAccessKey == "" {
            v.addError("storage.r2.secret_access_key", "is required when provider is r2")
        }
        if cfg.Storage.R2.Bucket == "" {
            v.addError("storage.r2.bucket", "is required when provider is r2")
        }
        
    case "local":
        if cfg.Storage.Local.BasePath == "" {
            v.addError("storage.local.base_path", "is required when provider is local")
        }
    }
}

func (v *Validator) validateEmail(cfg *Config) {
    validProviders := []string{"sendgrid", "smtp", "mock"}
    if !contains(validProviders, cfg.Email.Provider) {
        v.addError("email.provider", 
            fmt.Sprintf("must be one of: %s", strings.Join(validProviders, ", ")))
    }
    
    switch cfg.Email.Provider {
    case "sendgrid":
        if cfg.Email.Sendgrid.APIKey == "" {
            v.addError("email.sendgrid.api_key", "is required when provider is sendgrid")
        }
        if !strings.HasPrefix(cfg.Email.Sendgrid.APIKey, "SG.") {
            v.addError("email.sendgrid.api_key", "must start with 'SG.'")
        }
        
    case "smtp":
        if cfg.Email.SMTP.Host == "" {
            v.addError("email.smtp.host", "is required when provider is smtp")
        }
        if cfg.Email.SMTP.Port <= 0 || cfg.Email.SMTP.Port > 65535 {
            v.addError("email.smtp.port", "must be between 1 and 65535")
        }
        if cfg.Email.SMTP.Username == "" {
            v.addError("email.smtp.username", "is required when provider is smtp")
        }
        if cfg.Email.SMTP.Password == "" {
            v.addError("email.smtp.password", "is required when provider is smtp")
        }
    }
    
    // Validate email format
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if cfg.Email.From != "" && !emailRegex.MatchString(cfg.Email.From) {
        v.addError("email.from", "must be a valid email address")
    }
}

func (v *Validator) validateSecurity(cfg *Config) {
    // CORS validation
    if cfg.Security.CorsAllowedOrigins == "" {
        v.addError("security.cors_allowed_origins", "is required")
    }
    
    // Production should not use wildcard CORS
    if cfg.IsProduction() && cfg.Security.CorsAllowedOrigins == "*" {
        v.addError("security.cors_allowed_origins", "wildcard (*) not allowed in production")
    }
    
    // Rate limiting validation
    if cfg.Security.RateLimitRequests <= 0 {
        v.addError("security.rate_limit_requests", "must be greater than 0")
    }
    if cfg.Security.RateLimitWindow < time.Second {
        v.addError("security.rate_limit_window", "must be at least 1 second")
    }
}

func (v *Validator) addError(field, message string) {
    v.errors = append(v.errors, ValidationError{
        Field:   field,
        Message: message,
    })
}

func (v *Validator) buildError() error {
    var msgs []string
    for _, err := range v.errors {
        msgs = append(msgs, err.Error())
    }
    return fmt.Errorf("configuration validation failed with %d errors:\n- %s", 
        len(v.errors), strings.Join(msgs, "\n- "))
}

// HealthCheck validates external service connectivity
func (cfg *Config) HealthCheck(ctx context.Context) error {
    // Check database connectivity
    if err := cfg.checkDatabase(ctx); err != nil {
        return fmt.Errorf("database health check failed: %w", err)
    }
    
    // Check storage connectivity
    if err := cfg.checkStorage(ctx); err != nil {
        return fmt.Errorf("storage health check failed: %w", err)
    }
    
    // Check email connectivity
    if err := cfg.checkEmail(ctx); err != nil {
        return fmt.Errorf("email health check failed: %w", err)
    }
    
    return nil
}

func (cfg *Config) checkDatabase(ctx context.Context) error {
    // Implementation depends on your MongoDB driver
    // Example with mongo-driver:
    // client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Database.URI))
    // if err != nil {
    //     return err
    // }
    // defer client.Disconnect(ctx)
    // return client.Ping(ctx, nil)
    return nil
}

func (cfg *Config) checkStorage(ctx context.Context) error {
    // Implementation depends on storage provider
    return nil
}

func (cfg *Config) checkEmail(ctx context.Context) error {
    // Implementation depends on email provider
    return nil
}
```

### Integration with Application Startup

```go
// main.go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/yourorg/wedding-api/config"
    "github.com/yourorg/wedding-api/server"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }
    
    // Run health checks on external services
    if err := cfg.HealthCheck(ctx); err != nil {
        log.Fatalf("Health check failed: %v", err)
    }
    
    log.Printf("Configuration loaded successfully (env: %s)", cfg.Server.Environment)
    
    // Create and start server
    srv, err := server.New(cfg)
    if err != nil {
        log.Fatalf("Failed to create server: %v", err)
    }
    
    // Handle graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        log.Println("Shutdown signal received...")
        
        shutdownCtx, shutdownCancel := context.WithTimeout(ctx, cfg.Server.ShutdownTimeout)
        defer shutdownCancel()
        
        if err := srv.Shutdown(shutdownCtx); err != nil {
            log.Printf("Server shutdown error: %v", err)
        }
        
        cancel()
    }()
    
    // Start server (blocking)
    if err := srv.Start(); err != nil {
        log.Fatalf("Server error: %v", err)
    }
    
    log.Println("Server stopped gracefully")
}
```

---

## 7. Environment-Specific Behaviors

### Development Configuration

```go
// config/environments.go
package config

// DevelopmentBehavior defines development-specific settings
func (c *Config) DevelopmentBehavior() *EnvironmentBehavior {
    return &EnvironmentBehavior{
        // Logging
        LogLevel:        "debug",
        LogFormat:       "console",
        LogCaller:       true,
        LogStacktrace:   true,
        
        // Debugging
        DebugMode:       true,
        PProfEnabled:    true,
        PProfPort:       6060,
        
        // CORS - Permissive for local development
        CORSOrigins:     []string{"*"},
        CORSMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
        CORSHeaders:     []string{"*"},
        CORSCredentials: false,
        
        // Security - Relaxed for development
        RateLimitRequests: 1000,
        RateLimitWindow:   time.Minute,
        SecureHeaders:     false,
        
        // Performance
        EnableGzip:        false,
        CacheEnabled:      false,
        
        // Features
        HotReload:         true,
        AutoMigrate:       true,
        
        // Monitoring
        MetricsEnabled:    true,
        TracingEnabled:    false,
    }
}

// ProductionBehavior defines production-specific settings
func (c *Config) ProductionBehavior() *EnvironmentBehavior {
    return &EnvironmentBehavior{
        // Logging
        LogLevel:        "warn",
        LogFormat:       "json",
        LogCaller:       false,
        LogStacktrace:   false,
        
        // Debugging
        DebugMode:       false,
        PProfEnabled:    false,
        PProfPort:       0,
        
        // CORS - Strict in production
        CORSOrigins:     parseOrigins(c.Security.CorsAllowedOrigins),
        CORSMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        CORSHeaders:     []string{"Authorization", "Content-Type", "X-Request-ID"},
        CORSCredentials: true,
        
        // Security - Strict
        RateLimitRequests: c.Security.RateLimitRequests,
        RateLimitWindow:   c.Security.RateLimitWindow,
        SecureHeaders:     true,
        
        // Performance
        EnableGzip:        true,
        CacheEnabled:      true,
        
        // Features
        HotReload:         false,
        AutoMigrate:       false,
        
        // Monitoring
        MetricsEnabled:    c.Monitoring.MetricsEnabled,
        TracingEnabled:    c.Monitoring.TracingEnabled,
    }
}

// EnvironmentBehavior contains environment-specific behavior settings
type EnvironmentBehavior struct {
    LogLevel        string
    LogFormat       string
    LogCaller       bool
    LogStacktrace   bool
    DebugMode       bool
    PProfEnabled    bool
    PProfPort       int
    CORSOrigins     []string
    CORSMethods     []string
    CORSHeaders     []string
    CORSCredentials bool
    RateLimitRequests int
    RateLimitWindow   time.Duration
    SecureHeaders     bool
    EnableGzip        bool
    CacheEnabled      bool
    HotReload         bool
    AutoMigrate       bool
    MetricsEnabled    bool
    TracingEnabled    bool
}

// GetBehavior returns the appropriate behavior for the current environment
func (c *Config) GetBehavior() *EnvironmentBehavior {
    switch c.Server.Environment {
    case "development":
        return c.DevelopmentBehavior()
    case "production":
        return c.ProductionBehavior()
    case "staging":
        // Staging behaves like production
        return c.ProductionBehavior()
    case "test":
        return c.TestBehavior()
    default:
        return c.DevelopmentBehavior()
    }
}

// TestBehavior defines test-specific settings
func (c *Config) TestBehavior() *EnvironmentBehavior {
    return &EnvironmentBehavior{
        LogLevel:        "error",
        LogFormat:       "console",
        DebugMode:       false,
        CORSOrigins:     []string{"*"},
        RateLimitRequests: 10000,
        SecureHeaders:     false,
        EnableGzip:        false,
        CacheEnabled:      false,
        HotReload:         false,
        AutoMigrate:       true,
        MetricsEnabled:    false,
        TracingEnabled:    false,
    }
}

func parseOrigins(origins string) []string {
    if origins == "" || origins == "*" {
        return []string{"*"}
    }
    return strings.Split(origins, ",")
}
```

### Middleware Configuration Based on Environment

```go
// server/middleware.go
package server

import (
    "github.com/gin-gonic/gin"
    "github.com/yourorg/wedding-api/config"
)

// SetupMiddleware configures middleware based on environment
func SetupMiddleware(r *gin.Engine, cfg *config.Config) {
    behavior := cfg.GetBehavior()
    
    // Recovery middleware (always enabled)
    r.Use(gin.Recovery())
    
    // Logging middleware
    if behavior.LogLevel == "debug" {
        r.Use(gin.Logger())
    }
    
    // CORS middleware
    r.Use(corsMiddleware(behavior))
    
    // Rate limiting
    if cfg.IsProduction() || cfg.IsStaging() {
        r.Use(rateLimitMiddleware(cfg))
    }
    
    // Security headers
    if behavior.SecureHeaders {
        r.Use(securityHeadersMiddleware())
    }
    
    // Gzip compression
    if behavior.EnableGzip {
        r.Use(gzipMiddleware())
    }
    
    // Request ID
    r.Use(requestIDMiddleware())
    
    // Metrics
    if behavior.MetricsEnabled {
        r.Use(metricsMiddleware())
    }
}

func corsMiddleware(behavior *config.EnvironmentBehavior) gin.HandlerFunc {
    config := cors.Config{
        AllowOrigins:     behavior.CORSOrigins,
        AllowMethods:     behavior.CORSMethods,
        AllowHeaders:     behavior.CORSHeaders,
        AllowCredentials: behavior.CORSCredentials,
        MaxAge:           86400,
    }
    return cors.New(config)
}

func rateLimitMiddleware(cfg *config.Config) gin.HandlerFunc {
    limiter := tollbooth.NewLimiter(
        float64(cfg.Security.RateLimitRequests)/cfg.Security.RateLimitWindow.Seconds(),
        &limiter.ExpirableOptions{DefaultExpirationTTL: cfg.Security.RateLimitWindow},
    )
    limiter.SetBurst(cfg.Security.RateLimitBurst)
    
    return func(c *gin.Context) {
        httpError := tollbooth.LimitByRequest(limiter, c.Writer, c.Request)
        if httpError != nil {
            c.AbortWithStatusJSON(httpError.StatusCode, gin.H{
                "error": "Rate limit exceeded",
                "retry_after": httpError.Message,
            })
            return
        }
        c.Next()
    }
}
```

---

## 8. Code Examples

### Complete Application Bootstrap

```go
// cmd/server/main.go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    
    "github.com/yourorg/wedding-api/config"
    "github.com/yourorg/wedding-api/database"
    "github.com/yourorg/wedding-api/handlers"
    "github.com/yourorg/wedding-api/middleware"
    "github.com/yourorg/wedding-api/storage"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }
    
    // Setup logging
    setupLogging(cfg)
    
    // Validate configuration
    validator := config.NewValidator()
    if err := validator.Validate(cfg); err != nil {
        log.Fatalf("Configuration validation failed: %v", err)
    }
    
    // Health check external services
    ctx := context.Background()
    if err := cfg.HealthCheck(ctx); err != nil {
        log.Fatalf("Health check failed: %v", err)
    }
    
    // Initialize database
    db, err := database.Connect(cfg.Database)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Disconnect(ctx)
    
    // Initialize storage
    storageClient, err := storage.New(cfg.Storage)
    if err != nil {
        log.Fatalf("Failed to initialize storage: %v", err)
    }
    
    // Create Gin router
    behavior := cfg.GetBehavior()
    if !behavior.DebugMode {
        gin.SetMode(gin.ReleaseMode)
    }
    
    r := gin.New()
    
    // Setup middleware
    middleware.Setup(r, cfg)
    
    // Setup routes
    handlers.SetupRoutes(r, db, storageClient, cfg)
    
    // Create HTTP server
    srv := &http.Server{
        Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
        Handler:      r,
        ReadTimeout:  cfg.Server.RequestTimeout,
        WriteTimeout: cfg.Server.RequestTimeout,
    }
    
    // Start metrics server (separate port)
    if cfg.Monitoring.MetricsEnabled {
        go startMetricsServer(cfg)
    }
    
    // Start pprof server in development
    if behavior.PProfEnabled {
        go startPProfServer(behavior.PProfPort)
    }
    
    // Setup graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    
    go func() {
        <-sigChan
        log.Println("Shutdown signal received...")
        
        shutdownCtx, cancel := context.WithTimeout(ctx, cfg.Server.ShutdownTimeout)
        defer cancel()
        
        if err := srv.Shutdown(shutdownCtx); err != nil {
            log.Printf("Server shutdown error: %v", err)
        }
    }()
    
    // Start server
    log.Printf("Server starting on %s", srv.Addr)
    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("Server error: %v", err)
    }
    
    log.Println("Server stopped gracefully")
}

func setupLogging(cfg *config.Config) {
    behavior := cfg.GetBehavior()
    
    // Configure logrus or zap based on settings
    // This is a simplified example
    if behavior.LogFormat == "json" {
        log.SetFlags(0)
    } else {
        log.SetFlags(log.LstdFlags | log.Lshortfile)
    }
}

func startMetricsServer(cfg *config.Config) {
    mux := http.NewServeMux()
    mux.Handle(cfg.Monitoring.MetricsPath, promhttp.Handler())
    
    addr := fmt.Sprintf(":%d", cfg.Monitoring.MetricsPort)
    log.Printf("Metrics server starting on %s%s", addr, cfg.Monitoring.MetricsPath)
    
    if err := http.ListenAndServe(addr, mux); err != nil {
        log.Printf("Metrics server error: %v", err)
    }
}

func startPProfServer(port int) {
    addr := fmt.Sprintf("localhost:%d", port)
    log.Printf("pprof server starting on http://%s/debug/pprof/", addr)
    
    if err := http.ListenAndServe(addr, nil); err != nil {
        log.Printf("pprof server error: %v", err)
    }
}
```

### Docker Compose Setup

```yaml
# docker-compose.yml
version: '3.8'

services:
  api:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - APP_ENV=development
      - PORT=8080
      - MONGODB_URI=mongodb://mongo:27017/wedding_invitation
      - JWT_SECRET=dev-secret-32-chars-minimum-required
      - JWT_REFRESH_SECRET=dev-refresh-32-chars-minimum-required
      - STORAGE_PROVIDER=local
      - STORAGE_BASE_PATH=/app/uploads
      - LOG_LEVEL=debug
      - DEBUG=true
    volumes:
      - ./uploads:/app/uploads
      - .env:/app/.env:ro
    depends_on:
      - mongo
      - redis
    networks:
      - wedding-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  mongo:
    image: mongo:6
    ports:
      - "27017:27017"
    volumes:
      - mongo-data:/data/db
    networks:
      - wedding-network
    environment:
      - MONGO_INITDB_DATABASE=wedding_invitation

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    networks:
      - wedding-network

volumes:
  mongo-data:
  redis-data:

networks:
  wedding-network:
    driver: bridge
```

### Makefile for Environment Management

```makefile
# Makefile
.PHONY: dev prod test lint clean help

# Environment variables
ENV_FILE ?= .env
export $(shell grep -v '^#' $(ENV_FILE) | xargs)

# Default target
.DEFAULT_GOAL := help

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# Development commands
dev: ## Run in development mode with hot reload
	APP_ENV=development \
	DEBUG=true \
	LOG_LEVEL=debug \
	go run cmd/server/main.go

dev-docker: ## Run development environment with Docker Compose
	docker-compose up --build

dev-down: ## Stop development environment
	docker-compose down -v

# Configuration commands
config-check: ## Validate configuration
	@go run cmd/server/main.go --check-config

config-show: ## Display current configuration
	@go run cmd/server/main.go --show-config

env-init: ## Initialize environment from template
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "Created .env from template. Please edit with your values."; \
	else \
		echo ".env already exists. Remove it first to recreate."; \
	fi

env-validate: ## Validate environment file
	@bash scripts/validate-env.sh

# Testing commands
test: ## Run all tests
	APP_ENV=test go test -v ./...

test-unit: ## Run unit tests
	APP_ENV=test go test -v -short ./...

test-integration: ## Run integration tests
	APP_ENV=test go test -v -run Integration ./...

test-coverage: ## Run tests with coverage report
	APP_ENV=test go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Production commands
prod-build: ## Build production Docker image
	docker build -t wedding-api:latest -f Dockerfile.prod .

prod-deploy: ## Deploy to production (requires proper credentials)
	@echo "Deploying to production..."
	@bash scripts/deploy-production.sh

# Maintenance commands
lint: ## Run linter
	@golangci-lint run

fmt: ## Format code
	@go fmt ./...

clean: ## Clean build artifacts
	@rm -rf bin/ coverage.out coverage.html
	@go clean -cache

secrets-rotate: ## Rotate JWT secrets
	@go run cmd/tools/rotate-secrets.go

# Database commands
db-migrate: ## Run database migrations
	@go run cmd/migrate/main.go up

db-seed: ## Seed database with test data
	@go run cmd/seed/main.go

db-reset: ## Reset database (WARNING: destroys all data)
	@go run cmd/migrate/main.go reset

# Monitoring commands
logs: ## View application logs
	docker-compose logs -f api

metrics: ## Open metrics dashboard
	@open http://localhost:9090/metrics
```

---

## 9. Security Best Practices

### Never Commit Secrets

**Git Configuration:**

```bash
# .gitignore
# Environment files
.env
.env.*
!.env.example
!.env.template

# Secret files
*.pem
*.key
*.cert
secrets/
credentials/

# Local configuration
config.local.yaml
config.dev.yaml

# IDE and OS files
.DS_Store
.idea/
.vscode/
*.swp
```

**Pre-commit Hook:**

```bash
# .git/hooks/pre-commit
#!/bin/bash

# Check for secrets in staged files
if git diff --cached --name-only | xargs grep -l "SECRET\|API_KEY\|PASSWORD\|TOKEN" 2>/dev/null; then
    echo "ERROR: Potential secrets found in staged files!"
    echo "Please remove sensitive data before committing."
    exit 1
fi

# Check for .env files
if git diff --cached --name-only | grep -q "\.env$"; then
    echo "ERROR: .env files should not be committed!"
    exit 1
fi

exit 0
```

### Secret Rotation

**Automated Rotation Script:**

```bash
#!/bin/bash
# scripts/rotate-secrets.sh

set -e

ENVIRONMENT=${1:-staging}
echo "Rotating secrets for environment: $ENVIRONMENT"

# Generate new secrets
NEW_JWT_SECRET=$(openssl rand -base64 64)
NEW_JWT_REFRESH=$(openssl rand -base64 64)

# Update AWS Secrets Manager
if [ "$ENVIRONMENT" = "production" ]; then
    aws secretsmanager put-secret-value \
        --secret-id wedding-api/production/jwt \
        --secret-string "{\"JWT_SECRET\":\"$NEW_JWT_SECRET\",\"JWT_REFRESH_SECRET\":\"$NEW_JWT_REFRESH\"}"
    
    # Trigger rolling restart of pods
    kubectl rollout restart deployment/wedding-api -n production
else
    aws secretsmanager put-secret-value \
        --secret-id wedding-api/$ENVIRONMENT/jwt \
        --secret-string "{\"JWT_SECRET\":\"$NEW_JWT_SECRET\",\"JWT_REFRESH_SECRET\":\"$NEW_JWT_REFRESH\"}"
fi

echo "Secrets rotated successfully"
echo "Old tokens will expire naturally within 15 minutes (access) or 7 days (refresh)"
```

### Least Privilege Principle

**AWS IAM Policy for Application:**

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:PutObject",
                "s3:GetObject",
                "s3:DeleteObject"
            ],
            "Resource": "arn:aws:s3:::wedding-invitation-*-uploads/*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "s3:ListBucket"
            ],
            "Resource": "arn:aws:s3:::wedding-invitation-*-uploads"
        },
        {
            "Effect": "Allow",
            "Action": [
                "secretsmanager:GetSecretValue"
            ],
            "Resource": "arn:aws:secretsmanager:*:*:secret:wedding-api/*"
        },
        {
            "Effect": "Allow",
            "Action": [
                "cloudwatch:PutMetricData"
            ],
            "Resource": "*"
        }
    ]
}
```

### Environment Variable Security

**Docker Security:**

```dockerfile
# Dockerfile
# Don't bake secrets into image
FROM golang:1.21-alpine AS builder

# Build application
WORKDIR /app
COPY . .
RUN go build -o server cmd/server/main.go

# Runtime image
FROM alpine:latest
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -u 1000 -G appuser -D appuser

WORKDIR /app

# Copy binary
COPY --from=builder /app/server .

# Don't copy .env file!
# Use environment variables or secrets management

# Switch to non-root user
USER appuser

EXPOSE 8080

CMD ["./server"]
```

### Configuration Encryption

```go
// config/encryption.go
package config

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "fmt"
    "io"
)

// EncryptedValue represents an encrypted configuration value
type EncryptedValue struct {
    Ciphertext string `json:"ciphertext"`
    Nonce      string `json:"nonce"`
}

// Encrypt encrypts a plaintext value
func Encrypt(plaintext string, key []byte) (*EncryptedValue, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
        return nil, fmt.Errorf("failed to create cipher: %w", err)
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, fmt.Errorf("failed to create GCM: %w", err)
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, fmt.Errorf("failed to generate nonce: %w", err)
    }
    
    ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
    
    return &EncryptedValue{
        Ciphertext: base64.StdEncoding.EncodeToString(ciphertext),
        Nonce:      base64.StdEncoding.EncodeToString(nonce),
    }, nil
}

// Decrypt decrypts an encrypted value
func Decrypt(enc *EncryptedValue, key []byte) (string, error) {
    ciphertext, err := base64.StdEncoding.DecodeString(enc.Ciphertext)
    if err != nil {
        return "", fmt.Errorf("failed to decode ciphertext: %w", err)
    }
    
    block, err := aes.NewCipher(key)
    if err != nil {
        return "", fmt.Errorf("failed to create cipher: %w", err)
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", fmt.Errorf("failed to create GCM: %w", err)
    }
    
    nonceSize := gcm.NonceSize()
    if len(ciphertext) < nonceSize {
        return "", fmt.Errorf("ciphertext too short")
    }
    
    nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", fmt.Errorf("failed to decrypt: %w", err)
    }
    
    return string(plaintext), nil
}
```

---

## 10. Troubleshooting

### Common Configuration Issues

#### Issue: Application fails to start with "configuration validation failed"

**Symptoms:**
```
configuration validation failed with 2 errors:
- database.uri: is required
- auth.jwt_secret: must be at least 32 characters in production
```

**Solutions:**

1. **Check environment variables are loaded:**
```bash
# Verify env vars are set
env | grep -E "(MONGODB_URI|JWT_SECRET)"

# If using .env file, ensure it's loaded
export $(cat .env | xargs)
```

2. **Validate .env file format:**
```bash
# Check for syntax errors
bash -c 'export $(cat .env | xargs) && env | grep MONGODB'

# Look for common issues:
# - Spaces around = (should be VAR=value, not VAR = value)
# - Quotes inside values (use VAR="value" not VAR=value with spaces)
# - Special characters not escaped
```

3. **Check file permissions:**
```bash
# Ensure .env is readable
chmod 600 .env
ls -la .env
```

#### Issue: "connection refused" to database

**Symptoms:**
```
Failed to connect to database: connection refused
Health check failed: database health check failed: connection refused
```

**Solutions:**

1. **Verify MongoDB is running:**
```bash
# Check MongoDB status
mongosh --eval "db.adminCommand('ping')"

# Or using docker
docker-compose ps mongo
docker-compose logs mongo | tail -20
```

2. **Check connection string:**
```bash
# Test connection manually
mongosh "mongodb://localhost:27017/wedding_invitation_dev"

# Verify host and port
netstat -tlnp | grep 27017
telnet localhost 27017
```

3. **Common connection string fixes:**
```bash
# Local development
MONGODB_URI=mongodb://localhost:27017

# Docker network (from container to host)
MONGODB_URI=mongodb://host.docker.internal:27017

# Docker compose (service name)
MONGODB_URI=mongodb://mongo:27017/wedding_invitation
```

#### Issue: JWT authentication fails in production

**Symptoms:**
```
Token validation failed: signature is invalid
401 Unauthorized errors on all protected endpoints
```

**Solutions:**

1. **Verify secrets are set and long enough:**
```bash
# Check secret length
echo -n "$JWT_SECRET" | wc -c  # Should be >= 32

# Generate new secrets if needed
export JWT_SECRET=$(openssl rand -base64 64)
export JWT_REFRESH_SECRET=$(openssl rand -base64 64)
```

2. **Check for character encoding issues:**
```bash
# Ensure no special characters causing issues
printf '%s' "$JWT_SECRET" | od -c

# Regenerate if needed
export JWT_SECRET=$(openssl rand -hex 32)
```

3. **Verify secrets are the same across all instances:**
```bash
# In Kubernetes, check all pods have same secret
kubectl get secret wedding-api-secrets -o yaml
kubectl rollout restart deployment/wedding-api
```

#### Issue: CORS errors in browser

**Symptoms:**
```
Access to fetch at 'http://api.example.com/...' from origin 'http://localhost:3000' 
has been blocked by CORS policy
```

**Solutions:**

1. **Development - Allow all origins:**
```bash
CORS_ALLOWED_ORIGINS=*
```

2. **Production - Specify exact origins:**
```bash
CORS_ALLOWED_ORIGINS=https://app.example.com,https://admin.example.com
```

3. **Verify protocol and port match:**
```bash
# If frontend is on https://app.example.com:3000
CORS_ALLOWED_ORIGINS=https://app.example.com:3000
```

#### Issue: Rate limiting too aggressive

**Symptoms:**
```
429 Too Many Requests errors
API becomes unusable during normal operation
```

**Solutions:**

1. **Increase rate limits:**
```bash
# Development
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_WINDOW=1m

# Production (adjust based on monitoring)
RATE_LIMIT_REQUESTS=500
RATE_LIMIT_BURST=50
```

2. **Check for client retries:**
```bash
# Monitor rate limit hits in logs
grep "rate limit" app.log | tail -20
```

3. **Exclude health checks from rate limiting:**
```go
// In middleware setup
if c.Request.URL.Path == "/health" {
    c.Next()
    return
}
```

### Debug Commands

#### Check Current Configuration

```bash
# Show effective configuration
./server --show-config

# Show as JSON
./server --show-config --format=json

# Show specific section
./server --show-config --section=database
```

#### Validate Configuration Without Starting

```bash
# Quick validation
./server --check-config

# Verbose validation with all warnings
./server --check-config --verbose

# Validate specific environment
APP_ENV=production ./server --check-config
```

#### Environment Variable Debugging

```bash
# List all env vars seen by application
./server --dump-env

# Check if specific variable is set
./server --check-var=MONGODB_URI

# Show environment variable mappings
./server --show-mappings
```

### Configuration Testing

```go
// config/config_test.go
package config

import (
    "os"
    "testing"
    "time"
)

func TestLoad(t *testing.T) {
    // Set required test env vars
    os.Setenv("APP_ENV", "test")
    os.Setenv("MONGODB_URI", "mongodb://localhost:27017")
    os.Setenv("MONGODB_DATABASE", "test_db")
    
    cfg, err := Load()
    if err != nil {
        t.Fatalf("Failed to load config: %v", err)
    }
    
    if cfg.Server.Environment != "test" {
        t.Errorf("Expected environment 'test', got '%s'", cfg.Server.Environment)
    }
}

func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        setup   func()
        wantErr bool
    }{
        {
            name: "valid production config",
            setup: func() {
                os.Setenv("APP_ENV", "production")
                os.Setenv("MONGODB_URI", "mongodb://localhost:27017")
                os.Setenv("MONGODB_DATABASE", "prod")
                os.Setenv("JWT_SECRET", "this-is-a-32-char-secret-for-prod!!")
                os.Setenv("JWT_REFRESH_SECRET", "this-is-a-32-char-refresh-for-prod!!")
            },
            wantErr: false,
        },
        {
            name: "missing database URI",
            setup: func() {
                os.Setenv("APP_ENV", "development")
                os.Unsetenv("MONGODB_URI")
                os.Setenv("MONGODB_DATABASE", "test")
            },
            wantErr: true,
        },
        {
            name: "short JWT secret in production",
            setup: func() {
                os.Setenv("APP_ENV", "production")
                os.Setenv("MONGODB_URI", "mongodb://localhost:27017")
                os.Setenv("MONGODB_DATABASE", "prod")
                os.Setenv("JWT_SECRET", "short")
                os.Setenv("JWT_REFRESH_SECRET", "this-is-a-32-char-refresh-for-prod!!")
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Clear env
            os.Clearenv()
            
            // Setup test case
            tt.setup()
            
            // Reset global config
            globalConfig = nil
            configOnce = sync.Once{}
            
            cfg, err := Load()
            
            if tt.wantErr {
                if err == nil {
                    t.Errorf("Expected error, got nil")
                }
                return
            }
            
            if err != nil {
                t.Errorf("Unexpected error: %v", err)
                return
            }
            
            // Validate loaded config
            validator := NewValidator()
            if err := validator.Validate(cfg); err != nil {
                t.Errorf("Validation failed: %v", err)
            }
        })
    }
}
```

### Log Analysis

```bash
# Find configuration-related errors
grep -E "(config|env|secret)" app.log | grep -i error

# Monitor configuration changes
grep "Configuration" app.log

# Check startup sequence
grep -A 20 "Server starting" app.log
```

---

## Summary

This guide provides a comprehensive approach to environment configuration management for the Wedding Invitation API:

1. **12-Factor App compliance** ensures proper separation of config and code
2. **Viper integration** provides flexible, type-safe configuration management
3. **Environment-specific files** support development, staging, and production
4. **Secrets management** options cover local development through production
5. **Validation** ensures configuration correctness at startup
6. **Security best practices** protect sensitive data
7. **Troubleshooting** section helps resolve common issues

Always follow these principles:
- Never commit secrets to version control
- Use strong, unique secrets for each environment
- Rotate secrets regularly
- Validate configuration before starting the application
- Monitor and alert on configuration errors
- Use least privilege access for production credentials
