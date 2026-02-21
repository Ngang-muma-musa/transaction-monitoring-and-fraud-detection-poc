############################
# STEP 1: Build executable
############################
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git gcc g++ libc-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGET=api

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" \
    -o /bin/transaction-monitoring-and-fraud-detection-poc ./cmd/${TARGET}

############################
# STEP 2: Create minimal runtime image
############################
FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /bin/transaction-monitoring-and-fraud-detection-poc /usr/local/bin/app-binary

RUN chmod +x /usr/local/bin/app-binary

WORKDIR /app

EXPOSE 8080

# Copy static web assets from build stage so runtime can serve them
COPY --from=builder /app/web /app/web

ENTRYPOINT ["/usr/local/bin/app-binary"]