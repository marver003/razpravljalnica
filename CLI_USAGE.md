# Razpravljalnica CLI Commands

This project uses the **Cobra** CLI framework for all command-line interfaces. Each component (client, controlplane, node) provides a consistent command-line experience with structured flags and built-in help.

In addition, the project includes a **custom Go test runner** for running and organizing tests across API and internal packages.

---

## Client CLI

**Binary:** `client`

### Usage

```bash
client [flags]
```

### Flags

* `-c, --controlplane string` – Control Plane address (`host:port`)
  *(default: `localhost:12345`)*
* `-h, --help` – Show help

### Examples

```bash
./client
./client --controlplane 192.168.1.100:12345
./client -c example.com:12345
./client --help
```

---

## Control Plane CLI

**Binary:** `controlplane`

### Description

The control plane server for the Razpravljalnica distributed message board system.
Manages node registration, chain state, and discovery services.

### Usage

```bash
controlplane [flags]
```

### Flags

* `-p, --port string` – Port to listen on
  *(default: `12345`)*
* `-h, --help` – Show help

### Examples

```bash
./controlplane
./controlplane --port 8080
./controlplane -p 9999
./controlplane --help
```

---

## Node CLI

**Binary:** `node`

### Description

A node in the Razpravljalnica distributed message board system.
Stores messages, handles replication, and participates in the chain.

### Usage

```bash
node [flags]
```

### Flags

* `-p, --port string` – Port to listen on *(default: `54321`)*
* `-c, --controlplane string` – Control Plane address *(default: `localhost:12345`)*
* `-i, --id string` – Unique node ID *(default: `node-1`)*
* `-h, --help` – Show help

### Examples

```bash
./node
./node --port 54322 --controlplane localhost:12345 --id node-2
./node -p 54322 -c localhost:12345 -i node-2
./node --help
```

---

## Build Instructions

### Build all binaries

```bash
./build.bat
```

---

## Run Sequence (Example)

```bash
# Terminal 1: Control Plane
./controlplane --port 12345

# Terminal 2: Node 1
./node --port 54321 --id node-1 --controlplane localhost:12345

# Terminal 3: Node 2
./node --port 54322 --id node-2 --controlplane localhost:12345

# Terminal 4: Client
./client --controlplane localhost:12345
```

---

# Testing

The project includes a **custom Go test runner** (`test`) that provides a convenient interface for running tests across different parts of the codebase.

The test runner wraps `go test` and supports grouped execution, verbose output, and coverage reporting.

---

## Test Runner Usage

```bash
test <command>
```

### Available Commands

| Command    | Description                                                               |
| ---------- | ------------------------------------------------------------------------- |
| `all`      | Run **all tests** (API + internal packages)                               |
| `api`      | Run **API tests only** (`controlplane`, `razpravljalnica`, `replication`) |
| `internal` | Run **internal tests only** (`server`, `storage`, `subscription`)         |
| `verbose`  | Run all tests with **verbose output**                                     |
| `coverage` | Run all tests with **coverage report**                                    |
| `help`     | Show test runner help                                                     |

---

## Test Runner Examples

### Run all tests

```bash
test all
```

### Run only API tests

```bash
test api
```

### Run only internal tests

```bash
test internal
```

### Run all tests with verbose output

```bash
test verbose
```

### Run tests with coverage

```bash
test coverage
```

This command runs `go test -cover` for each package and prints the **statement coverage percentage**, indicating how much of the code was executed during testing.

---

## Coverage Notes

* Coverage is reported **per package**
* Coverage measures **executed Go statements**, not test quality
* A higher percentage indicates broader code execution but does **not guarantee correctness**
