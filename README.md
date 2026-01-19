# Playingfield

A real-time, lightweight project management platform.
Built with **Go (Echo framework)**, **PostgreSQL (Neon)**, and a **React (TypeScript)** frontend.

---

## 🌟 Key Features

### 💬 Real-Time Project Chat
* **Contextual Messaging:** Each project features a dedicated real-time chat room.
* **Smart UI Alignment:** Messages are intelligently aligned—your messages appear on the right ("Me") in blue, while teammates' messages appear on the left in gray.
* **Live Timestamps:** Every message is stamped with a human-readable time (e.g., 14:05) for better context.
* **History Persistence:** New members can see previous project discussions instantly upon joining.

### ⚡ Real-Time Synchronization (WebSockets)
* **Global Hub:** A custom WebSocket Hub manages concurrent client connections and room-based broadcasting.
* **Live Dashboard Updates:** * **Project Membership:** Projects appear/vanish from your dashboard instantly when you are added or removed by an owner.
    * **Global Deletion:** If an owner deletes a project, it is wiped from every member's screen in real-time.
* **Automatic Member Sync:** Live updates to member lists without requiring page refreshes.

### 🔐 Authentication & Security
* **JWT-Based Auth:** Secure registration and login with token-based identity.
* **Identity Integrity:** Handlers derive `user_id` exclusively from verified JWT claims, preventing "ID Spoofing."
* **Ownership Enforcement:** Destructive actions (deleting projects, removing members) are restricted to the project owner via backend middleware.

---

## 🛠 Tech Stack
* **Backend:** Go (Echo Framework), SQLC (Type-safe SQL), Gorilla WebSocket.
* **Frontend:** React 18, TypeScript, Vite, CSS3 (Glassmorphism).
* **Database:** PostgreSQL (Hosted on Neon.tech).
* **Communication:** REST API for state + WebSockets for reactivity.

## Future Goals
* Implement **Task creation from the UI**.
* Improve **error handling and logging** further.
* Implement **user role management** (admin vs regular users).
* Add **unit and integration tests** for the project domain.
* Add **Project group chats and 1 on 1 individual project member chat feature**.

---

## Quick Start
<details>
<summary><b>Quick Start!</b> (Click to expand)</summary>
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
</details>


<details>
<summary><b>Code Structure</b> (Click to expand)</summary>
* `internal/domain/user` – domain model, repository interfaces.
* `internal/domain/projects` – project domain, service, repository interface.
* `internal/infrastructure/postgres` – SQLC-based repository implementation, DB adapter.
* `cmd/server` – Echo server initialization and routing.

---
</details>

<details>
<summary><b>Known Issues & How I Solved Them</b> (Click to expand)</summary>

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
  
### 5. The "Vanishing List" Bug
* **Issue:** WebSocket updates triggered a UI toggle, closing the project list.
* **Solution:** Separated fetchProjects logic from the showProjects toggle state, allowing background refreshes without affecting UI visibility.

### 6. The "Room 0" Connection Spam
* **Issue:** Components re-rendering caused the WebSocket hook to reconnect multiple times, flooding the server with "User joined Room 0" warnings.
* **Solution:** Implemented `useRef` to maintain a stable socket connection and a **cleanup function** to close old sockets before new ones open. Added a "connection lock" to prevent React Strict Mode from double-connecting.

### 7. The "Null Map" Crash
* **Issue:** Fetching chat history for a brand-new project returned `null`, causing the frontend `.map()` function to crash the app.
* **Solution:** Implemented "Guard Clauses" in the frontend (`history || []`) and ensured the Go backend initializes empty slices instead of returning nil.


---------------
</details>

🛠 <b>Development History</b>
<details>
<summary><b>Jan 19, 2026: Project Membership & State Verification Testing</b> (Click to expand)</summary>
- Upgraded FakeRepository to track project-user relationships in-memory
- Added TestAddUserToProject with deep verification of repository state
- Implemented ListUsers in FakeRepository to support membership assertions
- Fixed type assertion issues with SQLC-generated pgtype.Text fields in tests
</details>

<details>
<summary><b>Jan 18, 2026: Key Architectural Achievements in Testing</b> (Click to expand)</summary>

* **Decoupled Architecture:** Refactored the Service layer to depend on a Repository interface, allowing for FakeRepository implementations that eliminate the need for a live database during test execution.
* **Dependency Inversion:** Successfully moved from concrete sqlc.Queries dependencies to abstract interfaces, preventing nil pointer panics and making the codebase "unit-testable."
* **Context Propagation:** Implemented context.Context throughout the stack to ensure request cancellation and timeouts are respected from the HTTP layer down to the database.
* **Middleware Validation:** Integrated tests for JWT Authentication and Role-Based Access Control (RBAC) to ensure protected routes are only accessible by authorized users.
</details>

