# Chessbox Notification - Backend

The backend service for **Chessbox Notification** is a high-performance Go application responsible for tracking chess tournaments on [chess-results.com](https://chess-results.com), detecting new round pairings in real-time, and dispatching push notifications to registered mobile devices.

---

## Architecture Overview

The backend is built following clean modular principles in Go, with decoupled packages handling specific domains:

```
backend/
├── cmd/
│   └── main.go                    # Application entrypoint
├── docker-compose.yml             # Local PostgreSQL service
├── go.mod / go.sum                # Dependencies & Go module definition
└── internal/
    ├── application/               # Application bootstrap, routing & config
    │   ├── app.go                 # Dependency injection & lifecycle coordinator
    │   ├── config.go              # Environment variable loader (.env)
    │   ├── health.go              # Health check handler
    │   └── router.go              # Chi HTTP router setup
    ├── database/                  # Storage abstractions & migration runner
    │   ├── migrate.go             # Auto-migrates database on startup
    │   ├── migrations/            # SQL migration files
    │   ├── postgres.go            # PostgreSQL connection pool
    │   └── sqlite.go              # SQLite driver support
    ├── device/                    # Mobile device registration & authentication
    │   ├── dto.go                 # Request/response transfer objects
    │   ├── handler.go             # HTTP handlers for device endpoints
    │   ├── model.go               # Device domain models
    │   ├── repository.go          # Database operations for devices
    │   └── service.go             # Device credentials & Argon2id hashing
    ├── middleware/                # HTTP middlewares
    │   ├── device_auth.go         # Bearer token authentication middleware
    │   └── log.go                 # Request/response structured logging
    ├── notification/              # Notification delivery & job worker
    │   ├── exponent.go            # Expo Push Notification provider
    │   ├── fcm.go                 # Firebase Cloud Messaging provider
    │   ├── job.go                 # Notification queue consumer worker
    │   ├── model.go / view.go     # Domain entities and representations
    │   ├── repository.go          # Query subscriptions by tournament
    │   └── service.go             # Notification dispatch orchestrator
    ├── subscription/              # User tournament subscriptions
    │   ├── dto.go / view.go       # DTOs and view models
    │   ├── handler.go             # HTTP handlers for subscriptions
    │   ├── model.go               # Subscription entity
    │   ├── repository.go          # Database operations for subscriptions
    │   └── service.go             # Subscription business logic
    ├── tournament/                # Web scraping & tournament tracking
    │   ├── dto.go / model.go      # Tournament data structures
    │   ├── handler.go             # Tournament search HTTP handler
    │   ├── job.go                 # Periodic polling cron job
    │   ├── provider.go            # HTML scraper & parser (goquery)
    │   ├── repository.go          # Tournament data persistence
    │   └── service.go             # Refresh coordinator & worker pool
    └── utils/
        └── request.go             # Generic JSON request/response helper
```

---

## Key Features

1. **Automated Tournament Scraping (`chess-results.com`)**:
   - Scrapes tournament details, active rounds, and total rounds using `goquery`.
   - Forces English localization (`?lan=1`) and expanded details (`?turdet=YES`) to ensure reliable parsing regardless of origin host.
   - Extracts round pairing links (`art=2`) and parses round counts dynamically.
   - Implements full tournament search against `https://s1.chess-results.com/TurnierSuche.aspx`, properly negotiating ASP.NET `__VIEWSTATE` form tokens.

2. **Background Cron & Concurrency Pipeline**:
   - Uses `robfig/cron/v3` to schedule polling every minute (`@every 1m`).
   - A concurrent worker pool (10 concurrent workers) processes active tournaments.
   - If a new round is detected (`updated.CurrentRound > tournament.CurrentRound`), the tournament is updated in the database and an event is queued to an internal Go channel (`notificationCh`).

3. **Push Notification Workers**:
   - A dedicated notification worker consumes events from `notificationCh`.
   - Queries all devices subscribed to the affected tournament with notifications enabled (`active = true`).
   - Sends notifications via the **Expo Push Notification Service** (Exponent) or **Firebase Cloud Messaging** (FCM).

4. **Zero-Login Device Authentication**:
   - Eliminates user/password friction: mobile devices automatically register an anonymous device token.
   - The raw credential token is hashed using **Argon2id** (with configurable salt, memory, iterations, and parallelism).
   - Subsequent requests authenticate using `Authorization: Bearer <deviceToken>`.

---

## Prerequisites

- **Go**: 1.26 or higher
- **Docker** and **Docker Compose** (for running PostgreSQL)

---

## Environment Configuration

Create a `.env` file in the `backend/` directory:

```bash
PORT=8080
DATABASE_URL="host=localhost user=postgres password=secret port=5432 sslmode=disable"
SECRET="your-argon2id-salt-secret-at-least-16-bytes"

# Optional: if using Firebase Cloud Messaging directly
GOOGLE_APPLICATION_CREDENTIALS="./.secrets/firebase-service-account.json"
```

### Environment Variables

| Variable | Description | Default / Example |
| :--- | :--- | :--- |
| `PORT` | Port the HTTP server listens on | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | `host=localhost user=postgres password=secret port=5432 sslmode=disable` |
| `SECRET` | Salt string used for Argon2id device credential hashing | Any random string |
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to Google Service Account JSON (if using FCM) | `./.secrets/firebase-service-account.json` |

---

## Database & Migrations

The application uses PostgreSQL with the `pgcrypto` extension for UUID generation. Migrations run **automatically** on application startup via `internal/database/migrate.go`.

### Database Schema

- **`devices`**:
  - `id` (UUID, Primary Key)
  - `credentials_hash` (TEXT, Unique) — Argon2id hash of device authentication token
  - `push_token` (TEXT, Unique) — Expo Push Token (`ExponentPushToken[...]`)
  - `platform` (VARCHAR(20)) — `android` / `ios`
  - `app_version` (VARCHAR(50))
  - `active` (BOOLEAN) — Global notification toggle for the device
  - `created_at`, `updated_at` (TIMESTAMPTZ)

- **`tournaments`**:
  - `id` (UUID, Primary Key)
  - `url` (TEXT, Unique) — Canonical chess-results tournament URL
  - `name` (TEXT) — Tournament title
  - `current_round` (INTEGER) — Most recently published round
  - `total_rounds` (INTEGER) — Total scheduled rounds
  - `last_checked_at`, `created_at`, `updated_at` (TIMESTAMPTZ)

- **`subscriptions`**:
  - `id` (UUID, Primary Key)
  - `device_id` (UUID, Foreign Key -> `devices.id` ON DELETE CASCADE)
  - `tournament_id` (UUID, Foreign Key -> `tournaments.id` ON DELETE CASCADE)
  - `enabled` (BOOLEAN, Default: `true`)
  - `UNIQUE(device_id, tournament_id)`

---

## Running Locally

### 1. Start the Database

```bash
docker compose up -d
```

This starts a PostgreSQL instance on port `5432` with password `secret`.

### 2. Run the Application

```bash
go run cmd/main.go
```

The server starts, executes any pending database migrations, starts the cron scheduler and notification workers, and binds to the configured port:

```
2026/09/18 13:46:00 Running migrations...
2026/09/18 13:46:00 Migrations finished
2026/09/18 13:46:00 Running server on port: 8080
```

### 3. Build Binary

```bash
make build-local
./bin/chessbox-notification
```

---

## Serverless Deployment (AWS Lambda & Terraform)

The backend can be built and deployed completely serverless on AWS using AWS Lambda (on ARM64 / Graviton), API Gateway HTTP API (v2), Amazon EventBridge, and an external PostgreSQL database provider (e.g. Neon, Supabase, Railway, CockroachDB).

### Architecture

- **API Lambda**: Interfaced via API Gateway HTTP API (`$default` catch-all route with `httpadapter.NewV2`). Runs with `APP_MODE=api`.
- **Worker Lambda**: Triggered periodically via Amazon EventBridge (`rate(1 minute)`). Scrapes tournaments from `chess-results.com` and dispatches push notifications to mobile devices. Runs with `APP_MODE=worker`.
- **External Database Integration**: The Lambda functions connect directly over TLS to your external PostgreSQL database (e.g. Neon, Supabase). Because the database is managed externally, there is no need for VPC, RDS, or NAT Gateway provisioning—yielding zero idle AWS infrastructure costs and ultra-low cold-start latency.
- **Embedded Migrations**: SQL migrations are embedded directly into the Go binary via `embed.FS` and executed safely on database connection.

### 1. Build Lambda Package

```bash
make build
```

This compiles a static Linux binary (`bootstrap`) targeted for `arm64` and packages it into `build/bootstrap.zip`.

### 2. Configure Terraform

```bash
cd terraform
cp terraform.tfvars.example terraform.tfvars
```

Set your external `database_url` in `terraform.tfvars`:
```hcl
database_url = "postgres://user:password@ep-sample-12345.us-east-1.aws.neon.tech/neondb?sslmode=require"
```

### 3. Deploy Infrastructure

```bash
make tf-init
make tf-plan
make tf-apply
```

Terraform outputs the public `api_endpoint` for your mobile app and the Lambda function names.

---

## API Reference

All responses follow a standard envelope:

```json
{
  "success": true,
  "message": "Human-readable status message",
  "data": { ... }
}
```

Errors return HTTP status 4xx/5xx with:

```json
{
  "success": false,
  "message": "Error description",
  "error": "Detailed error string"
}
```

---

### Health Check

#### `GET /v1/health`
Returns service health status.

- **Auth**: None
- **Response (200 OK)**:
```json
{
  "success": true,
  "message": "Healthy",
  "data": null
}
```

---

### Devices

#### `PUT /v1/devices`
Registers a new device or updates an existing device record.

- **Auth**: Optional. If the `Authorization: Bearer <deviceToken>` header is provided, the existing record is updated. If omitted or new, a new device token is generated.
- **Request Body**:
```json
{
  "pushToken": "ExponentPushToken[xxxxxxxxxxxxxxxxxxxxxx]",
  "platform": "android",
  "appVersion": "1.0.0",
  "active": true
}
```
- **Response (200 OK)**:
```json
{
  "success": true,
  "message": "Device updated successfuly",
  "data": {
    "deviceId": "c2b3d309-8bc3-4876-8877-3e1ee8ee306e",
    "deviceToken": "4b684ec2-4217-47b2-a634-8c65ce74d284"
  }
}
```

> [!IMPORTANT]
> Save the returned `deviceToken` on the client (e.g. via `expo-secure-store`). It must be passed in the `Authorization: Bearer <deviceToken>` header for all protected endpoints.

---

### Subscriptions

#### `POST /v1/subscriptions`
Subscribes the authenticated device to a tournament by URL. If the tournament does not exist yet, it is immediately scraped and registered.

- **Auth**: Required (`Bearer <deviceToken>`)
- **Request Body**:
```json
{
  "TournamentURL": "https://chess-results.com/tnr123456.aspx"
}
```
- **Response (200 OK)**:
```json
{
  "success": true,
  "message": "Subscribed to the tournament successfuly",
  "data": {
    "subscription_id": "8fa5bc52-16a7-4770-9828-98e35a116cf9"
  }
}
```

---

#### `GET /v1/subscriptions`
Lists all tournaments followed by the authenticated device.

- **Auth**: Required (`Bearer <deviceToken>`)
- **Response (200 OK)**:
```json
{
  "success": true,
  "message": "Listed subscriptions successfuly",
  "data": [
    {
      "id": "8fa5bc52-16a7-4770-9828-98e35a116cf9",
      "tournament_id": "97e68d04-1ec7-4ca2-8db4-9275b28d6c70",
      "tournament_name": "Open Internacional de Xadrez",
      "tournament_url": "https://chess-results.com/tnr123456.aspx",
      "tournament_round": 3,
      "tournament_total_rounds": 7
    }
  ]
}
```

---

#### `DELETE /v1/subscriptions/{id}`
Unsubscribes the authenticated device from a tournament.

- **Auth**: Required (`Bearer <deviceToken>`)
- **Parameters**: `id` — Subscription UUID
- **Response (200 OK)**: Empty data response or 200 with success status.

---

### Tournaments

#### `GET /v1/tournaments?q={query}`
Searches `chess-results.com` in real-time by tournament title or keyword.

- **Auth**: Required (`Bearer <deviceToken>`)
- **Query Parameters**:
  - `q`: Search string (e.g., `Festival de Xadrez`)
- **Response (200 OK)**:
```json
{
  "success": true,
  "message": "Search results found 2 tournaments",
  "data": [
    {
      "tournament_url": "https://s1.chess-results.com/tnr987654.aspx?lan=1",
      "tournament_name": "Torneio Aberto de Sao Paulo 2026",
      "start_date": "",
      "end_date": ""
    }
  ]
}
```
