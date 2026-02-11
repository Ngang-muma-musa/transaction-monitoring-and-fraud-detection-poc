############################
# STEP 1: Build executable
############################
FROM golang:1.24-alpine AS builder

# Install dependencies for building Go binaries
RUN apk add --no-cache git gcc g++ libc-dev

WORKDIR /app

# Copy dependency files and download first (cache optimization)
COPY go.mod go.sum ./
RUN go mod download

# Copy entire project
COPY . .

# Accept build target: "api" or "worker"
ARG TARGET=api

# Build binary based on target
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/transaction-monitoring-and-fraud-detection-poc ./cmd/${TARGET}

############################
# STEP 2: Create minimal runtime image
############################
FROM alpine:latest

RUN apk --no-cache add ca-certificates

# Copy compiled binary and version info
COPY --from=builder /bin/transaction-monitoring-and-fraud-detection-poc /bin/transaction-monitoring-and-fraud-detection-poc

WORKDIR /app

# Use port 8080 for API only
EXPOSE 8080

# Start application
ENTRYPOINT ["/bin/transaction-monitoring-and-fraud-detection-poc"]
