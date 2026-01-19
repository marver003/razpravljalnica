package main

import (
	"log"
	"net"

	cp "github.com/marver003/razpravljalnica/api/controlplane"
	"github.com/marver003/razpravljalnica/cmd/controlplane/ui"
	"github.com/marver003/razpravljalnica/internal/server"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var port string

var rootCmd = &cobra.Command{
	Use:   "controlplane",
	Short: "Razpravljalnica Control Plane Server",
	Long: `The control plane server for the Razpravljalnica distributed message board system.
Manages node registration, chain state, and provides discovery services.`,
	Run: runControlPlane,
}

func init() {
	rootCmd.Flags().StringVarP(&port, "port", "p", "12345", "Port to listen on")
}

func runControlPlane(cmd *cobra.Command, args []string) {
	cpServer := server.NewControlPlane()

	// run gRPC server in a goroutine
	go func() {
		lis, err := net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		s := grpc.NewServer()
		cp.RegisterControlPlaneServer(s, cpServer)

		log.Printf("Control Plane gRPC server listening on port %s", port)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// run TUI on the main thread
	ui.Start(cpServer)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
