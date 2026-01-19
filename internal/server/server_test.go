package server_test

import (
	"testing"

	"github.com/marver003/razpravljalnica/internal/server"
	"github.com/marver003/razpravljalnica/internal/storage"
	"github.com/marver003/razpravljalnica/internal/subscription"
)

func TestNewNode_Constructs(t *testing.T) {
	store := storage.NewStorage()
	sub := subscription.NewManager()

	n := server.NewNode(store, "test-node", "localhost:0", sub)
	if n == nil {
		t.Fatalf("NewNode returned nil")
	}
}

func TestNode_DifferentNodeIDs(t *testing.T) {
	store := storage.NewStorage()
	sub := subscription.NewManager()

	nodeIDs := []string{"node-1", "node-2", "node-3"}

	for _, id := range nodeIDs {
		t.Run(id, func(t *testing.T) {
			n := server.NewNode(store, id, "localhost:54321", sub)
			if n == nil {
				t.Fatalf("Failed to create node with ID %s", id)
			}
		})
	}
}

func TestNode_DifferentAddresses(t *testing.T) {
	store := storage.NewStorage()
	sub := subscription.NewManager()

	addresses := []string{
		"localhost:54321",
		"localhost:54322",
		"localhost:54323",
	}

	for _, addr := range addresses {
		t.Run(addr, func(t *testing.T) {
			n := server.NewNode(store, "test-node", addr, sub)
			if n == nil {
				t.Fatalf("Failed to create node with address %s", addr)
			}
		})
	}
}

func TestNode_WithDifferentStorages(t *testing.T) {
	sub := subscription.NewManager()

	for i := 0; i < 3; i++ {
		store := storage.NewStorage()
		n := server.NewNode(store, "test-node", "localhost:0", sub)
		if n == nil {
			t.Fatalf("Failed to create node with storage %d", i)
		}
	}
}

func TestNode_MultipleNodes(t *testing.T) {
	store := storage.NewStorage()
	sub := subscription.NewManager()

	nodes := make([]interface{}, 5)
	for i := 0; i < 5; i++ {
		nodes[i] = server.NewNode(store, "node", "localhost:0", sub)
		if nodes[i] == nil {
			t.Fatalf("Failed to create node %d", i)
		}
	}

	if len(nodes) != 5 {
		t.Errorf("Expected 5 nodes, got %d", len(nodes))
	}
}

func TestNode_Initialization(t *testing.T) {
	store := storage.NewStorage()
	sub := subscription.NewManager()

	n := server.NewNode(store, "init-test-node", "localhost:54321", sub)
	if n == nil {
		t.Fatalf("Node initialization failed")
	}
	t.Log("Node initialized successfully")
}

func TestNode_StressTestCreation(t *testing.T) {
	store := storage.NewStorage()
	sub := subscription.NewManager()

	for i := 0; i < 20; i++ {
		n := server.NewNode(store, "stress-test-node", "localhost:0", sub)
		if n == nil {
			t.Fatalf("Failed to create node at iteration %d", i)
		}
	}
	t.Log("Stress test: 20 node creations passed")
}

func TestNode_ConcurrentCreation(t *testing.T) {
	store := storage.NewStorage()
	sub := subscription.NewManager()
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			n := server.NewNode(store, "concurrent-node", "localhost:0", sub)
			if n == nil {
				t.Errorf("Failed to create node in goroutine %d", id)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
	t.Log("Concurrent node creation test passed")
}
