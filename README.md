# Playingfield

A real-time, collaborative project and task management application,
built with **Go (Echo framework)**, **PostgreSQL (Neon)**, and a **React (TypeScript)** frontend.

---

## 🌟 Key Features

### 💬 Real-Time Project Chat
* **Contextual Messaging:** Each project features a dedicated real-time chat room.
* **Smart UI Alignment:** Messages are intelligently aligned—your messages appear on the right ("Me") in blue, while teammates' messages appear on the left in gray.
* **Live Timestamps:** Every message is stamped with a human-readable time (e.g., 14:05) for better context.
* **History Persistence:** New members can see previous project discussions instantly upon joining.

### 📋 Collaborative Task Management
* **Kanban-Style Organization:** High-visibility board layout grouping tasks into `To Do`, `In Progress`, and `Done` columns for clear project tracking.
* **Granular Task Ownership:** Ability to create tasks with specific descriptions and assign them to any verified project member.
* **Signal-Driven Refresh:** Leverages a lightweight "Pulse" synchronization logic where task changes trigger instant UI re-validation across all collaborator screens via WebSockets.
* **Role-Based Task Control:** Strict authorization logic ensuring only Project Owners can create or delete tasks, while allowing assigned members to update task status.
* **Persistent History:** Every task is backed by a robust database schema, ensuring assignments and statuses are preserved across sessions. Kind of like a github commit. (Members can see what was changed, when it was changed).

### ⚡ Real-Time Synchronization (WebSockets)
* **Global Hub:** A custom WebSocket Hub manages concurrent client connections and room-based broadcasting.
* **Live Dashboard Updates:** * **Project/Task Membership:** Projects/Tasks appear/vanish from your dashboard instantly when you are added or removed by an owner.
 * **Global Deletion/Creation:** If an owner creates/deletes/updates a project/task, it is edited from every member's screen in real-time.
* **Automatic Member Sync:** Live updates to member lists without requiring page refreshes.

### 🔐 Authentication & Security
* **JWT-Based Auth:** Secure registration and login with token-based identity.
* **Identity Integrity:** Handlers derive `user_id` exclusively from verified JWT claims, preventing "ID Spoofing."
* **Ownership Enforcement:** Destructive actions (deleting projects/tasks, removing members) are restricted to the project owner via backend middleware.
Updating or creating actions are the same.

### 🛠️ Reliability & Observability
* **Centralized Error Translation:** Implemented a Global Error Handler that acts as a bridge between internal domain errors and HTTP responses. It ensures that internal truths (like specific DB failures) are logged for developers while users receive clean, safe, and actionable error messages.
* **Structured Telemetry (slog):** The system uses machine-readable JSON logging in production. It follows a "Leveled" approach where low-level database traces are hidden by default, and only actionable WARN or ERROR events trigger alerts.
* **Recursive Error Wrapping:** Using Go's %w verb to wrap errors as they move up the stack. This preserves the original error (like a DB connection timeout) while adding "Domain Context" (like 'failed to register user').
* **Defensive Mapping:** Data is strictly mapped between the Database (Postgres/sqlc) and the Domain (Go structs). This "Anti-Corruption Layer" ensures that database changes never break business rules.
* **Ownership Enforcement:** The system performs "Deep Validation" on every request. Beyond just checking if a user is logged in, the backend verifies ownership before allowing any Updates, Deletions, or Member Management actions.

---

## 🛠 Tech Stack
* **Backend:** Go (Echo Framework), SQLC (Type-safe SQL), Gorilla WebSocket.
* **Frontend:** React 18, TypeScript, Vite, CSS3 (Glassmorphism).
* **Database:** PostgreSQL (Hosted on Neon.tech).
* **Communication:** REST API for state + WebSockets for reactivity.

---

## Future Goals
* Implement **Task creation from the UI**. DONE
* Improve **error handling and logging** further.
* Implement **user role management** (admin vs regular users). 
* Add **unit and integration tests** for the project domain.   Mostly Done
* Add **Project group chats and 1 on 1 individual project member chat feature**. DONE

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

</details>

---

## Code Structure
<details>
<summary>(Click to expand)</summary>

* `internal/domain/user` – domain model, repository interfaces.
* `internal/domain/projects` – project domain, service, repository interface.
* `internal/infrastructure/postgres` – SQLC-based repository implementation, DB adapter.
* `cmd/server` – Echo server initialization and routing.

---
</details>

## Known Issues & How I Solved Them
<details>
<summary>(Click to expand)</summary>

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

### 8. Testing really helps =D
* **Issue**: While writing automated tests for "Unauthorized Access," I discovered a security vulnerability. The system allowed *any* logged-in user to add members to *any* project because the service lacked ownership context.
* **Solution**: 
    * Updated the Service signature to accept a `requesterID`.
    * Implemented **Object-Level Authorization**: The service now fetches the project and compares the `OwnerID` against the `requesterID` before performing any mutations.
    * Fixed a "Silent Bug" regarding parameter ordering (`userID` vs `projectID`) identified during unit testing.

