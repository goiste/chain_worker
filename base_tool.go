package chain_worker

import (
	"context"
	"fmt"
)

const bufSize = 8

// tool is an interface that must be implemented in the BaseTool implementations to do useful work
type tool[I, O any] interface {
	Do(data I) (O, error)
	Name() string
}

// BaseTool is an implementation of ChainTool with common handling logic
type BaseTool[I, O any] struct {
	tool tool[I, O]

	inChan  chan Message
	outChan chan Message
	errChan chan error
}

// NewBaseTool creates a new BaseTool object
func NewBaseTool[I, O any](tool tool[I, O]) *BaseTool[I, O] {
	return &BaseTool[I, O]{
		tool: tool,

		inChan:  make(chan Message, bufSize),
		outChan: make(chan Message, bufSize),
		errChan: make(chan error, bufSize),
	}
}

// Run starts handling
func (t *BaseTool[I, O]) Run(ctx context.Context) {
outer:
	for {
		select {
		case message, opened := <-t.InChan():
			if !opened {
				return
			}

			data, err := t.getDataFromMessage(message)
			if err != nil {
				t.ErrChan() <- fmt.Errorf("decode data error: %w", err)
				continue outer
			}

			result, err := t.tool.Do(data)
			if err != nil {
				t.ErrChan() <- fmt.Errorf("tool error: %w", err)
				continue outer
			}

			out := Message{
				ToolName: t.tool.Name(),
				Done:     true,
			}
			err = out.Encode(result)
			if err != nil {
				t.ErrChan() <- fmt.Errorf("encode data error: %w", err)
				continue outer
			}

			t.OutChan() <- out
		case <-ctx.Done():
			return
		}
	}
}

// Name returns tool name
func (t *BaseTool[I, O]) Name() string {
	return t.tool.Name()
}

// InChan returns input channel
func (t *BaseTool[I, O]) InChan() chan Message {
	return t.inChan
}

// OutChan returns output channel
func (t *BaseTool[I, O]) OutChan() chan Message {
	return t.outChan
}

// ErrChan returns error channel
func (t *BaseTool[I, O]) ErrChan() chan error {
	return t.errChan
}

// converts data from Message to object of required type
func (t *BaseTool[I, O]) getDataFromMessage(message Message) (I, error) {
	obj := new(I)
	err := message.Decode(obj)
	return *obj, err
}
