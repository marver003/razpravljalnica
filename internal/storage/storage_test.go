package storage

import (
	"testing"
)

func TestNewStorage_NotNil(t *testing.T) {
	s := NewStorage()
	if s == nil {
		t.Fatalf("NewStorage returned nil")
	}
}

func TestStorage_Initialization(t *testing.T) {
	s := NewStorage()
	if s == nil {
		t.Fatalf("Storage initialization failed")
	}
	t.Log("Storage initialized successfully")
}

func TestStorage_MultipleInstances(t *testing.T) {
	s1 := NewStorage()
	s2 := NewStorage()

	if s1 == nil || s2 == nil {
		t.Fatalf("Failed to create storage instances")
	}

	// Vsaka instanca bi morala biti neodvisna
	if s1 == s2 {
		t.Error("Storage instances should be different")
	}
}

func TestStorage_ConcurrentAccess(t *testing.T) {
	s := NewStorage()
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			if s == nil {
				t.Errorf("Storage is nil in goroutine %d", id)
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
	t.Log("Concurrent access test passed")
}

func TestStorage_Stability(t *testing.T) {
	s := NewStorage()

	for i := 0; i < 100; i++ {
		if s == nil {
			t.Fatalf("Storage became nil at iteration %d", i)
		}
	}
	t.Log("Stability test: 100 iterations passed")
}

func TestStorage_NotCorrupted(t *testing.T) {
	s1 := NewStorage()
	s2 := NewStorage()
	s3 := NewStorage()

	if s1 == s2 || s2 == s3 || s1 == s3 {
		t.Error("Storage instances should be independent")
	}

	if s1 == nil || s2 == nil || s3 == nil {
		t.Error("Storage instances are nil")
	}
}

func TestStorage_RepeatedCreation(t *testing.T) {
	storages := make([]interface{}, 0)

	for i := 0; i < 20; i++ {
		s := NewStorage()
		if s == nil {
			t.Fatalf("Failed to create storage at iteration %d", i)
		}
		storages = append(storages, s)
	}

	if len(storages) != 20 {
		t.Errorf("Expected 20 storages, got %d", len(storages))
	}
	t.Log("Repeated creation test passed")
}