### 9. Race Conditions on `client.Send`
* **Issue**: The hub was closing `client.Send` directly during `Unregister`, while the handler’s writer goroutine could still be sending, causing potential `panic: send on closed channel`.  
* **Solution**:  
    * Introduced a `done` channel owned by `Client`.  
    * Writer goroutine now listens on `done` and closes `Send` itself.  
    * Hub signals shutdown by closing `done` instead of `Send`.  
    * Encapsulation maintained by keeping `done` unexported and exposing only a getter method.

### 10. Handler Owning Connection Logic
* **Issue**: The handler managed both reading from the websocket and writing to it, forcing export of internal channels and complicating lifecycle management.  
* **Solution**:  
    * Writer goroutine remains in the handler for now, handling writes safely via the `done` channel.  
    * Handler upgrades the connection, creates the client, registers it with the hub, and starts the writer.  
    * Lifecycle management and connection cleanup are coordinated with the hub and `done` signaling.

### 11. Hub Coupled to Transport
* **Issue**: The hub contained logic that could block or panic when writing to clients directly. It also tracked websocket details unnecessarily.  
* **Solution**:  
    * Hub now only routes messages, adds/removes clients from rooms, and signals shutdown.  
    * Writes are fully handled by the handler’s writer goroutine; the hub uses non-blocking sends to prevent backpressure issues.  

### 12. Unsafe Channel Access Across Packages
* **Issue**: Accessing `done` directly from the handler required exporting the channel, which could allow accidental closure from external code.  
* **Solution**:  
    * Added `Client.DoneChan()` getter, providing read-only access for select statements in the handler.  
    * Maintained internal ownership, ensuring only the writer closes channels and prevents panics.  


</details>

---

## 🛠 <b>Development History</b>
<details><summary>(Click to expand)</summary>

<details>
<summary><b>Feb 9, 2026: Real-Time Chat UX & Identity Synchronization</b> (Click to expand) </summary>

### Phase 1: Backend Signal Chain & State Persistence
* **Stateful Read Receipts**: Engineered the backend `read_receipt` handler to update the database with `read_at` timestamps. Implemented a broadcast mechanism that notifies the original sender in real-time, allowing the frontend to transition message status from "Sent" to "Seen."
* **Identity-Enriched Typing Signals**: Refactored the WebSocket `typing` broadcast to be self-contained. The server now "stamps" the `SenderEmail` from the authenticated `claims` onto the outgoing JSON payload. This ensures that even users with no prior message history are correctly identified by email rather than an anonymous ID.
* **Contextual Signal Routing**: Hardened the Hub’s routing logic to distinguish between `ProjectID` (Room Broadcast) and `ReceiverID` (Direct Message). This ensures typing and read signals are strictly scoped to the relevant conversation, preventing data leakage across projects.

### Phase 2: Professional Chat UX & Layout Hardening
* **Non-Disruptive Scroll Logic**: Decoupled auto-scroll behavior from the typing state. The message container now only triggers a "scroll-to-bottom" upon the arrival of new `Message` objects, allowing users to browse history without the UI jumping when a peer starts typing.
* **Sticky "Typing Shelf" Implementation**: Introduced a dedicated, fixed-height UI shelf above the message input. This prevents the message list from shifting layout and ensures the "typing..." indicator remains anchored and visible regardless of the user's scroll position.
* **Stateful Hook Integration**: Refactored `useChat` and `useDirectChat` to synchronize the enriched backend signals. The hooks now maintain dedicated state for `typingUserEmail`, providing a cleaner interface for the UI components.

### Phase 3: Component Resilience & Logic Cleanup
* **Race-Condition Mitigation**: Implemented aggressive timeout clearing in the `handleInputChange` and `handleSend` functions. This ensures that the "is typing" status is immediately revoked upon message submission, eliminating "ghost" indicators.
* **Standardized Real-Time Payloads**: Unified the JSON response structure between Direct Messages and Project Chats, ensuring consistent handling of `user_typing`, `new_message`, and `message_read` event types across the entire application.
</details>

<details>
<summary><b>Feb 6, 2026: Cross-Domain Security & Real-Time Signal Chain</b> (Click to expand) </summary>

### Phase 1: Task Domain & Multi-Role Authorization
* **Advanced RBAC Implementation**: Developed the `TaskService` with a multi-role security model. Verified that status updates are restricted to either the **Project Owner** or the **Task Assignee**, while preventing unauthorized member interference.
* **Sentinel Error Standardization**: Replaced generic string-based errors with Domain Sentinel Errors (e.g., `ErrUnauthorized`, `ErrTaskNotFound`). This allows for precise error type-checking across the entire application stack.
* **Audit Trail Integration**: Wired `TaskActivity` logging into the service layer, ensuring every state change (Create/Update) is persisted for history tracking.
* **Cross-Service Membership Verification**: Implemented logic where the Task service consults the Project repository to verify user membership before granting access to task history or listings.

