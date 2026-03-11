## Module 8 – Security for Microservices (Auth, Tokens, mTLS Concepts)

### 1. Authentication vs Authorization

- **Authentication (AuthN)**: who are you?
  - Proving identity (passwords, tokens, OAuth, etc.).
- **Authorization (AuthZ)**: what are you allowed to do?
  - Permissions, roles, policies (RBAC/ABAC).

In microservices:

- Usually authenticate at the **edge** (gateway/BFF).
- Propagate identity as tokens/claims to downstream services (token relay).

---

### 2. Tokens (JWT, opaque) – practical view

- **JWT (JSON Web Token)**:
  - Signed (not encrypted) JSON object.
  - Contains claims: `sub` (subject/user), `exp`, `roles`, etc.
  - Self-contained: services can validate signature and expiry locally.
- **Opaque tokens**:
  - Random strings; must be looked up in a token store or introspection endpoint.

Tradeoffs:

- JWT:
  - Pros: no central lookup, good for distributed systems.
  - Cons: revocation is harder; must be careful with size and sensitive data.
- Opaque:
  - Pros: easy revocation, minimal data in token.
  - Cons: requires central check.

We’ll implement:

- Simple **HMAC-signed JWT-like token** (using `HMAC-SHA256`).
- Gateway that:
  - Validates token.
  - Extracts `user_id` and `roles`.
  - Forwards identity to downstream services via headers.

---

### 3. Service-to-service security and mTLS (concepts)

- **TLS**:
  - Client verifies server’s certificate.
  - Used for HTTPS.
- **mTLS (mutual TLS)**:
  - Both client and server present certificates.
  - Each side verifies the other’s identity.
- In microservices (especially in zero-trust-ish environments):
  - mTLS can be enforced by a service mesh or sidecars.
  - Individual services often rely on the mesh for certificate handling.

In this module:

- We will conceptually show:
  - How you would configure a Go `http.Client` and `http.Server` with TLS.
  - But not generate real certs (to keep it simple and runnable).

---

### 4. API security basics

- Validate input (schema/JSON structure).
- Enforce auth and authorization at the right layers.
- Rate limiting and throttling.
- Avoid leaking sensitive data in logs or error messages.

---

### 5. What we implement in Module 8

Single Go project `module8-security` that includes:

- **Auth service**:
  - `POST /login` – accepts `user_id` and returns a signed token.
- **Secure gateway**:
  - Validates tokens on incoming requests.
  - If valid, sets `X-User-ID` and `X-User-Roles` headers and forwards to a simple backend service.
- **Backend service**:
  - Reads `X-User-ID` and `X-User-Roles`.
  - Enforces simple role-based access:
    - `/admin` only for users with role `admin`.

This demonstrates:

- Token issuing and validation.
- Edge authentication & basic RBAC.
- Identity propagation to downstream services.

