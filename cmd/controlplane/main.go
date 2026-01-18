package main

import (
	"flag"
	"log"
	"net"

	cp "github.com/marver003/razpravljalnica/api/controlplane"
	"github.com/marver003/razpravljalnica/cmd/controlplane/ui"
	"github.com/marver003/razpravljalnica/internal/server"
	"google.golang.org/grpc"
)

func main() {
	port := flag.String("port", "12345", "Control Plane port")
	flag.Parse()

	cpServer := server.NewControlPlane()

	// run gRPC server in a goroutine
	go func() {
		lis, err := net.Listen("tcp", ":"+*port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		s := grpc.NewServer()
		cp.RegisterControlPlaneServer(s, cpServer)
		
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// run TUI on the main thread
	ui.Start(cpServer)
}
