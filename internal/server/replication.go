package server

import (
	"context"
	"log"

	repl "github.com/marver003/razpravljalnica/api/replication"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (n *Node) Replicate(ctx context.Context, op *repl.Operation) (*repl.AckMessage, error) {
	n.mu.RLock()
	next := n.nextInChain
	n.mu.RUnlock()

	// forward to next node if not tail
	if next != nil {
		_, err := next.Replicate(ctx, op)
		if err != nil {
			log.Printf("Replication to next node failed: %v. Waiting for reconfiguration...", err)
			return nil, status.Error(codes.Unavailable, "Chain broken, try again in a moment")
		}
	}

	// apply locally to storage
	n.storage.Apply(op)

	n.broadcastOperation(op)

	return &repl.AckMessage{Index: op.Index}, nil
}

func (n *Node) GetLogFrom(req *repl.GetLogRequest, stream repl.Replication_GetLogFromServer) error {
	ops := n.storage.GetOperationsFrom(req.FromIndex)
	for _, op := range ops {
		if err := stream.Send(op); err != nil {
			return err
		}
	}
	return nil
}
