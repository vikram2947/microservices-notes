## Module 2 – `orders` Service

### Run

```bash
cd module2-orders

# Optional: override default port 8082
export APP_PORT=8082

go run ./...
```

### Endpoints

- `GET /healthz`
- `POST /orders`
  - Body: `{"user_id": "usr_xxx", "amount": 100.5, "currency": "USD"}`
- `GET /orders/{id}`

