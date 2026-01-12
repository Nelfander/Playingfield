# Playingfield

A lightweight backend and frontend project management platform.
This project is built with **Go (Echo framework)**, **PostgreSQL/Neon**, and eventually a React frontend.

---

## Current Features (Achieved So Far)

### Authentication

* User registration and login with **JWT-based authentication**.
* Passwords securely hashed.
* Role and status fields enforced (`role = "user"`, `status = "active"`).
* Handlers never rely on client-sent `owner_id`; identity always comes from JWT claims.

### Projects

* Users can create projects with a **name and description**.
* Enforces **per-user uniqueness**: a user cannot have two projects with the same name, but different users can have projects with the same name.
* Users can list only their own projects; JWT required.
* Backend correctly returns HTTP 409 for duplicate project names.

### Database

* PostgreSQL/Neon schema enforces defaults (`role`, `status`) and per-user uniqueness.
* SQLC is used to generate type-safe queries.
* Clean separation between **repository**, **service**, and **handler** layers.

---

## Future Goals

* Implement a **minimal React frontend**:

  * Login page with JWT integration.
  * Projects list for the logged-in user.
  * Create new projects with real-time validation for duplicates.
* Add **tasks under projects**.
* Improve **error handling and logging** further.
* Implement **user role management** (admin vs regular users).
* Add **unit and integration tests** for the project domain.

---

## Quick Start

1. Clone the repository:

```bash
git clone https://github.com/Nelfander/Playingfield.git
cd Playingfield
```

2. Set up Neon/PostgreSQL database and update `.env` with the connection string.

3. Generate SQLC queries (if changed):

```bash
sqlc generate
```

4. Run the server:

```bash
go run ./cmd/server
```

5. Use PowerShell or Postman to test:

```powershell
# Login
$login = Invoke-RestMethod -Method POST -Uri http://localhost:880/login -ContentType "application/json" -Body '{"email":"me@example.com","password":"supersecret"}'
$token = $login.token

# Create project
Invoke-RestMethod -Method POST -Uri http://localhost:880/projects -Headers @{ Authorization = "Bearer $token" } -ContentType "application/json" -Body '{"name":"Ball","description":"First Ball project"}'

# List projects
Invoke-RestMethod -Method GET -Uri http://localhost:880/projects -Headers @{ Authorization = "Bearer $token" }
```

---

## Code Structure

* `internal/domain/user` – domain model, repository interfaces.
* `internal/domain/projects` – project domain, service, repository interface.
* `internal/infrastructure/postgres` – SQLC-based repository implementation, DB adapter.
* `cmd/server` – Echo server initialization and routing.

---

## Known Issues & How I Solved Them

### 1. Users created with empty role/status

* Issue: Old rows in Neon/PostgreSQL had `role=""` and `status=NULL`, breaking login and JWT logic.
* Solution:

  * Set **default values in DB**: `role TEXT NOT NULL DEFAULT 'user'`, `status TEXT NOT NULL DEFAULT 'active'`.
  * Updated **SQLC `CreateUser` query** to include `role` and `status`.
  * Updated Go repository to explicitly set `Role` and `Status` during user creation.
  * Re-registered users to clean the broken rows.

### 2. Project creation owned by the wrong user

* Issue: `owner_id` was sometimes taken from request instead of JWT, causing `0` or incorrect IDs.
* Solution:

  * Handlers now derive `owner_id` **exclusively from JWT claims**.
  * Both **CreateProject** and **ListProjects** enforce this invariant.
  * Removed `OwnerID` from client request structs.

### 3. Duplicate project names

* Issue: Initially, there was no constraint enforcing per-user uniqueness. Users could create multiple projects with the same name.
* Solution:

  * Added **database unique constraint** on `(owner_id, name)`.
  * Handled `duplicate key` errors in the `CreateProject` handler, returning `409 Conflict` with JSON error.
  * PowerShell commands now show a friendly error message instead of a generic 500.

### 4. Generic Internal Server Errors in PowerShell

* Issue: PowerShell throws `Invoke-RestMethod : 500 Internal Server Error` for any failed request.
* Solution:

  * Added **debug logging** in handlers to print real errors to server console.
  * Ensured handler returns **specific HTTP status codes** (`400`, `401`, `409`) with JSON error bodies.

---

## Architecture & Flow Diagram

```text
                        ┌───────────────┐
                        │    Client     │
                        │ (React / PS)  │
                        └───────┬───────┘
                                │
                                │ POST /users (register)
                                │ POST /login (login)
                                ▼
                        ┌───────────────┐
                        │   HTTP Server │
                        │   (Echo / Go) │
                        └───────┬───────┘
                                │
                                │ JWT Middleware
                                │ Extract user ID
                                ▼
                 ┌─────────────────────────┐
                 │ ProjectHandler / UserHandler │
                 │  - Validate requests        │
                 │  - Bind JSON               │
                 │  - Pass data to Service    │
                 └─────────┬───────────────┘
                           │
                           ▼
                   ┌───────────────┐
                   │   Service     │
                   │  - Business   │
                   │    logic      │
                   │  - Enforce    │
                   │    per-user   │
                   │    uniqueness │
                   └───────┬───────┘
                           │
                           ▼
                    ┌─────────────┐
                    │ Repository  │
                    │  (SQLC)     │
                    │ - SQL queries│
                    │ - Insert /   │
                    │   Fetch      │
                    └───────┬─────┘
                            │
                            ▼
                     ┌──────────────┐
                     │   PostgreSQL │
                     │   / Neon DB  │
                     │ - users      │
                     │ - projects   │
                     └──────────────┘

## 🛠 Development History

<details>
  <summary><b>Jan 12, 2026: Frontend & Security Integration</b> (Click to expand)</summary>

  #### Frontend Updates
  * **React Frontend Implemented:** * Full Login page with JWT authentication integration.
    * Dynamic Projects list with a "Load projects" toggle to manage visibility.
    * **Interactive Members & Tasks:** Each project includes "Show Members" and "Show Tasks" buttons with smooth slide-down animations.
    * **Empty State Handling:** Fixed crashes when backend returns `null` projects; added fallback messages ("No members yet").
  * **Polished UI:** * Centered layout with modern Glassmorphism (frosted glass effect).
    * High-resolution mountain background (Moraine Lake) with readability overlays.

  #### Project Users & Roles (Backend)
  * **Ownership Logic:** Only project owners are permitted to remove users from a project.
  * **JWT Claims:** Security checks are now enforced using role-based claims within the JWT.
  * **Error Handling:** API now returns descriptive error messages for unauthorized actions (401/403).

  #### Misc
  * **Verified Requests:** Example PowerShell scripts for adding/removing users are now fully functional.
  * **Live Integration:** The frontend is fully integrated with the backend API for real-time data display.
</details>

<details>
  <summary><b>Future Goals</b></summary>
  
  - [ ] Add Search/Filter functionality.
  - [ ] Implement Task creation from the UI.
  - [ ] Add database persistence (PostgreSQL/SQLite).
</details>