# Build Configuration Guide - GOEXPERIMENT=greenteagc

This document describes the build configuration for the DSL Onboarding POC, specifically the use of the experimental `greenteagc` garbage collector.

## Overview

This project uses **Go's experimental green tea garbage collector** (`GOEXPERIMENT=greenteagc`) to optimize performance for the DSL state machine workload.

## What is greenteagc?

`greenteagc` is an experimental garbage collection implementation in Go that:

- **Reduces pause times**: Minimizes stop-the-world GC pauses
- **Improves memory efficiency**: Better memory layout and allocation patterns
- **Optimizes for workloads**: Particularly beneficial for applications with frequent allocations (like this DSL parser)
- **Concurrent design**: Better scaling on multi-core systems

## Compiler Flags

### Primary Flag

```
GOEXPERIMENT=greenteagc
```

This environment variable must be set **before** running `go build`.

### Complete Build Command

```sh
GOEXPERIMENT=greenteagc go build -o dsl-poc .
```

### Breakdown

| Component | Purpose |
|-----------|---------|
| `GOEXPERIMENT=greenteagc` | Enable experimental green tea GC |
| `go build` | Go compiler command |
| `-o dsl-poc` | Output binary name |
| `.` | Build current directory |

## Build Methods

### Method 1: Direct Command (Explicit Control)

```sh
GOEXPERIMENT=greenteagc go build -o dsl-poc .
```

**Pros**: Full control, transparent  
**Cons**: Verbose, easy to forget the flag

### Method 2: Build Script (Recommended for Development)

```sh
chmod +x build.sh
./build.sh                    # Uses greenteagc by default
./build.sh --no-greenteagc    # Use standard GC if needed
./build.sh -o custom-binary   # Custom output name
```

**Features**:
- Colored output for better readability
- Automatic dependency management
- Error checking and helpful messages
- Flag-based control

### Method 3: Makefile (Recommended for CI/CD)

```sh
make build-greenteagc    # Build with greenteagc
make build               # Build with standard GC
make clean               # Clean up
make help                # Show all targets
```

**Features**:
- Standardized build process
- Dependency management via `install-deps` target
- Simple, idiomatic Go development

## Environment Variables

### Required for Build

```sh
# Not required for build itself, but needed at runtime
export DB_CONN_STRING="postgres://user:password@localhost:5432/db?sslmode=disable"
```

### Build-Time Variables

```sh
# Set greenteagc before building
export GOEXPERIMENT=greenteagc
go build -o dsl-poc .

# Or inline (recommended)
GOEXPERIMENT=greenteagc go build -o dsl-poc .
```

## Verification

### Verify greenteagc Build

To confirm your binary was built with greenteagc, check the build output:

```sh
$ ./build.sh
Building DSL POC...
Go version: go version go1.21.x darwin/amd64
Downloading dependencies...
Building with experimental greenteagc garbage collector...
go build -v -o dsl-poc .
...
✓ Build successful!
```

### Check Go Version Compatibility

Ensure you're using Go 1.21 or later:

```sh
go version
# Output: go version go1.21.x darwin/amd64
```

## Performance Characteristics

### When to Use greenteagc

✅ Use greenteagc when:
- Running production workloads
- Memory efficiency is important
- Working with frequent allocations
- Running on multi-core systems

❌ Consider standard GC when:
- Debugging garbage collection issues
- Testing for compatibility
- Working on very old Go versions (< 1.21)

## Troubleshooting

### Build Error: "GOEXPERIMENT: unknown experiment greenteagc"

**Cause**: Go version too old or experiment not available  
**Solution**: Update to Go 1.21+

```sh
go version  # Check current version
go get -u ...  # Update if needed
```

### Build Error: "go: unknown flag '-o'"

**Cause**: Incorrect build command syntax  
**Solution**: Ensure proper flag order

```sh
# Correct
GOEXPERIMENT=greenteagc go build -o dsl-poc .

# Also correct
go build -o dsl-poc .  # (without greenteagc)
```

### Build Succeeds but Binary Not Updated

**Cause**: Stale binary or build cache  
**Solution**: Clean and rebuild

```sh
make clean
make build-greenteagc

# Or
go clean -cache
GOEXPERIMENT=greenteagc go build -o dsl-poc .
```

## Performance Tuning

### Additional GC Flags (Optional)

These can be combined with `GOEXPERIMENT=greenteagc` for further tuning:

```sh
# Memory limit (Go 1.19+)
GOMEMLIMIT=1024MiB GOEXPERIMENT=greenteagc go build -o dsl-poc .

# Debug garbage collection
GODEBUG=gctrace=1 ./dsl-poc init-db

# Parallel GC
GOMAXPROCS=8 ./dsl-poc create --cbu="CBU-1234"
```

### Runtime Performance

Monitor GC behavior:

```sh
# Enable GC trace on binary run
GODEBUG=gctrace=1 ./dsl-poc create --cbu="CBU-1234"

# Output shows GC pauses, memory stats, etc.
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Build with greenteagc

on: [push, pull_request]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: make build-greenteagc
      - run: ./dsl-poc init-db
```

### Docker Build Example

```dockerfile
FROM golang:1.21-alpine as builder

WORKDIR /app
COPY . .

ENV GOEXPERIMENT=greenteagc
RUN go build -o dsl-poc .

FROM alpine:latest
COPY --from=builder /app/dsl-poc /usr/local/bin/
ENTRYPOINT ["dsl-poc"]
```

## Summary

| Aspect | Details |
|--------|---------|
| **Flag** | `GOEXPERIMENT=greenteagc` |
| **Minimum Go Version** | 1.21+ |
| **Recommended** | Yes, for production |
| **Build Methods** | Direct, Script, Makefile |
| **Performance Impact** | Improved GC efficiency |
| **Compatibility** | Requires Go 1.21+ |

## References

- [Go Experiments](https://pkg.go.dev/cmd/compile)
- [Go Garbage Collection](https://go.dev/blog/gc-guide)
- [GOEXPERIMENT Documentation](https://pkg.go.dev/runtime)