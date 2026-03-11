## Module 6 – Service Discovery & Config Demo (In-Memory)

This module demonstrates:

- A simple in-memory **service registry**.
- Config loaded from `config.json` with **env overrides**.
- A small **gateway** that routes to `users` via the registry instead of hardcoded URLs.

### 1. Config file + env overrides

`config.json`:

```json
{
  "discovery_port": 8300,
  "gateway_port": 8301
}
```

You can override ports via env vars:

```bash
export DISCOVERY_PORT=8400
export GATEWAY_PORT=8401
```

### 2. Run the demo

First, start the existing `users` service from Module 2:

```bash
cd module2-users
go run ./...
```

Then start this module:

```bash
cd module6-discovery
go run ./...
```

You now have:

- Discovery service on `http://localhost:8300` (or overridden port).
- Simple gateway on `http://localhost:8301` (or overridden port).

### 3. Check discovery

```bash
curl -s http://localhost:8300/discover/users
```

You should see a JSON describing the registered `users` instance.

### 4. Use gateway with discovery

Create a user via discovery-aware gateway:

```bash
curl -s -X POST http://localhost:8301/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com","name":"Alice"}'
```

The gateway:

- Looks up a **healthy** `users` instance in the registry.
- Forwards the request to that URL.

If the gateway encounters an error calling `users`, it marks that instance as `UNHEALTHY` in the registry.

