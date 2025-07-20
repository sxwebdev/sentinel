# Frontend build stage
FROM node:24-alpine AS frontend-builder

WORKDIR /app/frontend

# Copy frontend package files
COPY frontend/package.json frontend/pnpm-lock.yaml ./

# Install pnpm and dependencies
RUN npm install -g pnpm && pnpm install

# Copy frontend source
COPY frontend/ ./

# Build frontend
RUN pnpm run build

# Backend build stage
FROM golang:1.24-alpine AS backend-builder

# Install dependencies
RUN apk add --no-cache ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Copy built frontend from previous stage
COPY --from=frontend-builder /app/frontend/dist ./frontend/dist

# Build the application with embedded frontend
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix -ldflags="-w -s" -o sentinel ./cmd/server

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder stage
COPY --from=backend-builder /app/sentinel .

# Run the binary
CMD ["./sentinel"]
