# 1GO

1GO is a gym booking API, Go service for managing gyms, classes, class sessions, and user bookings. It provides JWT-based authentication, role-based access for admin operations, automatic PostgreSQL migrations on startup, and a small seed utility for bootstrapping an admin account and demo data.

## Features

- User registration and login
- JWT-protected endpoints
- Admin-only management for gyms, classes, and sessions
- Public listing endpoints for gyms, classes, and sessions
- Booking creation with wallet balance deduction
- Booking cancellation with refund handling
- Attendance tracking via QR codes
- Automatic background maintenance (e.g., auto-completing expired sessions)
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
│   ├── auth/        # JWT generation and parsing (including Attendance tokens)
│   ├── config/      # Environment-based configuration
│   ├── domain/      # Core entities
│   ├── handler/     # HTTP handlers
│   ├── infrastructure/
│   │   └── database/# DB connection and migration bootstrap
│   ├── middleware/  # Auth, RBAC, request logging
│   ├── repository/  # Database access layer
│   ├── usecase/     # Business logic
│   └── worker/      # Background maintenance workers
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
| `SEED_DEMO_DATA` | `false` | No | When `true`, also seeds demo gyms, classes, and sessions |
| `BYPASS_ATTENDANCE_TIME_CHECK` | `false` | No | When `true`, allows marking attendance before session starts (Testing only) |

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

### 1. Start PostgreSQL.
### 2. Clone the repository:

```bash
git clone https://github.com/yertaypert/gym-booking-api.git # or SSH
cd gym-booking-api
```

### 3. Create database:

```bash
psql -U postgres -c "CREATE DATABASE gym_booking;" # replace postgres with your superuser
```

### 4. Configure environment variables, copy example env:

```bash
cp .env.example .env
```

### 5. Run the API:

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

There are four roles in the system:

- `user`: Standard member, can book sessions and view own profile.
- `trainer`: Professional instructor, can be assigned to gyms.
- `gym_owner`: Can manage gyms, classes, and sessions for their own properties.
- `admin`: Full system access, including global management and role updates.

Authentication is based on a Bearer token:

```http
Authorization: Bearer <jwt>
```

Access rules:

- Public: browse gyms, gym details, classes, and sessions.
- Authenticated user: view own profile, top up wallet, create bookings, cancel own bookings, scan attendance.
- Gym Owner/Admin: manage gyms, classes, and sessions. View attendee lists and generate QR codes.
- Admin: manage all gyms and change user roles.


## Booking and Wallet Behavior

Booking creation is transactional:

- The user must have enough balance to cover the session price
- The session must have at least one available slot
- On success, the service creates a booking, decreases the user balance, decrements available slots, and records a payment transaction

Booking cancellation is also transactional:

- Users can cancel their own bookings
- Admins can cancel any booking
- On success, the booking is marked as cancelled, the session slot is restored, and the user receives a refund transaction

## Background Workers

The system includes a background worker manager that handles periodic maintenance tasks:

- **Session Auto-Completion:** Runs every minute to transition `active` sessions to `completed` status once their `end_time` has passed. This ensures data integrity and prevents modifications to past events.

## Attendance & QR Codes

Gym owners and admins can generate attendance tokens for their sessions. Users can then "scan" these tokens to mark themselves as attended.

1. **Generation:** Owner calls `/sessions/{sessionId}/attendance-qr` to get a temporary token.
2. **Scanning:** User calls `/attendance/scan` with the token/qr_code to mark attendance.
3. **Verification:** System verifies the user has a confirmed booking for that session and that the session has started (unless bypassed by `BYPASS_ATTENDANCE_TIME_CHECK=true`).

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
| **Classes** | `GET` | `/classes` | List all distinct class names |
| **Classes** | `GET` | `/classes/{name}` | List all gyms offering a class by name |
| **Classes** | `GET` | `/gyms/{id}/classes` | Return all classes for a specific gym |
| **Classes** | `POST` | `/gyms/{id}/classes` | Create a class for a gym (Admin required) |
| **Sessions** | `GET` | `/classes/{name}/sessions` | Search sessions by class name (supports `date`, `start_time`, `end_time` filters) |
| **Sessions** | `GET` | `/sessions/{id}` | Return detailed info for a specific session |
| **Sessions** | `GET` | `/gyms/{gymId}/classes/{classId}/sessions` | Return all sessions for a specific class in a gym |
| **Sessions** | `POST` | `/gyms/{gymId}/classes/{classId}/sessions` | Create a session for a class (Admin required) |
| **Bookings** | `POST` | `/sessions/{sessionId}/bookings` | Create a booking for a session |
| **Bookings** | `POST` | `/bookings/{bookingId}/cancel` | Cancel an existing booking |

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
