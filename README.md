# 1GO

1GO is a gym booking API, Go service for managing gyms, classes, class sessions, and user bookings. It provides JWT-based authentication, role-based access for admin operations, automatic PostgreSQL migrations on startup, and a small seed utility for bootstrapping an admin account and demo data.

## Features

- User registration and login
- JWT-protected endpoints
- Admin-only management for gyms, classes, and sessions
- Public listing endpoints for gyms, classes, and sessions
- Booking creation with wallet balance deduction
- Booking cancellation with refund handling
- PostgreSQL persistence with automatic migrations on application startup
- Seed command for creating the initial admin user and optional demo records

## Tech Stack

- Go 1.25
- PostgreSQL 16
- `net/http` standard library router
- `golang-migrate` for schema migrations
- JWT for authentication
- Docker and Docker Compose for local development

## Project Structure

```text
.
├── cmd/
│   ├── api/         # HTTP server entry point
│   └── seed/        # Seed admin user and optional demo data
├── database/
│   └── migrations/  # SQL migrations
├── internal/
│   ├── auth/        # JWT generation and parsing
│   ├── config/      # Environment-based configuration
│   ├── domain/      # Core entities
│   ├── handler/     # HTTP handlers
│   ├── infrastructure/
│   │   └── database/# DB connection and migration bootstrap
│   ├── middleware/  # Auth, RBAC, request logging
│   ├── repository/  # Database access layer
│   └── usecase/     # Business logic
├── pkg/
│   └── logger/      # Application logger setup
├── Dockerfile
└── docker-compose.yml
```

## Requirements

- Go 1.25+
- PostgreSQL 16+
- Docker Desktop and Docker Compose, if you want to run the stack in containers

## Environment Variables

The application reads configuration from environment variables and optionally from a local `.env` file.

| Variable | Default | Required | Description |
| --- | --- | --- | --- |
| `DB_HOST` | `localhost` | No | PostgreSQL host |
| `DB_PORT` | `5432` | No | PostgreSQL port |
| `DB_USER` | `postgres` | No | PostgreSQL username |
| `DB_PASSWORD` | `password` | No | PostgreSQL password |
| `DB_NAME` | `gym_booking` | No | PostgreSQL database name |
| `SERVER_PORT` | `:8080` | No | HTTP server listen address |
| `JWT_SECRET` | `super-secret-change-me` | No | Declared in config, but the current JWT implementation uses a hardcoded signing key |
| `SEED_ADMIN_EMAIL` | none | Yes for `cmd/seed` | Admin email created by the seed command |
| `SEED_ADMIN_PASSWORD` | none | Yes for `cmd/seed` | Admin password created by the seed command |
| `SEED_ADMIN_FULL_NAME` | `Initial Admin` | No | Admin full name for seed data |
| `SEED_DEMO_DATA` | `false` | No | When `true`, also seeds a demo gym, class, and session |

Example `.env`:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=gym_booking
SERVER_PORT=:8080
JWT_SECRET=super-secret-change-me

SEED_ADMIN_EMAIL=admin@example.com
SEED_ADMIN_PASSWORD=Admin123
SEED_ADMIN_FULL_NAME=Initial Admin
SEED_DEMO_DATA=true
```

## Running Locally

1. Start PostgreSQL.
2. Create a `.env` file in the project root.
3. Run the API:

```bash
go run ./cmd/api
```

The application connects to PostgreSQL, runs migrations automatically from `database/migrations`, and starts the HTTP server on `SERVER_PORT`.

## Running with Docker Compose

```bash
docker compose up --build
```

This starts:

- `app` on `http://localhost:8080`
- PostgreSQL on `localhost:5433`

The app container overrides `DB_HOST=postgres` so it can connect to the database service inside the Compose network.

## Seeding Data

Create the initial admin user:

```bash
go run ./cmd/seed
```

What the seed command does:

- Upserts an admin user using `SEED_ADMIN_EMAIL`
- Updates the password and full name if that admin already exists
- Optionally inserts demo data when `SEED_DEMO_DATA=true`


## Authentication and Roles

There are two roles in the system:

- `user`
- `admin`

Authentication is based on a Bearer token:

```http
Authorization: Bearer <jwt>
```

Access rules:

- Public: browse gyms, gym details, classes, and sessions
- Authenticated user: view own profile, create bookings, cancel own bookings
- Admin: all authenticated user capabilities plus create gyms, classes, and sessions

## Booking and Wallet Behavior

Booking creation is transactional:

- The user must have enough balance to cover the session price
- The session must have at least one available slot
- On success, the service creates a booking, decreases the user balance, decrements available slots, and records a payment transaction

Booking cancellation is also transactional:

