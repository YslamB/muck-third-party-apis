# Stage 1: Build the binary
FROM golang:1.23.4 AS builder

WORKDIR /build

# Copy Go source code
COPY . .

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app

# Stage 2: Create a minimal image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /build/app .


CMD ["./app"]
