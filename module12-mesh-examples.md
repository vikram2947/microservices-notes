## Module 12 – Mesh Config Examples (Istio-style, Conceptual)

These are **reference snippets** – you don’t have to run them, but you can quote/adapt them in interviews.

### 1. mTLS between services

```yaml
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
  namespace: my-namespace
spec:
  mtls:
    mode: STRICT
```

This enforces **mutual TLS** for all workloads in `my-namespace`. Proxies handle certificates and verification; apps just speak plain HTTP to localhost.

### 2. Traffic splitting (canary) between v1 and v2

```yaml
apiVersion: networking.istio.io/v1beta1
kind: DestinationRule
metadata:
  name: orders
spec:
  host: orders
  subsets:
    - name: v1
      labels:
        version: v1
    - name: v2
      labels:
        version: v2
---
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: orders
spec:
  hosts:
    - orders
  http:
    - route:
        - destination:
            host: orders
            subset: v1
          weight: 90
        - destination:
            host: orders
            subset: v2
          weight: 10
```

This sends **90%** of traffic to `v1` Pods (label `version: v1`) and **10%** to `v2` – a classic canary rollout.

### 3. Per-route timeout and retry policy

```yaml
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: payments
spec:
  hosts:
    - payments
  http:
    - match:
        - uri:
            prefix: /charges
      timeout: 2s
      retries:
        attempts: 3
        perTryTimeout: 500ms
        retryOn: 5xx,connect-failure,refused-stream
```

This applies:

- 2s overall timeout for `/charges`.
- Up to 3 retries with per-try timeout 500ms.
- Retry on specific failure types.

You can map this directly to the **retry + timeout** logic you implemented in Module 3, but here it’s managed by the mesh instead of application code.

