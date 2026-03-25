# Build stage
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy go mod files and vendor directory
COPY go.mod go.sum ./
COPY vendor/ ./vendor/

# Copy source code
COPY . .

# Build the binary using vendored dependencies
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /flight-search ./cmd/server

# Runtime stage
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /flight-search .
COPY --from=builder /app/mockdata ./mockdata
COPY --from=builder /app/.env .env

EXPOSE 3000

CMD ["./flight-search"]
