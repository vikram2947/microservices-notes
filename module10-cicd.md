## Module 10 – CI/CD & Delivery Safety

### 1. Pipeline stages (what every service should have)

- **Build**: compile Go binaries, build Docker images.
- **Test**: run unit + integration tests.
- **Static checks**: linters (`golangci-lint`), vulnerability scans (image and deps).
- **Package**: push images to registry.
- **Deploy**: apply manifests/Helm charts, with safeguards (e.g., only from `main`).

Interview angle: be able to describe this pipeline end-to-end and explain where quality/safety checks live.

---

### 2. Testing strategies

- **Unit tests**:
  - Functions, small components, no external I/O.
- **Integration tests**:
  - Talk to real (or test) DB/message broker; focus on service boundaries.
- **Contract tests**:
  - Ensure that service APIs match consumer expectations.
  - Often implemented as “consumer-driven contracts”.
- **End-to-end tests**:
  - Full flow across multiple services, closer to real traffic.

Rule: fast unit tests on every push; heavier tests in later pipeline stages or nightly.

---

### 3. Contract testing (high level)

- Producer (service) and consumer (client/another service) agree on:
  - Request/response shapes.
  - Status codes and error formats.
- Consumers define expectations as tests/contracts.
- Producers run those contracts as part of CI to avoid breaking changes.

We’ll add a simple HTTP contract test for `module1-service` as an example.

---

### 4. Schema evolution & migrations

- Avoid breaking changes:
  - Add fields (don’t remove) first.
  - Make new fields optional initially.
- **Expand/contract pattern**:
  - Expand: support both old and new schema/API.
  - Contract: remove old paths after all clients are migrated.

Database migrations:

- Apply schema changes in small, reversible steps.
- Keep application compatible with both pre- and post-migration states during rollout.

---

### 5. What we implement in Module 10

- A simple **GitHub Actions-style** CI definition file for:
  - Building and testing Go services.
  - Building Docker image for `module1-service`.
- Basic **Go tests** for `module1-service`:
  - Unit test for config loading.
  - HTTP handler test (contract-like check) ensuring `/healthz` returns expected status and body shape.

