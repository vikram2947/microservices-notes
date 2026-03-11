## Module 4 – Saga, Outbox & CQRS Demo (In-Memory)

This demo simulates:

- Orders service with local transactions + Outbox.
- Payment service consuming events.
- Saga orchestrator updating order status.
- CQRS read model (`summaries`) built from events.

All components run in a single Go process using in-memory stores and an in-memory event bus.

### Run

```bash
cd module4-saga

# Optional: override default port 8100
export APP_PORT=8100

go run ./...
```

### Create an order (starts the saga)

```bash
curl -s -X POST http://localhost:8100/orders \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"usr_123","amount":100.5,"currency":"USD"}'
```

Response (status `202 Accepted`) will look like:

```json
{
  "id": "ord_xxx",
  "user_id": "usr_123",
  "amount": 100.5,
  "currency": "USD",
  "status": "PENDING"
}
```

This writes:

- Order with status `PENDING` into the write model.
- `OrderCreated` event into an **Outbox**.

The outbox publisher then publishes the event to the in-memory event bus.

### Observe order state (write model)

```bash
curl -s http://localhost:8100/orders/ord_xxx
```

After a short time, you should see status change to either:

- `CONFIRMED` (payment succeeded), or
- `CANCELLED` (payment failed).

### Observe order summary (CQRS read model)

```bash
curl -s http://localhost:8100/summaries/ord_xxx
```

You will see:

- Basic order info.
- `paid: true/false`.
- Status reflecting the saga outcome.

### What to notice

- Order creation uses a **local transaction + Outbox** (simulated) to ensure that:
  - Order write and event enqueue happen together.
- Payment service and saga orchestrator react to events asynchronously:
  - This models **eventual consistency** across services.
- The `summaries` endpoint uses a **read model**:
  - Separate from the write model; built from events (CQRS style).

