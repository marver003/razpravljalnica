package subscription

import (
	"sync"

	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"
)

// TODO: naročnine, broadcastanje sporočil in lajkov

type Manager struct {
	mu   sync.RWMutex
	subs map[int64][]chan *pb.MessageEvent
}

func NewManager() *Manager {
	return &Manager{
		subs: make(map[int64][]chan *pb.MessageEvent),
	}
}

func (m *Manager) Subscribe(topicID int64) chan *pb.MessageEvent {
	ch := make(chan *pb.MessageEvent, 16)

	m.mu.Lock()
	m.subs[topicID] = append(m.subs[topicID], ch)
	m.mu.Unlock()

	return ch
}

func (m *Manager) Unsubscribe(topicID int64, ch chan *pb.MessageEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()

	subs := m.subs[topicID]
	for i, c := range subs {
		if c == ch {
			m.subs[topicID] = append(subs[:i], subs[i+1:]...)
			close(c)
			break
		}
	}
}

func (m *Manager) Publish(topicID int64, ev *pb.MessageEvent) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, ch := range m.subs[topicID] {
		select {
		case ch <- ev:
		default:
			// drop event for slow subscriber
		}
	}
}
