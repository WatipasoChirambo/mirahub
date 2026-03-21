FROM golang:1.22-alpine AS builder

RUN apk add --no-cache bash

WORKDIR /app
RUN go mod tidy
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o mirahub-app .

FROM alpine:latest
RUN apk add --no-cache bash

WORKDIR /root/
COPY --from=builder /app/mirahub-app ./mirahub-app

EXPOSE 8080

CMD ["./mirahub-app"]