package statemachine

import "fmt"

func NewStateMachine() *StateMachine {
	return &StateMachine{
		Current: StateStart,
		States: States{
			StateStart: {
				Action: &ProcessMessageAction{},
				Events: Events{
					EventProcessMessage: StateAwaitingInput,
					EventFunctionCall:   StateFunctionCall,
				},
			},
			StateFunctionCall: {
				Action: &FunctionCallAction{},
				Events: Events{
					NoOp: StateEnd,
				},
			},
		},
	}
}

// SendEvent processes an event, transitions the state machine to the next state, and executes any associated action.
func (sm *StateMachine) SendEvent(event EventType, eventCtx EventContext) error {
	sm.Mutex.Lock()
	defer sm.Mutex.Unlock()

	nextState, err := sm.getNextState(event)
	if err != nil {
		fmt.Println("Event rejected:", err)
		return ErrEventRejected // Custom error indicating the event was not accepted in the current state
	}

	// Transition to the next state
	sm.Previous = sm.Current
	sm.Current = nextState

	// Execute the action associated with the next state
	if action, ok := sm.States[nextState]; ok && action.Action != nil {
		// Execute the action and check for a follow-up event
		nextEvent := action.Action.Execute(eventCtx)

		// If the action results in a new event, handle it recursively
		if nextEvent != NoOp {
			return sm.SendEvent(nextEvent, eventCtx)
		}
	} else {
		fmt.Println("No action defined for state:", nextState)
		// Optionally, handle states without defined actions
	}

	return nil
}

// getNextState determines the next state based on the current state and the event
func (sm *StateMachine) getNextState(event EventType) (StateType, error) {
	if state, ok := sm.States[sm.Current]; ok {
		if nextState, ok := state.Events[event]; ok {
			return nextState, nil
		}
	}
	return StateEnd, ErrEventRejected
}
