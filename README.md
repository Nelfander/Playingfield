📌 Playingfield Backend

A production‑grade Go backend for user authentication and project management, built with:

✅ Echo HTTP framework
✅ PostgreSQL (Neon) backend (planned)
✅ sqlc for type‑safe database access (in progress)
✅ JWT authentication
✅ Clean / hexagonal‑inspired architecture
✅ Focus on tests and correctness

🧠 Features Implemented
User System

Register new users with hashed passwords

Login with email/password and return JWT

Me endpoint (GET /me) returns authenticated user info

JWT middleware and role support

Tests covering:

Registration

Login

Invalid credentials

Inactive account

JWT validation

📦 Projects Domain (Work‑in‑Progress)

Projects domain created with a fake repository for fast iteration

Projects can be created and listed via API

Projects are linked to authenticated users (JWT)
— no more anonymous owner_id = 0

🚀 Getting Started

1. Clone the repo

git clone https://github.com/Nelfander/Playingfield.git
cd Playingfield


2. Install dependencies

go mod tidy


3. Run the server

go run cmd/api/server.go


The server will start on port :880.

📡 API Endpoints (Current)
Health Check
GET /health


Check if server is running.

User Endpoints
Register
curl -X POST http://localhost:880/users \
  -H "Content-Type: application/json" \
  -d '{"email":"me@example.com","password":"supersecret"}'

Login
curl -X POST http://localhost:880/login \
  -H "Content-Type: application/json" \
  -d '{"email":"me@example.com","password":"supersecret"}'


Expected JSON response contains:

{
  "token": "JWT_TOKEN_HERE",
  "user": {
    "id": 1,
    "email": "me@example.com",
    "role": "user",
    "created_at": "2026-01-11T..."
  }
}

Get current user
curl http://localhost:880/me \
  -H "Authorization: Bearer <your_jwt_token>"

Projects Endpoints (Auth Required)
Create a project
curl -X POST http://localhost:880/projects \
  -H "Authorization: Bearer <your_jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{"name":"First Project","description":"This is mine"}'

List projects (for authenticated user)
curl http://localhost:880/projects \
  -H "Authorization: Bearer <your_jwt_token>"

🧱 Architecture Overview
cmd/api
    └── main.go                      // Entry point, server start
internal/
    ├── app/
    │    └── server.go               // HTTP server setup
    │    └── routes.go
    ├── domain/
    │    ├── user/                   // User entity, service, repository interface
    │    └── projects/               // Projects domain
    ├── interfaces/http/
    │    ├── handlers/               // HTTP handlers
    │    ├── middleware/             // JWT and role validation
    │    └── dto/                    // HTTP request/response structs
    ├── infrastructure/
    │    ├── auth/                   // JWT manager, password utils
    │    └── postgres/               // Postgres integration (in progress)
pkg/
    ├── config/                      // Env config loader
    └── logger/                      // Logger setup

🧪 Testing

All tests pass. Run them with:

go test ./...

You’ll see tests for:

User registration and login

JWT middleware

Inactive account handling

Me endpoint

Projects domain (fake repo)

🛠 Next Steps

✅ Enforce JWT for all protected endpoints
✅ Wire projects domain with real PostgreSQL via sqlc
✅ Add tasks under projects
✅ Add activities under tasks
✅ Build minimal React frontend (login + projects list)
✅ Add authorization rules (roles, permissions)
✅ Add unit + integration tests for new domains

🗂 Future Frontend MVP

The frontend will let an authenticated user:

Login

See account info

Create / view projects

Navigate to projects → tasks → activity log

React recommended (Vite + TS + Tailwind CSS) for a modern corporate‑style stack.

🙌 Contributing

This project is meant for real learning, real feedback loops, and real standards.
Feel free to open issues or PRs. All code must include tests.

📜 License

MIT License (same as code in this repo)

NEXT STEPS
1. implement projects, tasks, and activities domain logic
2. apply role-based permissions for projects/tasks/activities
3. add unit and integration tests for all layers   ✅
4. dockerize backend for deployment
5. add frontend (HTML/JS) to interact with backend