<details>
<summary><b>Jan 17, 2026: Tooling & Private Messaging Update</b> (Click to expand)</summary>

* **Fix (UX):** Resolved an issue where the `ChatBox` would trigger an outer page scroll on new messages by switching from `scrollIntoView` to direct `scrollTop` container manipulation.
* **Feature:** Implemented **Direct Messaging (1-on-1)** between project members.
* **Frontend:** Created `DirectMessageBox` and `useDirectChat` hook to handle private WebSocket events and history fetching for 1-on-1 conversations.
* **Architecture:** Updated `ProjectList` and `App.tsx` to support toggling between Project-wide chat and Private Member chat without visual conflicts.
</details>

<details>
<summary><b>Jan 16, 2026: The Identity & Context Update</b> (Click to expand)</summary>

* **Backend (SQL):** Optimized message retrieval by implementing `JOIN` queries between `messages` and `users` tables to fetch sender emails automatically.
* **Backend (Live Data):** Refactored the `Create` repository method using a SQL `WITH` clause to return the `sender_email` instantly for real-time WebSocket broadcasting.
* **Feature:** Added `GET /projects/:id` endpoint and handler to fetch specific project metadata.
* **UX:** Replaced "User ID" labels with actual "Sender Emails" and updated the chat header to display the **Project Name** instead of a raw ID.
* **Architecture:** Synchronized TypeScript interfaces across the `ChatBox` and `useChat` hook to ensure type safety for the new `sender_email` field.
</details>

<details>
<summary><b>Jan 15, 2026: The Chat & Stability Update</b> (Click to expand)</summary>

* **Feature:** Integrated `ChatBox` with project-specific WebSocket rooms.
* **UX:** Added "Me" vs "User ID" logic and right-to-left message alignment.
* **Stability:** Refactored `useWebSockets` hook with `useRef` and cleanup logic to stop connection spam during re-renders.
* **Fix:** Resolved `TypeError: Cannot read properties of null (reading 'map')` by adding array guards to API responses.
</details>

<details> <summary><b>Jan 14, 2026: The WebSocket Revolution</b> (Click to expand)</summary>

Real-Time Engine
Implemented a WebSocket Hub in Go to manage concurrent client connections.

Created a custom useWebSockets React hook to handle incoming signals (PROJECT_CREATED, PROJECT_DELETED, USER_ADDED, USER_REMOVED).

UI Stability: Refactored project fetching to allow "background refreshes," preventing the UI list from closing when updates arrive.

Membership Logic
Added AddUserToProject and RemoveUserFromProject with real-time broadcasting.

Refined ListProjects to ensure users see projects they own and projects where they are members.

</details>

<details> <summary><b>Jan 13, 2026: Ownership & Permissions</b> (Click to expand)</summary>

Backend (Go)
Updated LoginResponse DTO to include userId field for frontend permission handling.

Modified UserHandler to return the userId directly in the login response payload.

Frontend (React/TS)
Implemented strict ownership checks in ProjectList using currentUserId.

Fixed bug where project management buttons were visible to non-owners by ensuring ID type consistency.

Updated LoginForm to persist userId in localStorage upon successful authentication.

</details>

<details> <summary><b>Jan 12, 2026: Frontend & Security Integration</b> (Click to expand)</summary>

Frontend Updates
React Frontend Implemented: Login page with JWT authentication integration.

Interactive UI: Smooth slide-down animations for Members and Tasks.

Polished UI: Modern Glassmorphism effect with Moraine Lake background.

Project Users & Roles (Backend)
Ownership Logic: Only project owners are permitted to remove users.

JWT Claims: Security checks enforced using role-based claims within the JWT.

</details>


## 🧪 Testing  (Will add all of the tests here in the future)

The project utilizes Go's native testing toolchain and testify/assert to validate system integrity. I'm using Fake Repositories to ensure tests are fast and run without a live database.

<details> <summary> <b> Middleware & Security Testing </b> (Click to expand) </summary>

This suite verifies that the security layers correctly identify users and enforce access rules.

