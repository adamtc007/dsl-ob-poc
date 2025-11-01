# GREENTEAGC Quick Reference

## What is it?

`GOEXPERIMENT=greenteagc` enables Go's experimental green tea garbage collector for optimized performance.

## Quick Start

```bash
# Build with greenteagc (recommended)
GOEXPERIMENT=greenteagc go build -o dsl-poc .

# Or use the build script
./build.sh

# Or use make
make build-greenteagc
```

## Compiler Flag

| Property | Value |
|----------|-------|
| Environment Variable | `GOEXPERIMENT` |
| Setting | `greenteagc` |
| Required Go Version | 1.21+ |
| Placement | Before `go build` command |

## Three Ways to Build

### 1. Direct Command (Most Explicit)
```bash
GOEXPERIMENT=greenteagc go build -o dsl-poc .
```

### 2. Build Script (Most Convenient)
```bash
chmod +x build.sh
./build.sh                  # with greenteagc (default)
./build.sh --no-greenteagc  # with standard GC
```

### 3. Make (Best for CI/CD)
```bash
make build-greenteagc    # with greenteagc (recommended)
make build               # with standard GC
```

## Why Use It?

✅ Lower garbage collection pause times  
✅ Better memory efficiency  
✅ Improved performance for concurrent workloads  
✅ More predictable latency  

## Usage Pattern

```bash
# Setup
export DB_CONN_STRING="postgres://user:pass@localhost/db?sslmode=disable"

# Build (pick one method above)
GOEXPERIMENT=greenteagc go build -o dsl-poc .

# Initialize database
./dsl-poc init-db

# Run commands
./dsl-poc create --cbu="CBU-1234"
./dsl-poc add-products --cbu="CBU-1234" --products="CUSTODY,FUND_ACCOUNTING"
```

## Available Build Scripts

| Tool | Command | Purpose |
|------|---------|---------|
| Direct | `GOEXPERIMENT=greenteagc go build -o dsl-poc .` | Manual control |
| Script | `./build.sh` | User-friendly with options |
| Make | `make build-greenteagc` | Standardized build process |

## Verify Your Build

Check that greenteagc was used:
```bash
./build.sh
# Output should show: "Building with experimental greenteagc garbage collector..."
```

## Troubleshooting

| Issue | Solution |
|-------|----------|
| "unknown experiment: greenteagc" | Update Go to 1.21+ |
| Variable not recognized | Use correct syntax: `GOEXPERIMENT=greenteagc go build` |
| Build fails | Run `go clean -cache` then try again |

## Environment Variable Syntax

```bash
# Bash/Linux/macOS
GOEXPERIMENT=greenteagc go build -o dsl-poc .

# Windows CMD
set GOEXPERIMENT=greenteagc && go build -o dsl-poc .

# Windows PowerShell
$env:GOEXPERIMENT='greenteagc'; go build -o dsl-poc .
```

## Key Files

- `Makefile` - Standardized build targets with greenteagc support
- `build.sh` - Build script with colored output and options
- `BUILD.md` - Comprehensive build documentation
- `README.md` - General setup and usage instructions

## Next Steps

1. Run `make help` to see all available targets
2. Run `make build-greenteagc` to build
3. Set `DB_CONN_STRING` environment variable
4. Run `./dsl-poc init-db` to initialize database
5. Test with `./dsl-poc create --cbu="CBU-1234"`

## Performance Tips

```bash
# Monitor GC behavior at runtime
GODEBUG=gctrace=1 ./dsl-poc create --cbu="CBU-1234"

# Set memory limit (Go 1.19+)
GOMEMLIMIT=1024MiB GOEXPERIMENT=greenteagc go build -o dsl-poc .

# Parallel GC on multi-core systems
GOMAXPROCS=8 ./dsl-poc add-products --cbu="CBU-1234" --products="CUSTODY"
```

## References

See `BUILD.md` for comprehensive documentation on:
- Detailed compiler flags explanation
- CI/CD integration examples
- Performance tuning options
- Troubleshooting guide