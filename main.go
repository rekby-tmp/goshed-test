package main

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime"
	"sync"
	"time"
)

const (
	goroutinesCount    = 1
	iterationTimeout   = time.Second * 10 * 1000
	counterDotInterval = 10
	counterNewLine     = counterDotInterval * 100
	lockCount          = 1000
	testDuration       = time.Minute * 10
)

func iteration() error {
	var wg sync.WaitGroup

	for i := 0; i < goroutinesCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var m sync.Mutex
			for k := 0; k < lockCount; k++ {
				m.Lock()
				m.Unlock()
			}

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
		runtime.GC()
		return nil
	case <-timer.C:
		return fmt.Errorf("timeout")
	}

}

func main() {
	fmt.Println("CPU:", runtime.GOMAXPROCS(0))
	start := time.Now()
	counter := 0

	http.HandleFunc("/", pprof.Index)
	go func() {
		http.ListenAndServe("localhost:8080", nil)
	}()

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
