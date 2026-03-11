## Module 12 – Service Mesh & Advanced Traffic Management (Conceptual + Light Config)

### 1. What is a service mesh?

- **Service mesh**: infrastructure layer for service-to-service communication, usually implemented with:
  - Sidecar proxies (e.g., Envoy) next to each service instance.
  - A control plane (e.g., Istio, Linkerd) that manages config for proxies.
- Responsibilities:
  - mTLS between services.
  - Retries, timeouts, circuit breaking, rate limiting.
  - Traffic splitting (canary, A/B).
  - Rich metrics and tracing out of the box.

Key idea: move many network/resilience/security concerns **out of application code** into the mesh.

---

### 2. Benefits vs drawbacks

- **Benefits**:
  - Standardized, consistent policies (retries, timeouts, mTLS) across services.
  - Deep observability without changing app code much.
  - Advanced traffic management: canary, mirroring, gradual rollouts.
- **Drawbacks**:
  - Operational complexity (install, upgrade, debug).
  - Performance overhead (extra hops, proxy CPU/memory).
  - More moving parts (control plane outages can affect traffic).

Interview angle: emphasize that service mesh is powerful but should be adopted when you already have many services and need uniform policies, not as step 1.

---

### 3. Traffic management patterns

- **Canary releases**:
  - Route a small percentage of traffic to new version.
  - Increase gradually as confidence grows.
- **Traffic mirroring**:
  - Copy requests to a shadow service (no impact on users).
  - Useful for testing new versions.
- **Per-route policies**:
  - Different timeouts/retry budgets for different APIs.

In a mesh, these are usually configured via YAML CRDs (e.g., `VirtualService`, `DestinationRule` in Istio).

---

### 4. What we provide in Module 12

Because setting up a full mesh (Istio/Linkerd) requires cluster-level tools, here we:

- Keep things **conceptual and interview-focused**.
- Provide:
  - A small example of Envoy-like config for:
    - mTLS between two services.
    - Traffic splitting (canary) between v1 and v2 of a service.
  - Notes on how this relates to patterns you already implemented in code (retries, circuit breakers, rate limiting).

See `module12-mesh-examples.md` for concrete config snippets you can reference in interviews.

