package statemachine

import "fmt"

type ProcessMessageAction struct{}

func (a *ProcessMessageAction) Execute(eventCtx EventContext) EventType {
	fmt.Println("Processing user message")
	// Placeholder: add logic to interact with OpenAI or another processing mechanism.
	return EventProcessMessage
}

type FunctionCallAction struct{}

func (a *FunctionCallAction) Execute(eventCtx EventContext) EventType {
	fmt.Println("Executing a function call")
	// Placeholder: implement the logic to handle a function call.
	return NoOp
}
