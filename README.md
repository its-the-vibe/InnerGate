# InnerGate

A simple service written in Go that acts as a reverse HTTP proxy for internally hosted services. This is useful for multiplexing multiple incoming webhooks using a single externally exposed endpoint (e.g., via ngrok).

## Features

- **Route-based proxying**: Forward requests to different backend services based on URL paths
- **Header preservation**: All headers (including authentication headers) are forwarded to backend services
- **Lightweight**: Built with Go and runs in a minimal Docker container (using scratch base image)
- **Easy configuration**: Simple JSON-based configuration file

## Configuration

Create a `config.json` file with your routes:

```json
[
  {
    "name": "github-webhook",
    "path": "github-webhook",
    "target": "http://localhost:8000/webhook"
  },
  {
    "name": "slack-webhook",
    "path": "slack-webhook",
    "target": "http://localhost:8001/webhook"
  }
]
```

Each route has:
- `name`: A descriptive name for the route
- `path`: The URL path to match (without leading slash)
- `target`: The backend service URL to forward requests to

## Usage

### Using Docker Compose (Recommended)

1. Create your `config.json` file
2. Run:
```bash
docker-compose up -d
```

The service will be available on port 8080.

### Using Docker

Build the image:
```bash
docker build -t innergate .
```

Run the container:
```bash
docker run -p 8080:8080 -v $(pwd)/config.json:/config.json:ro innergate
```

### Running Locally

Build the binary:
```bash
go build -o innergate .
```

Run:
```bash
./innergate
```

### Environment Variables

- `CONFIG_PATH`: Path to configuration file (default: `config.json`)
- `PORT`: Port to listen on (default: `8080`)

## Example

With the example configuration above:
- Requests to `http://localhost:8080/github-webhook` → forwarded to `http://localhost:8000/webhook`
- Requests to `http://localhost:8080/slack-webhook` → forwarded to `http://localhost:8001/webhook`

All headers (including authentication headers like `Authorization`, `X-GitHub-Event`, etc.) are preserved and forwarded to the backend services.

## Use Case

This is particularly useful when you need to expose multiple internal services through a single external endpoint, such as:
- Multiple webhook receivers behind a single ngrok tunnel
- Centralizing access to internal services
- Simplifying external access management
