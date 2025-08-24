# Build stage
FROM golang:1.23-alpine AS builder

# Install git for go mod download
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o goalhero-payment-jobs .

# Final stage
FROM scratch

# Copy timezone data and certificates
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /app/goalhero-payment-jobs /goalhero-payment-jobs


# Expose port
EXPOSE 8081

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/goalhero-payment-jobs", "-health-check"] || exit 1

# Run the binary
ENTRYPOINT ["/goalhero-payment-jobs"]