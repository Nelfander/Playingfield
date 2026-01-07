PLAYINGFIELD BACKEND
Overview

Playingfield is a production-ready Go backend using Echo, PostgreSQL (Neon), and sqlc.
Architecture follows a clean/hexagonal-inspired monolith pattern

Tech Stack

-Language: Go
-HTTP Framework: Echo
-Database: PostgreSQL (Neon)
-SQL Layer: sqlc (type-safe queries)
-Driver: pgx/v5
-Authentication: JWT
-Configuration: .env files

PROJECT STRUCTURE
cmd/api                   - Main entrypoint
internal/app               - Server bootstrap and routes
internal/domain            - Business logic
internal/interfaces/http   - Handlers, middleware, DTOs
internal/infrastructure    - Postgres, auth, worker
pkg                        - Config, logger

GETTING STARTED
1. Copy .env.example to .env and fill in Neon credentials
2. Install dependencies: go mod tidy
3. Run server: go run ./cmd/api
4. Test health endpoint: Invoke-RestMethod http://localhost:880/health
5. Register a user: Invoke-RestMethod -Method POST -Uri http://localhost:880/users -ContentType "application/json" -Body '{"email":"test@example.com","password":"123456"}'
6. Login:  Invoke-RestMethod -Method POST -Uri http://localhost:880/login -ContentType "application/json" -Body '{"email":"test@example.com","password":"123456"}'

CURRENT FEATURES
- Echo server with /users route
- Health check endpoint /health
- PostgreSQL connection via pgxpool
- Type-safe sqlc-generated queries and repositories
- Layered architecture: Repository → Service → Handler
- JWT auth scaffolding ready

NEXT STEPS
1. Complete domain logic: projects, tasks, activities
2. Add JWT auth to protected endpoints
3. Implement user roles (admin, regular)
4. Background worker tasks
5. Unit and integration tests
6. Dockerize backend for deployment