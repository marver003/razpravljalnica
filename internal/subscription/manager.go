package subscription

import (
	"log"
	"sync"

	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"
)

// TODO: dodat treba balancer

type Subscriber struct {
	UserID   int64
	TopicIDs map[int64]bool
	Stream   pb.MessageBoard_SubscribeTopicServer
}

type Manager struct {
	mu          sync.Mutex
	subscribers map[int64]*Subscriber // userID -> subscriber
}

func NewManager() *Manager {
	return &Manager{
		subscribers: make(map[int64]*Subscriber),
	}
}

func (m *Manager) Add(sub *Subscriber) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.subscribers[sub.UserID] = sub
}

func (m *Manager) Remove(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.subscribers, userID)
}

func (m *Manager) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.subscribers)
}

func (m *Manager) Broadcast(event *pb.MessageEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, sub := range m.subscribers {
		if sub.TopicIDs[event.Message.TopicId] {
			err := sub.Stream.Send(event)
			if err != nil {
				log.Printf("Failed to send to subscriber: %v", err)
			} else {
				log.Printf("Successfully broadcasted message %d to a sub", event.Message.Id)
			}
		}
	}
}
