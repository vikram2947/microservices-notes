## Module 5 – API Gateway & Backend-for-Frontend (BFF)

### 1. API Gateway – what it is and why

- **API Gateway** sits at the edge and:
  - Routes requests to downstream services (users, orders, payments, etc.).
  - Applies cross-cutting concerns:
    - Authentication/authorization.
    - Rate limiting and basic throttling.
    - Logging, metrics, tracing.
    - Simple request/response transformations.
- It is **not**:
  - A place for heavy domain logic of every service.
  - A giant monolith that “knows everything”.

Rule: keep the gateway focused on **edge concerns**; put core business logic in the services.

---

### 2. Backend-for-Frontend (BFF)

- **BFF**: a specialized backend tailored to a specific client (e.g., web app, mobile app).
  - Same services underneath, but:
    - Aggregates data in a client-friendly shape.
    - Hides complexity of multiple service calls.
    - Can apply client-specific behavior (pagination defaults, caching, feature flags).
- Example:
  - Web BFF exposing `/me/dashboard` which:
    - Gets user profile from `users` service.
    - Gets recent orders from `orders` service.
    - Returns a single JSON tailored for the web UI.

Often, BFF is implemented **inside** or **next to** the API Gateway, but conceptually it is about **client-specific APIs**.

---

### 3. Gateway patterns

- **Routing**:
  - Path-based: `/users/**` → `users` service, `/orders/**` → `orders` service, etc.
- **Auth at the edge**:
  - Validate JWTs / tokens once at the gateway.
  - Inject authenticated user info as headers for downstream services.
- **Aggregation**:
  - `/me/overview` calls 2–3 services and merges results.
  - This is where BFF-like behavior often lives.
- **Rate limiting & throttling**:
  - Protects downstreams from overload.
  - Done per API key, user, or route.

---

### 4. What we implement in Module 5

One Go project `module5-gateway` that:

- Acts as a **simple API gateway + BFF** for the earlier module 2 services:
  - Routes:
    - `/users/**` → `module2-users` (8081).
    - `/orders/**` → `module2-orders` (8082).
    - `/payments/**` → `module2-payments` (8083).
- Adds:
  - Fake “auth” via `X-User-ID` header (required for some endpoints).
  - Rate limiting middleware at the edge.
- Provides a **BFF-style aggregated endpoint**:
  - `GET /me/overview`:
    - Reads `X-User-ID`.
    - Calls `users` and `orders` services.
    - Returns a combined JSON with user info + list of orders.

This demonstrates:

- Centralized edge concerns (routing, rate limiting, auth check).
- Aggregation/BFF pattern where the client only calls one endpoint.

