# Copilot Instructions for InnerGate

## Repository Overview

InnerGate is a lightweight reverse HTTP proxy service written in Go that multiplexes multiple incoming webhooks through a single externally exposed endpoint. It's designed for simplicity, performance, and ease of deployment.

### Purpose
- Route requests from a single external endpoint (e.g., ngrok) to multiple internal services
- Preserve all headers (including authentication) when forwarding requests
- Provide a minimal, containerized solution for webhook multiplexing

## Project Structure

```
InnerGate/
├── main.go              # Main application code with proxy logic
├── go.mod               # Go module definition
├── config.json          # Runtime configuration (not in git)
├── config.example.json  # Example configuration file
├── Dockerfile           # Multi-stage Docker build
├── docker-compose.yml   # Docker Compose setup
└── README.md           # User documentation
```

## Building and Running

### Local Development
```bash
# Build the binary
go build -o innergate .

# Run locally (requires config.json)
./innergate

# Run with custom config
CONFIG_PATH=/path/to/config.json PORT=9000 ./innergate
```

### Docker
```bash
# Build Docker image
docker build -t innergate .

# Run with Docker Compose (recommended)
docker-compose up -d
```

### Testing
The project currently does not have automated tests. When adding functionality:
- Manual testing should be performed by running the service and sending HTTP requests
- Test both exact path matches and prefix matches
- Verify headers are preserved
- Test error cases (invalid config, missing routes, unreachable backends)

## Coding Standards and Conventions

### Go Conventions
- Follow standard Go formatting (use `gofmt` or `go fmt`)
- Use Go 1.24+ features as specified in `go.mod`
- Keep the codebase simple and maintainable
- Use standard library packages whenever possible

### Code Style
- Clear, descriptive variable and function names
- Add comments for complex logic or non-obvious behavior
- Keep functions focused and single-purpose
- Use error wrapping with `fmt.Errorf` for better error context

### Configuration
- All runtime configuration is in JSON format
- Configuration file path via `CONFIG_PATH` env var (default: `config.json`)
- Port configuration via `PORT` env var (default: `8080`)
- Never commit actual `config.json` to git (use `config.example.json` for examples)

## Contribution Guidelines

### Making Changes
1. Keep changes minimal and focused
2. Maintain backward compatibility with existing config format
3. Update README.md if adding new features or changing behavior
4. Update config.example.json if adding new configuration options

### Docker Considerations
- The final image uses `scratch` base for minimal size
- SSL certificates are included for HTTPS backend support
- Binary is statically compiled with `CGO_ENABLED=0`
- Any changes to build process must maintain these characteristics

### Documentation
- Update README.md for user-facing changes
- Add inline comments for complex code logic
- Keep configuration examples up to date

## Technical Principles

1. **Simplicity**: Keep the codebase simple and easy to understand
2. **Minimal dependencies**: Use Go standard library whenever possible
3. **Container-first**: Ensure changes work well in Docker environments
4. **Header preservation**: All HTTP headers must be forwarded to backends
5. **Performance**: Keep the proxy lightweight and fast
6. **Error handling**: Always log errors with context for debugging

## Acceptance Criteria for Changes

When implementing new features or fixing bugs:
- [ ] Code follows Go conventions and formatting
- [ ] Changes are minimal and focused on the issue
- [ ] README.md is updated if behavior changes
- [ ] config.example.json is updated if config schema changes
- [ ] Manual testing has been performed
- [ ] Error cases are handled and logged appropriately
- [ ] Docker build succeeds and container runs correctly

## Common Tasks

### Adding a New Route Feature
1. Update the `Route` struct if needed
2. Modify routing logic in `ServeHTTP` method
3. Update config.example.json with new fields
4. Document the feature in README.md

### Modifying Proxy Behavior
1. Changes should be made in the `proxyRequest` method
2. Ensure headers are still preserved
3. Test with various backend services
4. Consider edge cases (timeouts, errors, redirects)

### Docker/Deployment Changes
1. Update Dockerfile if build process changes
2. Update docker-compose.yml for new environment variables
3. Test the full build and run process
4. Verify minimal image size is maintained