### Phase 2: HTTP Error Bridging & Reliability
* **Global Error Translator Optimization**: Enhanced the `CustomHTTPErrorHandler` to recognize Task-domain sentinels. Successfully bridged internal logic errors to appropriate HTTP status codes (403 Forbidden, 404 Not Found, 409 Conflict).
* **Stateful Test Seeding**: Refactored HTTP integration tests to use "Service-style" seeding. By manually populating membership tables in fakes, we ensured tests accurately reflect the real-world behavior of the Service/Repository relationship.
* **Panic-Resilient Assertions**: Hardened handler tests to verify response body content before indexing, preventing test-suite crashes during 500-error regressions.

### Phase 3: WebSocket Integration Testing
* **Full-Circuit Verification**: Conducted the first "Full-Circuit" integration test. Proved that a Service-layer action (creating a task) successfully navigates the Hub's concurrency loop to deliver a real-time signal to a mock Client's `Send` channel.
* **Hub Concurrency Hardening**: Verified the `Hub`'s `sync.RWMutex` and `select` logic to ensure non-blocking broadcasts. Confirmed that the Hub correctly manages registration and unregistration without leaking goroutines.
* **Signal Mapping**: Established a standardized real-time message format (e.g., `TASK_CREATED:ID`) to allow frontend consumers to selectively refresh UI components based on the incoming event stream.
</details>

<details>
<summary><b>Feb 5, 2026: Logging Standardization & Architecture Polish</b> (Click to expand) </summary>

### Phase 1: Observability & Technical Context
* **Structured Logging (Slog) Implementation**: Standardized all repositories (User, Project, Task, Message) with `slog.Error`. This ensures every database-level failure is logged with structured metadata (IDs, Emails, Actions) for precise debugging without cluttering the business logic.
* **Separation of Concerns (Logging)**: Established a clear logging hierarchy:
    * **Repositories**: Log technical/infrastructure failures (e.g., SQL timeouts, connection loss).
    * **Services**: Log high-value business events (e.g., successful project creation) and security warnings.
    * **Handlers**: Remain "log-silent" for standard HTTP, leveraging the global translator for cleaner terminal output.
* **Error Context Wrapping**: Refactored Repository methods to use `fmt.Errorf` with the `%w` verb. This preserves the "Technical Context" (where the DB failed) while allowing the Service layer to perform business-level erro    r translation.

### Phase 2: WebSocket & Infrastructure Hardening
* **WS Lifecycle Management**: Modernized the `WSHandler` and `Hub` to use `slog`. Implemented structured connect/disconnect logging to track real-time user activity without using legacy `fmt.Printf` or `log.Println`.
* **Zero-Allocation Signaling**: Finalized the Hub's shutdown logic using `chan struct{}` for 0-byte signaling, ensuring an efficient, thread-safe cleanup of all active client connections during server shutdown.
* **Redundancy Cleanup**: Conducted a global audit to remove "Double Logging." Optimized the flow so a single event generates exactly one relevant log entry, ensuring production logs remain readable and searchable.
</details>

<details>
<summary><b>Feb 4, 2026: Sentinel Errors & Repository Alignment</b> (Click to expand) </summary>

### Phase 1: Domain-Driven Error Architecture
* **Sentinel Error Implementation (Projects)**: Transitioned from brittle string-based error matching to **Sentinel Errors** (e.g., `ErrProjectNotFound`, `ErrUnauthorized`, `ErrDuplicateProject`). This establishes a formal "Error Contract" across the repository, service, and interface layers.
* **Smart Error Wrapping**: Implemented the `%w` wrapping pattern in the Service layer. This allows the system to provide high-context dynamic messages (like specific duplicate project names) while still allowing the global translator to identify the root error via `errors.Is`.
* **The "Bouncer" Refactor**: Upgraded the `CustomHTTPErrorHandler` from high-overhead string searching to high-performance pointer comparison. The translator now efficiently maps internal "Sentinels" to precise HTTP status codes (403 Forbidden, 404 Not Found, 409 Conflict).

### Phase 2: Full-Stack Test Fidelity
* **Mock-Real Alignment**: Synchronized the `FakeRepository` to return the exact same sentinel errors as the production PostgreSQL driver. This ensures unit tests are "high-fidelity," catching logic bugs that would previously only appear in integration environments.
* **State & Security Verification**: Hardened the Project test suite with mandatory **State Checks**. Tests now verify both the HTTP response code and the actual database state in the `fakeRepo`, ensuring that "Unauthorized" attempts result in zero data changes.
* **Constraint Simulation**: Enhanced the Fake Repository to simulate database unique-key constraints (e.g., duplicate name checks), allowing the test suite to validate complex conflict resolution logic without a live database.
</details>

