package main

import (
	"sync"
	"time"
)

func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()

	select {
	case <-c:
		return false // completed normally

	case <-time.After(timeout):
		return true // timed out
	}

}
