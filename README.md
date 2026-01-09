PLAYINGFIELD BACKEND
Overview

Playingfield is a production-ready Go backend using Echo, PostgreSQL (Neon), and sqlc.
Architecture follows a clean/hexagonal-inspired monolith pattern
interfaces/http → domain → infrastructure.


Tech Stack

-Language: Go
-HTTP Framework: Echo
-Database: PostgreSQL (Neon)
-SQL Layer: sqlc (type-safe queries)
-Driver: pgx/v5
-Authentication: JWT
-Configuration: .env files

PROJECT STRUCTURE
cmd/api                   - main entrypoint
internal/app               - server bootstrap and route registration
internal/domain            - business logic, user service, repository interfaces
internal/interfaces/http   - HTTP handlers, middleware, DTOs
internal/infrastructure    - database (Postgres), auth, workers
pkg                        - config, logger utilities

GETTING STARTED
1. Copy .env.example to .env and fill in Neon credentials and JWT_SECRET
2. Install dependencies: go mod tidy
3. Run the server: go run ./cmd/api
4. Test health endpoint:
   Invoke-RestMethod http://localhost:880/health
5. Register a user:
   Invoke-RestMethod -Method POST -Uri http://localhost:880/users -ContentType "application/json" -Body '{"email":"test@example.com","password":"123456"}'
6. Login a user:
   Invoke-RestMethod -Method POST -Uri http://localhost:880/login -ContentType "application/json" -Body '{"email":"test@example.com","password":"123456"}'
7. Access current user info:
   Invoke-RestMethod -Method GET -Uri http://localhost:880/me -Headers @{Authorization="Bearer <TOKEN_HERE>"}
8. Admin route (requires "admin" role):
   Invoke-RestMethod -Method GET -Uri http://localhost:880/admin -Headers @{Authorization="Bearer <TOKEN_HERE>"}

CURRENT FEATURES
- echo server with `/users`, `/login`, `/me`, `/admin`, `/health` endpoints
- postgreSQL connection via pgxpool
- sqlc-generated repositories
- repository → service → handler wiring
- JWT auth with role-based middleware
- admin seeding available

NEXT STEPS
1. implement projects, tasks, and activities domain logic
2. apply role-based permissions for projects/tasks/activities
3. add unit and integration tests for all layers
4. dockerize backend for deployment
5. add frontend (HTML/JS) to interact with backend