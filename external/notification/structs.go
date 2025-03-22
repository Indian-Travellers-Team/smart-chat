package notification

// MessagePair represents user and bot messages.
type MessagePair struct {
	User string `json:"user"`
	Bot  string `json:"bot"`
}

// Payload is the JSON body sent to the notification service.
type Payload struct {
	ConversationID uint        `json:"conversation_id"`
	Mobile         string      `json:"mobile"`
	MessagePair    MessagePair `json:"message_pair"`
}
