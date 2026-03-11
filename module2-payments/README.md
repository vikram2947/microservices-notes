## Module 2 – `payments` Service

### Run

```bash
cd module2-payments

# Optional: override default port 8083
export APP_PORT=8083

go run ./...
```

### Endpoints

- `GET /healthz`
- `POST /payments`
  - Body: `{"order_id": "ord_xxx", "amount": 100.5, "currency": "USD"}`
- `GET /payments/{id}`

