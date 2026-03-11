## Module 5 – Simple API Gateway + BFF Demo

This gateway sits in front of the Module 2 services:

- `module2-users` (default `http://localhost:8081`)
- `module2-orders` (default `http://localhost:8082`)
- `module2-payments` (default `http://localhost:8083`)

It demonstrates:

- Routing by path.
- Edge auth check via `X-User-ID` header.
- Rate limiting.
- A BFF-style aggregated endpoint `/me/overview`.

### 1. Start backend services (from repo root)

In three terminals:

```bash
cd module2-users
go run ./...
```

```bash
cd module2-orders
go run ./...
```

```bash
cd module2-payments
go run ./...
```

### 2. Start the gateway

```bash
cd module5-gateway

# Optional overrides:
export APP_PORT=8200
export USERS_BASE_URL=http://localhost:8081
export ORDERS_BASE_URL=http://localhost:8082
export PAYMENTS_BASE_URL=http://localhost:8083

go run ./...
```

Gateway listens on `http://localhost:8200`.

### 3. Basic routing examples

Create a user via the gateway:

```bash
curl -s -X POST http://localhost:8200/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","name":"Alice"}'
```

You should see the same response as hitting `module2-users` directly.

### 4. Auth-protected routes (`X-User-ID` required)

`/orders` and `/payments` are protected by a simple auth middleware that requires `X-User-ID`.

Example (replace `USR_ID_HERE` with ID from user creation):

```bash
curl -s -X POST http://localhost:8200/orders \
  -H 'Content-Type: application/json' \
  -H 'X-User-ID: USR_ID_HERE' \
  -d '{"user_id":"USR_ID_HERE","amount":100.5,"currency":"USD"}'
```

### 5. BFF-style endpoint: `/me/overview`

This endpoint aggregates data for the current user.

```bash
curl -s http://localhost:8200/me/overview \
  -H "X-User-ID: USR_ID_HERE"
```

It will:

- Fetch the user from `module2-users`.
- Return a combined JSON with user info and a placeholder for “orders info” (explaining that in a real system we would also fetch orders here).

### 6. Rate limiting

- The gateway uses a simple token bucket limiter.
- If you send many requests quickly, you may receive `429 Too Many Requests`.

