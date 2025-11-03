# Getting Started - DSL Onboarding POC

## Quick Start (5 minutes)

### 1. Prerequisites
```bash
# Verify Go 1.21+ is installed
go version

# Verify PostgreSQL is running
psql --version
```

### 2. Clone and Navigate
```bash
cd dsl-ob-poc
```

### 3. Set Database Connection
```bash
export DB_CONN_STRING="postgres://user:password@localhost:5432/your_db?sslmode=disable"
```

### 4. Build with greenteagc (Recommended)
```bash
# Option A: Using the build script (easiest)
chmod +x build.sh
./build.sh

# Option B: Using make
make build-greenteagc

# Option C: Direct command
GOEXPERIMENT=greenteagc go build -o dsl-poc .
```

### 5. Initialize Database
```bash
./dsl-poc init-db
```

### 6. Create Your First Case
```bash
./dsl-poc create --cbu="CBU-1234"
```

You should see output like:
```
Created new case version: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6
---
(case.create
  (cbu.id "CBU-1234")
  (nature-purpose "UCITS equity fund domiciled in LU")
)
---
```

### 7. Add Products (State Transition)
```bash
./dsl-poc add-products --cbu="CBU-1234" --products="CUSTODY,FUND_ACCOUNTING"
```

You should see a new version with the products appended:
```
Created new case version: x9y8z7w6-v5u4-t3s2-r1q0-p9o8n7m6l5k4
---
(case.create
  (cbu.id "CBU-1234")
  (nature-purpose "UCITS equity fund domiciled in LU")
)

(products.add "CUSTODY" "FUND_ACCOUNTING")
---
```

✅ **You're done!** You've successfully created an immutable, versioned DSL record.

---

## Build Methods Explained

### Method 1: Build Script (Recommended for Development)
```bash
chmod +x build.sh
./build.sh                    # Uses greenteagc by default
./build.sh --no-greenteagc    # Use standard GC if needed
./build.sh -o custom-name     # Custom binary name
./build.sh -h                 # Show help
```

**Why:** User-friendly, colored output, automatic error handling, shows what's happening.

### Method 2: Makefile (Recommended for CI/CD)
```bash
make build-greenteagc    # Build with greenteagc
make build               # Build with standard GC
make help                # Show all available targets
make clean               # Clean build artifacts
```

**Why:** Standardized, idiomatic Go development, perfect for automation.

### Method 3: Direct Command (Manual Control)
```bash
GOEXPERIMENT=greenteagc go build -o dsl-poc .
```

**Why:** Full control, useful for scripting and understanding what's happening.

---

## What is greenteagc?

`GOEXPERIMENT=greenteagc` enables Go's experimental green tea garbage collector.

### Benefits
- ✅ Lower garbage collection pause times (60% reduction)
- ✅ Better memory efficiency
- ✅ Improved throughput (~4% better)
- ✅ More predictable latency
- ✅ Optimized for concurrent workloads

### Requirements
- Go 1.21 or later
- Must be set **before** running `go build`
- Experimental feature (requires explicit opt-in)

### When to Use
- **Production:** Yes, recommended
- **Development:** Yes, for better performance
- **Debugging:** Maybe not, standard GC is easier to troubleshoot

---

## Available Commands

### Create a New Case
```bash
./dsl-poc create --cbu="CBU-1234"
```

Creates the initial DSL state for a CBU. This is version 1.

### Add Products
```bash
./dsl-poc add-products --cbu="CBU-1234" --products="CUSTODY,FUND_ACCOUNTING"
```

Appends products to an existing case, creating a new immutable version.

### Initialize Database (One-Time)
```bash
./dsl-poc init-db
```

Creates the `kyc-dsl` schema and `dsl_ob` table in PostgreSQL. Only run once.

### Help
```bash
./dsl-poc help
```

Shows usage information and available commands.

---

## Troubleshooting

### Error: "DB_CONN_STRING environment variable is not set"
```bash
# Solution: Set the environment variable
export DB_CONN_STRING="postgres://user:password@localhost:5432/db?sslmode=disable"
```

### Error: "unknown GOEXPERIMENT: greenteagc"
```bash
# Solution: Update Go to 1.21 or later
go version
# If below 1.21, download from: https://golang.org/dl/

# Fallback: Build without greenteagc
./build.sh --no-greenteagc
# or
make build
```

### Error: "failed to connect to database"
```bash
# Solution: Check PostgreSQL connection
# 1. Ensure PostgreSQL is running
# 2. Verify connection string is correct
# 3. Check database user has permissions

# Test connection:
psql $DB_CONN_STRING -c "SELECT 1;"
```

### Binary doesn't work after build
```bash
# Solution: Clean cache and rebuild
go clean -cache
./build.sh
# or
make clean
make build-greenteagc
```

---

## Project Structure

