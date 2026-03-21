FROM golang:1.26-alpine AS builder

RUN apk add --no-cache postgresql-client bash

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o mirahub-app main.go

FROM alpine:latest
RUN apk add --no-cache postgresql-client bash
WORKDIR /root/
COPY --from=builder /app/mirahub-app .
COPY wait-for-db.sh .
RUN chmod +x wait-for-db.sh
EXPOSE 8080
CMD ["./wait-for-db.sh", "db", "./mirahub-app"]