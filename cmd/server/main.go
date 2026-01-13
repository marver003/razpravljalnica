package main

import (
	"flag"
	"fmt"
	"net"

	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"
	"github.com/marver003/razpravljalnica/internal/server"
	"github.com/marver003/razpravljalnica/internal/storage"
	"github.com/marver003/razpravljalnica/internal/subscription"

	"google.golang.org/grpc"
)

// gRPC server, tuki se registrira /internal/server/MessageBoard.go in začne poslušat na localhost:<port>, default port je 12345
func main() {
	portPtr := flag.Int("p", 12345, "port number")

	flag.Parse()

	if *portPtr < 1024 || *portPtr > 65535 {
		panic("Port-Range Error: port range 1024-65535")
	}

	addr := fmt.Sprintf("localhost:%d", *portPtr)

	grpcServer := grpc.NewServer()

	store := storage.NewStorage()
	subs := subscription.NewManager()

	srv := server.NewServerMessageBoard(store, subs)

	pb.RegisterMessageBoardServer(grpcServer, srv)

	// hostname, err := os.Hostname()
	// if err != nil {
	// 	panic(err)
	// }

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	fmt.Printf("gRPC server listening at %s\n", addr)

	if err := grpcServer.Serve(listener); err != nil {
		panic(err)
	}

}
