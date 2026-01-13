package storage

import (
	"sync"

	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"
)

// Storage contains all in-memory data structures used by the application.
// This file only provides the data types and a constructor; it does not
// implement any business logic (as requested).
type Storage struct {
	mu sync.RWMutex

	// auto-increment counters
	nextUserID    int64
	nextTopicID   int64
	nextMessageID int64
	nextEventSeq  int64

	// primary data stores
	Users           map[int64]*pb.User
	Topics          map[int64]*pb.Topic
	Messages        map[int64]*pb.Message   // message id -> Message
	MessagesByTopic map[int64][]*pb.Message // topic id -> ordered slice of Messages
}

// NewStorage initializes and returns an empty Storage instance.
func NewStorage() *Storage {
	return &Storage{
		Users:           make(map[int64]*pb.User),
		Topics:          make(map[int64]*pb.Topic),
		Messages:        make(map[int64]*pb.Message),
		MessagesByTopic: make(map[int64][]*pb.Message),
	}
}
