# Build stage
FROM --platform=$BUILDPLATFORM golang:1.27.0-alpine AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /build

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY *.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -trimpath -ldflags="-s -w" -o innergate .

# Final stage (distroless)
FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=builder /build/innergate /innergate

# Expose default port
EXPOSE 8080

# Set default config path
ENV CONFIG_PATH=/config.json

USER nonroot:nonroot

# Run the binary
ENTRYPOINT ["/innergate"]
