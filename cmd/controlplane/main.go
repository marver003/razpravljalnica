package main

import (
	"flag"
	"log"
	"net"

	cp "github.com/marver003/razpravljalnica/api/controlplane"
	"github.com/marver003/razpravljalnica/internal/server"
	"google.golang.org/grpc"
)

func main() {
	port := flag.String("port", "8080", "Control Plane port")
	flag.Parse()

	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	cp.RegisterControlPlaneServer(s, server.NewControlPlane())

	log.Printf("Control Plane listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
