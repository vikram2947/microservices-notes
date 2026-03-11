## Module 4 – Data, Transactions & Consistency Patterns

### 1. Why distributed transactions (2PC) are avoided

- Traditional **2-phase commit (2PC)** tries to make a multi-resource operation look like a single atomic transaction.
- In microservices this is usually a bad idea:
  - Tight coupling between services and data stores.
  - Coordinator becomes a single point of failure.
  - Hard to scale and reason about in failure scenarios.
- Instead, we embrace:
  - **Local transactions** inside each service.
  - **Asynchronous coordination** across services (Sagas, Outbox, events).

Interview phrasing: “We avoid 2PC across services; we use local transactions + Sagas and Outbox to achieve reliability with eventual consistency.”

---

### 2. Consistency models (practical view)

- **Strong consistency** (within a service / DB transaction):
  - After a successful write, all subsequent reads see the new value.
- **Eventual consistency** (across services):
  - After a write in one service, other services may see stale data for a brief time.
  - System will converge if messages/events are delivered and processed.

Key microservices rule:

- Aim for **strong consistency within a single service / aggregate**.
- Accept **eventual consistency between services**, but design for:
  - Idempotent operations.
  - Retries.
  - Clear states like `PENDING`, `CONFIRMED`, `FAILED`.

---

### 3. Saga pattern (orchestration vs choreography)

- **Saga**: sequence of local transactions in different services, coordinated to implement a business process.
  - If a step fails, previous steps are compensated by “undo” actions.
- **Orchestration-style**:
  - A central **orchestrator** sends commands to services and listens for replies.
  - Easier to reason about the flow in one place.
- **Choreography-style**:
  - No central brain; services react to events from others.
  - Can become complex with many services (harder to see the full flow).

Example process:

1. Create order.
2. Reserve payment or charge.
3. If payment fails → cancel order.

We’ll implement a **simple orchestrator saga** in Go for this example.

---

### 4. Outbox pattern & transactional messaging

Problem:

- Service writes to its DB and publishes a message.
- If you do:
  - `INSERT ...` then `PUBLISH event` in separate operations, one can succeed and the other fail → inconsistency.

**Outbox pattern**:

- Within a single local DB transaction:
  - Write domain data (e.g., `orders` table).
  - Write an **outbox record** (e.g., `order_created` event) into an `outbox` table.
- A background **outbox worker**:
  - Reads outbox records.
  - Publishes messages/events.
  - Marks records as sent.

This gives **atomic “write + enqueue message”** within one service without 2PC across services.

In this module we’ll simulate:

- Orders service with an in-memory store and an **outbox** slice.
- Background goroutine publishing events to an in-memory “bus”.

---

### 5. CQRS & read models

- **CQRS (Command Query Responsibility Segregation)**:
  - Separate models/paths for **writes (commands)** and **reads (queries)**.
- Often combined with events:
  - Write-side handles commands and emits events.
  - Read-side listens to events and builds **read models / projections** optimized for queries.

Benefits:

- Simpler write model with strong invariants.
- Read models tuned for specific UI/query needs.

Tradeoffs:

- Eventual consistency for reads.
- More moving parts (events, projections).

In this module:

- We will maintain:
  - An `orders` write model.
  - A separate `order_summary` read model built from events.

---

### 6. What we implement in Module 4

Single Go project `module4-saga` that simulates:

- **Orders service** (write model + outbox):
  - Create order with status `PENDING`.
  - Writes an outbox event `OrderCreated`.
- **Payment service** (simulated):
  - Handles `OrderCreated` events.
  - Marks payment as `CONFIRMED` or `FAILED`.
  - Emits `PaymentCompleted` events.
- **Saga orchestrator**:
  - Listens to events, updates order status to `CONFIRMED` or `CANCELLED`.
- **CQRS read model**:
  - Builds a simple `order_summary` map for fast reads.

HTTP endpoints:

- `POST /orders` – create order (starts saga).
- `GET /orders/{id}` – show current order state (from write model).
- `GET /summaries/{id}` – show read-model projection.

The flow is fully in-memory to keep it easy to run and inspect, but models real patterns you’d apply with databases and message brokers.