<details>
<summary><b>Feb 3, 2026: Global Error Translation & Handler Decoupling</b> (Click to expand) </summary>

### Phase 1: Global Error Architecture
* **The "Translator" Pattern**: Implemented a centralized `CustomHTTPErrorHandler` to decouple domain errors from HTTP status codes. This removed high-entropy `if/else` logic from handlers, allowing the service layer to define the "Internal Truth" while the translator manages the "External Persona."
* **Security-Minded Responses**: Standardized API error responses to prevent account enumeration and data leaking. The system now maps specific internal failures (e.g., `wrong password`) to generic, safe public messages (`Invalid email or password`) while preserving technical detail in server-side logs.
* **Echo Middleware Integration**: Fully integrated the error handler into the Echo framework's lifecycle, enabling automatic recovery and consistent JSON error formats across all endpoints, including the built-in binder and logger.

### Phase 2: Test Suite Evolution
* **Error-Aware Unit Testing**: Refactored the HTTP test suite to integrate with the global error handler. Transitioned tests from asserting on raw handler returns to validating the final "Translated" response, ensuring the baseline tests accurately reflect the production user experience.
* **Enhanced Client-Side Debugging**: Improved `c.Bind` error propagation, providing frontend developers with detailed syntax offsets and unmarshalling errors rather than generic "invalid request" messages.
</details>

<details>
<summary><b>Feb 2, 2026: Persistence Layer & Production-Grade Mapping</b> (Click to expand) </summary>

### Phase 1: Infrastructure & Repository Refactoring
* **The "Senior" Repository Pattern**: Refactored Postgres repositories to act as a strict bridge between `sqlc` generated code and the Domain layer. Implemented manual mapping to ensure database-specific types (like `pgtype.Text` or `pgtype.Int8`) never leak into the business logic.
* **Contextual Persistence**: Standardized `context.Context` propagation across the entire stack (Postgres and Mock repositories). This enables system-wide query cancellation, deadline enforcement, and consistent request tracing.
* **Smart Mapping & Nil Safety**: Developed robust mapping functions to handle nullable database columns, transforming them into safe Go pointers to prevent `nil` dereference panics during runtime.

### Phase 2: High-Volume Observability
* **Log Level Strategy**: Implemented a "Value-Based" logging strategy. High-frequency events (Messages) are relegated to `DEBUG` to prevent disk-thrashing, while high-value events (Project Creation, User Registration) are promoted to `INFO`.
* **Error Wrapping (%w)**: Integrated `fmt.Errorf("db: description: %w", err)` across all repository methods. This creates a "breadcrumb trail" in logs, making it instantly clear if a failure originated in the database, the service, or the handler.
</details>

<details>
<summary><b>Feb 1, 2026: Modern Observability (log/slog)</b> (Click to expand) </summary>

### Phase 1: Structured Logging Migration (`slog`)
* **Standard Library Modernization**: Migrated the entire application logging from `fmt/log` to the Go 1.21+ `log/slog` package. This moves the project from "string-based" logs to structured, machine-readable telemetry.
* **Global Observability Setup**: Implemented a centralized `slog.Handler` in `server.go` using `slog.SetDefault`. This ensures that all downstream packages (Services, Repositories, Hub) share a unified logging format and level control without extra dependency injection.
* **Leveled Telemetry**: Introduced differentiated log levels (`DEBUG`, `INFO`, `WARN`, `ERROR`) allowing for high-verbosity development traces while maintaining clean, actionable logs in production environments.

### Phase 2: Error Traceability
* **Structured Context Pairs**: Updated core error paths to use key-value pairs (e.g., `logger.Error("msg", "error", err)`). This allows for instant filtering by error type or resource ID in log aggregators.
* **Fail-Fast vs. Resilient Startup**: Standardized the use of `os.Exit(1)` for critical infrastructure failures (Database, Config) while using `logger.Warn` for non-critical bootstrap tasks like admin seeding.
</details>

<details>
<summary><b>Jan 31, 2026: Graceful Shutdown</b> (Click to expand) 🛡️</summary>

### 🏗️ Infrastructure & Database Mapping

- **Type-Safe Data Access**: Migrated from manual row scanning to **SQLC**. This ensures compile-time safety for all Postgres queries and automatically handles complex PostgreSQL types using `pgtype`.
- **Relational Data Loading**: Optimized `Create` operations for Projects and Messages using **SQL CTEs (Common Table Expressions)** with `RETURNING` joins. This allows the API to return "rich" entities (e.g., including sender emails or owner names) in a single database round-trip.

### 🛡️ System Stability & Lifecycle

- **Graceful Shutdown**: Implemented a robust server lifecycle management system using Go `channels` and `os/signal`. 
  - The server now listens for `SIGINT` and `SIGTERM` signals.
  - Uses `context.WithTimeout` to allow active HTTP requests 10 seconds to "drain" before the process exits.
