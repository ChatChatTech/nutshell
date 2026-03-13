# User Management REST API — Requirements

## Objective

Build a complete REST API for user management that supports CRUD operations, authentication, and rate limiting.

## Functional Requirements

### FR-1: User CRUD
- `POST /api/v1/users` — Create user (name, email, role required)
- `GET /api/v1/users` — List users with pagination (page, per_page)
- `GET /api/v1/users/:id` — Get single user by ID
- `PUT /api/v1/users/:id` — Update user fields
- `DELETE /api/v1/users/:id` — Soft-delete user (set deleted_at)

### FR-2: Authentication
- JWT-based authentication using HS256
- `POST /api/v1/auth/login` — Returns access + refresh tokens
- Access token expires in 15 minutes, refresh in 7 days
- Protected routes require `Authorization: Bearer <token>` header

### FR-3: Rate Limiting
- Token bucket algorithm, 100 requests/minute per IP
- Return `429 Too Many Requests` with `Retry-After` header
- Rate limit info in response headers: `X-RateLimit-Remaining`, `X-RateLimit-Reset`

### FR-4: Input Validation
- Email format validation (RFC 5322)
- Name: 2-100 characters, alphanumeric + spaces
- Role: must be one of ["admin", "member", "viewer"]

## Non-Functional Requirements

- Response time < 200ms for all endpoints (p99)
- Structured JSON error responses: `{"error": "message", "code": "ERROR_CODE"}`
- Request/response logging with correlation IDs
- Database connection pooling (max 20 connections)

## Tech Stack

- Go 1.22+ with Gin framework
- PostgreSQL 15+ with golang-migrate
- github.com/golang-jwt/jwt/v5 for JWT