```
dsl-ob-poc/
├── main.go                 # Entry point & CLI routing
├── go.mod / go.sum         # Go dependencies
├── sql/
│   └── init.sql           # Database schema
├── internal/
│   ├── cli/               # Command handlers
│   │   ├── create.go      # Create case command
│   │   └── add_products.go # Add products command
│   ├── dsl/               # DSL generation
│   │   └── dsl.go         # S-expression builders
│   ├── store/             # Database layer
│   │   └── store.go       # PostgreSQL operations
│   └── mocks/             # Test data
│       └── data.go        # Mock CBU data
├── Makefile               # Build automation
├── build.sh               # Build script
├── README.md              # Full documentation
├── BUILD.md               # Build configuration
├── GREENTEAGC.md          # Quick reference
└── COMPILER_FLAGS.md      # Technical details
```

---

## Documentation Guide

| Document | Best For | Read Time |
|----------|----------|-----------|
| **GETTING_STARTED.md** (this file) | First-time setup | 5 min |
| **README.md** | Project overview & usage | 10 min |
| **GREENTEAGC.md** | Quick compiler flag reference | 3 min |
| **BUILD.md** | Detailed build configuration | 15 min |
| **COMPILER_FLAGS.md** | Complete technical reference | 20 min |

---

## Next Steps

### For Local Development
1. ✅ Build with `./build.sh`
2. ✅ Initialize database with `./dsl-poc init-db`
3. ✅ Create test cases with `./dsl-poc create --cbu="..."`
4. ✅ Query PostgreSQL to verify immutable versions:
   ```sql
   SELECT version_id, cbu_id, created_at FROM "dsl-ob-poc".dsl_ob ORDER BY created_at DESC;
   ```

### For Production Deployment
1. Read `BUILD.md` for production build options
2. Read `COMPILER_FLAGS.md` for advanced configuration
3. Use `make build-greenteagc` in CI/CD pipeline
4. Monitor GC behavior: `GODEBUG=gctrace=1 ./dsl-poc ...`

### For Contributing
1. Read `README.md` for full project context
2. Check `internal/` directory structure
3. Follow existing code patterns in `cli/` and `dsl/` packages
4. Add new commands in `internal/cli/` package

---

## Common Workflows

### Development Loop
```bash
# Setup (one time)
export DB_CONN_STRING="postgres://user:pass@localhost/db?sslmode=disable"
make build-greenteagc
./dsl-poc init-db

# Development (repeat as needed)
./dsl-poc create --cbu="TEST-001"
./dsl-poc add-products --cbu="TEST-001" --products="CUSTODY"
./dsl-poc add-products --cbu="TEST-001" --products="FUND_ACCOUNTING"

# Verify in database
psql $DB_CONN_STRING -c 'SELECT version_id, dsl_text FROM "dsl-ob-poc".dsl_ob WHERE cbu_id = '\''TEST-001'\'';'
```

### Testing a New Build
```bash
# Clean build
make clean
make build-greenteagc

# Verify binary exists
ls -lh dsl-poc

# Quick test
export DB_CONN_STRING="..."
./dsl-poc init-db
./dsl-poc create --cbu="CBU-TEST"
```

### Comparing Standard vs Experimental GC
```bash
# Build with greenteagc
make build-greenteagc
time ./dsl-poc create --cbu="CBU-1"

# Build with standard GC
make build
time ./dsl-poc create --cbu="CBU-2"

# Compare times (greenteagc should be faster or similar)
```

---

## Performance Tips

### Monitor Garbage Collection
```bash
# See detailed GC statistics
GODEBUG=gctrace=1 ./dsl-poc create --cbu="CBU-1234"

# Output shows GC pauses, memory allocation, etc.
```

### Limit Memory Usage (Go 1.19+)
```bash
# Restrict heap to 1GB
GOMEMLIMIT=1024MiB ./dsl-poc add-products --cbu="CBU-1234" --products="CUSTODY"
```

### Utilize Multiple Cores
```bash
# Use all available CPU cores
GOMAXPROCS=8 ./dsl-poc create --cbu="CBU-1234"
```

---

## Getting Help

### See All Available Commands
```bash
./dsl-poc help
```

### View Build Targets
```bash
make help
```

### Show Build Script Options
```bash
./build.sh -h
```

### Documentation Hierarchy
1. **Quick answers:** `GREENTEAGC.md`
2. **Build issues:** `BUILD.md`
3. **Technical details:** `COMPILER_FLAGS.md`
4. **Project context:** `README.md`

---

## Key Takeaways

✅ **Always build with:** `./build.sh` or `make build-greenteagc`

✅ **greenteagc provides:** Better performance, lower latency, experimental feature

✅ **Requires:** Go 1.21+, PostgreSQL running, DB_CONN_STRING set

✅ **Immutable versioning:** Each command creates a new, unchangeable version

✅ **Three build methods:** Script (easiest), Make (CI/CD), Direct (manual)

---

## Questions?

- **"How do I...?"** → Check `README.md`
- **"What's greenteagc?"** → Read `GREENTEAGC.md`
- **"Build not working?"** → See `BUILD.md` troubleshooting
- **"I want details"** → Read `COMPILER_FLAGS.md`

Welcome aboard! Happy building! 🚀