- **WebSocket Hub Management**: Added a coordinated shutdown sequence for the WebSocket Hub.
  - When the server stops, the Hub executes a `cleanup()` routine to gracefully close all active client connections and clear memory-resident "Project Rooms."
  - Prevents goroutine leaks and dangling TCP connections.

### 🛠️ Technical Stack Additions

- **Concurrency Patterns**: Utilized the "Signal Channel" and "Stop Channel" patterns to manage background worker loops.
</details>

<details>
<summary><b>Jan 30, 2026: Domain-Driven Message Testing Infrastructure</b> (Click to expand) 🧪</summary>

### Phase 1: Stateful Fake Repository Implementation
* **In-Memory Logic Simulation**: Developed `FakeRepository` for the Messaging domain using stateful Go slices. This allows tests to simulate database persistence, chronological message retrieval, and bi-directional DM history without a live PostgreSQL instance.
* **Complex Relationship Mocking**: Enhanced the Projects Fake Repository to support real-time membership lookups, enabling the test suite to verify "Shared Project" constraints for private communications.

### Phase 2: Service Layer Authorization Testing
* **Logic Gate Validation**: Implemented comprehensive unit tests for `SendProjectMessage` and `SendDirectMessage`. Verified that unauthorized users are strictly blocked from project channels and that direct messages are restricted to verified project collaborators.
* **Nil-Pointer Resilience**: Hardened the Service layer with proactive nil-checks for the WebSocket Hub, ensuring the application remains stable during testing environments or partial infrastructure failures.

### Phase 3: Integration with WebSocket Hub
* **Real-time Path Execution**: Integrated the `ws.Hub` into the test suite using background goroutines. This ensures that the message "broadcast" logic is actually executed and exercised during tests, providing higher confidence in the real-time delivery pipeline.
</details>

<details>
<summary><b>Jan 29, 2026: Key Architectural Achievements in WebSocket Refactor</b> (Click to expand) 🌟🌟</summary>

* **Safe Channel Ownership:** Introduced a `done` channel (unexported) in `Client` to signal shutdown safely. Writer goroutine owns the `Send` channel, preventing "send on closed channel" panics.  
* **Minimal Handler Refactor:** The WebSocket handler now only wires the connection, registers the client, and starts the writer goroutine. No direct channel or goroutine management is required in the hub.  
* **Getter Method for Lifecycle Signaling:** Added `Client.DoneChan()` to allow other packages (like the handler) to listen for shutdown signals without exposing internal channels, maintaining encapsulation.  
* **Buffered Send Channel & Non-blocking Writes:** All writes to `Send` use select with default to avoid blocking the hub when a client is slow. This prevents hub-level blocking and ensures smooth broadcast even under load.  
* **Robust Project Room Management:** Clients are added/removed from project rooms safely with mutex protection, and empty rooms are cleaned up automatically, preventing memory leaks.
</details>

<details>
<summary><b>Jan 26, 2026: Real-Time Task Infrastructure & Collaborative UI</b> (Click to expand) </summary>

### Phase 1: Task Board Frontend Architecture
* **Componentized Kanban System**: Developed a full-scale `TaskBoard` and `TaskColumn` infrastructure. Implemented logical grouping of tasks by status (`To Do`, `In Progress`, `Done`) with dynamic filtering.
* **Member-Aware Assignment UI**: Integrated project member data into the task creation flow, allowing for real-id assignment and visual tracking of task owners within the board.

### Phase 2: Reactive State Synchronization (The "Tick" System)
* **Signal-Based Update Architecture**: Implemented a lightweight "Pulse" mechanism (`taskRefreshTick`) for real-time updates. Rather than pushing heavy data payloads over WebSockets, the backend emits a versioning signal that triggers optimized client-side re-validation.
* **WebSocket Event Consolidation**: Standardized broadcast logic for `TASK_CREATED`, `TASK_UPDATED`, and `TASK_DELETED`. All mutation events now feed into a unified "Signal" bus, ensuring all collaborators maintain a synchronized view without manual polling.

### Phase 3: Project Ownership & RBAC Hardening
* **Verified Mutation Gates**: Hardened the Project controller to enforce strict **Project Owner** authorization. "Edit" and "Delete" operations now perform server-side verification against JWT claims before executing database writes.
* **Idempotent Update Service**: Refined the `PUT /projects/:id` endpoint to handle partial updates, ensuring metadata changes are persisted without disrupting established project-member relationships.

### Phase 4: Optimized Domain Hydration & UI Logic
* **On-Demand Membership Mapping**: Implemented a lazy-loading strategy for project metadata. Member lists and task boards are now hydrated only when the domain section is activated, significantly reducing initial payload size.
* **Global Interaction Layer**: Developed a universal UI feedback system using CSS filters and transforms, providing tactile hover states and "lift" effects for all interactive elements to improve the demo's professional feel.

