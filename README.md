# 🏨 Hotel Microservice

> Hotel and review management service for the Hotel Reservation Platform.

## Overview

The Hotel Microservice manages **hotel entities** and their **user reviews**. It provides CRUD operations for hotels (restricted to admin users), public listing/search capabilities, and a review system that automatically recalculates hotel ratings. It also integrates with the Media Service for hotel image uploads.

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.25 |
| Router | [go-chi/chi](https://github.com/go-chi/chi) v5 |
| Database | PostgreSQL 16 |
| DB Driver | [pgx](https://github.com/jackc/pgx) v5 |
| Auth | JWT verification (RSA-256 public key) |
| UUID | Google UUID v7 (time-sortable) |
| Container | Docker (multi-stage Alpine build) |

## Architecture

```
app/
├── cmd/api/          # Application entrypoint
│   └── main.go
├── internal/
│   ├── client/       # Media Service HTTP client
│   ├── config/       # YAML config loader
│   ├── database/     # PostgreSQL connection pool
│   ├── handler/      # HTTP handlers, routing, JWT middleware
│   ├── helper/       # Validators, error types, response helpers
│   ├── logging/      # Structured slog logger
│   ├── models/       # Domain entities (Hotel, Review, DTOs)
│   ├── repo/         # Repository interface + PostgreSQL implementation
│   └── service/      # Business logic (CRUD + rating recalculation)
├── sql/
│   └── migrations/   # SQL migrations
├── config.yaml
├── Dockerfile
└── go.mod
```

## API Endpoints

### Public Routes (No Authentication)

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Liveness probe |
| `GET` | `/ready` | Readiness probe |
| `GET` | `/hotels` | List hotels (filter by `?city=`) |
| `GET` | `/hotels/{id}` | Get hotel by ID |
| `GET` | `/hotels/{id}/reviews` | List reviews for a hotel |

### Protected Routes (JWT Required)

| Method | Path | Role | Description |
|---|---|---|---|
| `POST` | `/hotels` | Admin | Create a new hotel |
| `PUT` | `/hotels/{id}` | Admin | Update hotel details |
| `DELETE` | `/hotels/{id}` | Admin | Delete a hotel |
| `POST` | `/hotels/{id}/reviews` | User | Submit a review |

## Data Model

### `hotels` Table

| Column | Type | Description |
|---|---|---|
| `id` | UUID v7 | Primary key |
| `admin_id` | UUID | FK → Users service (admin who created) |
| `name` | VARCHAR | Hotel name |
| `city` | VARCHAR | City location |
| `description` | TEXT | Hotel description |
| `rating` | FLOAT | Average rating (auto-calculated) |
| `lat` | FLOAT | Latitude coordinate |
| `lng` | FLOAT | Longitude coordinate |
| `created_at` | TIMESTAMP | Record creation time |
| `updated_at` | TIMESTAMP | Last update time |

### `reviews` Table

| Column | Type | Description |
|---|---|---|
| `id` | UUID v7 | Primary key |
| `hotel_id` | UUID | FK → `hotels.id` |
| `user_id` | UUID | FK → Users service |
| `rating` | INT | Rating (1–5) |
| `comment` | TEXT | Review comment |
| `created_at` | TIMESTAMP | Review creation time |

## Flow Diagram

```mermaid
flowchart TD
    A["Client Request"] --> B{"Route Type?"}
    B -->|Public| C{"Endpoint?"}
    B -->|Protected| D["JWT Middleware"]
    
    C -->|GET /hotels| E["Parse city, limit, offset"]
    E --> E1["ListHotels from DB"]
    E1 --> E2["Return JSON array"]
    
    C -->|GET /hotels/id| F["GetHotelByID"]
    F --> F1["Return Hotel JSON"]
    
    C -->|GET /hotels/id/reviews| G["ListReviewsByHotelID"]
    G --> G1["Return Reviews JSON"]
    
    D --> D1{"Token Valid?"}
    D1 -->|No| D2["401 Unauthorized"]
    D1 -->|Yes| D3["Extract Claims"]
    
    D3 --> H{"Endpoint?"}
    H -->|POST /hotels| I{"Is Admin?"}
    I -->|No| I0["403 Forbidden"]
    I -->|Yes| I1["Decode CreateHotelRequest"]
    I1 --> I2["Generate UUID v7"]
    I2 --> I3["Insert Hotel"]
    I3 --> I4{"Has Files?"}
    I4 -->|Yes| I5["Upload to Media Service"]
    I4 -->|No| I6["201 Created"]
    I5 --> I6
    
    H -->|POST /hotels/id/reviews| J["Decode CreateReviewRequest"]
    J --> J1["Validate rating 1-5"]
    J1 --> J2["Insert Review"]
    J2 --> J3["Recalculate Hotel Rating"]
    J3 --> J4["201 Created"]
    
    H -->|PUT /hotels/id| K["Decode UpdateHotelRequest"]
    K --> K1["Update Hotel in DB"]
    K1 --> K2["200 OK"]
    
    H -->|DELETE /hotels/id| L["Delete Hotel"]
    L --> L1["204 No Content"]
```

## Use Case Diagram

```mermaid
graph LR
    subgraph Actors
        Guest["🧑 Guest"]
        User["👤 Authenticated User"]
        Admin["🔑 Admin"]
    end
    
    subgraph "Hotel Microservice"
        UC1["Browse Hotels"]
        UC2["View Hotel Details"]
        UC3["View Reviews"]
        UC4["Submit Review"]
        UC5["Create Hotel"]
        UC6["Update Hotel"]
        UC7["Delete Hotel"]
    end
    
    Guest --> UC1
    Guest --> UC2
    Guest --> UC3
    User --> UC1
    User --> UC2
    User --> UC3
    User --> UC4
    Admin --> UC5
    Admin --> UC6
    Admin --> UC7
```

## State Diagram

```mermaid
stateDiagram-v2
    [*] --> NonExistent
    NonExistent --> Active : Admin creates hotel
    Active --> Active : Update details
    Active --> Active : New review (rating recalc)
    Active --> [*] : Admin deletes hotel
    
    state Active {
        [*] --> Listed
        Listed --> DetailView : GET /hotels/id
        DetailView --> ReviewListing : GET /hotels/id/reviews
        ReviewListing --> DetailView : Navigate back
        DetailView --> Listed : Navigate back
    }
```

## Package Diagram

```mermaid
graph TB
    subgraph "cmd/api"
        Main["main.go"]
    end
    
    subgraph "internal"
        subgraph "handler"
            Handlers["handlers.go"]
            Routing["routing.go"]
            MW["middleware.go (JWT)"]
        end
        
        subgraph "service"
            SVC["service.go"]
        end
        
        subgraph "repo"
            RepoIF["repo.go (interface)"]
            DBRepo["database_repo.go"]
        end
        
        subgraph "models"
            Models["models.go"]
        end
        
        subgraph "client"
            MediaClient["media_client.go"]
        end
        
        subgraph "helper"
            Helper["validators, errors"]
        end
        
        subgraph "config"
            Config["config.go"]
        end
        
        subgraph "database"
            DB["connection.go"]
        end
    end
    
    Main --> Config
    Main --> DB
    Main --> SVC
    Main --> Handlers
    
    Handlers --> SVC
    Handlers --> Helper
    Handlers --> Models
    Routing --> Handlers
    Routing --> MW
    
    SVC --> RepoIF
    SVC --> Models
    SVC --> MediaClient
    
    DBRepo -.->|implements| RepoIF
    DBRepo --> DB
    DBRepo --> Models
    
    MediaClient -->|HTTP| ExternalMedia["Media Service"]
```

## Configuration

```yaml
server:
  host: "0.0.0.0"
  port: 8080

logging:
  level: "info"
  format: "json"
```

### Environment Variables

| Variable | Description |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string |

### Volume Mounts (Docker)

| Host Path | Container Path | Description |
|---|---|---|
| `./keys/public.pem` | `/app/keys/public.pem` | JWT verification key |

## Port Mapping

| Context | Port |
|---|---|
| Internal (container) | `8080` |
| External (host) | `8084` |
| Database (host) | `5435` → `5432` |
