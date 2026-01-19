package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "all":
		runAllTests()
	case "api":
		runAPITests()
	case "internal":
		runInternalTests()
	case "verbose":
		runAllTestsVerbose()
	case "coverage":
		runTestsWithCoverage()
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`
Razpravljalnica Test Runner

Usage:
  test <command>

Commands:
  all         Run all tests (API + Internal)
  api         Run API tests only (controlplane, razpravljalnica, replication)
  internal    Run internal tests only (server, storage, subscription)
  verbose     Run all tests with verbose output
  coverage    Run all tests with coverage report
  help        Show this help message

Examples:
  test all
  test api
  test internal
  test verbose
  test coverage
`)
}

func runAllTests() {
	fmt.Println("=== Running All Tests (API + Internal) ===\n")
	packages := []string{
		"./api/controlplane",
		"./api/razpravljalnica",
		"./api/replication",
		"./internal/server",
		"./internal/storage",
		"./internal/subscription",
	}

	for _, pkg := range packages {
		fmt.Printf("\n--- Testing %s ---\n", pkg)
		cmd := exec.Command("go", "test", pkg)
		runCommand(cmd)
	}
}

func runAPITests() {
	fmt.Println("=== Running API Tests ===\n")
	packages := []string{
		"./api/controlplane",
		"./api/razpravljalnica",
		"./api/replication",
	}

	for _, pkg := range packages {
		fmt.Printf("\n--- Testing %s ---\n", pkg)
		cmd := exec.Command("go", "test", "-v", pkg)
		runCommand(cmd)
	}
}

func runInternalTests() {
	fmt.Println("=== Running Internal Tests ===\n")
	packages := []string{
		"./internal/server",
		"./internal/storage",
		"./internal/subscription",
	}

	for _, pkg := range packages {
		fmt.Printf("\n--- Testing %s ---\n", pkg)
		cmd := exec.Command("go", "test", "-v", pkg)
		runCommand(cmd)
	}
}

func runAllTestsVerbose() {
	fmt.Println("=== Running All Tests (Verbose) ===\n")
	packages := []string{
		"./api/controlplane",
		"./api/razpravljalnica",
		"./api/replication",
		"./internal/server",
		"./internal/storage",
		"./internal/subscription",
	}

	for _, pkg := range packages {
		fmt.Printf("\n--- Testing %s ---\n", pkg)
		cmd := exec.Command("go", "test", "-v", pkg)
		runCommand(cmd)
	}
}

func runTestsWithCoverage() {
	fmt.Println("=== Running All Tests with Coverage ===\n")

	packages := []string{
		"./api/controlplane",
		"./api/razpravljalnica",
		"./api/replication",
		"./internal/server",
		"./internal/storage",
		"./internal/subscription",
	}

	count := 0

	for _, pkg := range packages {
		fmt.Printf("\n--- Coverage for %s ---\n", pkg)
		cmd := exec.Command("go", "test", "-v", "-cover", pkg)
		output := runCommandGetOutput(cmd)

		// Parse coverage percentage
		if strings.Contains(output, "coverage:") {
			parts := strings.Split(output, "coverage: ")
			if len(parts) > 1 {
				coverStr := strings.Fields(parts[1])[0]
				fmt.Printf("Coverage: %s\n", coverStr)
				count++
			}
		}
	}

	fmt.Printf("\n=== Total Packages Tested: %d ===\n", count)
}

func runCommand(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Printf("Tests failed with exit code: %d\n", exitErr.ExitCode())
		} else {
			fmt.Printf("Error running tests: %v\n", err)
		}
	}
}

func runCommandGetOutput(cmd *exec.Cmd) string {
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	return string(output)
}
