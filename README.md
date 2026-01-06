PLAYINGFIELD BACKEND

OVERVIEW
Playingfield is a Go backend using Echo, PostgreSQL (Neon), and sqlc.
Architecture: clean/hexagonal-inspired monolith:
interfaces/http -> domain -> infrastructure

TECH STACK
- Go
- Echo HTTP framework
- PostgreSQL (Neon)
- SQLC for type-safe queries
- pgx/v5 driver
- JWT auth scaffolding
- .env for configuration

PROJECT STRUCTURE
cmd/api             - main entrypoint
internal/app        - server setup and routes
internal/domain     - business logic
internal/interfaces/http - handlers, middleware, DTOs
internal/infrastructure   - postgres, auth, worker
pkg                 - config, logger

GETTING STARTED
1. Copy .env.example to .env and fill in Neon credentials
2. Install dependencies: go mod tidy
3. Run server: go run ./cmd/api
4. Test health endpoint: Invoke-RestMethod http://localhost:880/health
5. Test /users POST:
   curl -X POST http://localhost:880/users -H "Content-Type: application/json" -d '{"email":"test@example.com","password":"123456"}'

CURRENT FEATURES
- Echo server with /users route
- Health check endpoint /health
- PostgreSQL connection via pgxpool
- sqlc-generated repositories
- Repository -> Service -> Handler wiring
- JWT auth scaffolding ready

NEXT STEPS
1. Complete domain logic: projects, tasks, activities
2. Add JWT auth to protected endpoints
3. Implement background worker tasks
4. Add unit and integration tests
5. Dockerize backend for deployment