# ---- Build Stage ----
FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy go module files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /sync ./cmd/sync

# ---- Runtime Stage ----
FROM alpine:3.22

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binaries from the builder stage
COPY --from=builder /server /app/server
COPY --from=builder /sync /app/sync

# Copy SQL init scripts (useful for reference or manual migration)
COPY structure.sql /app/structure.sql

EXPOSE 8080

CMD ["/app/server"]
