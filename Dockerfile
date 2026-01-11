## Builder stage: compile the Go API binary
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install git and certificates if needed by Go modules
RUN apk add --no-cache git ca-certificates

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build a statically linked Linux binary
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -ldflags="-s -w" -o /app/api ./cmd/api


## Runtime stage: minimal image for running the compiled binary
FROM alpine:3.19

WORKDIR /app

# Ensure certificates are available for outbound HTTPS (e.g. Astra DB)
RUN apk add --no-cache ca-certificates && update-ca-certificates

COPY --from=builder /app/api /app/api

# Default port; override via PORT env if your app reads it
ENV PORT=8080
EXPOSE 8080

CMD ["/app/api"]


