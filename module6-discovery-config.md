## Module 6 – Service Discovery, Configuration & Networking

### 1. Service discovery – why it exists

- In small setups you can hardcode URLs: `http://users-service:8081`.
- In real systems:
  - Service instances come and go (scaling, failures, deployments).
  - IPs and ports are dynamic.
- **Service discovery** solves:
  - How a client finds *where* a service currently lives.

Common approaches:

- **DNS-based** (Kubernetes Services, load balancers).
- **Registry-based** (Consul, Eureka, etcd): services register themselves; clients query registry.

We’ll simulate a **simple in-process registry** and client-side discovery logic.

---

### 2. Configuration & secrets – 12-factor style

- **Configuration**:
  - Things that change per environment: ports, DB URLs, feature flags, external endpoints.
  - Should be externalized (not hard-coded in code).
- **Secrets**:
  - Credentials, tokens, keys.
  - Must not be committed to source control.

Typical practices:

- Environment variables as the simplest form (what we’ve been using).
- Config files (YAML/JSON/TOML) per environment.
- Centralized config service or secret manager (conceptually).

We’ll:

- Load configuration from both **env vars** and a **JSON config file**, with env vars overriding.

---

### 3. Networking basics that matter to services

- **Connection pooling**:
  - Reuse TCP connections instead of opening a new one per request.
- **Keep-alive**:
  - HTTP/1.1 persistent connections; HTTP/2 multiplexing.
- **Timeouts**:
  - Connection timeout, read/write timeouts, overall request timeout.
- **TLS vs mTLS**:
  - TLS: server presents certificate; client verifies server.
  - mTLS: both sides present certificates; mutual validation (used in zero-trust-ish setups).

We’ll continue using:

- HTTP client with timeouts and pooling (like earlier modules).

---

### 4. What we implement in Module 6

Single Go project `module6-discovery` that demonstrates:

- **Service registry**:
  - In-memory registry where services register at startup:
    - `users`, `orders`, `payments`.
  - Simple lookup by service name.
- **Config loading**:
  - Load base config from `config.json`.
  - Override with environment variables (`DISCOVERY_PORT`, etc.).
- **Gateway using discovery**:
  - Instead of hardcoded URLs, gateway resolves `users`, `orders`, `payments` via registry.
- **Health-aware discovery** (simple):
  - Registry keeps track of health status.
  - Gateway prefers only `healthy` instances when routing.

This mirrors how DNS-based or registry-based discovery works conceptually, without requiring Kubernetes or external tools.

