## Module 1 – Minimal Go Microservice

### Prerequisites

- Go 1.22+ installed

### Run the service

```bash
cd module1-service

# Optional: override default port 8080
export APP_PORT=8080

go run ./...
```

You should see logs like:

```text
[module1-service] ... starting server on :8080
```

### Test the endpoints

Health:

```bash
curl -i http://localhost:8080/healthz
curl -i http://localhost:8080/readyz
```

Hello:

```bash
curl -i http://localhost:8080/hello
```

### Graceful shutdown

- While the server is running, press `Ctrl+C`.
- The server will stop accepting new requests and shut down gracefully within a few seconds.

