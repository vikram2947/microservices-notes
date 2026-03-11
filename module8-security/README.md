## Module 8 – Security Demo (Tokens, Edge Auth, RBAC)

This module runs three components in a single binary:

- **Auth service** on `AUTH_PORT` (default `8600`).
- **Gateway** on `GATEWAY_PORT` (default `8601`).
- **Backend service** on `BACKEND_PORT` (default `8602`).

### 1. Run

```bash
cd module8-security

export AUTH_SECRET="dev-secret-change-me"   # optional; default is insecure
export AUTH_PORT=8600
export GATEWAY_PORT=8601
export BACKEND_PORT=8602

go run ./...
```

### 2. Get a token from the auth service

Issue a token for a normal user:

```bash
curl -s -X POST http://localhost:8600/login \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"user-123","roles":["user"]}'
```

Issue a token for an admin:

```bash
curl -s -X POST http://localhost:8600/login \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"admin-1","roles":["admin"]}'
```

Each response returns:

```json
{"token":"..."}
```

### 3. Call backend via gateway (edge auth + RBAC)

Use the `token` from the auth service:

#### Profile (requires any authenticated user)

```bash
TOKEN=PASTE_TOKEN_HERE

curl -s http://localhost:8601/profile \
  -H "Authorization: Bearer $TOKEN"
```

Gateway:

- Validates the token (signature + expiry).
- Extracts `user_id` and roles.
- Forwards to backend with `X-User-ID` and `X-User-Roles`.

#### Admin (requires `admin` role)

```bash
ADMIN_TOKEN=PASTE_ADMIN_TOKEN_HERE

curl -s http://localhost:8601/admin \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

If you call `/admin` with a non-admin token, you should get `403 Forbidden`.

### 4. What this demonstrates

- **Token issuing** with signed, expiring tokens.
- **Edge authentication** in the gateway:
  - Validates tokens.
  - Enforces required roles for endpoints.
  - Propagates identity via headers to backend.
- **Backend authorization (RBAC)**:
  - Checks headers set by the gateway and enforces access control.

