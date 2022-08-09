package chain_worker

import (
	"context"
	"fmt"
)

// generator is an interface that must be implemented in the BaseGeneratorTool implementations to do useful work
type generator[I, O any] interface {
	Generate(data []I) func() (O, bool)
	Name() string
}

// BaseGeneratorTool is an implementation of ChainTool with the option to process a slice from the input channel
// by one element and then send each processed element to the output channel
type BaseGeneratorTool[I, O any] struct {
	*BaseTool[I, O]
	tool generator[I, O]
}

// NewBaseGeneratorTool creates a new BaseGeneratorTool
func NewBaseGeneratorTool[I, O any](tool generator[I, O]) *BaseGeneratorTool[I, O] {
	return &BaseGeneratorTool[I, O]{
		BaseTool: NewBaseTool[I, O](nil),
		tool:     tool,
	}
}

// Run starts handling
func (t *BaseGeneratorTool[I, O]) Run(ctx context.Context) {
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

			iterator := t.tool.Generate(data)

			s, ok := iterator()
			for ok {
				out := Message{ToolName: t.tool.Name()}

				err = out.Encode(s)
				if err != nil {
					t.ErrChan() <- fmt.Errorf("encode data error: %w", err)
					continue outer
				}

				s, ok = iterator()
				out.Done = !ok

				t.OutChan() <- out
			}
		case <-ctx.Done():
			return
		}
	}
}

// converts data from Message to slice of required type
func (t *BaseGeneratorTool[I, O]) getDataFromMessage(message Message) ([]I, error) {
	var obj []I
	err := message.Decode(&obj)
	return obj, err
}
