## Module 9 – Container & Kubernetes Usage for `module1-service`

### 1. Build and run Docker image

From the repo root:

```bash
cd module1-service

docker build -t module1-service:latest .

docker run --rm -p 8080:8080 -e APP_PORT=8080 module1-service:latest
```

Then, in another terminal:

```bash
curl -s http://localhost:8080/healthz
```

### 2. Run on Kubernetes (kind/minikube)

Assuming you have a local Kubernetes cluster and `kubectl`:

```bash
cd module1-service

kubectl apply -f k8s-deployment.yaml
```

Check resources:

```bash
kubectl get pods
kubectl get svc module1-service
kubectl get ingress module1-service
```

Depending on your Ingress controller and DNS, you can access it via the host configured in the Ingress (e.g., `module1.local`) or via `kubectl port-forward` on the Service.

