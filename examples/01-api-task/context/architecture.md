# Architecture

## System Diagram

```
┌─────────────┐    ┌──────────────┐    ┌─────────────────┐
│   Client     │───►│  Gin Router  │───►│  PostgreSQL 15  │
│  (HTTP/JSON) │    │  + Middleware │    │  (userservice)  │
└─────────────┘    └──────────────┘    └─────────────────┘
                         │
                   ┌─────┴──────┐
                   │ Middleware  │
                   │ Chain:      │
                   │ 1. Logger   │
                   │ 2. RateLimit│
                   │ 3. JWT Auth │
                   │ 4. Handler  │
                   └────────────┘
```

## Directory Structure (Target)

```
src/
├── main.go              # Entry point, router setup
├── config/
│   └── config.go        # Environment-based configuration
├── handlers/
│   ├── user.go          # CRUD handlers
│   └── auth.go          # Login/token handlers
├── middleware/
│   ├── auth.go          # JWT verification
│   ├── ratelimit.go     # Token bucket rate limiter
│   └── logger.go        # Request logging with correlation ID
├── models/
│   └── user.go          # User struct + validation
├── repository/
│   └── user.go          # Database operations
└── migrations/
    └── 001_users.sql    # Initial schema
```

## Dependency Direction

```
handlers → models → (no deps)
handlers → repository → models
middleware → config
main → handlers, middleware, config
```

No circular dependencies. Repository layer is the only code that touches the database.
