FROM golang:1.25-alpine AS builder

RUN apk add --no-cache bash

WORKDIR /app

# Copy module files first
COPY go.mod go.sum ./

# Tidy & download dependencies
RUN go mod tidy
RUN go mod download

# Now copy the source code
COPY . .

# Build the binary
RUN go build -o mirahub-app .

# Minimal runtime image
FROM alpine:latest
RUN apk add --no-cache bash

WORKDIR /root/
COPY --from=builder /app/mirahub-app ./mirahub-app

EXPOSE 8080

CMD ["./mirahub-app"]