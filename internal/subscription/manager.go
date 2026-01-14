package subscription

import (
	"sync"

	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"
)

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

func (m *Manager) Broadcast(event *pb.MessageEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, sub := range m.subscribers {
		if sub.TopicIDs[event.Message.TopicId] {
			_ = sub.Stream.Send(event)
		}
	}
}
