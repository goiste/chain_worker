package chain_worker

import (
	"context"
	"fmt"
)

const bufSize = 8

type tool[I, O any] interface {
	Do(data I) (O, error)
	Name() string
}

type BaseTool[I, O any] struct {
	tool tool[I, O]

	inChan  chan Message
	outChan chan Message
	errChan chan error
}

func NewBaseTool[I, O any](tool tool[I, O]) *BaseTool[I, O] {
	return &BaseTool[I, O]{
		tool: tool,

		inChan:  make(chan Message, bufSize),
		outChan: make(chan Message, bufSize),
		errChan: make(chan error, bufSize),
	}
}

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

func (t *BaseTool[I, O]) Name() string {
	return t.tool.Name()
}

func (t *BaseTool[I, O]) InChan() chan Message {
	return t.inChan
}

func (t *BaseTool[I, O]) OutChan() chan Message {
	return t.outChan
}

func (t *BaseTool[I, O]) ErrChan() chan error {
	return t.errChan
}

func (t *BaseTool[I, O]) getDataFromMessage(message Message) (I, error) {
	obj := new(I)
	err := message.Decode(obj)
	return *obj, err
}
