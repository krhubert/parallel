// Package parallel provide API for running function in parallel.
package parallel

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

type taskInput interface{}

// Task representation.
type Task struct {
	n int
	f reflect.Value

	hasError bool
	numOut   int
	input    []taskInput
	result   []interface{}
	err      error

	errChan    chan error
	inputChan  chan taskInput
	resultChan chan interface{}

	mx               *sync.Mutex // muxtex for accessing *ChanClosed value
	mxAppend         *sync.Mutex // muxtex for append to result[]
	appendWg         sync.WaitGroup
	inputChanClosed  bool
	resultChanClosed bool
	errChanClosed    bool
}

// NewTask create task that could run parallel workers.
// N is the number of workers run in parallel and f is pointer to function.
// Acceptable functions declartaions are those which takes 0 or 1 paramentr and raturn 0, 1 or 2 value,
// where number of return value defines format:
// 	1 - can be error or any other type
// 	2 - 2nd value must be an error
// for example:
// 	func foo() {}
// 	func foo(m *MyType) interface{}
// 	func foo(s string) (m *MyType1, error)
// 	func foo(i int) error
//
// When function return error, workers STOP execute.
func NewTask(n int, f interface{}) (*Task, error) {
	fVal, fType := valueAndType(f)
	if fVal.Kind() != reflect.Func {
		return nil, fmt.Errorf("task: f paramentr must be function")
	}
	if fType.NumIn() > 1 {
		return nil, fmt.Errorf("task: f function lenght of paramentr must be 0 or 1")
	}
	if fType.NumOut() > 2 {
		return nil, fmt.Errorf("task: f function lenght of return paramentr must be 2 or less")
	}

	numOut := fType.NumOut()

	var t = &Task{
		n:        n,
		f:        fVal,
		numOut:   numOut,
		input:    make([]taskInput, 0),
		result:   make([]interface{}, 0),
		mx:       &sync.Mutex{},
		mxAppend: &sync.Mutex{},
	}

	t.appendWg.Add(1)
	if numOut > 0 {
		if _, b := fType.Out(numOut - 1).MethodByName("Error"); b {
			t.hasError = true
		}
	}

	return t, nil
}

// Feed task with input which will be passed to workers.
func (t *Task) Feed(input interface{}) {
	t.input = append(t.input, input)
}

// Run task. It return error if any, otherwise array of function returned values and nil.
func (t *Task) Run() ([]interface{}, error) {
	t.initializeChannels()
	t.collectResultsFromWorkers()
	t.sendInputToWorkers()
	t.waitForError()
	t.startWorkers()

	if t.err != nil {
		return nil, t.err
	}

	// TODO: write faster version
	t.appendWg.Wait()
	return t.result, nil
}

// call function which is stored in reflect.Value and pass return value to next step.
func (t *Task) reflectCallFunction(input taskInput) error {
	p := reflect.ValueOf(input)
	result := t.f.Call([]reflect.Value{p})
	return t.parseAndStoreResult(result)
}

func (t *Task) parseAndStoreResult(result []reflect.Value) error {
	if t.hasError {
		err := result[t.numOut-1].Interface()
		if err != nil {
			t.mx.Lock()
			defer t.mx.Unlock()
			if !t.errChanClosed {
				t.errChan <- err.(error)
			}
			return errors.New("")
		}
		result = result[:len(result)-1]
	}

	if len(result) == 1 {
		t.mx.Lock()
		defer t.mx.Unlock()

		if !t.resultChanClosed {
			t.resultChan <- result[0].Interface()
		}
	}

	return nil
}

func (t *Task) startWorkers() {
	var wg sync.WaitGroup
	wg.Add(t.n)
	for i := 0; i < t.n; i++ {
		go func() {
			t.doWork()
			wg.Done()
		}()
	}
	wg.Wait()
	t.deinitializeChannels()
}

// Main loop of workers. It reads input from channel and calls function.
func (t *Task) doWork() {
	for input := range t.inputChan {
		if t.reflectCallFunction(input) != nil {
			return
		}
	}
}

// loop for sending input to workers. It stops sending if worker return error.
func (t *Task) sendInputToWorkers() {
	go func() {
		defer func() { recover() }()

		for _, i := range t.input {
			t.inputChan <- i
		}

		close(t.inputChan)
		t.inputChanClosed = true
	}()
}

func (t *Task) waitForError() {
	go func() {
		for err := range t.errChan {
			t.err = err
			t.deinitializeChannels()
		}
	}()
}

func (t *Task) collectResultsFromWorkers() {
	go func() {
		for res := range t.resultChan {
			t.result = append(t.result, res)
		}
		t.appendWg.Done()
	}()
}

func (t *Task) initializeChannels() {
	t.errChan = make(chan error)
	t.inputChan = make(chan taskInput)
	t.resultChan = make(chan interface{})
}

func (t *Task) deinitializeChannels() {
	t.mx.Lock()
	defer t.mx.Unlock()

	if !t.inputChanClosed {
		t.inputChanClosed = true
		close(t.inputChan)
	}
	if !t.resultChanClosed {
		t.resultChanClosed = true
		close(t.resultChan)
	}
	if !t.errChanClosed {
		t.errChanClosed = true
		close(t.errChan)
	}
}

func valueAndType(f interface{}) (v reflect.Value, t reflect.Type) {
	v = reflect.ValueOf(f)
	t = v.Type()
	return
}
