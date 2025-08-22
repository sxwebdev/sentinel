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
FROM golang:1.25.0-alpine AS backend-builder

# Define build arguments for version, commit, and date.
ARG VERSION="unknown"
ARG COMMIT_HASH="unknown"
ARG BUILD_DATE="unknown"

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
RUN go build -trimpath -ldflags="-w -s -X 'main.version=${VERSION}' -X 'main.commitHash=${COMMIT_HASH}' -X 'main.buildDate=${BUILD_DATE}'" -o bin/sentinel ./cmd/sentinel

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /root/

# Copy binary from builder stage
COPY --from=backend-builder /app/bin/sentinel .

# Run the binary
CMD ["./sentinel", "start"]
