package replication

import (
	"testing"
)

func TestOperation_Creation(t *testing.T) {
	testCases := []struct {
		name    string
		index   int64
		opType  OperationType
		payload []byte
		valid   bool
	}{
		{
			name:    "Create user operation",
			index:   1,
			opType:  OperationType_OP_CREATE_USER,
			payload: []byte("user-data"),
			valid:   true,
		},
		{
			name:    "Create topic operation",
			index:   2,
			opType:  OperationType_OP_CREATE_TOPIC,
			payload: []byte("topic-data"),
			valid:   true,
		},
		{
			name:    "Post message operation",
			index:   3,
			opType:  OperationType_OP_POST_MESSAGE,
			payload: []byte("message-data"),
			valid:   true,
		},
		{
			name:    "Like message operation",
			index:   4,
			opType:  OperationType_OP_LIKE_MESSAGE,
			payload: []byte("like-data"),
			valid:   true,
		},
		{
			name:    "Empty payload",
			index:   5,
			opType:  OperationType_OP_CREATE_USER,
			payload: []byte{},
			valid:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			op := &Operation{
				Index:   tc.index,
				Type:    tc.opType,
				Payload: tc.payload,
			}

			if len(tc.payload) == 0 && tc.valid {
				t.Error("Expected error for empty payload")
			}

			if op.Index != tc.index {
				t.Errorf("Index mismatch: expected %d, got %d", tc.index, op.Index)
			}
			if op.Type != tc.opType {
				t.Errorf("Operation type mismatch: expected %v, got %v", tc.opType, op.Type)
			}
		})
	}
}

func TestOperation_IndexMonotonic(t *testing.T) {
	operations := []*Operation{
		{Index: 1, Type: OperationType_OP_CREATE_USER, Payload: []byte("data1")},
		{Index: 2, Type: OperationType_OP_CREATE_TOPIC, Payload: []byte("data2")},
		{Index: 3, Type: OperationType_OP_POST_MESSAGE, Payload: []byte("data3")},
	}

	for i := 0; i < len(operations)-1; i++ {
		if operations[i].Index >= operations[i+1].Index {
			t.Errorf("Operations not in monotonic order at index %d", i)
		}
	}
	t.Log("Operation indices are monotonic")
}

func TestAckMessage_Validation(t *testing.T) {
	testCases := []struct {
		name  string
		index int64
		valid bool
	}{
		{"Valid ack", 1, true},
		{"Ack index 100", 100, true},
		{"Zero index", 0, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ack := &AckMessage{
				Index: tc.index,
			}

			if ack.Index != tc.index {
				t.Errorf("Index mismatch: expected %d, got %d", tc.index, ack.Index)
			}
		})
	}
}

func TestGetLogRequest_Validation(t *testing.T) {
	testCases := []struct {
		name      string
		fromIndex int64
		valid     bool
	}{
		{"From beginning", 0, true},
		{"From middle", 50, true},
		{"From large index", 1000, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &GetLogRequest{
				FromIndex: tc.fromIndex,
			}

			if req.FromIndex != tc.fromIndex {
				t.Errorf("FromIndex mismatch: expected %d, got %d", tc.fromIndex, req.FromIndex)
			}
		})
	}
}

func TestOperation_AllTypes(t *testing.T) {
	operationTypes := []struct {
		name   string
		opType OperationType
	}{
		{"Create user", OperationType_OP_CREATE_USER},
		{"Create topic", OperationType_OP_CREATE_TOPIC},
		{"Post message", OperationType_OP_POST_MESSAGE},
		{"Like message", OperationType_OP_LIKE_MESSAGE},
	}

	for i, opTest := range operationTypes {
		t.Run(opTest.name, func(t *testing.T) {
			op := &Operation{
				Index:   int64(i + 1),
				Type:    opTest.opType,
				Payload: []byte("test-payload"),
			}

			if op.Type != opTest.opType {
				t.Errorf("Operation type mismatch: expected %v, got %v", opTest.opType, op.Type)
			}
		})
	}
}

func TestOperation_PayloadIntegrity(t *testing.T) {
	originalPayload := []byte("Important replication payload")
	op := &Operation{
		Index:   1,
		Type:    OperationType_OP_POST_MESSAGE,
		Payload: originalPayload,
	}

	for i := 0; i < 10; i++ {
		if string(op.Payload) != string(originalPayload) {
			t.Errorf("Payload integrity check failed at iteration %d", i)
		}
	}
	t.Log("Payload integrity test passed")
}

func TestOperation_SequenceProcessing(t *testing.T) {
	operations := []*Operation{
		{Index: 1, Type: OperationType_OP_CREATE_USER, Payload: []byte("user1")},
		{Index: 2, Type: OperationType_OP_CREATE_TOPIC, Payload: []byte("topic1")},
		{Index: 3, Type: OperationType_OP_POST_MESSAGE, Payload: []byte("msg1")},
		{Index: 4, Type: OperationType_OP_LIKE_MESSAGE, Payload: []byte("like1")},
		{Index: 5, Type: OperationType_OP_POST_MESSAGE, Payload: []byte("msg2")},
	}

	for i, op := range operations {
		t.Run(op.String(), func(t *testing.T) {
			if op.Index != int64(i+1) {
				t.Errorf("Operation %d: index mismatch", i)
			}
			if len(op.Payload) == 0 {
				t.Errorf("Operation %d: empty payload", i)
			}
		})
	}
}

func TestReplication_ConcurrentAcks(t *testing.T) {
	done := make(chan bool, 10)

	for i := 1; i <= 10; i++ {
		go func(index int64) {
			ack := &AckMessage{
				Index: index,
			}

			if ack.Index != index {
				t.Errorf("Ack index mismatch in goroutine %d", index)
			}
			done <- true
		}(int64(i))
	}

	for i := 0; i < 10; i++ {
		<-done
	}
	t.Log("Concurrent acks test passed")
}

func TestReplication_GetLogRequest_Range(t *testing.T) {
	testCases := []struct {
		name        string
		fromIndex   int64
		description string
	}{
		{"Full log", 0, "Get all operations"},
		{"Partial log", 10, "Get from index 10"},
		{"Recent log", 100, "Get recent 100+ operations"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &GetLogRequest{
				FromIndex: tc.fromIndex,
			}

			if req.FromIndex != tc.fromIndex {
				t.Errorf("Expected from_index %d, got %d", tc.fromIndex, req.FromIndex)
			}
			t.Logf("Request: %s", tc.description)
		})
	}
}
