@echo off
REM filepath: c:\Users\nejc\Desktop\faks\VPSA\projekt\razpravljalnica\build.bat

echo Building binaries...
go build -o client.exe ./cmd/client
go build -o controlplane.exe ./cmd/controlplane
go build -o node.exe ./cmd/node
echo All binaries built successfully!

setlocal enabledelayedexpansion

REM Build test runner
echo.
echo Building test runner...
go build -o test.exe .\cmd\test
echo Test runner built successfully
