package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

const (
	goroutinesCount    = 2000
	iterationTimeout   = time.Second * 10
	counterDotInterval = 10
	counterNewLine     = counterDotInterval * 100
	testDuration       = time.Minute * 10
)

func iteration() error {
	var wg sync.WaitGroup
	var m sync.Mutex

	for i := 0; i < goroutinesCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			m.Lock()
			m.Unlock()

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
		return nil
	case <-timer.C:
		return fmt.Errorf("timeout")
	}

}

func main() {
	fmt.Println("CPU:", runtime.GOMAXPROCS(0))
	start := time.Now()
	counter := 0
	for {
		if time.Since(start) > testDuration {
			fmt.Println()
			fmt.Println("OK")
			return
		}
		if err := iteration(); err != nil {
			fmt.Println()
			fmt.Printf("\ncounter: %d\n", counter)
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
