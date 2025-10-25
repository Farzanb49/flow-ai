# Pack CLI Bundling for Flow Deploy

The `flow deploy` command now supports bundling the `pack` CLI to eliminate external dependencies.

## Option 1: Bundle Pack CLI (Recommended)

### Download Pack CLI

Download the appropriate `pack` binary for your target platform:

```bash
# For Linux (most common for containers)
curl -sSL "https://github.com/buildpacks/pack/releases/latest/download/pack-v0.32.0-linux.tgz" | tar -xz -C /tmp
cp /tmp/pack /path/to/your/binary/directory/

# For macOS
curl -sSL "https://github.com/buildpacks/pack/releases/latest/download/pack-v0.32.0-macos.tgz" | tar -xz -C /tmp
cp /tmp/pack /path/to/your/binary/directory/

# For Windows
curl -sSL "https://github.com/buildpacks/pack/releases/latest/download/pack-v0.32.0-windows.zip" -o /tmp/pack.zip
unzip /tmp/pack.zip -d /tmp
cp /tmp/pack.exe /path/to/your/binary/directory/
```

### Build with Pack CLI

When building your Go application, include the `pack` binary:

```bash
# Build the Go application
go build -o flow ./cmd/deployer

# Copy pack binary to the same directory
cp pack flow-pack

# Or create a release script that bundles both
```

### Docker Build

For containerized deployments, include the pack binary in your Docker image:

```dockerfile
FROM golang:1.21-alpine AS builder

# Install pack CLI
RUN apk add --no-cache curl tar
RUN curl -sSL "https://github.com/buildpacks/pack/releases/latest/download/pack-v0.32.0-linux.tgz" | tar -xz -C /usr/local/bin/

# Build your application
WORKDIR /app
COPY . .
RUN go build -o flow ./cmd/deployer

FROM alpine:latest
RUN apk add --no-cache ca-certificates docker-cli
COPY --from=builder /usr/local/bin/pack /usr/local/bin/pack
COPY --from=builder /app/flow /usr/local/bin/flow
ENTRYPOINT ["/usr/local/bin/flow"]
```

## Option 2: Use Go Libraries (Alternative)

If you prefer to avoid bundling binaries, you can implement a pure Go solution using container libraries:

### Add Dependencies

```bash
go get github.com/google/go-containerregistry
go get github.com/buildpacks/libcnb
```

### Implementation

The current implementation falls back to system `pack` CLI if bundled version is not found. For a pure Go solution, you would need to:

1. Use `libcnb` to implement buildpack detection and build logic
2. Use `go-containerregistry` for image building and pushing
3. Implement the buildpack lifecycle directly

This approach is more complex but eliminates all external dependencies.

## Current Behavior

The `flow deploy` command will:

1. First look for a bundled `pack` binary in the same directory as the executable
2. Fall back to system `pack` CLI if found in PATH
3. Return an error if neither is available

This provides flexibility while reducing external dependencies for most use cases.
