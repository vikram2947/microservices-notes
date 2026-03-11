## Module 2 – `users` Service

### Run

```bash
cd module2-users

# Optional: override default port 8081
export APP_PORT=8081

go run ./...
```

### Endpoints

- `GET /healthz`
- `POST /users`
  - Body: `{"email": "user@example.com", "name": "Alice"}`
- `GET /users/{id}`

