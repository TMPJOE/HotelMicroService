# Hotel Microservice Blueprint

A lightweight Go microservice built with a clean architecture pattern, featuring PostgreSQL integration, structured logging, JWT authentication, rate limiting, circuit breaker pattern, and HTTP request handling via `chi` router.

## Architecture

The project follows a layered architecture:

```
cmd/api/main.go → Entry point, wires dependencies
internal/handler → HTTP handlers, routing, and middleware
internal/service → Business logic layer
internal/repo → Data access layer
internal/database → Database connection management
internal/logging → Structured logging setup
internal/models → Domain models
internal/helper → Utility functions
internal/config → YAML configuration loader
```

## Tech Stack

- **Router**: [go-chi/chi/v5](https://github.com/go-chi/chi)
- **Logging**: [go-chi/httplog/v3](https://github.com/go-chi/httplog) + `log/slog`
- **Database**: [jackc/pgx/v5](https://github.com/jackc/pgx) (PostgreSQL connection pool)
- **JWT Authentication**: [golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt)
- **Validation**: [go-playground/validator/v10](https://github.com/go-playground/validator)
- **Password Hashing**: [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto)
- **UUID Generation**: [google/uuid](https://github.com/google/uuid)

## Features

### Security
- **JWT Authentication**: RSA-based token validation with configurable issuer and expiration
- **Security Headers**: X-Content-Type-Options, X-Frame-Options, X-XSS-Protection, HSTS, CSP
- **Request ID**: Unique request tracking for debugging and logging

### Resilience
- **Rate Limiting**: Token bucket algorithm with configurable requests/second and burst
- **Circuit Breaker**: Automatic failure detection with half-open state for recovery
- **Graceful Shutdown**: 30-second timeout for in-flight requests

### Configuration
- **YAML Config**: All settings loaded from `config.yaml` with environment variable expansion
- **No hardcoded values**: Server port, timeouts, rate limits all configurable

## Prerequisites

- Go 1.25.7+
- PostgreSQL database
- Docker & Docker Compose (optional, for local development)
- RSA key pair for JWT signing (`public.pem`, `private.pem`)

## Getting Started

### 1. Generate JWT Keys

```bash
# Generate private key
openssl genrsa -out private.pem 2048

# Generate public key
openssl rsa -in private.pem -pubout -out public.pem
```

### 2. Set Environment Variables

```bash
export DATABASE_URL="postgres://user:password@localhost:5432/dbname?sslmode=disable"
```

### 3. Configure the Service

Edit `config.yaml` to customize:
- Server host/port and timeouts
- Logging level and format
- Rate limiting parameters
- Circuit breaker settings
- Health check paths

### 4. Run the Service

```bash
go run app/cmd/api/main.go
```

The server starts on `localhost:8080` (or configured port).

### 5. Test the Health Endpoint

```bash
curl http://localhost:8080/health
```

Response:
```json
{"status": "ok"}
```

## Docker

### Build the Image

```bash
docker build -t microservice-blueprint .
```

### Run with Docker

```bash
docker run -p 8080:8080 \
  -e DATABASE_URL="postgres://user:password@host:5432/dbname?sslmode=disable" \
  -v /path/to/keys:/app/keys \
  microservice-blueprint
```

### Docker Compose

Use `docker-compose.yml` to spin up dependencies (e.g., PostgreSQL):

```bash
docker-compose up -d
```

## Project Structure

| Path | Description |
|------|-------------|
| `app/cmd/api/main.go` | Application entry point. Wires together database, repository, service, and handler layers, then starts the HTTP server. |
| `app/internal/config/` | YAML configuration loader with environment variable expansion. |
| `app/internal/database/` | Database connection pool initialization using `pgx`. |
| `app/internal/handler/` | HTTP handlers, request routing (`chi`), and middleware (security, JWT, rate limiting). |
| `app/internal/service/` | Business logic layer. Defines service interfaces and implements use cases. |
| `app/internal/repo/` | Data access layer. Handles all database queries and transactions. |
| `app/internal/logging/` | Structured JSON logger configuration using `slog` and `httplog`. |
| `app/internal/models/` | Domain models and data structures shared across layers. |
| `app/internal/helper/` | Utility/helper functions including comprehensive error definitions. |
| `app/sql/` | SQL migration files and queries. |
| `config.yaml` | Service configuration file. |
| `Dockerfile` | Multi-stage Docker build with healthcheck. |

## API Endpoints

### Public Routes (No Authentication)

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check endpoint. Returns service health status. |
| `GET` | `/ready` | Readiness check. Verifies database connectivity. |
| `GET` | `/hotels` | List hotels with optional `?city=...` and pagination. |
| `GET` | `/hotels/{id}` | Fetch details for a specific hotel by ID. |
| `GET` | `/hotels/{id}/reviews` | List reviews for a specific hotel. |

### Protected Routes (JWT Required)

The following routes require a valid JWT token passed in the `Authorization: Bearer <token>` header.

| Method | Path | Description | Access |
|----------|------|-------------|--------|
| `POST` | `/hotels` | Create a new hotel. Supports both JSON and multipart/form-data for file uploads. | Admin Only |
| `PUT` | `/hotels/{id}` | Update an existing hotel. | Admin Only (Owner) |
| `DELETE` | `/hotels/{id}` | Delete a hotel. | Admin Only (Owner) |
| `POST` | `/hotels/{id}/reviews` | Submit a review (1-5 stars) for a hotel. | Any Authenticated User |

### Creating a Hotel with Files

To create a hotel with image uploads, send a `multipart/form-data` POST request:

```bash
curl -X POST http://localhost:8080/hotels \
  -H "Authorization: Bearer <your-jwt-token>" \
  -F "name=Grand Hotel" \
  -F "city=Paris" \
  -F "description=Luxury hotel in the heart of Paris" \
  -F "lat=48.8566" \
  -F "lng=2.3522" \
  -F "files=@/path/to/image1.jpg" \
  -F "files=@/path/to/image2.png"
```

The service will:
1. Create the hotel record in the database
2. Upload each file to the media service (which stores them in MinIO/S3)
3. Return the created hotel object

This architecture ensures the media service remains protected behind the hotel service's authentication layer.

## Configuration Reference

### config.yaml

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s

logging:
  level: "info"
  format: "json"

rate_limit:
  enabled: true
  requests_per_second: 100
  burst: 200

circuit_breaker:
  enabled: true
  max_failures: 5
  timeout: 30s

health:
  path: "/health"
  ready_path: "/ready"
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string (required) |
| `MEDIA_SERVICE_URL` | URL of the media service for file uploads (default: `http://media-service:8080`) |

## Adding New Features

1. **Models**: Define structs in `app/internal/models/models.go`
2. **Repository**: Add data access methods to `app/internal/repo/repo.go`
3. **Service**: Add business logic methods to `app/internal/service/service.go` (update the `Service` interface)
4. **Handler**: Add HTTP handler functions to `app/internal/handler/handlers.go`
5. **Routing**: Register new routes in `app/internal/handler/routing.go`
6. **Configuration**: Add any new config options to `config.yaml` and `app/internal/config/config.go`
7. **Client (optional)**: Add external service clients to `app/internal/client/` for communicating with other microservices

## Media Service Integration

The hotel service includes a media client (`app/internal/client/media_client.go`) that communicates with the media microservice for file uploads. This architecture:

- **Hides the media service**: The media service has no authentication, so it's not exposed to external clients
- **Handles multipart uploads**: The `POST /hotels` endpoint accepts both JSON and `multipart/form-data` requests
- **Automatically uploads files**: When files are included in a hotel creation request, they're uploaded to the media service before the hotel is created
- **Uses environment configuration**: Set `MEDIA_SERVICE_URL` environment variable to configure the media service endpoint

### Media Client Architecture

```
Client → Hotel Service (JWT Auth) → Media Service → MinIO/S3
```

1. Client sends `multipart/form-data` request with files to hotel service
2. Hotel service validates JWT token and admin permissions
3. Hotel service creates hotel record in database
4. Hotel service uploads each file to media service with `asset_type=hotel` and `asset_id=<hotel_id>`
5. Media service stores files in MinIO/S3 and records metadata in database
6. Hotel service returns created hotel object

## Error Handling

The service uses a comprehensive error system defined in `app/internal/helper/util.go`:

- **General errors**: `ErrInternalServer`, `ErrUnauthorized`, `ErrForbidden`, `ErrNotFound`, etc.
- **Database errors**: `ErrDBConnection`, `ErrDBQuery`, `ErrRecordNotFound`, `ErrDuplicateEntry`, etc.
- **Authentication errors**: `ErrInvalidCredentials`, `ErrInvalidToken`, `ErrTokenExpired`, etc.
- **Service errors**: `ErrServiceUnavailable`, `ErrCreateFailed`, `ErrProcessingFailed`, etc.

Use `helper.MapError()` in the repository layer to convert raw database errors to application sentinel errors.
