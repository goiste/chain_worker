package chain_worker

import (
	"fmt"
	"reflect"
	"sync"

	"golang.org/x/net/context"
)

const (
	InputName = "input"
)

// ChainTool is an interface which implementation must be included in chained tool
//
// Implemented by BaseTool and BaseGeneratorTool
//
// see usage example: https://github.com/goiste/chain_worker/tree/main/example
type ChainTool interface {
	Run(ctx context.Context)
	Name() string

	InChan() chan Message
	OutChan() chan Message
	ErrChan() chan error
}

// Worker handles chained tools
type Worker[T any] struct {
	*sync.WaitGroup

	listeners map[string][]ChainTool
	tools     map[string]ChainTool

	input  chan Message
	output chan Message
	errors chan error

	outNames map[string]func() interface{}
}

// New returns a new Worker with initial input data
func New[T any](inputData []T) *Worker[T] {
	wrk := &Worker[T]{
		WaitGroup: new(sync.WaitGroup),

		listeners: make(map[string][]ChainTool),
		tools:     make(map[string]ChainTool),

		input:  make(chan Message, len(inputData)),
		output: make(chan Message, bufSize),
		errors: make(chan error, bufSize),
	}

	for _, in := range inputData {
		message := Message{ToolName: InputName}
		err := message.Encode(in)
		if err != nil {
			fmt.Printf("error encoding input data %#v: %v", in, err)
			continue
		}
		wrk.input <- message
	}

	close(wrk.input)

	return wrk
}

// Subscribe adds new tools to handle toolName output
func (w *Worker[T]) Subscribe(toolName string, tools ...ChainTool) {
	w.listeners[toolName] = append(w.listeners[toolName], tools...)
	for _, t := range tools {
		w.tools[t.Name()] = t
	}
}

// SetOutput sets Worker to send some tools output to Worker output
//
// outNames uses key as tool name and value as function to produce a pointer to object of type corresponding to output type
// (for using in Message Decode() method)
func (w *Worker[T]) SetOutput(outNames map[string]func() interface{}) {
	w.outNames = outNames
}

// Run starts the Worker
func (w *Worker[T]) Run(ctx context.Context) (results []interface{}, errs []error) {
	wCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, t := range w.tools {
		go t.Run(wCtx)
	}

	for m := range w.input {
		for _, t := range w.listeners[m.ToolName] {
			w.Add(1)
			t.InChan() <- m
		}
	}

	outChannels := make([]chan Message, 0, len(w.tools))
	errChannels := make([]chan error, 0, len(w.tools))
	for _, t := range w.tools {
		outChannels = append(outChannels, t.OutChan())
		errChannels = append(errChannels, t.ErrChan())
	}

	go w.handleOutChannels(wCtx, getCases(outChannels), w.output)
	go w.handleErrorChannels(wCtx, getCases(errChannels), w.errors)

	go func() {
	outer:
		for {
			select {
			case m := <-w.output:
				dataFunc := w.outNames[m.ToolName]
				if dataFunc == nil {
					w.doneTask(m)
					continue outer
				}

				data := dataFunc()
				err := m.Decode(data)
				if err != nil {
					errs = append(errs, fmt.Errorf("error decoding data: %w", err))
					w.doneTask(m)
					continue outer
				}

				results = append(results, data)

				w.doneTask(m)
			case err := <-w.errors:
				errs = append(errs, err)
				w.Done()
			case <-wCtx.Done():
				return
			}
		}
	}()

	w.Wait()

	return
}

func (w *Worker[T]) handleOutChannels(ctx context.Context, cases []reflect.SelectCase, resultChan chan Message) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, value, ok := reflect.Select(cases)
			if !ok {
				continue
			}

			msg := value.Interface().(Message)

			toolName := msg.ToolName

			for _, t := range w.listeners[toolName] {
				if t.Name() == toolName {
					continue
				}
				w.Add(1)
				t.InChan() <- msg
			}

			if _, exists := w.outNames[toolName]; exists {
				resultChan <- msg
			} else if msg.Done {
				w.Done()
			}
		}
	}
}

func (w *Worker[T]) handleErrorChannels(ctx context.Context, cases []reflect.SelectCase, errChan chan error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, value, ok := reflect.Select(cases)
			if !ok {
				continue
			}

			err := value.Interface().(error)
			errChan <- err
		}
	}
}

func getCases[T any](channels []chan T) []reflect.SelectCase {
	cases := make([]reflect.SelectCase, len(channels))
	for i, ch := range channels {
		cases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		}
	}
	return cases
}

func (w *Worker[T]) doneTask(message Message) {
	if message.Done {
		w.Done()
	}
}
