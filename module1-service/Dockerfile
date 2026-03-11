## Module 9 – Dockerfile for module1-service

# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

# Final minimal image
FROM alpine:3.20

WORKDIR /app

ENV APP_PORT=8080

COPY --from=builder /app/server /app/server

EXPOSE 8080

ENTRYPOINT ["/app/server"]

