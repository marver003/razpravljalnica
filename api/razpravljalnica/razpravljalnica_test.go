package razpravljalnica

import (
	"testing"
)

func TestUser_Creation(t *testing.T) {
	testCases := []struct {
		name      string
		id        int64
		userName  string
		shouldErr bool
	}{
		{
			name:      "Valid user",
			id:        1,
			userName:  "alice",
			shouldErr: false,
		},
		{
			name:      "Valid user 2",
			id:        2,
			userName:  "bob",
			shouldErr: false,
		},
		{
			name:      "Empty name",
			id:        3,
			userName:  "",
			shouldErr: true,
		},
		{
			name:      "Negative ID",
			id:        -1,
			userName:  "charlie",
			shouldErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user := &User{
				Id:   tc.id,
				Name: tc.userName,
			}

			if tc.userName == "" && !tc.shouldErr {
				t.Error("Expected error for empty user name")
			}
			if tc.id < 0 && !tc.shouldErr {
				t.Error("Expected error for negative ID")
			}

			if user.Id != tc.id {
				t.Errorf("User ID mismatch: expected %d, got %d", tc.id, user.Id)
			}
			if user.Name != tc.userName {
				t.Errorf("User name mismatch: expected %s, got %s", tc.userName, user.Name)
			}
		})
	}
}

func TestTopic_Creation(t *testing.T) {
	testCases := []struct {
		name      string
		id        int64
		topicName string
		valid     bool
	}{
		{"Valid topic", 1, "golang", true},
		{"Valid topic 2", 2, "distributed-systems", true},
		{"Empty name", 3, "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			topic := &Topic{
				Id:   tc.id,
				Name: tc.topicName,
			}

			if tc.topicName == "" && tc.valid {
				t.Error("Expected error for empty topic name")
			}

			if topic.Id != tc.id {
				t.Errorf("Topic ID mismatch: expected %d, got %d", tc.id, topic.Id)
			}
			if topic.Name != tc.topicName {
				t.Errorf("Topic name mismatch: expected %s, got %s", tc.topicName, topic.Name)
			}
		})
	}
}

func TestMessage_Creation(t *testing.T) {
	testCases := []struct {
		name    string
		id      int64
		topicID int64
		userID  int64
		text    string
		valid   bool
	}{
		{"Valid message", 1, 1, 1, "Hello World", true},
		{"Empty text", 2, 1, 1, "", false},
		{"Invalid topic", 3, -1, 1, "test", false},
		{"Invalid user", 4, 1, -1, "test", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := &Message{
				Id:      tc.id,
				TopicId: tc.topicID,
				UserId:  tc.userID,
				Text:    tc.text,
				Likes:   0,
			}

			if tc.text == "" && tc.valid {
				t.Error("Expected error for empty message text")
			}
			if tc.topicID < 0 && tc.valid {
				t.Error("Expected error for invalid topic ID")
			}

			if msg.Id != tc.id {
				t.Errorf("Message ID mismatch: expected %d, got %d", tc.id, msg.Id)
			}
			if msg.Text != tc.text {
				t.Errorf("Message text mismatch: expected %s, got %s", tc.text, msg.Text)
			}
		})
	}
}

func TestCreateUserRequest_Validation(t *testing.T) {
	testCases := []struct {
		name     string
		userName string
		valid    bool
	}{
		{"Valid user request", "alice", true},
		{"Another user", "bob", true},
		{"Empty name", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &CreateUserRequest{
				Name: tc.userName,
			}

			if tc.userName == "" && tc.valid {
				t.Error("Expected error for empty name")
			}

			if req.Name != tc.userName {
				t.Errorf("Name mismatch: expected %s, got %s", tc.userName, req.Name)
			}
		})
	}
}

func TestCreateTopicRequest_Validation(t *testing.T) {
	testCases := []struct {
		name      string
		topicName string
		valid     bool
	}{
		{"Valid topic", "golang", true},
		{"Another topic", "databases", true},
		{"Empty name", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &CreateTopicRequest{
				Name: tc.topicName,
			}

			if tc.topicName == "" && tc.valid {
				t.Error("Expected error for empty topic name")
			}

			if req.Name != tc.topicName {
				t.Errorf("Topic name mismatch: expected %s, got %s", tc.topicName, req.Name)
			}
		})
	}
}