### Phase 5: Full-Stack Interface Alignment
* **Type-Safe Contract Synchronization**: Aligned backend DTOs with Frontend TypeScript interfaces, ensuring strict compile-time safety across the network boundary.
* **Unified Response Formatting**: Standardized error and success handling across the Project and Task services for predictable UI notification behavior.

</details>

<details>
<summary><b>Jan 25, 2026: Task Management Backend completion!</b> (Click to expand) 🏗️</summary>
- **Task Management System**: Full CRUD for tasks with project-level authorization.
- **Activity Logging**: Every task creation and update is now automatically logged in a `task_activities` audit trail.
- **RESTful Task Routing**: Implemented nested resource routing for projects and direct task access.

## 🛠 API Progress (Tasks)

### Projects & Tasks
| Method | Endpoint | Description | Auth |
| :--- | :--- | :--- | :--- |
| GET | `/projects/:id/tasks` | List all tasks in a project | JWT (Member) |
| POST | `/tasks` | Create a new task | JWT (Owner) |
| PUT | `/tasks/:id` | Update task details/status | JWT (Owner/Assignee) |
| DELETE | `/tasks/:id` | Delete a task | JWT (Owner) |
| GET | `/tasks/:id/history` | View audit log for a task | JWT (Member) |

### Real-time Updates
- `TASK_CREATED:{project_id}`
- `TASK_UPDATED:{project_id}:{task_id}`
- `TASK_DELETED:{project_id}:{task_id}`

</details>

<details>
<summary><b>Jan 24, 2026: Task Management & Audit Infrastructure</b> (Click to expand) 🏗️</summary>

### Phase 1: Database Audit & History Architecture
* **Task Schema Implementation**: Designed the `tasks` table with a focus on simplicity, supporting single-assignee ownership and project-level isolation.
* **Full Audit Logging (The Activity Ledger)**: Created the `task_activities` table. This acts as an immutable record of "who did what and when," providing a complete history of task creation, status changes, and assignments.

### Phase 2: Domain-Level Security & Authorization
* **Owner-Locked Creation**: Implemented logic in the `Task Service` that requires a "Project Owner" role to create tasks. The service now cross-references the `Project Repository` to verify authority before any data is written.
* **Multi-Role Update Logic**: Developed a robust authorization gate for task updates. Modifications are now strictly limited to either the **Project Owner** or the **Assigned Member**, preventing unauthorized changes by other project members.

### Phase 3: Real-Time Event Synchronization
* **Hub-Driven Notifications**: Integrated the `ws.Hub` directly into the Task service. Successful creation and updates now trigger immediate broadcasts (`TASK_CREATED`, `TASK_UPDATED`), ensuring all collaborators see project changes without manual refreshes.
* **Standardized Broadcast Messaging**: Aligned Task notification strings with existing Project and Message patterns (`TYPE:ID`) to maintain a predictable API for the frontend.

### Phase 4: Data Integrity & Fault Tolerance
* **Strict History Constraints**: Opted for a "Strict Integrity" model where task operations return an error if the history log fails to write. This ensures the "Full History" requirement is never compromised by partial database successes.
* **Interface-Driven Task Repository**: Defined a clean `Repository` interface for Tasks, fully decoupling the business rules from the underlying SQLC implementation and keeping the domain pure.
</details>

<details>
<summary><b>Jan 23, 2026: The "Grand Refactor" - Domain Purity & System-Wide Cleanup</b> (Click to expand) 🌟</summary>

### Phase 1: Standardizing Domain Architecture
* **Global Interface Decoupling**: Refactored the **User**, **Project**, and **Message** services to depend exclusively on interfaces. No service layer now "leaks" SQLC or raw database logic, making the entire system 100% unit-testable.
* **Service-to-Service Communication**: Implemented a "Waiter-to-Waiter" pattern where the Message service asks the Project Repository for authorization checks (like membership or shared projects) rather than reaching into the database directly.

### Phase 2: System Plumbing & Dependency Injection
* **Clean Wiring in `app.Run()`**: Streamlined the initialization of the server. Standardized how repositories are injected into services, ensuring a single source of truth for database connections.
* **Postgres Adapter Optimization**: Cleaned up the `postgres` package to act as a clean wrapper for SQLC, hiding the complexity of `pgtype` and raw SQL parameters from the business logic.

### Phase 3: Project "Cleanup"
* **Dead Code Elimination**: Identified and deleted redundant files and "ghost" structs that were left over from earlier iterations, significantly reducing the project's cognitive load.
* **Consistent Error Handling**: Standardized error wrapping (e.g., `fmt.Errorf("...: %w", err)`) across all services to ensure that when something breaks, the logs tell a clear, traceable story.
* **Fake Repo Synchronization**: Updated all `FakeRepository` implementations (User, Project, and Message) to match the new interface signatures, fixing global compiler errors and preparing the ground for the next phase of testing.

