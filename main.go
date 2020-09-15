package main

import (
	"errors"
	"fmt"
	"time"

	future "github.com/surabhisuman/go-future/future"
)

// calculates fibonacci number with some delay
func findFibonacci(delay time.Duration, n int) (interface{}, error) {
	time.Sleep(delay)
	if n < 0 {
		return nil, errors.New("Fibonacci cannot be calculated for negative integers")
	}
	fib := make([]int, n+1)
	for i := 0; i <= n; i++ {
		if i == 1 || i == 0 {
			fib[i] = 1
		} else {
			fib[i] = fib[i-1] + fib[i-2]
		}
	}
	return fib[n], nil
}

// timeoutFuture function that takes a delay(for calculating handler) and a timeout(future timeout) and gives output of a future. future states are shown here
func timeoutFuture(delay time.Duration, timeout time.Duration, num int) {
	f := future.New(timeout, func() (interface{}, error) {
		return findFibonacci(delay, num)
	})
	fmt.Println("Future is", f.GetState())
	for !f.Isdone() {
		// this is just to block function execution until future execution is done
	}
	val, err := f.Result()
	fmt.Println("Future is", f.GetState())
	print(val, err, num)
}

// this function builds a future and then cancels it. future states are shown here
func cancelFuture(delay time.Duration, timeout time.Duration, num int) {
	f := future.New(timeout, func() (interface{}, error) {
		return findFibonacci(delay, num)
	})
	fmt.Println("Future is", f.GetState())
	f.Cancel()
	fmt.Println("Future is", f.GetState())
	val, err := f.Result()
	print(val, err, num)
}

// this function adds a done callback to a future and shows its state
func nextFuture(delay time.Duration, timeout time.Duration, num int, nextDelay time.Duration, nextTimeout time.Duration, nextNum int) {
	f := future.New(timeout, func() (interface{}, error) {
		return findFibonacci(delay, num)
	})
	fmt.Println("Future is", f.GetState())
	newFuture := f.AddDoneCallback(nextTimeout, func(interface{}) (interface{}, error) {
		return findFibonacci(nextDelay, nextNum)
	})
	fmt.Println("Next Future is", newFuture.GetState())
	fmt.Println("Future is", f.GetState())
	f.Result() // wait for first future to finish
	fmt.Println("Future is", f.GetState())
	newFuture.Result() // wait for newFuture to finish
	fmt.Println("Next Future is", newFuture.GetState())
}

// this function just prints the future output and error(if any)
func print(val interface{}, err error, num int) {
	if err != nil {
		fmt.Println("got error", err)
	} else {
		fmt.Println("Fib[", num, "] =", val)
	}
	fmt.Println("\n")
}

func main() {
	num := 5
	time1 := 1 * time.Second
	time2 := 5 * time.Second
	time3 := 2 * time.Second
	timeoutFuture(time1, time1, num)             // delay in calculating handler is less than timeout. So, it should give proper result
	timeoutFuture(time2, time1, num)             // delay in calculating handler is greater than timeout. So, result should give timeout error
	timeoutFuture(time1, time2, -1)              // should throw custom error
	cancelFuture(time1, time2, num)              // future is cancelled, throws custom error
	nextFuture(time3, time2, 2, time1, time2, 5) // future is created and a callback future is attached to it
}
