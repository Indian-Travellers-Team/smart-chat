package conversation

import (
	"smart-chat/internal/models"

	"github.com/sashabaranov/go-openai"
	"gorm.io/gorm"
)

type State string

const (
	ConversationStateStart        State = "Start"
	ConversationStateRespond      State = "BotResponse"
	ConversationStateFunctionCall State = "FunctionCall"
	ConversationStateEnd          State = "END"
)

type ConversationState struct {
	ConversationID      uint
	State               State
	ConversationHistory []openai.ChatCompletionMessage
}

var workflowNextState = map[State]State{
	ConversationStateFunctionCall: ConversationStateStart,
	ConversationStateRespond:      ConversationStateEnd,
}

var workflowActionBasedState = map[models.MessageType]State{
	models.MessageTypeUserSent:     ConversationStateRespond,
	models.MessageTypeFunctionCall: ConversationStateFunctionCall,
}

func NewConversationState(db *gorm.DB) *ConversationState {
	cs := &ConversationState{
		State:               ConversationStateStart,
		ConversationHistory: make([]openai.ChatCompletionMessage, 0),
	}
	return cs
}

func (cs *ConversationState) InitState(conversationID uint, messages []openai.ChatCompletionMessage) {
	cs.ConversationID = conversationID
	cs.State = ConversationStateStart
	cs.ConversationHistory = messages
}

func (cs *ConversationState) NextState(messageType models.MessageType) {
	cs.State = workflowNextState[workflowActionBasedState[messageType]]
}

func (cs *ConversationState) EndState() {
	cs.State = ConversationStateEnd
}

func (cs *ConversationState) AddToHistory(message openai.ChatCompletionMessage) {
	cs.ConversationHistory = append(cs.ConversationHistory, message)
}
