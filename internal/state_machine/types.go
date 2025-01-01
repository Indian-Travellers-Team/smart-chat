package statemachine

import (
	"errors"
	"sync"
)

type StateType string
type EventType string

const (
	StateStart           StateType = "Start"
	StateAwaitingInput   StateType = "AwaitingInput"
	StateProcessing      StateType = "Processing"
	StateFunctionCall    StateType = "FunctionCall"
	StateEnd             StateType = "End"
	EventUserMessage     EventType = "UserMessage"
	EventProcessMessage  EventType = "ProcessMessage"
	EventFunctionCall    EventType = "FunctionCall"
	EventEndConversation EventType = "EndConversation"
	NoOp                 EventType = "NoOp"
)

var ErrEventRejected = errors.New("event rejected")

type EventContext interface{}

type Action interface {
	Execute(eventCtx EventContext) EventType
}

type Events map[EventType]StateType

type State struct {
	Action Action
	Events Events
}

type States map[StateType]State

type StateMachine struct {
	Previous StateType
	Current  StateType
	States   States
	Mutex    sync.Mutex
}
