# ---------- Builder stage ----------
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache bash

WORKDIR /app

# Copy go mod files first (for caching dependencies)
COPY go.mod go.sum ./

# Download dependencies only (no tidy yet)
RUN go mod download

# Now copy the full source code
COPY . .

# Tidy AFTER code is present
RUN go mod tidy

# Build the binary
RUN go build -o mirahub-app .

# ---------- Runtime stage ----------
FROM alpine:latest

RUN apk add --no-cache bash

WORKDIR /root/

# Copy compiled binary
COPY --from=builder /app/mirahub-app .

EXPOSE 8080

CMD ["./mirahub-app"]