# Frontend Integration Guide

## Table of Contents

1. [Integration Overview](#integration-overview)
2. [API Client Architecture](#api-client-architecture)
3. [Authentication Integration](#authentication-integration)
4. [Current Frontend Changes Needed](#current-frontend-changes-needed)
5. [JavaScript API Client Class](#javascript-api-client-class)
6. [API Endpoints Mapping](#api-endpoints-mapping)
7. [Error Handling](#error-handling)
8. [Loading States and UX](#loading-states-and-ux)
9. [Real-World Examples](#real-world-examples)
10. [Security Considerations](#security-considerations)
11. [Testing the Integration](#testing-the-integration)
12. [Complete Code Examples](#complete-code-examples)

---

## Integration Overview

This guide covers integrating the wedding invitation frontend with the Go backend API. The current frontend uses `localStorage` for data persistence, which needs to be replaced with API calls to enable multi-user support, cloud storage, and real-time collaboration.

### Current Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     CURRENT (Static)                             │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐        │
│  │   HTML      │────▶│ JavaScript  │────▶│ localStorage│        │
│  │   (Forms)   │     │ (Logic)     │     │ (Storage)   │        │
│  └─────────────┘     └─────────────┘     └─────────────┘        │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Target Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     TARGET (API-Driven)                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────┐     ┌─────────────┐     ┌───────────────────┐  │
│  │   HTML      │────▶│  API Client │────▶│   Go Backend API  │  │
│  │   (Forms)   │     │ (HTTP/JSON) │     │  (REST/HTTPS)     │  │
│  └─────────────┘     └─────────────┘     └───────────────────┘  │
│                              │                                    │
│                              ▼                                    │
│                       ┌─────────────┐                            │
│                       │ localStorage│ (Token cache only)         │
│                       │ (Minimal)   │                            │
│                       └─────────────┘                            │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Key Changes

| Aspect | Current | Target |
|--------|---------|--------|
| **Storage** | localStorage for all data | MongoDB via API |
| **Auth** | None | JWT with refresh tokens |
| **User Data** | Shared browser data | User-specific |
| **Images** | URLs only | File uploads to S3/R2 |
| **RSVP** | Alert only | Saved to database |
| **Collaboration** | None | Multi-user support |

---

## API Client Architecture

### Base URL Configuration

The API client supports multiple environments through configuration:

```javascript
// api-client.js - Configuration
const API_CONFIG = {
  development: {
    baseURL: 'http://localhost:8080/api/v1',
    timeout: 30000,
    retries: 3
  },
  staging: {
    baseURL: 'https://api-staging.wedding-app.com/api/v1',
    timeout: 30000,
    retries: 3
  },
  production: {
    baseURL: 'https://api.wedding-app.com/api/v1',
    timeout: 30000,
    retries: 3
  }
};

// Auto-detect environment
const ENV = window.location.hostname === 'localhost' ? 'development' : 
            window.location.hostname.includes('staging') ? 'staging' : 'production';

const config = API_CONFIG[ENV];
```

### Fetch vs Axios

This implementation uses native `fetch` API for zero dependencies, but includes an Axios adapter pattern for easy migration:

**Why Native Fetch:**
- Zero dependencies (no bundle size increase)
- Native browser support (all modern browsers)
- Streaming support for large file uploads
- Modern Promise-based API

**When to Use Axios:**
- If you need request/response interceptors (can be added to fetch)
- If you need automatic JSON transforms (handled manually here)
- If you need request cancellation (AbortController available)
- If you need older browser support (IE11)

### Interceptors Architecture

```javascript
// Interceptor pattern for fetch
class APIClient {
  constructor(config) {
    this.config = config;
    this.requestInterceptors = [];
    this.responseInterceptors = [];
    this.errorInterceptors = [];
  }

  // Add request interceptor
  addRequestInterceptor(interceptor) {
    this.requestInterceptors.push(interceptor);
  }

  // Add response interceptor
  addResponseInterceptor(interceptor) {
    this.responseInterceptors.push(interceptor);
  }

  // Add error interceptor
  addErrorInterceptor(interceptor) {
    this.errorInterceptors.push(interceptor);
  }

  async executeRequest(url, options) {
    // 1. Apply request interceptors
    let modifiedOptions = options;
    for (const interceptor of this.requestInterceptors) {
      modifiedOptions = await interceptor(modifiedOptions);
    }

    // 2. Make request
    let response;
    try {
      response = await fetch(url, modifiedOptions);
    } catch (error) {
      // Network errors
      return this.handleNetworkError(error);
    }

    // 3. Apply response interceptors
    for (const interceptor of this.responseInterceptors) {
      response = await interceptor(response);
    }

    // 4. Handle errors
    if (!response.ok) {
      return this.handleHTTPError(response);
    }

    return response;
  }
}
```

### Auth Interceptor

Automatically adds JWT token to requests:

```javascript
// Auth interceptor implementation
const authInterceptor = async (options) => {
  const token = TokenManager.getAccessToken();
  
  if (token) {
    options.headers = {
      ...options.headers,
      'Authorization': `Bearer ${token}`
    };
  }
  
  return options;
};
```

### Error Interceptor

Handles token refresh on 401 errors:

```javascript
// Error interceptor with token refresh
const errorInterceptor = async (response) => {
  if (response.status === 401) {
    // Try to refresh token
    const refreshed = await TokenManager.refreshAccessToken();
    
    if (refreshed) {
      // Retry original request with new token
      const token = TokenManager.getAccessToken();
      const newOptions = {
        ...response.requestOptions,
        headers: {
          ...response.requestOptions.headers,
          'Authorization': `Bearer ${token}`
        }
      };
      return fetch(response.url, newOptions);
    } else {
      // Refresh failed, redirect to login
      window.location.href = '/login.html';
    }
  }
  
  return response;
};
```

---

## Authentication Integration

### JWT Storage Strategy

**Recommended Approach: HTTP-only Cookies + Memory Storage**

The backend uses HTTP-only cookies for tokens (most secure), but we still need to track authentication state:

```javascript
// Token Manager - Handles JWT operations
class TokenManager {
  constructor() {
    // Store minimal auth state (not tokens!)
    this.isAuthenticated = false;
    this.user = null;
    this.tokenExpiry = null;
  }

  // Check if user is authenticated
  isLoggedIn() {
    return this.isAuthenticated;
  }

  // Get user info (from memory or API)
  getUser() {
    return this.user;
  }

  // Set auth state after login
  setAuthenticated(userData) {
    this.isAuthenticated = true;
    this.user = userData;
    // Store minimal info in localStorage for persistence
    localStorage.setItem('auth_state', JSON.stringify({
      isAuthenticated: true,
      userId: userData.id
    }));
  }

  // Clear auth state on logout
  clearAuth() {
    this.isAuthenticated = false;
    this.user = null;
    localStorage.removeItem('auth_state');
    // Call logout API to clear cookies
  }

  // Check if token needs refresh (5 min before expiry)
  shouldRefreshToken() {
    if (!this.tokenExpiry) return false;
    const fiveMinutes = 5 * 60 * 1000;
    return Date.now() > (this.tokenExpiry - fiveMinutes);
  }
}

// Singleton instance
const tokenManager = new TokenManager();
```

**Why NOT localStorage for tokens:**
- XSS vulnerability: JavaScript can read localStorage
- No automatic cookie transmission
- Must manually attach to every request
- No SameSite protection

### Login Form Integration

```javascript
// Login form handler (login.html)
document.getElementById('loginForm').addEventListener('submit', async (e) => {
  e.preventDefault();
  
  const email = document.getElementById('email').value;
  const password = document.getElementById('password').value;
  const rememberMe = document.getElementById('rememberMe').checked;
  
  // Show loading state
  const submitBtn = e.target.querySelector('button[type="submit"]');
  const originalText = submitBtn.textContent;
  submitBtn.disabled = true;
  submitBtn.textContent = 'Logging in...';
  
  try {
    const response = await apiClient.post('/auth/login', {
      email,
      password,
      device_info: navigator.userAgent
    });
    
    // Backend sets HTTP-only cookies automatically
    // Store minimal auth state
    tokenManager.setAuthenticated(response.user);
    
    // Redirect to dashboard
    const redirectUrl = new URLSearchParams(window.location.search).get('redirect') || '/dashboard.html';
    window.location.href = redirectUrl;
    
  } catch (error) {
    // Show error message
    showError(error.message || 'Login failed. Please check your credentials.');
    submitBtn.disabled = false;
    submitBtn.textContent = originalText;
  }
});
```

### Registration Form Integration

```javascript
// Registration form handler (register.html)
document.getElementById('registerForm').addEventListener('submit', async (e) => {
  e.preventDefault();
  
  const data = {
    email: document.getElementById('email').value,
    password: document.getElementById('password').value,
    name: document.getElementById('name').value,
    device_info: navigator.userAgent
  };
  
  // Validate password strength
  if (!validatePassword(data.password)) {
    showError('Password must be at least 8 characters with uppercase, lowercase, number, and special character.');
    return;
  }
  
  try {
    const response = await apiClient.post('/auth/register', data);
    
    // Show success message
    showSuccess('Registration successful! Please check your email to verify your account.');
    
    // Redirect to verification pending page
    setTimeout(() => {
      window.location.href = '/verify-pending.html?email=' + encodeURIComponent(data.email);
    }, 2000);
    
  } catch (error) {
    showError(error.message || 'Registration failed. Please try again.');
  }
});
```

### Token Refresh Logic

```javascript
// Automatic token refresh
class TokenRefresher {
  constructor(apiClient) {
    this.apiClient = apiClient;
    this.refreshPromise = null;
  }

  // Refresh access token using refresh token cookie
  async refreshAccessToken() {
    // Prevent multiple concurrent refresh requests
    if (this.refreshPromise) {
      return this.refreshPromise;
    }

    this.refreshPromise = this.doRefresh();
    
    try {
      const result = await this.refreshPromise;
      return result;
    } finally {
      this.refreshPromise = null;
    }
  }

  async doRefresh() {
    try {
      const response = await fetch(`${API_CONFIG.baseURL}/auth/refresh`, {
        method: 'POST',
        credentials: 'include', // Important: sends cookies
        headers: {
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        throw new Error('Token refresh failed');
      }

      const data = await response.json();
      
      // Update auth state
      tokenManager.setAuthenticated(data.user);
      
      return true;
    } catch (error) {
      console.error('Token refresh failed:', error);
      tokenManager.clearAuth();
      return false;
    }
  }

  // Setup automatic refresh interval
  startAutoRefresh(intervalMinutes = 10) {
    setInterval(() => {
      if (tokenManager.isLoggedIn() && tokenManager.shouldRefreshToken()) {
        this.refreshAccessToken();
      }
    }, intervalMinutes * 60 * 1000);
  }
}

const tokenRefresher = new TokenRefresher(apiClient);
```

### Logout Handling

```javascript
// Logout handler
async function logout() {
  try {
    // Call logout API to invalidate tokens
    await apiClient.post('/auth/logout', {}, { credentials: 'include' });
  } catch (error) {
    console.error('Logout API call failed:', error);
    // Continue with local logout even if API fails
  }
  
  // Clear local auth state
  tokenManager.clearAuth();
  
  // Redirect to home
  window.location.href = '/';
}

// Attach to logout button
document.getElementById('logoutBtn')?.addEventListener('click', logout);
```

---

## Current Frontend Changes Needed

### 1. Replace localStorage with API Calls

**Current Pattern (localStorage):**
```javascript
// generator.js - Current
function saveWeddingData(key, data) {
  localStorage.setItem(`wedding-${key}`, JSON.stringify(data));
  
  // Update wedding list
  const list = JSON.parse(localStorage.getItem('wedding-list') || '[]');
  if (!list.includes(key)) {
    list.push(key);
    localStorage.setItem('wedding-list', JSON.stringify(list));
  }
}
```

**New Pattern (API):**
```javascript
// api-client.js - New
async function createWedding(weddingData) {
  const response = await apiClient.post('/weddings', {
    couple_name: `${weddingData.groomName} & ${weddingData.brideName}`,
    slug: generateSlug(weddingData),
    event_date: new Date(weddingData.weddingDate),
    venue: {
      name: weddingData.venueName,
      address: weddingData.venueAddress
    },
    theme: weddingData.theme,
    rsvp_config: {
      deadline: new Date(weddingData.rsvpDate),
      collect_email: true
    },
    // ... other fields
  });
  
  return response; // Returns { id, slug, ... }
}
```

### 2. Update generator.html

**Changes needed:**

```javascript
// In generator.js, replace the generateInvitation function:

async function generateInvitation() {
  const formData = collectFormData();
  
  // Show loading overlay
  document.getElementById('loadingOverlay').classList.add('active');
  
  try {
    // 1. Create wedding first
    const wedding = await apiClient.post('/weddings', {
      couple_name: `${formData.groomName} & ${formData.brideName}`,
      event_date: formData.weddingDate,
      venue: {
        name: formData.venueName,
        address: formData.venueAddress
      },
      theme: formData.theme,
      story: formData.loveStory,
      rsvp_config: {
        deadline: formData.rsvpDate,
        collect_email: true
      }
    });
    
    // 2. Upload images if any
    if (formData.galleryImages.length > 0) {
      await uploadGalleryImages(wedding.id, formData.galleryImages);
    }
    
    // 3. Create events
    for (const event of formData.events) {
      await apiClient.post(`/weddings/${wedding.id}/events`, event);
    }
    
    // 4. Show success
    showGeneratedResult(wedding);
    
  } catch (error) {
    showError('Failed to create invitation: ' + error.message);
  } finally {
    document.getElementById('loadingOverlay').classList.remove('active');
  }
}
```

### 3. Update manage.html

**Replace localStorage-based listing:**

```javascript
// In manage.html, replace the loadInvitations function:

async function loadInvitations() {
  const list = document.getElementById('invitationList');
  
  // Show loading state
  list.innerHTML = '<div class="loading">Loading invitations...</div>';
  
  try {
    // Fetch from API instead of localStorage
    const weddings = await apiClient.get('/weddings');
    
    if (weddings.length === 0) {
      list.innerHTML = `
        <div class="empty-state">
          <div class="empty-state-icon">💕</div>
          <h2>No Invitations Yet</h2>
          <p>Create your first beautiful wedding invitation</p>
          <a href="generator.html" class="btn-create">Create Invitation</a>
        </div>
      `;
      return;
    }
    
    list.innerHTML = weddings.map(wedding => {
      const weddingDate = wedding.event_date ? 
        new Date(wedding.event_date).toLocaleDateString() : 'Date TBA';
      
      return `
        <div class="invitation-card" data-id="${wedding.id}">
          <div class="invitation-info">
            <h3>${wedding.couple_name}</h3>
            <p>${weddingDate}</p>
            <span class="invitation-theme">${formatThemeName(wedding.theme)}</span>
          </div>
          <div class="invitation-actions">
            <a href="/w/${wedding.slug}" class="btn btn-view" target="_blank">View</a>
            <a href="/edit.html?id=${wedding.id}" class="btn btn-edit">Edit</a>
            <button class="btn btn-delete" onclick="deleteInvitation('${wedding.id}')">Delete</button>
          </div>
        </div>
      `;
    }).join('');
    
  } catch (error) {
    list.innerHTML = `
      <div class="error-state">
        <p>Failed to load invitations: ${error.message}</p>
        <button onclick="loadInvitations()" class="btn">Retry</button>
      </div>
    `;
  }
}

async function deleteInvitation(id) {
  if (!confirm('Are you sure you want to delete this invitation?')) return;
  
  try {
    await apiClient.delete(`/weddings/${id}`);
    loadInvitations(); // Refresh list
  } catch (error) {
    alert('Failed to delete: ' + error.message);
  }
}
```

### 4. Update RSVP Forms in Themes

**Replace alert-only RSVP with API submission:**

```javascript
// In each theme's script.js, replace RSVP handler:

// Old (alert only):
// onsubmit="event.preventDefault(); alert('Thank you for RSVPing!');"

// New (API submission):
document.getElementById('rsvpForm').addEventListener('submit', async (e) => {
  e.preventDefault();
  
  const submitBtn = e.target.querySelector('button[type="submit"]');
  const originalText = submitBtn.textContent;
  
  // Get wedding slug from URL
  const slug = window.location.pathname.split('/').pop();
  
  const rsvpData = {
    guest_name: document.getElementById('guestName').value,
    email: document.getElementById('guestEmail').value,
    attending: document.getElementById('rsvpStatus').value === 'yes',
    dietary_restrictions: document.getElementById('dietary')?.value || '',
    plus_one: document.getElementById('plusOne')?.checked || false,
    plus_one_name: document.getElementById('plusOneName')?.value || '',
    message: document.getElementById('guestMessage')?.value || ''
  };
  
  submitBtn.disabled = true;
  submitBtn.textContent = 'Submitting...';
  
  try {
    await apiClient.post(`/weddings/${slug}/rsvp`, rsvpData);
    
    // Show success
    showRSVPSuccess();
    
    // Clear form
    e.target.reset();
    
  } catch (error) {
    showRSVPError(error.message || 'Failed to submit RSVP. Please try again.');
  } finally {
    submitBtn.disabled = false;
    submitBtn.textContent = originalText;
  }
});
```

---

## JavaScript API Client Class

### Complete Implementation

```javascript
/**
 * Wedding API Client
 * 
 * A production-ready HTTP client for the Wedding Invitation API.
 * Features:
 * - Automatic authentication
 * - Token refresh
 * - Retry logic with exponential backoff
 * - Request/response interceptors
 * - File upload support
 * - Error handling
 */

class WeddingAPIClient {
  constructor(config = {}) {
    this.config = {
      baseURL: config.baseURL || 'http://localhost:8080/api/v1',
      timeout: config.timeout || 30000,
      retries: config.retries || 3,
      retryDelay: config.retryDelay || 1000,
      credentials: config.credentials || 'include', // Send cookies
      ...config
    };

    this.requestInterceptors = [];
    this.responseInterceptors = [];
    this.errorInterceptors = [];
    this.isRefreshing = false;
    this.refreshSubscribers = [];

    // Setup default interceptors
    this.setupDefaultInterceptors();
  }

  // Setup default interceptors
  setupDefaultInterceptors() {
    // Auth interceptor - adds token to requests
    this.addRequestInterceptor(async (config) => {
      // Tokens are in HTTP-only cookies, but we can add additional headers if needed
      config.headers = {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        ...config.headers
      };
      return config;
    });

    // Error interceptor - handles 401 and token refresh
    this.addErrorInterceptor(async (error) => {
      const { response, requestConfig } = error;

      // Handle 401 Unauthorized
      if (response && response.status === 401 && !requestConfig._retry) {
        return this.handleTokenRefresh(requestConfig);
      }

      throw error;
    });
  }

  // Add request interceptor
  addRequestInterceptor(interceptor) {
    this.requestInterceptors.push(interceptor);
  }

  // Add response interceptor
  addResponseInterceptor(interceptor) {
    this.responseInterceptors.push(interceptor);
  }

  // Add error interceptor
  addErrorInterceptor(interceptor) {
    this.errorInterceptors.push(interceptor);
  }

  // HTTP GET
  async get(endpoint, options = {}) {
    return this.request(endpoint, { ...options, method: 'GET' });
  }

  // HTTP POST
  async post(endpoint, data, options = {}) {
    return this.request(endpoint, { 
      ...options, 
      method: 'POST', 
      body: data 
    });
  }

  // HTTP PUT
  async put(endpoint, data, options = {}) {
    return this.request(endpoint, { 
      ...options, 
      method: 'PUT', 
      body: data 
    });
  }

  // HTTP PATCH
  async patch(endpoint, data, options = {}) {
    return this.request(endpoint, { 
      ...options, 
      method: 'PATCH', 
      body: data 
    });
  }

  // HTTP DELETE
  async delete(endpoint, options = {}) {
    return this.request(endpoint, { ...options, method: 'DELETE' });
  }

  // File upload with progress
  async upload(endpoint, file, onProgress = null, options = {}) {
    const formData = new FormData();
    formData.append('file', file);

    // Add any additional fields
    if (options.fields) {
      Object.keys(options.fields).forEach(key => {
        formData.append(key, options.fields[key]);
      });
    }

    const fetchOptions = {
      method: 'POST',
      body: formData,
      credentials: this.config.credentials,
      // Don't set Content-Type - browser will set with boundary
    };

    // Apply request interceptors
    let modifiedOptions = fetchOptions;
    for (const interceptor of this.requestInterceptors) {
      modifiedOptions = await interceptor(modifiedOptions);
    }

    // Remove Content-Type if set by interceptor (FormData needs browser to set it)
    if (modifiedOptions.headers && modifiedOptions.headers['Content-Type']) {
      delete modifiedOptions.headers['Content-Type'];
    }

    const url = `${this.config.baseURL}${endpoint}`;
    
    try {
      const response = await fetch(url, modifiedOptions);
      return this.handleResponse(response);
    } catch (error) {
      throw this.handleError(error);
    }
  }

  // Main request method with retry logic
  async request(endpoint, options = {}) {
    const url = `${this.config.baseURL}${endpoint}`;
    const attemptRequest = async (attempt = 1) => {
      try {
        // Build request config
        let requestConfig = {
          method: options.method || 'GET',
          credentials: this.config.credentials,
          headers: options.headers || {},
          signal: options.signal
        };

        // Add body for non-GET requests
        if (options.body && options.method !== 'GET') {
          if (options.body instanceof FormData) {
            // Don't set Content-Type for FormData
            requestConfig.body = options.body;
          } else {
            requestConfig.body = JSON.stringify(options.body);
            requestConfig.headers['Content-Type'] = 'application/json';
          }
        }

        // Apply request interceptors
        for (const interceptor of this.requestInterceptors) {
          requestConfig = await interceptor(requestConfig);
        }

        // Make request with timeout
        const response = await this.fetchWithTimeout(url, requestConfig);

        // Handle 204 No Content
        if (response.status === 204) {
          return null;
        }

        // Apply response interceptors
        let modifiedResponse = response;
        for (const interceptor of this.responseInterceptors) {
          modifiedResponse = await interceptor(modifiedResponse);
        }

        // Check for HTTP errors
        if (!modifiedResponse.ok) {
          const error = await this.createHTTPError(modifiedResponse);
          error.requestConfig = requestConfig;
          
          // Apply error interceptors
          for (const interceptor of this.errorInterceptors) {
            try {
              return await interceptor(error);
            } catch (e) {
              // Interceptor re-threw, continue to next
            }
          }
          throw error;
        }

        return await this.handleResponse(modifiedResponse);

      } catch (error) {
        // Don't retry on client errors (4xx)
        if (error.status >= 400 && error.status < 500) {
          throw error;
        }

        // Retry on network errors or server errors (5xx)
        if (attempt < this.config.retries) {
          const delay = this.config.retryDelay * Math.pow(2, attempt - 1);
          await this.sleep(delay);
          return attemptRequest(attempt + 1);
        }

        throw error;
      }
    };

    return attemptRequest();
  }

  // Fetch with timeout
  async fetchWithTimeout(url, options) {
    const controller = new AbortController();
    const id = setTimeout(() => controller.abort(), this.config.timeout);

    try {
      const response = await fetch(url, {
        ...options,
        signal: controller.signal
      });
      return response;
    } finally {
      clearTimeout(id);
    }
  }

  // Handle token refresh
  async handleTokenRefresh(requestConfig) {
    // Prevent multiple refresh requests
    if (this.isRefreshing) {
      return new Promise((resolve, reject) => {
        this.refreshSubscribers.push({ resolve, reject, requestConfig });
      });
    }

    this.isRefreshing = true;

    try {
      const refreshResponse = await fetch(`${this.config.baseURL}/auth/refresh`, {
        method: 'POST',
        credentials: 'include'
      });

      if (!refreshResponse.ok) {
        throw new Error('Token refresh failed');
      }

      // Retry original request
      this.refreshSubscribers.forEach(({ resolve, reject, requestConfig }) => {
        this.request(requestConfig.url, requestConfig)
          .then(resolve)
          .catch(reject);
      });

    } catch (error) {
      // Refresh failed, redirect to login
      this.refreshSubscribers.forEach(({ reject }) => reject(error));
      window.location.href = '/login.html?redirect=' + encodeURIComponent(window.location.pathname);
    } finally {
      this.isRefreshing = false;
      this.refreshSubscribers = [];
    }
  }

  // Handle response parsing
  async handleResponse(response) {
    const contentType = response.headers.get('content-type');
    
    if (contentType && contentType.includes('application/json')) {
      return await response.json();
    }
    
    return await response.text();
  }

  // Create HTTP error from response
  async createHTTPError(response) {
    let errorData;
    try {
      errorData = await response.json();
    } catch {
      errorData = { error: response.statusText };
    }

    const error = new Error(errorData.error || errorData.message || `HTTP ${response.status}`);
    error.status = response.status;
    error.statusText = response.statusText;
    error.data = errorData;
    error.response = response;

    return error;
  }

  // Handle network/fetch errors
  handleError(error) {
    if (error.name === 'AbortError') {
      return new Error('Request timeout');
    }
    
    if (error.message === 'Failed to fetch') {
      return new Error('Network error. Please check your connection.');
    }

    return error;
  }

  // Utility: Sleep
  sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

// Create and export singleton instance
const apiClient = new WeddingAPIClient({
  baseURL: window.location.hostname === 'localhost' 
    ? 'http://localhost:8080/api/v1'
    : 'https://api.wedding-app.com/api/v1'
});

// Export for module systems
if (typeof module !== 'undefined' && module.exports) {
  module.exports = { WeddingAPIClient, apiClient };
}
```

---

## API Endpoints Mapping

### Authentication Endpoints

| Method | Endpoint | Description | Request Body | Response |
|--------|----------|-------------|--------------|----------|
| POST | `/auth/register` | Register new user | `{email, password, name, device_info}` | `{user}` |
| POST | `/auth/login` | Login user | `{email, password, device_info}` | `{user}` |
| POST | `/auth/refresh` | Refresh access token | - | `{message}` |
| POST | `/auth/logout` | Logout user | - | `{message}` |
| POST | `/auth/verify-email` | Verify email | `{token}` | `{message}` |
| POST | `/auth/forgot-password` | Request password reset | `{email}` | `{message}` |
| POST | `/auth/reset-password` | Reset password | `{token, new_password}` | `{message}` |
| POST | `/auth/logout-all` | Logout all sessions | - | `{message}` |

### Wedding CRUD Endpoints

| Method | Endpoint | Description | Auth | Request Body |
|--------|----------|-------------|------|--------------|
| GET | `/weddings` | List user's weddings | Yes | - |
| POST | `/weddings` | Create new wedding | Yes | `{couple_name, event_date, venue, theme, ...}` |
| GET | `/weddings/:id` | Get wedding by ID | Yes | - |
| GET | `/weddings/slug/:slug` | Get wedding by slug (public) | No | - |
| PUT | `/weddings/:id` | Update wedding | Yes | `{...fields}` |
| PATCH | `/weddings/:id` | Partial update | Yes | `{...fields}` |
| DELETE | `/weddings/:id` | Delete wedding | Yes | - |

### Wedding Sub-resources

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/weddings/:id/events` | List wedding events | Yes |
| POST | `/weddings/:id/events` | Add event | Yes |
| PUT | `/weddings/:id/events/:eventId` | Update event | Yes |
| DELETE | `/weddings/:id/events/:eventId` | Delete event | Yes |
| GET | `/weddings/:id/guests` | List guests | Yes |
| POST | `/weddings/:id/guests` | Add guest | Yes |
| POST | `/weddings/:id/guests/import` | Bulk import guests | Yes |
| GET | `/weddings/:id/gallery` | List gallery images | Yes/No |
| POST | `/weddings/:id/gallery` | Upload image | Yes |
| DELETE | `/weddings/:id/gallery/:imageId` | Delete image | Yes |

### RSVP Endpoints

| Method | Endpoint | Description | Auth | Request Body |
|--------|----------|-------------|------|--------------|
| POST | `/weddings/:slug/rsvp` | Submit RSVP | No | `{guest_name, email, attending, ...}` |
| POST | `/rsvp/public` | Public RSVP (with code) | No | `{wedding_code, ...}` |
| GET | `/weddings/:id/rsvps` | List RSVPs | Yes | - |
| GET | `/weddings/:id/rsvp-stats` | Get RSVP statistics | Yes | - |

### File Upload Endpoints

| Method | Endpoint | Description | Auth | Content-Type |
|--------|----------|-------------|------|--------------|
| POST | `/upload/image` | Upload image | Yes | `multipart/form-data` |
| POST | `/upload/cover` | Upload cover image | Yes | `multipart/form-data` |
| POST | `/upload/gallery` | Upload to gallery | Yes | `multipart/form-data` |
| DELETE | `/upload/:key` | Delete uploaded file | Yes | - |

### User Endpoints

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/me` | Get current user | Yes |
| PUT | `/me` | Update profile | Yes |
| PUT | `/me/password` | Change password | Yes |
| DELETE | `/me` | Delete account | Yes |

---

## Error Handling

### HTTP Status Codes

| Code | Meaning | Frontend Action |
|------|---------|-----------------|
| 200 | OK | Continue |
| 201 | Created | Show success, redirect |
| 204 | No Content | Continue (no body) |
| 400 | Bad Request | Show validation errors |
| 401 | Unauthorized | Redirect to login |
| 403 | Forbidden | Show "access denied" |
| 404 | Not Found | Show "not found" page |
| 409 | Conflict | Show conflict message |
| 422 | Unprocessable | Show validation errors |
| 429 | Too Many Requests | Show retry later |
| 500 | Server Error | Show error, retry |
| 502/503 | Service Unavailable | Show maintenance |

### Error Response Format

```json
{
  "error": "Human-readable error message",
  "code": "ERROR_CODE",
  "details": {
    "field": "Specific field error"
  },
  "request_id": "req_12345"
}
```

### Frontend Error Handling Pattern

```javascript
// Error handling utility
class ErrorHandler {
  static handle(error, context = '') {
    console.error(`[${context}]`, error);

    // Network errors
    if (!error.status) {
      return {
        type: 'network',
        message: 'Unable to connect. Please check your internet connection.',
        action: 'retry'
      };
    }

    // Authentication errors
    if (error.status === 401) {
      return {
        type: 'auth',
        message: 'Your session has expired. Please log in again.',
        action: 'redirect',
        redirectUrl: '/login.html'
      };
    }

    // Validation errors
    if (error.status === 400 || error.status === 422) {
      return {
        type: 'validation',
        message: error.data?.error || 'Please check your input and try again.',
        errors: error.data?.details || {},
        action: 'show_fields'
      };
    }

    // Rate limiting
    if (error.status === 429) {
      return {
        type: 'rate_limit',
        message: 'Too many requests. Please wait a moment and try again.',
        action: 'retry_with_delay',
        retryAfter: error.data?.retry_after || 60
      };
    }

    // Server errors
    if (error.status >= 500) {
      return {
        type: 'server',
        message: 'Something went wrong on our end. Please try again later.',
        action: 'retry'
      };
    }

    // Default
    return {
      type: 'unknown',
      message: error.message || 'An unexpected error occurred.',
      action: 'notify'
    };
  }

  static showError(container, errorInfo) {
    const errorContainer = typeof container === 'string' 
      ? document.getElementById(container) 
      : container;

    if (!errorContainer) {
      alert(errorInfo.message);
      return;
    }

    errorContainer.innerHTML = `
      <div class="error-message error-${errorInfo.type}">
        <span class="error-icon">⚠️</span>
        <span class="error-text">${errorInfo.message}</span>
        ${errorInfo.action === 'retry' ? 
          '<button onclick="window.location.reload()" class="btn-retry">Retry</button>' : ''}
      </div>
    `;
    errorContainer.style.display = 'block';
  }

  static showFieldErrors(form, errors) {
    // Clear previous errors
    form.querySelectorAll('.field-error').forEach(el => el.remove());
    form.querySelectorAll('.input-error').forEach(el => {
      el.classList.remove('input-error');
    });

    // Show new errors
    Object.keys(errors).forEach(field => {
      const input = form.querySelector(`[name="${field}"], [id="${field}"]`);
      if (input) {
        input.classList.add('input-error');
        
        const errorEl = document.createElement('span');
        errorEl.className = 'field-error';
        errorEl.textContent = errors[field];
        input.parentNode.appendChild(errorEl);
      }
    });
  }
}
```

### User Feedback Components

```javascript
// Toast notification system
class ToastManager {
  constructor() {
    this.container = this.createContainer();
  }

  createContainer() {
    let container = document.getElementById('toast-container');
    if (!container) {
      container = document.createElement('div');
      container.id = 'toast-container';
      container.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        z-index: 9999;
        display: flex;
        flex-direction: column;
        gap: 10px;
      `;
      document.body.appendChild(container);
    }
    return container;
  }

  show(message, type = 'info', duration = 5000) {
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.style.cssText = `
      padding: 15px 20px;
      border-radius: 8px;
      color: white;
      font-family: var(--font-ui);
      font-size: 0.9rem;
      animation: slideIn 0.3s ease;
      max-width: 300px;
      word-wrap: break-word;
      ${this.getToastStyles(type)}
    `;
    toast.textContent = message;

    this.container.appendChild(toast);

    setTimeout(() => {
      toast.style.animation = 'slideOut 0.3s ease';
      setTimeout(() => toast.remove(), 300);
    }, duration);
  }

  getToastStyles(type) {
    const colors = {
      success: 'background: #27ae60;',
      error: 'background: #e74c3c;',
      warning: 'background: #f39c12;',
      info: 'background: #3498db;'
    };
    return colors[type] || colors.info;
  }

  success(message) {
    this.show(message, 'success');
  }

  error(message) {
    this.show(message, 'error');
  }

  warning(message) {
    this.show(message, 'warning');
  }

  info(message) {
    this.show(message, 'info');
  }
}

const toast = new ToastManager();
```

---

## Loading States and UX

### Loading Indicators

```javascript
// Loading state manager
class LoadingManager {
  constructor() {
    this.loadingElements = new Map();
  }

  // Show loading on button
  button(button, text = 'Loading...') {
    const originalText = button.textContent;
    const originalDisabled = button.disabled;
    
    button.textContent = text;
    button.disabled = true;
    button.classList.add('loading');

    const key = button.id || button;
    this.loadingElements.set(key, { element: button, originalText, originalDisabled });

    return () => {
      button.textContent = originalText;
      button.disabled = originalDisabled;
      button.classList.remove('loading');
      this.loadingElements.delete(key);
    };
  }

  // Show loading overlay
  overlay(container, message = 'Loading...') {
    const overlay = document.createElement('div');
    overlay.className = 'loading-overlay';
    overlay.innerHTML = `
      <div class="spinner"></div>
      <p class="loading-message">${message}</p>
    `;
    overlay.style.cssText = `
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background: rgba(255,255,255,0.9);
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      z-index: 100;
    `;

    container.style.position = 'relative';
    container.appendChild(overlay);

    return () => overlay.remove();
  }

  // Show skeleton loading
  skeleton(container, template) {
    const skeleton = document.createElement('div');
    skeleton.className = 'skeleton-loading';
    skeleton.innerHTML = template;
    skeleton.style.cssText = `
      animation: pulse 1.5s ease-in-out infinite;
    `;

    const originalContent = container.innerHTML;
    container.innerHTML = '';
    container.appendChild(skeleton);

    return () => {
      container.innerHTML = originalContent;
    };
  }

  // Cleanup all loading states
  cleanup() {
    this.loadingElements.forEach(({ element, originalText, originalDisabled }) => {
      element.textContent = originalText;
      element.disabled = originalDisabled;
      element.classList.remove('loading');
    });
    this.loadingElements.clear();
  }
}

const loadingManager = new LoadingManager();
```

### CSS for Loading States

```css
/* Loading button styles */
button.loading {
  position: relative;
  color: transparent !important;
}

button.loading::after {
  content: '';
  position: absolute;
  width: 16px;
  height: 16px;
  top: 50%;
  left: 50%;
  margin-left: -8px;
  margin-top: -8px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

/* Spinner animation */
@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Skeleton loading */
.skeleton {
  background: linear-gradient(90deg, #f0f0f0 25%, #e0e0e0 50%, #f0f0f0 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  border-radius: 4px;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* Toast animations */
@keyframes slideIn {
  from {
    transform: translateX(100%);
    opacity: 0;
  }
  to {
    transform: translateX(0);
    opacity: 1;
  }
}

@keyframes slideOut {
  from {
    transform: translateX(0);
    opacity: 1;
  }
  to {
    transform: translateX(100%);
    opacity: 0;
  }
}
```

### Form Submission States

```javascript
// Form submission handler with states
async function handleFormSubmit(form, submitHandler) {
  const submitBtn = form.querySelector('button[type="submit"]');
  const errorContainer = form.querySelector('.form-errors') || document.createElement('div');
  
  // Clear previous errors
  if (errorContainer.classList.contains('form-errors')) {
    errorContainer.innerHTML = '';
    errorContainer.style.display = 'none';
  }

  // Set loading state
  const stopLoading = loadingManager.button(submitBtn);
  
  try {
    // Call submit handler
    const result = await submitHandler();
    
    // Success
    toast.success('Saved successfully!');
    return result;
    
  } catch (error) {
    // Handle error
    const errorInfo = ErrorHandler.handle(error, 'form_submit');
    
    if (errorInfo.type === 'validation') {
      ErrorHandler.showFieldErrors(form, errorInfo.errors);
    } else {
      ErrorHandler.showError(errorContainer, errorInfo);
    }
    
    throw error;
  } finally {
    stopLoading();
  }
}
```

### Optimistic Updates

```javascript
// Optimistic update pattern
class OptimisticUpdater {
  constructor() {
    this.pendingUpdates = new Map();
  }

  async update(elementId, optimisticData, apiCall, onRollback) {
    const element = document.getElementById(elementId);
    if (!element) return;

    // Store original data
    const originalData = element.dataset.original || element.innerHTML;
    element.dataset.original = originalData;

    // Apply optimistic update
    this.applyUpdate(element, optimisticData);
    element.classList.add('optimistic');

    try {
      // Make API call
      const result = await apiCall();
      
      // Success - update with real data
      element.classList.remove('optimistic');
      element.classList.add('confirmed');
      setTimeout(() => element.classList.remove('confirmed'), 1000);
      
      return result;
      
    } catch (error) {
      // Rollback on error
      this.rollback(element, originalData, onRollback);
      throw error;
    }
  }

  applyUpdate(element, data) {
    if (typeof data === 'object') {
      Object.keys(data).forEach(key => {
        const child = element.querySelector(`[data-field="${key}"]`);
        if (child) {
          child.textContent = data[key];
        }
      });
    } else {
      element.innerHTML = data;
    }
  }

  rollback(element, originalData, onRollback) {
    element.innerHTML = originalData;
    element.classList.remove('optimistic');
    element.classList.add('rollback');
    setTimeout(() => element.classList.remove('rollback'), 1000);
    
    if (onRollback) {
      onRollback();
    }
  }
}

const optimisticUpdater = new OptimisticUpdater();
```

---

## Real-World Examples

### Example 1: Creating a Wedding

```javascript
// Complete wedding creation flow
async function createWeddingFlow(formData) {
  const loadingOverlay = document.getElementById('loadingOverlay');
  loadingOverlay.classList.add('active');

  try {
    // Step 1: Create wedding
    console.log('Creating wedding...');
    const wedding = await apiClient.post('/weddings', {
      couple_name: `${formData.groomName} & ${formData.brideName}`,
      event_date: new Date(formData.weddingDate).toISOString(),
      venue: {
        name: formData.venueName,
        address: formData.venueAddress,
        city: formData.venueCity,
        country: formData.venueCountry
      },
      theme: formData.theme,
      story: formData.loveStory,
      settings: {
        is_public: true,
        allow_rsvp: true,
        rsvp_deadline: new Date(formData.rsvpDate).toISOString(),
        guest_limit: formData.guestLimit || 0
      }
    });

    console.log('Wedding created:', wedding.id);

    // Step 2: Upload images if any
    if (formData.galleryImages?.length > 0) {
      console.log('Uploading images...');
      await uploadImages(wedding.id, formData.galleryImages);
    }

    // Step 3: Add events
    if (formData.events?.length > 0) {
      console.log('Adding events...');
      await Promise.all(
        formData.events.map(event => 
          apiClient.post(`/weddings/${wedding.id}/events`, {
            name: event.name,
            description: event.description,
            start_time: new Date(`${formData.weddingDate}T${event.time}`).toISOString(),
            venue_name: event.venue || formData.venueName
          })
        )
      );
    }

    // Step 4: Show success
    showSuccessScreen(wedding);

  } catch (error) {
    console.error('Failed to create wedding:', error);
    
    const errorInfo = ErrorHandler.handle(error, 'create_wedding');
    ErrorHandler.showError('errorContainer', errorInfo);
    
  } finally {
    loadingOverlay.classList.remove('active');
  }
}

async function uploadImages(weddingId, images) {
  const uploadedImages = [];
  
  for (let i = 0; i < images.length; i++) {
    const image = images[i];
    
    try {
      // For file uploads
      if (image instanceof File) {
        const result = await apiClient.upload(
          '/upload/gallery',
          image,
          (progress) => {
            console.log(`Upload ${i + 1}/${images.length}: ${progress}%`);
          },
          { fields: { wedding_id: weddingId } }
        );
        uploadedImages.push(result);
      } 
      // For URLs
      else if (typeof image === 'string') {
        await apiClient.post(`/weddings/${weddingId}/gallery`, {
          url: image,
          type: 'external'
        });
      }
    } catch (error) {
      console.error(`Failed to upload image ${i + 1}:`, error);
      // Continue with other images
    }
  }

  return uploadedImages;
}
```

### Example 2: Updating Wedding Details

```javascript
// Update wedding with optimistic UI
async function updateWeddingDetails(weddingId, updates) {
  const saveBtn = document.getElementById('saveBtn');
  const form = document.getElementById('weddingForm');

  // Optimistic update
  const originalValues = {};
  Object.keys(updates).forEach(key => {
    const field = form.querySelector(`[name="${key}"]`);
    if (field) {
      originalValues[key] = field.value;
      field.classList.add('saving');
    }
  });

  const stopLoading = loadingManager.button(saveBtn, 'Saving...');

  try {
    const result = await apiClient.patch(`/weddings/${weddingId}`, updates);
    
    // Success
    toast.success('Changes saved!');
    
    // Remove saving state
    Object.keys(updates).forEach(key => {
      const field = form.querySelector(`[name="${key}"]`);
      if (field) {
        field.classList.remove('saving');
        field.classList.add('saved');
        setTimeout(() => field.classList.remove('saved'), 2000);
      }
    });

    return result;

  } catch (error) {
    // Rollback optimistic changes
    Object.keys(originalValues).forEach(key => {
      const field = form.querySelector(`[name="${key}"]`);
      if (field) {
        field.value = originalValues[key];
        field.classList.remove('saving');
        field.classList.add('error');
      }
    });

    throw error;
  } finally {
    stopLoading();
  }
}
```

### Example 3: Submitting RSVP

```javascript
// RSVP submission with validation
async function submitRSVP(weddingSlug, formData) {
  const form = document.getElementById('rsvpForm');
  const submitBtn = form.querySelector('button[type="submit"]');

  // Client-side validation
  const errors = validateRSVP(formData);
  if (Object.keys(errors).length > 0) {
    ErrorHandler.showFieldErrors(form, errors);
    return;
  }

  const stopLoading = loadingManager.button(submitBtn, 'Sending...');

  try {
    await apiClient.post(`/weddings/${weddingSlug}/rsvp`, {
      guest_name: formData.name,
      email: formData.email,
      attending: formData.attending === 'yes',
      number_of_guests: parseInt(formData.guests) || 1,
      dietary_restrictions: formData.dietary || '',
      message: formData.message || ''
    });

    // Show success modal
    showRSVPSuccessModal();
    
    // Reset form
    form.reset();

  } catch (error) {
    const errorInfo = ErrorHandler.handle(error, 'rsvp_submit');
    
    if (errorInfo.type === 'validation') {
      ErrorHandler.showFieldErrors(form, errorInfo.errors);
    } else {
      toast.error(errorInfo.message);
    }
  } finally {
    stopLoading();
  }
}

function validateRSVP(data) {
  const errors = {};
  
  if (!data.name || data.name.length < 2) {
    errors.name = 'Please enter your full name';
  }
  
  if (!data.email || !isValidEmail(data.email)) {
    errors.email = 'Please enter a valid email address';
  }
  
  if (!data.attending) {
    errors.attending = 'Please select your attendance';
  }
  
  return errors;
}
```

### Example 4: Listing User's Weddings

```javascript
// Load weddings with pagination and search
async function loadWeddingsList(options = {}) {
  const container = document.getElementById('weddingList');
  const { page = 1, search = '', sort = 'created_at' } = options;

  // Show skeleton loading
  container.innerHTML = `
    <div class="skeleton-list">
      ${Array(3).fill('<div class="skeleton-card"></div>').join('')}
    </div>
  `;

  try {
    const weddings = await apiClient.get('/weddings', {
      params: { page, search, sort }
    });

    if (weddings.length === 0 && page === 1) {
      container.innerHTML = `
        <div class="empty-state">
          <h3>No weddings yet</h3>
          <p>Create your first wedding invitation</p>
          <a href="/generator.html" class="btn-primary">Create Wedding</a>
        </div>
      `;
      return;
    }

    // Render weddings
    container.innerHTML = weddings.map(wedding => `
      <div class="wedding-card" data-id="${wedding.id}">
        <div class="wedding-info">
          <h3>${escapeHtml(wedding.couple_name)}</h3>
          <p class="wedding-date">
            ${formatDate(wedding.event_date)}
          </p>
          <span class="theme-badge">${formatTheme(wedding.theme)}</span>
        </div>
        <div class="wedding-actions">
          <a href="/w/${wedding.slug}" class="btn-view" target="_blank">View</a>
          <a href="/edit.html?id=${wedding.id}" class="btn-edit">Edit</a>
          <button class="btn-delete" onclick="deleteWedding('${wedding.id}')">Delete</button>
        </div>
      </div>
    `).join('');

    // Update pagination
    updatePagination(page, weddings.length >= 20);

  } catch (error) {
    container.innerHTML = `
      <div class="error-state">
        <p>Failed to load weddings</p>
        <button onclick="loadWeddingsList(${JSON.stringify(options)})" class="btn-retry">
          Try Again
        </button>
      </div>
    `;
  }
}
```

---

## Security Considerations

### XSS Prevention

```javascript
// HTML escaping utility
function escapeHtml(text) {
  const div = document.createElement('div');
  div.textContent = text;
  return div.innerHTML;
}

// Sanitize user input before DOM insertion
function sanitizeInput(input) {
  return input
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#x27;')
    .replace(/\//g, '&#x2F;');
}

// Safe template literal
function html(strings, ...values) {
  return strings.reduce((result, string, i) => {
    const value = values[i];
    if (value === undefined) return result + string;
    return result + string + escapeHtml(String(value));
  }, '');
}

// Usage
const userInput = '<script>alert("xss")</script>';
const safeHtml = html`<div>${userInput}</div>`;
// Result: <div>&lt;script&gt;alert("xss")&lt;/script&gt;</div>
```

### CSRF Protection

Since the API uses HTTP-only cookies with `SameSite=Strict`, CSRF protection is handled automatically by modern browsers. For additional protection:

```javascript
// CSRF Token handling (if needed for non-cookie auth)
class CSRFProtection {
  constructor() {
    this.token = null;
  }

  // Get CSRF token from meta tag or cookie
  getToken() {
    if (!this.token) {
      const meta = document.querySelector('meta[name="csrf-token"]');
      this.token = meta?.content;
    }
    return this.token;
  }

  // Add token to request headers
  addToken(headers) {
    const token = this.getToken();
    if (token) {
      headers['X-CSRF-Token'] = token;
    }
    return headers;
  }
}

// Add to API client if using CSRF tokens
apiClient.addRequestInterceptor(async (config) => {
  config.headers = csrfProtection.addToken(config.headers);
  return config;
});
```

### HTTPS Enforcement

```javascript
// Enforce HTTPS in production
if (window.location.protocol === 'http:' && 
    window.location.hostname !== 'localhost' &&
    window.location.hostname !== '127.0.0.1') {
  window.location.href = window.location.href.replace('http:', 'https:');
}

// API client HTTPS validation
class SecureAPIClient extends WeddingAPIClient {
  constructor(config) {
    super(config);
    
    if (this.config.baseURL.startsWith('http:') && 
        !this.config.baseURL.includes('localhost')) {
      console.error('WARNING: API URL should use HTTPS in production');
      this.config.baseURL = this.config.baseURL.replace('http:', 'https:');
    }
  }
}
```

### Content Security Policy

Add to HTML `<head>`:

```html
<meta http-equiv="Content-Security-Policy" content="
  default-src 'self';
  script-src 'self' 'unsafe-inline' https://cdn.example.com;
  style-src 'self' 'unsafe-inline' https://fonts.googleapis.com;
  font-src 'self' https://fonts.gstatic.com;
  img-src 'self' https: data: blob:;
  connect-src 'self' https://api.wedding-app.com;
  frame-ancestors 'none';
  base-uri 'self';
  form-action 'self';
">
```

### Secure Storage

```javascript
// Secure localStorage wrapper
class SecureStorage {
  constructor(prefix = 'wedding_app_') {
    this.prefix = prefix;
  }

  set(key, value) {
    try {
      const serialized = JSON.stringify(value);
      localStorage.setItem(this.prefix + key, serialized);
    } catch (e) {
      console.error('Storage failed:', e);
    }
  }

  get(key) {
    try {
      const item = localStorage.getItem(this.prefix + key);
      return item ? JSON.parse(item) : null;
    } catch (e) {
      console.error('Storage read failed:', e);
      return null;
    }
  }

  remove(key) {
    localStorage.removeItem(this.prefix + key);
  }

  clear() {
    Object.keys(localStorage)
      .filter(key => key.startsWith(this.prefix))
      .forEach(key => localStorage.removeItem(key));
  }
}

// Use only for non-sensitive data (UI preferences, etc.)
// NEVER store tokens or passwords
const secureStorage = new SecureStorage();
```

---

## Testing the Integration

### Mock API for Development

```javascript
// mock-api.js - Development mock server
class MockAPI {
  constructor() {
    this.weddings = new Map();
    this.users = new Map();
    this.currentUser = null;
  }

  // Mock responses
  async mockRequest(endpoint, options) {
    // Simulate network delay
    await this.delay(200);

    // Parse endpoint
    const [resource, id, subResource] = endpoint.split('/').filter(Boolean);

    switch (resource) {
      case 'auth':
        return this.handleAuth(id, options);
      case 'weddings':
        return this.handleWeddings(id, subResource, options);
      case 'upload':
        return this.handleUpload(options);
      default:
        throw new Error('Not found');
    }
  }

  async handleAuth(action, options) {
    switch (action) {
      case 'login':
        const { email } = JSON.parse(options.body);
        this.currentUser = { id: '1', email, name: 'Test User' };
        return { user: this.currentUser };
      
      case 'register':
        return { message: 'Registration successful' };
      
      case 'refresh':
        return { message: 'Token refreshed' };
      
      default:
        throw new Error('Unknown auth action');
    }
  }

  async handleWeddings(id, subResource, options) {
    switch (options.method) {
      case 'GET':
        if (id) {
          return this.weddings.get(id);
        }
        return Array.from(this.weddings.values());
      
      case 'POST':
        const data = JSON.parse(options.body);
        const wedding = {
          id: String(Date.now()),
          ...data,
          created_at: new Date().toISOString()
        };
        this.weddings.set(wedding.id, wedding);
        return wedding;
      
      case 'PUT':
      case 'PATCH':
        const updates = JSON.parse(options.body);
        const existing = this.weddings.get(id);
        const updated = { ...existing, ...updates };
        this.weddings.set(id, updated);
        return updated;
      
      case 'DELETE':
        this.weddings.delete(id);
        return null;
      
      default:
        throw new Error('Method not allowed');
    }
  }

  delay(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
  }
}

// Use mock in development
const mockAPI = new MockAPI();

if (window.location.search.includes('mock=true')) {
  // Override fetch with mock
  const originalFetch = window.fetch;
  window.fetch = async (url, options) => {
    if (url.includes('/api/')) {
      const endpoint = url.split('/api/')[1];
      const result = await mockAPI.mockRequest(endpoint, options);
      
      return {
        ok: true,
        status: 200,
        json: async () => result,
        text: async () => JSON.stringify(result)
      };
    }
    return originalFetch(url, options);
  };
}
```

### Unit Testing

```javascript
// api-client.test.js
// Using Jest or similar testing framework

describe('WeddingAPIClient', () => {
  let client;
  let fetchMock;

  beforeEach(() => {
    fetchMock = jest.fn();
    global.fetch = fetchMock;
    
    client = new WeddingAPIClient({
      baseURL: 'http://localhost:8080/api/v1'
    });
  });

  describe('GET requests', () => {
    it('should make GET request and return JSON', async () => {
      const mockData = { id: '1', couple_name: 'Test Wedding' };
      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => mockData
      });

      const result = await client.get('/weddings/1');
      
      expect(result).toEqual(mockData);
      expect(fetchMock).toHaveBeenCalledWith(
        'http://localhost:8080/api/v1/weddings/1',
        expect.any(Object)
      );
    });

    it('should handle 404 errors', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        status: 404,
        statusText: 'Not Found',
        json: async () => ({ error: 'Wedding not found' })
      });

      await expect(client.get('/weddings/999'))
        .rejects.toThrow('Wedding not found');
    });
  });

  describe('POST requests', () => {
    it('should send JSON body', async () => {
      const weddingData = { couple_name: 'John & Jane' };
      
      fetchMock.mockResolvedValueOnce({
        ok: true,
        status: 201,
        headers: new Map([['content-type', 'application/json']]),
        json: async () => ({ id: '1', ...weddingData })
      });

      await client.post('/weddings', weddingData);

      expect(fetchMock).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(weddingData),
          headers: expect.objectContaining({
            'Content-Type': 'application/json'
          })
        })
      );
    });
  });

  describe('Retry logic', () => {
    it('should retry on network error', async () => {
      fetchMock
        .mockRejectedValueOnce(new Error('Network error'))
        .mockRejectedValueOnce(new Error('Network error'))
        .mockResolvedValueOnce({
          ok: true,
          status: 200,
          headers: new Map(),
          json: async () => ({ success: true })
        });

      const result = await client.get('/weddings');
      
      expect(result).toEqual({ success: true });
      expect(fetchMock).toHaveBeenCalledTimes(3);
    });

    it('should not retry on 4xx errors', async () => {
      fetchMock.mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: async () => ({ error: 'Bad request' })
      });

      await expect(client.get('/weddings'))
        .rejects.toThrow();
      
      expect(fetchMock).toHaveBeenCalledTimes(1);
    });
  });
});
```

### Integration Testing

```javascript
// integration.test.js

describe('Wedding Creation Flow', () => {
  beforeAll(async () => {
    // Login
    await page.goto('http://localhost:3000/login.html');
    await page.fill('#email', 'test@example.com');
    await page.fill('#password', 'password123');
    await page.click('button[type="submit"]');
    await page.waitForNavigation();
  });

  it('should create a wedding', async () => {
    await page.goto('http://localhost:3000/generator.html');
    
    // Fill form
    await page.fill('#groomName', 'John');
    await page.fill('#brideName', 'Jane');
    await page.fill('#weddingDate', '2026-06-15');
    await page.fill('#venueName', 'Test Venue');
    
    // Submit
    await page.click('.btn-generate');
    
    // Wait for success
    await page.waitForSelector('.generated-container.active', { timeout: 10000 });
    
    // Verify success message
    const successText = await page.textContent('.generated-container h2');
    expect(successText).toContain('Ready');
  });

  it('should appear in manage list', async () => {
    await page.goto('http://localhost:3000/manage.html');
    
    await page.waitForSelector('.invitation-card');
    
    const cards = await page.$$('.invitation-card');
    expect(cards.length).toBeGreaterThan(0);
    
    // Verify content
    const name = await page.textContent('.invitation-info h3');
    expect(name).toContain('John');
  });
});
```

---

## Complete Code Examples

### Complete API Client Module

```javascript
/**
 * @file api-client.js
 * @description Complete API client for Wedding Invitation Backend
 * @version 1.0.0
 */

(function(global) {
  'use strict';

  // Configuration
  const DEFAULT_CONFIG = {
    baseURL: window.location.hostname === 'localhost' 
      ? 'http://localhost:8080/api/v1'
      : 'https://api.wedding-app.com/api/v1',
    timeout: 30000,
    retries: 3,
    retryDelay: 1000,
    credentials: 'include'
  };

  // Error Classes
  class APIError extends Error {
    constructor(message, status, data = null) {
      super(message);
      this.name = 'APIError';
      this.status = status;
      this.data = data;
    }
  }

  class NetworkError extends Error {
    constructor(message) {
      super(message);
      this.name = 'NetworkError';
    }
  }

  class ValidationError extends Error {
    constructor(message, errors) {
      super(message);
      this.name = 'ValidationError';
      this.errors = errors;
    }
  }

  // Token Manager
  class TokenManager {
    constructor() {
      this.authState = null;
      this.loadAuthState();
    }

    loadAuthState() {
      try {
        const saved = localStorage.getItem('auth_state');
        if (saved) {
          this.authState = JSON.parse(saved);
        }
      } catch (e) {
        console.error('Failed to load auth state:', e);
      }
    }

    setAuthenticated(user) {
      this.authState = { isAuthenticated: true, user };
      localStorage.setItem('auth_state', JSON.stringify({
        isAuthenticated: true,
        userId: user.id
      }));
    }

    clearAuth() {
      this.authState = null;
      localStorage.removeItem('auth_state');
    }

    isAuthenticated() {
      return this.authState?.isAuthenticated || false;
    }

    getUser() {
      return this.authState?.user || null;
    }
  }

  // Main API Client
  class WeddingAPIClient {
    constructor(config = {}) {
      this.config = { ...DEFAULT_CONFIG, ...config };
      this.tokenManager = new TokenManager();
      this.requestInterceptors = [];
      this.responseInterceptors = [];
      this.isRefreshing = false;
      this.refreshSubscribers = [];

      this.setupDefaultInterceptors();
    }

    setupDefaultInterceptors() {
      // Auth interceptor
      this.addRequestInterceptor(async (config) => {
        config.headers = {
          'Content-Type': 'application/json',
          'Accept': 'application/json',
          ...config.headers
        };
        return config;
      });

      // Error handler
      this.addErrorInterceptor(async (error) => {
        if (error.status === 401 && !error.config?._retry) {
          return this.handleTokenRefresh(error.config);
        }
        throw error;
      });
    }

    // HTTP Methods
    async get(endpoint, options = {}) {
      return this.request(endpoint, { ...options, method: 'GET' });
    }

    async post(endpoint, data, options = {}) {
      return this.request(endpoint, { ...options, method: 'POST', body: data });
    }

    async put(endpoint, data, options = {}) {
      return this.request(endpoint, { ...options, method: 'PUT', body: data });
    }

    async patch(endpoint, data, options = {}) {
      return this.request(endpoint, { ...options, method: 'PATCH', body: data });
    }

    async delete(endpoint, options = {}) {
      return this.request(endpoint, { ...options, method: 'DELETE' });
    }

    // File Upload
    async upload(endpoint, file, onProgress, options = {}) {
      const formData = new FormData();
      formData.append('file', file);

      if (options.fields) {
        Object.entries(options.fields).forEach(([key, value]) => {
          formData.append(key, value);
        });
      }

      const url = `${this.config.baseURL}${endpoint}`;
      
      return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();

        xhr.upload.addEventListener('progress', (e) => {
          if (e.lengthComputable && onProgress) {
            const progress = (e.loaded / e.total) * 100;
            onProgress(progress);
          }
        });

        xhr.addEventListener('load', () => {
          if (xhr.status >= 200 && xhr.status < 300) {
            try {
              resolve(JSON.parse(xhr.responseText));
            } catch {
              resolve(xhr.responseText);
            }
          } else {
            reject(new APIError('Upload failed', xhr.status));
          }
        });

        xhr.addEventListener('error', () => {
          reject(new NetworkError('Upload failed'));
        });

        xhr.open('POST', url);
        xhr.withCredentials = true;
        xhr.send(formData);
      });
    }

    // Main request method
    async request(endpoint, options = {}) {
      const url = `${this.config.baseURL}${endpoint}`;
      
      const attemptRequest = async (attempt = 1) => {
        try {
          let config = {
            method: options.method || 'GET',
            credentials: this.config.credentials,
            headers: options.headers || {}
          };

          if (options.body && config.method !== 'GET') {
            config.body = JSON.stringify(options.body);
          }

          // Apply request interceptors
          for (const interceptor of this.requestInterceptors) {
            config = await interceptor(config);
          }

          const response = await this.fetchWithTimeout(url, config);

          if (response.status === 204) {
            return null;
          }

          if (!response.ok) {
            const error = await this.createError(response);
            error.config = { ...config, _retry: options._retry };
            
            for (const interceptor of this.errorInterceptors) {
              try {
                return await interceptor(error);
              } catch (e) {
                continue;
              }
            }
            throw error;
          }

          return await this.parseResponse(response);

        } catch (error) {
          if (error.status >= 400 && error.status < 500) {
            throw error;
          }

          if (attempt < this.config.retries) {
            await this.delay(this.config.retryDelay * Math.pow(2, attempt - 1));
            return attemptRequest(attempt + 1);
          }

          throw error;
        }
      };

      return attemptRequest();
    }

    async fetchWithTimeout(url, config) {
      const controller = new AbortController();
      const timeout = setTimeout(() => controller.abort(), this.config.timeout);

      try {
        return await fetch(url, { ...config, signal: controller.signal });
      } finally {
        clearTimeout(timeout);
      }
    }

    async handleTokenRefresh(config) {
      if (this.isRefreshing) {
        return new Promise((resolve, reject) => {
          this.refreshSubscribers.push({ resolve, reject, config });
        });
      }

      this.isRefreshing = true;

      try {
        await this.post('/auth/refresh');
        
        this.refreshSubscribers.forEach(({ resolve, reject, config }) => {
          this.request(config.url, { ...config, _retry: true })
            .then(resolve)
            .catch(reject);
        });

      } catch (error) {
        this.tokenManager.clearAuth();
        this.refreshSubscribers.forEach(({ reject }) => reject(error));
        window.location.href = '/login.html';
      } finally {
        this.isRefreshing = false;
        this.refreshSubscribers = [];
      }
    }

    async parseResponse(response) {
      const contentType = response.headers.get('content-type');
      if (contentType?.includes('application/json')) {
        return await response.json();
      }
      return await response.text();
    }

    async createError(response) {
      let data;
      try {
        data = await response.json();
      } catch {
        data = { error: response.statusText };
      }

      return new APIError(
        data.error || data.message || `HTTP ${response.status}`,
        response.status,
        data
      );
    }

    // Interceptors
    addRequestInterceptor(interceptor) {
      this.requestInterceptors.push(interceptor);
    }

    addResponseInterceptor(interceptor) {
      this.responseInterceptors.push(interceptor);
    }

    addErrorInterceptor(interceptor) {
      this.errorInterceptors.push(interceptor);
    }

    // Utilities
    delay(ms) {
      return new Promise(resolve => setTimeout(resolve, ms));
    }

    // Auth helpers
    isAuthenticated() {
      return this.tokenManager.isAuthenticated();
    }

    getUser() {
      return this.tokenManager.getUser();
    }

    logout() {
      this.post('/auth/logout').catch(() => {});
      this.tokenManager.clearAuth();
      window.location.href = '/';
    }
  }

  // Create global instance
  const apiClient = new WeddingAPIClient();

  // Export
  global.WeddingAPIClient = WeddingAPIClient;
  global.apiClient = apiClient;
  global.APIError = APIError;
  global.NetworkError = NetworkError;
  global.ValidationError = ValidationError;

})(typeof window !== 'undefined' ? window : global);
```

### Usage Example: Complete Page Integration

```html
<!-- dashboard.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Dashboard | Wedding Invitations</title>
  <link rel="stylesheet" href="styles.css">
</head>
<body>
  <div id="app">
    <header class="header">
      <h1>Your Weddings</h1>
      <div id="userInfo"></div>
      <button id="logoutBtn" class="btn-secondary">Logout</button>
    </header>

    <main class="main">
      <div id="weddingList" class="wedding-list">
        <div class="loading-spinner">Loading...</div>
      </div>
      
      <a href="/generator.html" class="btn-primary fab">
        + Create New
      </a>
    </main>

    <div id="toastContainer"></div>
  </div>

  <script src="api-client.js"></script>
  <script>
    // Check authentication
    if (!apiClient.isAuthenticated()) {
      window.location.href = '/login.html?redirect=' + 
        encodeURIComponent(window.location.pathname);
    }

    // Display user info
    const user = apiClient.getUser();
    document.getElementById('userInfo').textContent = `Hello, ${user?.name || 'User'}`;

    // Logout handler
    document.getElementById('logoutBtn').addEventListener('click', () => {
      apiClient.logout();
    });

    // Load weddings
    async function loadWeddings() {
      const container = document.getElementById('weddingList');
      
      try {
        const weddings = await apiClient.get('/weddings');
        
        if (weddings.length === 0) {
          container.innerHTML = `
            <div class="empty-state">
              <h2>No weddings yet</h2>
              <p>Create your first beautiful invitation</p>
              <a href="/generator.html" class="btn-primary">Get Started</a>
            </div>
          `;
          return;
        }

        container.innerHTML = weddings.map(w => `
          <div class="wedding-card">
            <div class="wedding-info">
              <h3>${escapeHtml(w.couple_name)}</h3>
              <p>${new Date(w.event_date).toLocaleDateString()}</p>
              <span class="badge">${w.theme}</span>
            </div>
            <div class="wedding-actions">
              <a href="/w/${w.slug}" class="btn" target="_blank">View</a>
              <a href="/edit.html?id=${w.id}" class="btn btn-secondary">Edit</a>
              <button class="btn btn-danger" data-id="${w.id}">Delete</button>
            </div>
          </div>
        `).join('');

        // Attach delete handlers
        container.querySelectorAll('.btn-danger').forEach(btn => {
          btn.addEventListener('click', async () => {
            const id = btn.dataset.id;
            if (confirm('Delete this wedding?')) {
              try {
                await apiClient.delete(`/weddings/${id}`);
                btn.closest('.wedding-card').remove();
                showToast('Wedding deleted', 'success');
              } catch (error) {
                showToast('Failed to delete', 'error');
              }
            }
          });
        });

      } catch (error) {
        container.innerHTML = `
          <div class="error-state">
            <p>Failed to load weddings</p>
            <button onclick="loadWeddings()" class="btn">Try Again</button>
          </div>
        `;
      }
    }

    // Utility functions
    function escapeHtml(text) {
      const div = document.createElement('div');
      div.textContent = text;
      return div.innerHTML;
    }

    function showToast(message, type = 'info') {
      const container = document.getElementById('toastContainer');
      const toast = document.createElement('div');
      toast.className = `toast toast-${type}`;
      toast.textContent = message;
      container.appendChild(toast);
      setTimeout(() => toast.remove(), 3000);
    }

    // Initialize
    loadWeddings();
  </script>
</body>
</html>
```

---

## Summary

This guide provides a comprehensive approach to integrating the wedding invitation frontend with the Go backend API. Key takeaways:

1. **Use HTTP-only cookies** for JWT storage (handled by backend)
2. **Replace localStorage** with API calls for all data persistence
3. **Implement proper error handling** with user-friendly messages
4. **Add loading states** for better UX
5. **Use interceptors** for authentication and error handling
6. **Implement retry logic** for network resilience
7. **Follow security best practices** (XSS prevention, HTTPS, CSP)
8. **Test thoroughly** with mocks and integration tests

The provided `api-client.js` is production-ready and can be included in any theme or page to immediately start making API calls.