- Users can cancel their own bookings
- Admins can cancel any booking
- On success, the booking is marked as cancelled, the session slot is restored, and the user receives a refund transaction


## API Endpoints

| Category | Method | Endpoint | Description |
| :--- | :--- | :--- | :--- |
| **Auth** | `POST` | `/register` | Register a new user account |
| **Auth** | `POST` | `/login` | Login and obtain JWT token |
| **Auth** | `GET` | `/me` | Return the authenticated user |
| **Auth** | `GET` | `/admin/me` | Return the authenticated admin user |
| **Gyms** | `GET` | `/gyms` | Return all gyms |
| **Gyms** | `GET` | `/gyms/{id}` | Return a single gym by ID |
| **Gyms** | `POST` | `/gyms` | Create a gym (Admin required) |
| **Classes** | `GET` | `/gyms/{id}/classes` | Return all classes for a gym |
| **Classes** | `POST` | `/gyms/{id}/classes` | Create a class for a gym (Admin required) |
| **Sessions** | `GET` | `/gyms/{gymId}/classes/{classId}/sessions` | Return all sessions for a class |
| **Sessions** | `POST` | `/gyms/{gymId}/classes/{classId}/sessions` | Create a session for a class |

Base URL:

```text
http://localhost:8080
```

### Auth

#### `POST /register`

Creates a regular user account.

Request:

```json
{
  "email": "user@example.com",
  "password": "Password123",
  "full_name": "Jane Doe"
}
```

Validation:

- `email` must be valid
- `full_name` must not be empty
- `password` must be at least 8 characters and include uppercase, lowercase, and a digit

Response:

- `201 Created` with plain text `user created`
- `409 Conflict` if email already exists

#### `POST /login`

Authenticates a user and returns a JWT.

Request:

```json
{
  "email": "user@example.com",
  "password": "Password123"
}
```

Response:

```json
{
  "token": "<jwt>"
}
```

#### `GET /me`

Returns the authenticated user.

Auth:

- Bearer token required

#### `GET /admin/me`

Returns the authenticated admin user.

Auth:

- Bearer token required
- Admin role required

### Gyms

#### `GET /gyms`

Returns all gyms.

#### `GET /gyms/{id}`

Returns a single gym by ID.

#### `POST /gyms`

Creates a gym.

Auth:

- Bearer token required
- Admin role required

Request:

```json
{
  "name": "Downtown Gym",
  "address": "123 Main St",
  "description": "Open 24/7 demo gym"
}
```

### Classes

#### `GET /gyms/{id}/classes`

Returns all classes for a gym.

#### `POST /gyms/{id}/classes`

Creates a class for a gym.

Auth:

- Bearer token required
- Admin role required

Request:

```json
{
  "name": "Yoga",
  "max_capacity": 20
}
```

### Sessions

#### `GET /gyms/{gymId}/classes/{classId}/sessions`

Returns all sessions for a class in a gym.

#### `POST /gyms/{gymId}/classes/{classId}/sessions`

Creates a session for a class.

Auth:

- Bearer token required
- Admin role required

Request:

```json
{
  "start_time": "2026-04-06T10:00:00Z",
  "end_time": "2026-04-06T11:00:00Z",
  "price": 15
}
```

Notes:

- `start_time` and `end_time` must be RFC3339 timestamps
- `price` must be valid according to business rules in the use case layer

### Bookings

#### `POST /sessions/{sessionId}/bookings`

Creates a booking for the authenticated user.

Auth:

- Bearer token required

Success response:

```json
{
  "booking_id": 1,
  "message": "Booking created successfully"
}
```

Possible failure conditions include:

- session not found
- no available slots
- insufficient balance
- duplicate booking for the same session

#### `POST /bookings/{bookingId}/cancel`

Cancels an existing booking.

Auth:

- Bearer token required

Success response:

```json
{
  "message": "Booking cancelled successfully"
}
```

## Example Workflow

1. Start PostgreSQL and run the API.
2. Seed an admin account with `go run ./cmd/seed`.
3. Log in as admin and create a gym.
4. Create a class under that gym.
5. Create a session for the class.
6. Register a normal user and log in.
7. Ensure the user has enough balance in the database.
8. Create a booking for the session.
9. Cancel the booking to verify refund behavior.

## Database Schema

The current schema contains these main tables:

- `users`
- `gyms`
- `classes`
- `class_sessions`
- `bookings`
- `transactions`

It also uses PostgreSQL enum types for:

- `user_role`
- `session_status`
- `booking_status`
- `transaction_type`

## Testing

Run the full test suite with:

```bash
go test ./...
```

Verified in this repository on April 5, 2026.

## Notes

- Migrations run automatically every time the API starts.
- The HTTP server uses Go's standard `http.ServeMux` with method-aware routes.
