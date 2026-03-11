## Module 9 – Containers, Orchestration & Kubernetes Basics

### 1. Containers & images (for Go services)

- **Image**: packaged filesystem + metadata.
- **Container**: running instance of an image.
- For Go services:
  - Build statically linked binary.
  - Use multi-stage builds:
    - Stage 1: build.
    - Stage 2: minimal runtime image (e.g., `scratch` or small base).

Key ideas:

- Use small images → faster pulls, smaller attack surface.
- One main process per container.

---

### 2. Kubernetes essentials (mental model)

- **Pod**: smallest deployable unit; usually 1 container (sometimes sidecars).
- **Deployment**: manages replicated Pods and rolling updates.
- **Service**: stable virtual IP + DNS for a set of Pods.
  - Types: `ClusterIP` (internal), `NodePort`, `LoadBalancer`.
- **Ingress**: HTTP(S) routing from outside the cluster into Services.
- **ConfigMaps / Secrets**:
  - Externalize configuration and sensitive data.
- **Probes**:
  - `livenessProbe`: when to restart a container.
  - `readinessProbe`: when a Pod is ready for traffic.

---

### 3. Deployment strategies (high-level)

- **Rolling update**:
  - K8s default: gradually replace old Pods with new ones.
- **Blue/Green**:
  - Two environments; switch traffic at once.
- **Canary**:
  - Send small percentage of traffic to new version; increase gradually.

In this module we will use simple **Deployments + Services** and mention how canary/blue-green build on them.

---

### 4. What we implement in Module 9

Using one of our simple services (`module1-service`), we will:

- Add a **Dockerfile** with a multi-stage build.
- Add Kubernetes manifests in a `k8s/` folder:
  - `Deployment` for the service.
  - `Service` (ClusterIP).
  - `Ingress` example (assuming an Ingress controller).
- Show how to use:
  - Environment variables via `Deployment`.
  - Basic liveness and readiness probes.

These manifests are designed to be used with a local cluster (like kind/minikube), but are also good interview references.

