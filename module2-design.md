## Module 2 – Service Design & Boundaries (DDD-Lite)

### 1. How to slice services (what is a good boundary?)

- **Goal**: each service owns a **business capability** and its **data**, and can be deployed independently.
- Typical good boundaries:
  - `users` (identity, profiles, preferences)
  - `orders` (order lifecycle)
  - `payments` (payment intent, capture, refunds)
- Avoid:
  - Splitting by technical layers only (`order-api`, `order-db`, `order-worker`).
  - Sharing a single database schema across all services.

**Rule of thumb**: If two teams can work on capabilities mostly independently, they are good candidates for separate services.

---

### 2. Bounded Context (DDD-lite)

- **Bounded Context**: a boundary within which a particular model is valid and consistent.
  - Example: “User” in `users` (account, email, password) vs “User” in `billing` (payer ID, billing address).
- In microservices:
  - Each service is usually one Bounded Context.
  - Services may use different names or shapes for “the same” real-world concept (and that’s fine).

Interview phrasing: “We model each microservice as a Bounded Context that owns its own data and language.”

---

### 3. Aggregates and invariants (just enough)

- **Aggregate**: a cluster of domain objects treated as a single unit for consistency.
  - Has a root entity (Aggregate Root).
  - Invariants are enforced inside the aggregate boundary.
- Example (orders):
  - Aggregate: `Order` with items.
  - Invariant: total price must be the sum of items; status transitions allowed only in valid sequences.

In microservices, we typically ensure strong consistency *inside* a service/aggregate, and use async/event-based mechanisms *across* services (later modules).

---

### 4. Domain Events (as a modeling tool)

- **Domain Event**: something meaningful that happened in the domain, named in past tense.
  - `UserRegistered`, `OrderPlaced`, `PaymentAuthorized`.
- They:
  - Decouple producers from consumers.
  - Capture important points in the business process.

In this module we focus on **API contracts**; we will make heavy use of domain events in later modules (Sagas, Outbox, etc.).

---

### 5. API and contract design (HTTP JSON, simple)

- **Resource-oriented APIs**:
  - `POST /users` – create user
  - `GET /users/{id}` – get user
  - `POST /orders` – create order
  - `GET /orders/{id}` – get order
- Design principles:
  - Stable identifiers (`user_id`, `order_id`).
  - Clear separation of concerns: `orders` service references `user_id`, but does not own full user profile.
  - Responses include enough data for the caller, but not full cross-service joins.

We will implement:

- `users` service: manage users in its own data store (in-memory for now).
- `orders` service: manage orders, referencing `user_id`, in its own store.
- `payments` service: manage payment intents, referencing `order_id`, in its own store.

---

### 6. Data ownership and no shared DB

- **Database-per-service** (logical): each service has its own schema and persists its own data.
- Services communicate via APIs or events, **not by reaching into each other’s tables**.
- For now we’ll use in-memory maps per service to simulate separate data stores.

Interview angle: be ready to say “Each service owns its own data; other services access that data only via its API, not by joining directly on the database.”

---

### 7. What we implement in Module 2

Three small Go services:

- `module2-users`:
  - `POST /users` – create user
  - `GET /users/{id}` – fetch user
- `module2-orders`:
  - `POST /orders` – create order (takes `user_id`)
  - `GET /orders/{id}` – fetch order
- `module2-payments`:
  - `POST /payments` – create payment intent (takes `order_id`)
  - `GET /payments/{id}` – fetch payment

All use:

- Separate in-memory stores to simulate distinct databases.
- Own `APP_PORT` env var and logging.

Later modules will wire these together with real interaction patterns (sync/async, Sagas, etc.).

