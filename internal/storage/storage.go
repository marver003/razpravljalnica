package storage

import (
	"sync"
	"time"
)

type User struct {
	Id   int64
	Name string
}

type Topic struct {
	Id   int64
	Name string
}

type Message struct {
	Id        int64
	TopicId   int64
	UserId    int64
	Text      string
	CreatedAt time.Time
	Likes     int32
}

type Storage struct {
	mu sync.RWMutex

	// auto-increment counters
	nextUserID    int64
	nextTopicID   int64
	nextMessageID int64

	// data
	users    map[int64]*User
	topics   map[int64]*Topic
	messages map[int64]*Message

	// topic_id -> ordered messages
	messagesByTopic map[int64][]*Message
}

func NewStorage() *Storage {
	return &Storage{
		nextUserID:      1,
		nextTopicID:     1,
		nextMessageID:   1,
		users:           make(map[int64]*User),
		topics:          make(map[int64]*Topic),
		messages:        make(map[int64]*Message),
		messagesByTopic: make(map[int64][]*Message),
	}
}

//
// Users
//

func (s *Storage) CreateUser(name string) *User {
	s.mu.Lock()
	defer s.mu.Unlock()

	user := &User{
		Id:   s.nextUserID,
		Name: name,
	}
	s.nextUserID++

	s.users[user.Id] = user
	return user
}

func (s *Storage) UserExists(userId int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.users[userId]
	return ok
}

//
// Topics
//

func (s *Storage) CreateTopic(name string) *Topic {
	s.mu.Lock()
	defer s.mu.Unlock()

	topic := &Topic{
		Id:   s.nextTopicID,
		Name: name,
	}
	s.nextTopicID++

	s.topics[topic.Id] = topic
	return topic
}

func (s *Storage) ListTopics() []*Topic {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Topic, 0, len(s.topics))
	for _, t := range s.topics {
		result = append(result, t)
	}
	return result
}

func (s *Storage) TopicExists(topicId int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.topics[topicId]
	return ok
}

//
// Messages
//

func (s *Storage) PostMessage(topicId, userId int64, text string) *Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := &Message{
		Id:        s.nextMessageID,
		TopicId:   topicId,
		UserId:    userId,
		Text:      text,
		CreatedAt: time.Now(),
		Likes:     0,
	}
	s.nextMessageID++

	s.messages[msg.Id] = msg
	s.messagesByTopic[topicId] = append(s.messagesByTopic[topicId], msg)

	return msg
}

func (s *Storage) GetMessages(topicId int64, fromMessageId int64, limit int32) []*Message {

	s.mu.RLock()
	defer s.mu.RUnlock()

	msgs := s.messagesByTopic[topicId]
	result := make([]*Message, 0)

	for _, m := range msgs {
		if m.Id >= fromMessageId {
			result = append(result, m)
			if limit > 0 && int32(len(result)) >= limit {
				break
			}
		}
	}

	return result
}

func (s *Storage) MessageExists(messageId int64) bool {

	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.messages[messageId]

	return ok
}

func (s *Storage) LikeMessage(messageId int64) *Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := s.messages[messageId]
	msg.Likes++
	return msg
}
