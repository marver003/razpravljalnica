package storage

import (
	"log"
	"sync"

	//"time"

	pb "github.com/marver003/razpravljalnica/api/razpravljalnica"
	repl "github.com/marver003/razpravljalnica/api/replication"
	"google.golang.org/protobuf/proto"
)

type Storage struct {
	mu sync.RWMutex

	// Log operacij za recovery in sinhronizacijo
	operations []*repl.Operation
	lastIndex  int64

	// Dejanski podatki (uporabljamo PB tipe za konsistentnost)
	users           map[int64]*pb.User
	topics          map[int64]*pb.Topic
	messages        map[int64]*pb.Message
	messagesByTopic map[int64][]*pb.Message
}

func NewStorage() *Storage {
	return &Storage{
		users:           make(map[int64]*pb.User),
		topics:          make(map[int64]*pb.Topic),
		messages:        make(map[int64]*pb.Message),
		messagesByTopic: make(map[int64][]*pb.Message),
		operations:      make([]*repl.Operation, 0),
	}
}

func (s *Storage) Apply(op *repl.Operation) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Idempotentnost: Preveri, če smo to operacijo že izvedli
	if op.Index <= s.lastIndex && len(s.operations) > 0 {
		return
	}

	// 2. Dodaj v log
	s.operations = append(s.operations, op)
	s.lastIndex = op.Index

	// 3. Izvrši spremembo (Pretvori bytes nazaj v objekte)
	switch op.Type {
	case repl.OperationType_OP_CREATE_USER:
		log.Printf("STORAGE: Creating user at index %d", op.Index)
		var req pb.CreateUserRequest
		if proto.Unmarshal(op.Payload, &req) == nil {
			s.users[op.Index] = &pb.User{Id: op.Index, Name: req.Name}
		}

	case repl.OperationType_OP_CREATE_TOPIC:
		log.Printf("STORAGE: Applying new topic at index %d", op.Index)
		var req pb.CreateTopicRequest
		if proto.Unmarshal(op.Payload, &req) == nil {
			s.topics[op.Index] = &pb.Topic{Id: op.Index, Name: req.Name}
		}

	case repl.OperationType_OP_POST_MESSAGE:
		log.Printf("STORAGE: Posting new message at index %d", op.Index)
		var req pb.PostMessageRequest
		if proto.Unmarshal(op.Payload, &req) == nil {
			msg := &pb.Message{
				Id:      op.Index, // Index operacije je unikaten ID sporočila!
				TopicId: req.TopicId,
				UserId:  req.UserId,
				Text:    req.Text,
				Likes:   0,
			}
			s.messages[op.Index] = msg
			s.messagesByTopic[req.TopicId] = append(s.messagesByTopic[req.TopicId], msg)
		}

	case repl.OperationType_OP_LIKE_MESSAGE:
		log.Printf("STORAGE: Liking a message at index %d", op.Index)
		var req pb.LikeMessageRequest
		if proto.Unmarshal(op.Payload, &req) == nil {
			if msg, ok := s.messages[req.MessageId]; ok {
				msg.Likes++
			}
		}
	}
}

func (s *Storage) GetOperationsFrom(index int64) []*repl.Operation {
	log.Printf("STORAGE: GetOperationFrom index %d", index)
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*repl.Operation
	for _, op := range s.operations {
		if op.Index >= index {
			result = append(result, op)
		}
	}
	return result
}

// GetMessage vrne posamezno sporočilo po ID-ju v formatu, ki ga pričakuje gRPC
func (s *Storage) GetMessage(id int64) *pb.Message {
	log.Printf("STORAGE: Get Message id %d", id)
	s.mu.RLock()
	defer s.mu.RUnlock()

	msg, ok := s.messages[id]
	if !ok {
		return nil
	}

	return msg
}

/*
//
// Users
//

	func (s *Storage) CreateUser(name string) *pb.User {
		s.mu.Lock()
		defer s.mu.Unlock()

		user := &pb.User{
			Id:   s.nextUserID,
			Name: name,
		}
		s.nextUserID++

		s.users[user.Id] = user
		return user
	}
*/
func (s *Storage) UserExists(userId int64) bool {
	log.Printf("STORAGE: UserExists userId %d", userId)
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.users[userId]
	return ok
}

/*
//
// Topics
//

	func (s *Storage) CreateTopic(name string) *pb.Topic {
		s.mu.Lock()
		defer s.mu.Unlock()

		topic := &pb.Topic{
			Id:   s.nextTopicID,
			Name: name,
		}
		s.nextTopicID++

		s.topics[topic.Id] = topic
		return topic
	}
*/
func (s *Storage) ListTopics() []*pb.Topic {
	log.Printf("STORAGE: ListTopics")
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*pb.Topic, 0, len(s.topics))
	for _, t := range s.topics {
		result = append(result, t)
	}
	return result
}

func (s *Storage) TopicExists(topicId int64) bool {
	log.Printf("STORAGE: TopicExists topicId %d", topicId)
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.topics[topicId]
	return ok
}

/*
//
// Messages
//

	func (s *Storage) PostMessage(topicId, userId int64, text string) *pb.Message {
		s.mu.Lock()
		defer s.mu.Unlock()

		msg := &pb.Message{
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
*/
func (s *Storage) GetMessages(topicId int64, fromMessageId int64, limit int32) []*pb.Message {
	log.Printf("STORAGE: GetMessages topicId %d; fromMessageId %d; limit %d", topicId, fromMessageId, limit)
	s.mu.RLock()
	defer s.mu.RUnlock()

	msgs := s.messagesByTopic[topicId]
	result := make([]*pb.Message, 0)

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
	log.Printf("STORAGE: MessageExists messageId", messageId)
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.messages[messageId]

	return ok
}

/*
func (s *Storage) LikeMessage(messageId int64) *pb.Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	msg := s.messages[messageId]
	msg.Likes++
	return msg
}
*/

func (s *Storage) GetUser(id int64) *pb.User {
	log.Printf("STORAGE: GetUser id %d", id)
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.users[id]
}

func (s *Storage) GetTopic(id int64) *pb.Topic {
	log.Printf("STORAGE: GetTopic id %d", id)
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.topics[id]
}
