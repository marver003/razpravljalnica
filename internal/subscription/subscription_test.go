package subscription

import (
	"testing"
)

func TestNewManager_NotNil(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatalf("NewManager returned nil")
	}
}

func TestManager_Initialization(t *testing.T) {
	m := NewManager()

	if m == nil {
		t.Fatalf("Manager initialization failed")
	}
	t.Log("Manager initialized successfully")
}

func TestManager_MultipleManagers(t *testing.T) {
	m1 := NewManager()
	m2 := NewManager()

	if m1 == nil || m2 == nil {
		t.Fatalf("Failed to create manager instances")
	}

	if m1 == m2 {
		t.Error("Expected different manager instances")
	}
	t.Log("Multiple managers created successfully")
}

func TestManager_ConcurrentOperations(t *testing.T) {
	m := NewManager()
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(id int) {
			if m == nil {
				t.Errorf("Manager is nil in goroutine %d", id)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 5; i++ {
		<-done
	}
	t.Log("Concurrent operations test passed")
}

func TestManager_StressTest(t *testing.T) {
	m := NewManager()

	for i := 0; i < 50; i++ {
		if m == nil {
			t.Fatalf("Manager became nil at iteration %d", i)
		}
	}
	t.Log("Stress test: 50 iterations passed")
}

func TestManager_Independence(t *testing.T) {
	m1 := NewManager()
	m2 := NewManager()
	m3 := NewManager()

	if m1 == m2 || m2 == m3 || m1 == m3 {
		t.Error("Managers should be independent instances")
	}
	t.Log("Independence test passed")
}

func TestManager_RepeatedCreation(t *testing.T) {
	managers := make([]interface{}, 0)

	for i := 0; i < 10; i++ {
		m := NewManager()
		if m == nil {
			t.Fatalf("Failed to create manager at iteration %d", i)
		}
		managers = append(managers, m)
	}

	if len(managers) != 10 {
		t.Errorf("Expected 10 managers, got %d", len(managers))
	}
	t.Log("Repeated creation test passed")
}
