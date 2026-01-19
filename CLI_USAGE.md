# Razpravljalnica CLI Commands

This project now uses **Cobra** CLI framework for all command-line interfaces. Each component (client, controlplane, node) has been refactored to provide better command-line experience with improved flag handling and help documentation.

## Client CLI

**Binary:** `client`


### Usage
```bash
client [flags]
```

### Flags
- `-c, --controlplane string` - Control Plane address (host:port) (default: "localhost:12345")
- `-h, --help` - Help for client

### Examples
```bash
# Connect to local control plane
./client

# Connect to remote control plane
./client --controlplane 192.168.1.100:12345

# Short form
./client -c example.com:12345

# View help
./client --help
```

---

## Control Plane CLI

**Binary:** `controlplane`

### Description
The control plane server for the Razpravljalnica distributed message board system. Manages node registration, chain state, and provides discovery services.

### Usage
```bash
controlplane [flags]
```

### Flags
- `-p, --port string` - Port to listen on (default: "12345")
- `-h, --help` - Help for controlplane

### Examples
```bash
# Run on default port 12345
./controlplane

# Run on custom port
./controlplane --port 8080

# Short form
./controlplane -p 9999

# View help
./controlplane --help
```

---

## Node CLI

**Binary:** `node`

### Description
A node in the Razpravljalnica distributed message board system. Stores messages, handles replication, and participates in the chain.

### Usage
```bash
node [flags]
```

### Flags
- `-p, --port string` - Port to listen on (default: "54321")
- `-c, --controlplane string` - Control Plane address (host:port) (default: "localhost:12345")
- `-i, --id string` - Unique ID for this node (default: "node-1")
- `-h, --help` - Help for node

### Examples
```bash
# Run with default settings
./node

# Specify all parameters
./node --port 54322 --controlplane localhost:12345 --id node-2

# Short form
./node -p 54322 -c localhost:12345 -i node-2

# Join specific control plane
./node -c remote-cp.example.com:12345 -i node-prod-1

# View help
./node --help
```

---

## Build and Run Instructions

### Build all binaries
```bash
cd razpravljalnica
go build -o client.exe ./cmd/client
go build -o controlplane.exe ./cmd/controlplane
go build -o node.exe ./cmd/node
```
### Or run 
```bash
./build.bat
```

### Run sequence (example)
```bash
# Terminal 1: Start control plane
./controlplane --port 12345

# Terminal 2: Start node 1
./node --port 54321 --id node-1 --controlplane localhost:12345

# Terminal 3: Start node 2
./node --port 54322 --id node-2 --controlplane localhost:12345

# Terminal 4: Start client
./client --controlplane localhost:12345
```

---

