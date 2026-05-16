# ==========================================
# Stage 1: Build
# ==========================================
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Cache dependency downloads
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate Swagger docs (required by main.go)
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g cmd/http/main.go -o ./docs --parseDependency --parseInternal --exclude vendor

# Build the binaries
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/http
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/migrate ./cmd/migrate

# ==========================================
# Stage 2: Runtime
# ==========================================
FROM alpine:3.21

RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    python3 \
    py3-pip \
    bash \
    curl \
    && rm -rf /var/cache/apk/*

# Install common Python packages for script execution
RUN pip3 install --no-cache-dir --break-system-packages \
    requests \
    httpx \
    pydantic

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/bin/server /app/server
COPY --from=builder /app/bin/migrate /app/migrate

# Copy config, migration, and script files
COPY --from=builder /app/files /app/files

# Expose HTTP port
EXPOSE 8080

ENTRYPOINT ["/app/server"]
CMD ["--config", "/app/files/config/app.yaml"]