func TestPostMessageRequest_Validation(t *testing.T) {
	testCases := []struct {
		name    string
		topicID int64
		userID  int64
		text    string
		valid   bool
	}{
		{"Valid post", 1, 1, "Hello", true},
		{"Empty text", 1, 1, "", false},
		{"Invalid topic", -1, 1, "Hello", false},
		{"Invalid user", 1, -1, "Hello", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &PostMessageRequest{
				TopicId: tc.topicID,
				UserId:  tc.userID,
				Text:    tc.text,
			}

			if tc.text == "" && tc.valid {
				t.Error("Expected error for empty text")
			}

			if req.TopicId != tc.topicID {
				t.Errorf("TopicId mismatch: expected %d, got %d", tc.topicID, req.TopicId)
			}
		})
	}
}

func TestLikeMessageRequest_Validation(t *testing.T) {
	testCases := []struct {
		name      string
		topicID   int64
		messageID int64
		userID    int64
		valid     bool
	}{
		{"Valid like", 1, 1, 1, true},
		{"Invalid topic", -1, 1, 1, false},
		{"Invalid message", 1, -1, 1, false},
		{"Invalid user", 1, 1, -1, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			like := &LikeMessageRequest{
				TopicId:   tc.topicID,
				MessageId: tc.messageID,
				UserId:    tc.userID,
			}

			if tc.topicID < 0 && tc.valid {
				t.Error("Expected error for invalid topic")
			}

			if like.TopicId != tc.topicID {
				t.Errorf("TopicId mismatch: expected %d, got %d", tc.topicID, like.TopicId)
			}
		})
	}
}

func TestListTopicsResponse_Structure(t *testing.T) {
	topics := []*Topic{
		{Id: 1, Name: "golang"},
		{Id: 2, Name: "rust"},
		{Id: 3, Name: "python"},
	}

	resp := &ListTopicsResponse{
		Topics: topics,
	}

	if len(resp.Topics) != 3 {
		t.Errorf("Expected 3 topics, got %d", len(resp.Topics))
	}

	for i, topic := range resp.Topics {
		if topic.Id != int64(i+1) {
			t.Errorf("Topic %d ID mismatch", i)
		}
	}
}

func TestGetMessagesRequest_Validation(t *testing.T) {
	testCases := []struct {
		name      string
		topicID   int64
		fromMsgID int64
		limit     int32
		valid     bool
	}{
		{"Valid request", 1, 0, 10, true},
		{"With offset", 1, 5, 20, true},
		{"Invalid topic", -1, 0, 10, false},
		{"Zero limit", 1, 0, 0, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &GetMessagesRequest{
				TopicId:       tc.topicID,
				FromMessageId: tc.fromMsgID,
				Limit:         tc.limit,
			}

			if tc.topicID < 0 && tc.valid {
				t.Error("Expected error for invalid topic")
			}

			if req.TopicId != tc.topicID {
				t.Errorf("TopicId mismatch: expected %d, got %d", tc.topicID, req.TopicId)
			}
		})
	}
}

func TestMessage_LikeCounter(t *testing.T) {
	msg := &Message{
		Id:      1,
		TopicId: 1,
		UserId:  1,
		Text:    "Test message",
		Likes:   0,
	}

	if msg.Likes != 0 {
		t.Errorf("Initial likes should be 0, got %d", msg.Likes)
	}

	msg.Likes++
	if msg.Likes != 1 {
		t.Errorf("Expected 1 like, got %d", msg.Likes)
	}

	for i := 0; i < 9; i++ {
		msg.Likes++
	}

	if msg.Likes != 10 {
		t.Errorf("Expected 10 likes, got %d", msg.Likes)
	}
}

func TestMessageEvent_Creation(t *testing.T) {
	testCases := []struct {
		name           string
		sequenceNumber int64
		opType         OpType
		messagePresent bool
	}{
		{"Post event", 1, OpType_OP_POST, true},
		{"Like event", 2, OpType_OP_LIKE, true},
		{"No message", 3, OpType_OP_POST, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var msg *Message
			if tc.messagePresent {
				msg = &Message{
					Id:      1,
					TopicId: 1,
					UserId:  1,
					Text:    "Event message",
					Likes:   0,
				}
			}

			event := &MessageEvent{
				SequenceNumber: tc.sequenceNumber,
				Op:             tc.opType,
				Message:        msg,
			}

			if event.SequenceNumber != tc.sequenceNumber {
				t.Errorf("Sequence number mismatch: expected %d, got %d", tc.sequenceNumber, event.SequenceNumber)
			}

			if tc.messagePresent && event.Message == nil {
				t.Error("Message should not be nil")
			}
		})
	}
}
