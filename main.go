package main

import (
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

const (
	addCountInIteration = 100
	goroutinesCount     = 1000
	iterationTimeout    = time.Second
	counterDotInterval  = 10
	counterNewLine      = counterDotInterval * 100
)

func iteration() error {
	var m sync.Mutex
	var wg sync.WaitGroup
	s := 0

	for i := 0; i < goroutinesCount; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for i := 0; i < addCountInIteration; i++ {
				m.Lock()
				s++
				m.Unlock()
			}
			runtime.Gosched()
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < addCountInIteration; i++ {
				m.Lock()
				s--
				m.Unlock()
			}
			runtime.Gosched()
		}()
	}

	completed := make(chan struct{})
	go func() {
		wg.Wait()
		close(completed)
	}()

	timer := time.NewTimer(iterationTimeout)
	defer timer.Stop()

	select {
	case <-completed:
		if s != 0 {
			return fmt.Errorf("s is not zero: %d", s)
		}
		return nil
	case <-timer.C:
		return fmt.Errorf("timeout")
	}

}

func main() {
	counter := 0
	for {
		if err := iteration(); err != nil {
			log.Println()
			log.Printf("counter: %d", counter)
			panic(err)
		}
		counter++
		if counter%counterDotInterval == 0 {
			fmt.Print(".")
		}
		if counter%counterNewLine == 0 {
			fmt.Println()
		}
	}
}
