@echo off
go build -o client.exe ./cmd/client
go build -o controlplane.exe ./cmd/controlplane
go build -o node.exe ./cmd/node
echo All binaries built successfully!