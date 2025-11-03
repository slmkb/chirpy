# Chirpy

Chirpy is a backend server for a simple microblogging platform, written in Go. It allows users to post and view short-form messages called “chirps,” similar to tweets, with APIs for user authentication, content management, and administrative functions.

## In-Depth Description

Chirpy provides a RESTful HTTP API for managing user accounts and chirps. The application features user registration, login with JWT-based authentication, password updates, session management, and secure endpoints for interacting with chirps (creating, viewing, and deleting).

Chirps are short text messages, each associated with a user and time of creation. To protect against inappropriate content, chirps are filtered for a set of profanities before being stored. Messages exceeding 140 characters are rejected.

For administrative and developmental operations, Chirpy exposes endpoints for checking the server’s health, viewing system metrics, and resetting certain states. Administrative routes include a metrics page to monitor file server hits and a reset feature for test environments.

## Features

- User registration and login (JWT-based authentication)
- Secure password hashing and password change endpoints
- Post new chirps (short messages, max 140 characters)
- Profanity filtering for chirps
- View all chirps, with optional filtering and sorting
- View individual chirps by ID
- Delete chirps (accessible to content authors)
- Webhook endpoint for user role upgrades
- Token refresh and revoke mechanisms
- Admin endpoints:
  - Health check (`/api/healthz`)
  - Metrics display for admin users
  - System reset for testing or dev
- Static file serving for any web frontend

## Example Endpoints

- `POST /api/users`: Create a new user
- `POST /api/login`: Authenticate and retrieve a JWT
- `POST /api/chirps`: Create a new chirp (requires authentication)
- `GET /api/chirps`: List all chirps
- `GET /api/chirps/{id}/`: Get a chirp by its ID
- `DELETE /api/chirps/{id}`: Delete a chirp (if you are the author)
- `POST /api/refresh`: Refresh JWT
- `POST /api/revoke`: Revoke authentication token
- `GET /admin/metrics`: Show basic usage metrics
- `POST /admin/reset`: Reset the database (development use)

## Getting Started

```bash
git clone https://github.com/slmkb/chirpy.git
cd chirpy
go build
./chirpy
```

The default port is `8080`. Requires a PostgreSQL database and several environment variables (`DB_URL`, `JWT_SECRET`, `POLKA_API_KEY`).

## License

(Please specify the license.)

---

Made with Go.