### Phase 4: Direct Messaging Security
* **Shared Project Constraint**: Implemented the `UsersShareProject` logic. This enforces a privacy rule: users can only send Direct Messages if they have a "social connection" through at least one shared project, preventing platform-wide spam.
</details>

<details>
<summary><b>Jan 22, 2026: Real-time Project Updates & SQLC Migration</b> (Click to expand)</summary>

### Phase 1: Database & Repository Evolution
* **SQLC Integration**: Migrated the Project Update logic from raw SQL strings to type-safe code generation using `sqlc`. Defined the `UpdateProject` query to allow modifications of project names and descriptions.
* **FakeRepo Sync**: Updated the `FakeRepository` to mirror the new generated interfaces, ensuring that automated tests remain fast and database-agnostic while still validating business rules.

### Phase 2: Secure Update Logic & Real-time Sync
* **The "Owner-Only" Guard**: Implemented the `UpdateProject` service method with strict authorization. The system now validates that only the project creator can modify project details, returning a `403 Forbidden` for unauthorized attempts.
* **WebSocket Integration**: Connected the Update event to the global `Hub`. When a project is renamed, a broadcast signal (`PROJECT_UPDATED:ID`) is sent to all connected clients, ensuring data consistency across the platform.

### Phase 3: Inline-Edit Frontend
* **UX Transformation**: Developed an "Inline-Edit" mode in the React `ProjectList`. This allows owners to toggle between viewing project info and a live edit form without leaving the page.
* **Zero-Refresh UI**: Integrated the new WebSocket signal into the frontend `useWebSockets` hook. The application now automatically re-fetches project data the moment a broadcast is received, providing an "instant" feel for all users.
</details>

<details>
<summary><b>Jan 21, 2026: Project Authorization & Membership</b> (Click to expand)</summary>

### Phase 1: Testing & Security Discovery
* **The Problem**: While writing automated tests for "Unauthorized Access," I discovered a security vulnerability. The system allowed *any* logged-in user to add members to *any* project because the service lacked ownership context.
* **The Refactor**: 
    * Updated the Service signature to accept a `requesterID`.
    * Implemented **Object-Level Authorization**: The service now fetches the project and compares the `OwnerID` against the `requesterID` before performing any mutations.
    * Fixed a "Silent Bug" regarding parameter ordering (`userID` vs `projectID`) identified during unit testing.

### Phase 2: Robust Membership Logic
* **Goal**: Implement secure removal and state verification.
* **Outcome**: Added `RemoveUserFromProject` with the same ownership guards. Updated the `FakeRepository` to handle slice manipulation, allowing for "Deep Verification" (checking if the user was actually removed from memory after the API call).
</details>

<details>
<summary><b>Jan 20, 2026: Project Membership & Security Enforcement</b> (Click to expand)</summary>
* Added TestRemoveUserFromProject to verify successful member deletion
* Added TestRemoveUserFromProject_Unauthorized to enforce ownership rules
* Verified data persistence and side-effects using stateful repository checks
</details>

<details>
<summary><b>Jan 19, 2026: Project Membership & State Verification Testing</b> (Click to expand)</summary>
* Upgraded FakeRepository to track project-user relationships in-memory
* Added TestAddUserToProject with deep verification of repository state
* Implemented ListUsers in FakeRepository to support membership assertions
* Fixed type assertion issues with SQLC-generated pgtype.Text fields in tests
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
</details>

---

## 🧪 Testing Strategy

The project employs a tiered testing strategy using Go's native toolchain and the `testify/assert` library. By utilizing **Stateful Fake Repositories**, the suite ensures high execution speed and 100% data consistency without the overhead of a live database.

---

### 🏎️ Concurrency & Race Safety
The entire suite is verified using the **Go Race Detector** to ensure thread-safety in high-concurrency environments (like WebSockets).

