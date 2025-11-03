# Onboarding DSL - Go POC

This application is a proof-of-concept for the client onboarding DSL, with all state logic and persistence handled in Go.

## Prerequisites

1.  Go (1.21+)
2.  PostgreSQL (running and accessible)

## Setup

1.  **Set Environment Variable:**
    You must provide a connection string to your PostgreSQL database.

    ```sh
    export DB_CONN_STRING="postgres://user:password@localhost:5432/your_db?sslmode=disable"
    ```

    *Note: The user must have permissions to create schemas and tables.*

2.  **Install Dependencies:**
    ```sh
    go mod tidy
    ```

3.  **Build the CLI:**

    **Recommended: Build with experimental green tea garbage collector**
    ```sh
    GOEXPERIMENT=greenteagc go build -o dsl-poc .
    ```
    
    The `greenteagc` experiment provides improved garbage collection performance for this workload.

    **Alternative: Build with standard Go garbage collector**
    ```sh
    go build -o dsl-poc .
    ```

    **Or use the provided build script:**
    ```sh
    chmod +x build.sh
    ./build.sh                  # Uses greenteagc by default
    ./build.sh --no-greenteagc  # Uses standard GC
    ```

    **Or use make:**
    ```sh
    make build-greenteagc       # Recommended
    make build                  # Standard GC
    make help                   # Show all available targets
    ```

4.  **Initialize the Database:**
    Run the `init-db` command. This only needs to be done once. It creates the `"dsl-ob-poc"` schema and the `"dsl_ob"` table.

    ```sh
    ./dsl-poc init-db
    ```
    *Output: "Database initialized successfully."*

## Running the State Machine

The CLI allows you to run state transitions. Each command creates a **new, immutable version** of the DSL in the database.

### 1. Create a New Case

This is the first state, `CREATE`.

```sh
./dsl-poc create --cbu="CBU-1234"
```
*Output:*
```
Created new case version: a1b2c3d4-....
---
(case.create
  (cbu.id "CBU-1234")
  (nature-purpose "UCITS equity fund domiciled in LU")
)
---
```

### 2. Add Products (State Change)

This demonstrates a state change. It finds the *latest* version for "CBU-1234", appends the new DSL command, and saves it as a **new version**.

```sh
./dsl-poc add-products --cbu="CBU-1234" --products="CUSTODY,FUND_ACCOUNTING"
```
*Output:*
```
Created new case version: e5f6g7h8-....
---
(case.create
  (cbu.id "CBU-1234")
  (nature-purpose "UCITS equity fund domiciled in LU")
)

(products.add "CUSTODY" "FUND_ACCOUNTING")
---
```

You can now check your `dsl_ob` table in PostgreSQL to see both immutable versions.

## Build Configuration

This project supports the experimental `greenteagc` garbage collector for improved performance:

- **GOEXPERIMENT=greenteagc**: Enables the experimental green tea garbage collector
  - Optimized for workloads with frequent allocations
  - Recommended for this DSL state machine POC
  - Requires Go 1.21+

### Build Methods

| Method | Command | Notes |
|--------|---------|-------|
| Direct | `GOEXPERIMENT=greenteagc go build -o dsl-poc .` | Manual control |
| Script | `./build.sh` | Supports options, colored output |
| Make | `make build-greenteagc` | Standardized, includes dependencies |

## Troubleshooting

**Error: "DB_CONN_STRING environment variable is not set"**
```sh
export DB_CONN_STRING="postgres://user:password@localhost:5432/your_db?sslmode=disable"
```

**Error: "failed to connect to database"**
- Ensure PostgreSQL is running
- Verify connection string is correct
- Check that the database user has permissions to create schemas and tables

**Build fails with unknown GOEXPERIMENT**
- Ensure you're using Go 1.21 or later
- Check: `go version`
```