### What it tests:
1. **Authentication (JWT)**: Ensures the JWTMiddleware correctly extracts and validates tokens from the Authorization header.
2. **Context Injection**: Verifies that user claims (ID, Email, Role) are correctly injected into the Echo context for use by handlers.
3. **RBAC (Role-Based Access Control)**: Validates that the RequireRole middleware allows admin access while returning 403 Forbidden for standard users.
4. **Failure Handling**: Ensures malformed tokens or missing headers result in proper 401 Unauthorized responses.

### Files:

 internal/interfaces/http/middleware/test/required_role_test.go
 internal/interfaces/http/tests/auth_middleware_test.go
</details>

<details> <summary> <b> API Handler & Integration Testing </b> (Click to expand) </summary> 
These tests verify the "Social" integration between the HTTP layer, Business Services, and the Repository.

### What it tests:
1. **User Registration**: Validates the flow from JSON request to the hashing of passwords and final storage.
2. **Login Logic**: Verifies that the system correctly checks credentials and account status (active vs. inactive).
3. **Repository Interfacing**: Uses a FakeRepository to simulate database behavior (auto-incrementing IDs, duplicate email checks) in memory.
4. **Data Integrity**: Ensures that the /me endpoint correctly retrieves the authenticated user's profile information.

### Usage:

1. Run all backend integration tests using:
   ```bash
   go test ./internal/interfaces/http/tests/... -v
   
</details>

<details> <summary> <b> WebSocket Integration Testing </b> (Click to expand) </summary>

This directory contains utility scripts to verify the real-time communication engine of the Playingfield API.

## Chat Tester (`test_chat.go`)

This script performs a full end-to-end integration test of the WebSocket flow. It bypasses the need for a browser to verify that the "plumbing" of the backend is sound.

### What it tests:
1. **Authentication**: Performs a standard HTTP Login to retrieve a JWT.
2. **Upgrade**: Connects to the `/ws` endpoint and upgrades the connection.
3. **Registration**: Verifies the Hub correctly maps the User ID to the active connection.
4. **Authorization**: Attempts to send a message to a specific project.
5. **Persistence**: The server must save the message to Postgres before broadcasting.
6. **Targeted Broadcast**: Verifies the Hub routes the notification back to the sender (and other members).

### Usage:
1. Ensure the server is running (`go run cmd/app/main.go`).
2. Update the `testEmail` and `testPass` constants in the script to match a valid user.
3. Run the script:
   ```bash
   go run scripts/test_chat.go

</details>

<details> <summary> <b> Project Management & Domain Testing </b> (Click to expand) </summary>

This suite validates the project lifecycle and ensures that resource ownership is strictly enforced at the service level.

### What it tests:
1. **Ownership Enforcement**: Verifies that the DeleteProject and RemoveUserFromProject actions correctly identify the requester and block unauthorized users with appropriate errors.
2. **Automated Provisioning**: Ensures that when a project is created, the system automatically assigns the creator as the "Owner" and sets up initial permissions.
3. **Repository Abstraction**: Utilizes a fully implemented FakeRepository to simulate complex database operations (like fetching project details and verifying ownership) without requiring a Postgres instance.
4. **Clean Architecture Mapping**: Validates that data is correctly translated from infrastructure-specific types (like pgtype.Text) into clean Domain models before reaching the HTTP layer.
5. **Real-time Event Triggers**: Checks that successful project modifications (creation, deletion, or adding members) correctly trigger broadcasts to the WebSocket Hub.

### Files:
internal/interfaces/http/tests/project_handler_test.go
internal/domain/projects/fake_repository.go

###Usage:
1. Run all project-related tests:  
   ```bash
   go test ./internal/interfaces/http/tests/ -v -run Project

</details>

<details> <summary> <b> Project Membership & State Verification </b> (Click to expand) </summary> 

These tests ensure that project collaboration logic is sound and that data persists correctly through the service layers.

### What it tests:
1. **Member Invitation**: Validates that a project owner can successfully add new users to a project with specific roles.
2. **Stateful Mocking**: Uses an upgraded `FakeRepository` that simulates a SQL "Join Table" in-memory, allowing tests to verify that data was actually stored.
3. **Authorization Context**: Verifies that the system correctly identifies the project owner using JWT claims before allowing membership changes.
4. **Side-Effect Verification**: Instead of just checking HTTP status codes, these tests perform "Deep Verification" by querying the repository state after the API call.

### Files:
- `internal/interfaces/http/tests/project_handler_test.go`
- `internal/domain/projects/fake_repository.go`

### Usage:
1. Run the specific membership test:
   ```bash
   go test ./internal/interfaces/http/tests/ -v -run TestAddUserToProject

</details>



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

