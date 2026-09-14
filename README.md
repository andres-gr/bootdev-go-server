# Chirpy

A simple social media API built with Go. Users can create accounts, post chirps (short messages), and authenticate with JWT tokens.

## Prerequisites

- Go 1.27+
- PostgreSQL database
- `goose` for database migrations

## Setup

### 1. Install dependencies

```bash
go mod tidy
```

### 2. Configure environment

Create a `.env` file with the following variables:

```env
DB_URL=postgres://username:password@localhost:5432/chirpy?sslmode=disable
JWT_SECRET=your-secret-key-here
POLKA_KEY=your-polka-api-key
PLATFORM=dev
```

### 3. Set up the database

Create the database and run migrations:

```bash
createdb chirpy
go run ./cmd/migrate up
```

## Running

```bash
go run .
```

The server will start on `http://localhost:8080`.

## Architecture

```
server/
├── main.go                    # Entry point, HTTP server setup
├── routehandlers.go           # All API route handlers
├── middleware.go              # Middleware (metrics, auth)
├── utils.go                   # Response helpers, user formatting
├── models.go                  # Data models
├── internal/
│   ├── auth/                  # JWT, password hashing, token generation
│   ├── database/              # SQLC-generated database layer
│   └── ...
└── sql/
    ├── schema/                # Database migrations
    └── queries/               # SQL queries
```

## API Endpoints

### Authentication

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/users` | Create a new user account |
| POST | `/api/login` | Login with email and password |
| POST | `/api/refresh` | Refresh JWT token using refresh token |
| POST | `/api/revoke` | Revoke a refresh token |

### Users

| Method | Path | Description |
|--------|------|-------------|
| PUT | `/api/users` | Update user email and password (requires JWT) |

### Chirps

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/chirps` | Create a chirp (requires JWT) |
| GET | `/api/chirps` | List all chirps. Query params: `author_id`, `sort` (asc/desc) |
| GET | `/api/chirps/{id}` | Get a single chirp by ID |
| DELETE | `/api/chirps/{id}` | Delete a chirp (requires JWT, must own chirp) |

### Health

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/healthz` | Health check endpoint |

### Admin

| Method | Path | Description |
|--------|------|-------------|
| GET | `/admin/metrics` | Show visit counter (HTML) |
| POST | `/admin/reset` | Reset database (dev only) |

### Webhooks

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/polka/webhooks` | Receive Polka webhook events (user upgrades) |

### Static Files

| Method | Path | Description |
|--------|------|-------------|
| GET | `/app/*` | Serve static frontend files |

## Authentication Flow

1. **Register** → POST `/api/users` returns JWT + refresh token
2. **Login** → POST `/api/login` returns JWT + refresh token
3. **Refresh** → POST `/api/refresh` with refresh token returns new JWT
4. **Revoke** → POST `/api/revoke` with refresh token invalidates it

All protected endpoints require `Authorization: Bearer <token>` header.

## Data Validation

- **Password:** 3-16 characters, no spaces
- **Chirps:** Max 140 characters, profanity filtered
- **Email:** Required, must be non-empty

## Notes

- JWT tokens expire after 1 hour
- Refresh tokens expire after 60 days
- Admin reset only works when `PLATFORM=dev`
- Chirp deletion requires ownership verification