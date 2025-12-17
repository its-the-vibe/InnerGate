# Build stage
FROM golang:1.25.5-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY *.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o innergate .

# Final stage - using scratch for minimal image
FROM scratch

# Copy the binary from builder
COPY --from=builder /build/innergate /innergate

# Copy SSL certificates for HTTPS requests (if needed)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Expose default port
EXPOSE 8080

# Set default config path
ENV CONFIG_PATH=/config.json

# Run the binary
ENTRYPOINT ["/innergate"]
