package future

import (
	"context"
)

type futureInterface interface {
	cancel()
	cancelled() bool
	running() bool
	Isdone() bool
	result() (interface{}, error)
	exception() error
	add_done_callback(func(interface{}) (interface{}, error)) futureInterface
}

type futureImpl struct {
	done       chan struct{}
	cancelChan <-chan struct{}
	cancelFunc context.CancelFunc
	val        interface{}
	err        error
}

func (f *futureImpl) cancelled() bool {
	select {
	case <-f.cancelChan:
		return true //already cancelled
	default:
		return false // running or finished
	}
}

func (f *futureImpl) cancel() {
	select {
	case <-f.done:
		return //already finished
	case <-f.cancelChan:
		return //already cancelled
	default:
		f.cancelFunc() // or add custom errors with errors.new("New error")
	}
}

func (f *futureImpl) running() bool {
	select {
	case <-f.cancelChan:
		return false // already cancelled
	case <-f.done:
		return false //already finished
	default:
		return true
	}
}

func (f *futureImpl) Isdone() bool {
	select {
	case <-f.cancelChan:
		return true //already cancelled
	case <-f.done:
		return true
	default:
		return false // still running
	}
}

func (f *futureImpl) result() (interface{}, error) {
	select {
	case <-f.done:
		return f.val, f.err
	case <-f.cancelChan:
		return f.val, f.err
	}
	return nil, nil
}

func (f *futureImpl) add_done_callback(next func(interface{}) (interface{}, error)) futureInterface {
	nextFutureInFunc := func() (interface{}, error) {
		result, err := f.result()
		if f.cancelled() || err != nil {
			return result, err
		}
		return next(result)
	}
	nextFuture := newUtil(f.cancelChan, f.cancelFunc, nextFutureInFunc)
	return nextFuture
}

func (f *futureImpl) exception() error {
	return f.err
}

func New(inFunc func() (interface{}, error)) futureInterface {
	ctx := context.Background()
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	return newUtil(cancelCtx.Done(), cancelFunc, inFunc)
}

func newUtil(cancelChan <-chan struct{}, cancelFunc context.CancelFunc, inFunc func() (interface{}, error)) futureInterface {
	f := futureImpl{
		done:       make(chan struct{}),
		cancelChan: cancelChan,
		cancelFunc: cancelFunc,
	}
	go func() {
		go func() {
			f.val, f.err = inFunc()
			close(f.done)
		}()
		select {
		case <-f.done:
			//do nothing
		case <-f.cancelChan:
			//do nothing, val = nil and err = nil
		}
	}()
	return &f
}