* **CGO Enabled:** Configured with MinGW-w64 to support runtime memory analysis.
* **Thread-Safe Fakes:** Repositories utilize `sync.RWMutex` to prevent data races during parallel test execution.
* **Verification:** Run the full race-detection suite with:
  ```bash
  $env:CGO_ENABLED = "1"; go test -race ./...

---


### 🧩 Domain & Unit Testing (Logic Layer)
These tests focus on core business rules in isolation. They sit within the domain packages to verify that the "brain" of the application works correctly.

<details>
<summary><b>💬 Messaging & Authorization Logic</b></summary>

Validated within `internal/domain/messages/`.

* **Logic Gates:** Verifies that project messages are only accepted from verified members.
* **Social Constraints:** Ensures Direct Messages (DMs) are restricted to users who share at least one project.
* **Stateful Persistence:** Uses a `FakeRepository` to simulate message storage and chronological retrieval.
* **Nil-Resilience:** Validates that service methods handle infrastructure (WebSocket Hub) availability gracefully.
* **Execution:** `go test -v ./internal/domain/messages`
</details>

<details>
<summary><b>🏗️ Project Lifecycle & Ownership</b></summary>

Validated within `internal/domain/projects/`.

* **Ownership Guardrails:** Ensures only the project creator can delete resources or manage members.
* **Auto-Provisioning:** Validates that the system correctly assigns roles upon project creation.
* **Service-Based Seeding:** Uses the Service layer in tests to ensure realistic system states (e.g., owners are automatically members).
* **Member Management:** Tests the "Join Table" logic in-memory to ensure member lists are accurate.
* **Execution:** `go test -v ./internal/domain/projects`
</details>

<details>
<summary><b>📋 Task Management & Cross-Domain Security</b></summary>

Validated within `internal/domain/tasks/`.

* **Multi-Role Authorization:** Verifies the "VIP lanes" for updates—ensuring only Project Owners OR the specific Task Assignee can modify status.
* **Audit Trails:** Validates that every task action (Create/Update) automatically triggers a `TaskActivity` log entry.
* **Cross-Domain Integrity:** Tests the service's ability to verify project-level permissions before performing task-level actions.
* **Context Respect:** Ensures all repository methods correctly honor context deadlines and cancellations.
* **Execution:** `go test -v ./internal/domain/tasks`
</details>

---

### 🌐 HTTP & Integration Testing (API Layer)

These tests verify the "Social" integration between the HTTP layer, Middleware, and the Service layer.

<details>
<summary><b>🔐 Middleware & Security</b></summary>

Validated within `internal/interfaces/http/middleware/`.

* **JWT Integrity:** Ensures `JWTMiddleware` correctly extracts and validates tokens from headers.
* **Context Injection:** Verifies that user identity (ID, Role) is correctly passed to the internal logic.
* **RBAC Enforcement:** Validates that `RequireRole` blocks unauthorized access to sensitive routes.
* **Execution:** `go test -v ./internal/interfaces/http/middleware/...`
</details>

<details>
<summary><b>🚀 API Handler Endpoints & Error Translation</b></summary>

Validated within `internal/interfaces/http/tests/`.

* **Centralized Error Mapping:** Verifies that Domain Sentinel errors (like `ErrUnauthorized`) are correctly translated into standard HTTP codes (403, 404, 409).
* **End-to-End Persistence:** Tests the full flow from an HTTP request through the Service layer into the Fake Repository.
* **Security Resilience:** Tests that unauthorized API attempts return clean, safe error messages without leaking system internals.
* **JSON Binding:** Validates strict structural binding for complex entities like Tasks and Projects.
* **Execution:** `go test -v ./internal/interfaces/http/tests/...`
</details>

---

### ⚡ Real-Time Integration (WebSocket)
Testing for the communication engine, verifying that messages are not only saved but correctly routed.

<details>
<summary><b>📡 Full-Circuit Broadcast Integration</b></summary>

Validated through service-to-hub integration tests.

* **Live Signal Chain:** Proves the "Whole Circuit"—from a Service action (creating a task) through the Hub's concurrency loop to a Client's receiver channel.
* **Concurrency Safety:** Validates the Hub’s `sync.RWMutex` logic under simulated client registrations and broadcasts.
* **Room-Based Isolation:** Ensures the Hub correctly manages `ProjectRooms` for targeted data delivery.
* **Execution:** `go test -v ./internal/domain/tasks -run TestTaskService_WebSocketIntegration` , `go test -v ./internal/domain/tasks -run TestTaskService_WebSocketBroadcast`
</details>

<details>
<summary><b>📡 WebSocket Hub & Chat Tester</b></summary>

* **Connection Mapping:** Verifies the `Hub` correctly maps User IDs to active WebSocket connections.
* **Targeted Broadcasting:** Validates that messages sent to a project are routed strictly to that project's members.
* **Direct Messaging (P2P):** Ensures private messages are routed strictly to the sender and receiver.
* **E2E Script:** A dedicated utility (`scripts/test_chat.go`) to verify the full "Plumbing" from Auth -> Upgrade -> Broadcast.
* **Execution:** `go run scripts/test_chat.go`
</details>

---

## WebSocket Flow
<details>

1. **Handler** upgrades HTTP connection, creates `Client`, and registers it with the `Hub`.
2. **Writer Goroutine** (currently in the handler) listens on `Client.Send` and `Client.DoneChan()`.
3. **Read Loop** reads websocket messages and forwards them to the hub or services.
4. **Hub** routes messages to clients or project rooms, using non-blocking sends to avoid blocking slow clients.
5. **Shutdown**: Hub signals client via `done` channel; writer goroutine closes `Send` and websocket safely.

</details>

---

## Architecture & Flow Diagram
<details>
<summary>(Click to expand)</summary>


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






