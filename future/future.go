package future

import (
	"context"
	"errors"
	"sync"
	"time"
)

type future struct {
	done       chan struct{}
	cancelChan <-chan struct{}
	cancelFunc context.CancelFunc
	val        interface{}
	err        error
	mutex      sync.Mutex
}

func (f *future) Cancelled() bool {
	select {
	case <-f.cancelChan:
		return true //already cancelled
	default:
		return false // running or finished
	}
}

func (f *future) Cancel() {
	select {
	case <-f.done:
		return //already finished
	case <-f.cancelChan:
		return //already cancelled
	default:
		// f.mutex.Lock()
		f.err = errors.New("cancelled by user")
		// f.mutex.Unlock()
		f.cancelFunc()
	}
}

func (f *future) Running() bool {
	select {
	case <-f.cancelChan:
		return false // already cancelled
	case <-f.done:
		return false //already finished
	default:
		return true
	}
}

func (f *future) Isdone() bool {
	select {
	case <-f.cancelChan:
		return true //already cancelled
	case <-f.done:
		return true // done
	default:
		return false // still running
	}
}

// gets states of future through interface functions
func (f *future) GetState() string {
	if f.Cancelled() {
		return "Cancelled"
	} else if f.Isdone() {
		return "Done"
	} else {
		return "Running"
	}
}

// gives the value of handler and the error
func (f *future) Result() (interface{}, error) {
	select {
	case <-f.done:
		return f.val, f.err
	case <-f.cancelChan:
		return f.val, f.err
	}
}

// adds a next future to execute on completion of current future
func (f *future) AddDoneCallback(timeout time.Duration, next func(interface{}) (interface{}, error)) *future {
	nextFutureHandler := func() (interface{}, error) {
		result, err := f.Result()
		if f.Cancelled() || err != nil {
			return result, err
		}
		return next(result)
	}
	nextFuture := createFutureWithContext(timeout, f.cancelChan, f.cancelFunc, nextFutureHandler)
	return nextFuture
}

// returns exceptions for current future
func (f *future) exception() error {
	return f.err
}

// New creates a new future
func New(timeout time.Duration, handler func() (interface{}, error)) *future {
	ctx := context.Background()
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	return createFutureWithContext(timeout, cancelCtx.Done(), cancelFunc, handler)
}

func createFutureWithContext(timeout time.Duration, cancelChan <-chan struct{}, cancelFunc context.CancelFunc, handler func() (interface{}, error)) *future {
	f := future{
		done:       make(chan struct{}),
		cancelChan: cancelChan,
		cancelFunc: cancelFunc,
	}
	go func() {
		if timeout.Milliseconds() == 0 {
			return
		}
		// fmt.Println(time.Now())
		<-time.After(timeout)
		// fmt.Println(time.Now())
		f.Cancel()
		if f.Cancelled() {
			f.mutex.Lock()
			f.err = errors.New("future timed out")
			f.mutex.Unlock()
		}
	}()
	go func() {
		val, err := handler()
		// fmt.Println(f.GetState())
		if !f.Cancelled() {
			f.mutex.Lock()
			f.val, f.err = val, err
			f.mutex.Unlock()
		}
		close(f.done)
	}()
	return &f
}
