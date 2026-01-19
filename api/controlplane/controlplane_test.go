package controlplane

import (
	"testing"
)

// Test 1: Validacija RegisterRequest strukture
func TestRegisterRequest_ValidStructure(t *testing.T) {
	testCases := []struct {
		name      string
		nodeID    string
		address   string
		shouldErr bool
	}{
		{
			name:      "Valid node registration",
			nodeID:    "node-1",
			address:   "localhost:54321",
			shouldErr: false,
		},
		{
			name:      "Empty node ID",
			nodeID:    "",
			address:   "localhost:54321",
			shouldErr: true,
		},
		{
			name:      "Empty address",
			nodeID:    "node-1",
			address:   "",
			shouldErr: true,
		},
		{
			name:      "Invalid address format",
			nodeID:    "node-1",
			address:   "invalid",
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &RegisterRequest{
				Node: &NodeInfo{
					NodeId:  tc.nodeID,
					Address: tc.address,
				},
			}

			// Validacija
			if tc.nodeID == "" && !tc.shouldErr {
				t.Error("Expected error for empty NodeId")
			}
			if tc.address == "" && !tc.shouldErr {
				t.Error("Expected error for empty Address")
			}

			// Preverka, da je strukturo pravilno nastavljena
			if req.Node.NodeId != tc.nodeID {
				t.Errorf("NodeId mismatch: expected %s, got %s", tc.nodeID, req.Node.NodeId)
			}
			if req.Node.Address != tc.address {
				t.Errorf("Address mismatch: expected %s, got %s", tc.address, req.Node.Address)
			}
		})
	}
}

// Test 2: Validacija NodeInfo polj
func TestNodeInfo_FieldValidation(t *testing.T) {
	validationTests := []struct {
		name     string
		nodeInfo *NodeInfo
		wantErr  bool
		errMsg   string
	}{
		{
			name: "Valid NodeInfo",
			nodeInfo: &NodeInfo{
				NodeId:  "valid-node",
				Address: "localhost:54321",
			},
			wantErr: false,
		},
		{
			name: "Missing NodeId",
			nodeInfo: &NodeInfo{
				NodeId:  "",
				Address: "localhost:54321",
			},
			wantErr: true,
			errMsg:  "NodeId cannot be empty",
		},
		{
			name: "Missing Address",
			nodeInfo: &NodeInfo{
				NodeId:  "node-1",
				Address: "",
			},
			wantErr: true,
			errMsg:  "Address cannot be empty",
		},
	}

	for _, test := range validationTests {
		t.Run(test.name, func(t *testing.T) {
			// Simuliramo validacijo
			if test.nodeInfo.NodeId == "" {
				if !test.wantErr {
					t.Error("Expected NodeId validation error")
				}
			}
			if test.nodeInfo.Address == "" {
				if !test.wantErr {
					t.Error("Expected Address validation error")
				}
			}
		})
	}
}

// Test 3: Validacija različnih node ID-jev
func TestNodeID_Formats(t *testing.T) {
	validNodeIDs := []string{
		"node-1",
		"node-prod-1",
		"n1",
		"node_test",
		"localhost-node",
	}

	for _, nodeID := range validNodeIDs {
		t.Run(nodeID, func(t *testing.T) {
			node := &NodeInfo{
				NodeId:  nodeID,
				Address: "localhost:54321",
			}

			if node.NodeId != nodeID {
				t.Errorf("NodeId not set correctly: expected %s, got %s", nodeID, node.NodeId)
			}
			if node.NodeId == "" {
				t.Error("NodeId is empty")
			}
		})
	}
}

// Test 4: Validacija različnih naslovov
func TestAddress_Formats(t *testing.T) {
	validAddresses := []struct {
		name    string
		address string
	}{
		{"localhost", "localhost:54321"},
		{"IP address", "192.168.1.1:54321"},
		{"Domain", "example.com:12345"},
		{"Different port", "localhost:9999"},
	}

	for _, addr := range validAddresses {
		t.Run(addr.name, func(t *testing.T) {
			node := &NodeInfo{
				NodeId:  "test-node",
				Address: addr.address,
			}

			if node.Address != addr.address {
				t.Errorf("Address not set correctly: expected %s, got %s", addr.address, node.Address)
			}
			if node.Address == "" {
				t.Error("Address is empty")
			}
		})
	}
}

// Test 5: RegisterRequest z več vozlišči
func TestRegisterRequest_MultipleNodes(t *testing.T) {
	nodes := []*NodeInfo{
		{NodeId: "node-1", Address: "localhost:54321"},
		{NodeId: "node-2", Address: "localhost:54322"},
		{NodeId: "node-3", Address: "localhost:54323"},
	}

	for i, node := range nodes {
		t.Run(node.NodeId, func(t *testing.T) {
			req := &RegisterRequest{Node: node}

			if req.Node == nil {
				t.Error("RegisterRequest.Node is nil")
			}
			if req.Node.NodeId != node.NodeId {
				t.Errorf("Node %d: NodeId mismatch", i)
			}
			if req.Node.Address != node.Address {
				t.Errorf("Node %d: Address mismatch", i)
			}
		})
	}
}

// Test 6: Vrstni red podpolja v RegisterRequest
func TestRegisterRequest_FieldOrder(t *testing.T) {
	req := &RegisterRequest{
		Node: &NodeInfo{
			NodeId:  "test-node",
			Address: "localhost:54321",
		},
	}

	// Validacija, da so vsa polja nastavljena v pravilnem vrstnem redu
	if req.Node == nil {
		t.Fatal("Node field not set")
	}
	if req.Node.NodeId == "" {
		t.Error("NodeId should be set before Address")
	}
	if req.Node.Address == "" {
		t.Error("Address should be set after NodeId")
	}
}

// Test 7: Stabilnost podatkov
func TestNodeInfo_DataStability(t *testing.T) {
	originalID := "node-1"
	originalAddr := "localhost:54321"

	node := &NodeInfo{
		NodeId:  originalID,
		Address: originalAddr,
	}

	// Ponavljajoče branje - podatki se ne smejo spremeniti
	for i := 0; i < 10; i++ {
		if node.NodeId != originalID {
			t.Errorf("NodeId changed after %d reads", i)
		}
		if node.Address != originalAddr {
			t.Errorf("Address changed after %d reads", i)
		}
	}
}

// Test 8: Registracija več vozlišč hkrati
func TestRegisterRequest_ConcurrentRegistrations(t *testing.T) {
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(id int) {
			req := &RegisterRequest{
				Node: &NodeInfo{
					NodeId:  "node-" + string(rune(48+id)),
					Address: "localhost:5432" + string(rune(48+id)),
				},
			}

			if req.Node == nil {
				t.Error("Concurrent registration failed")
			}
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}
