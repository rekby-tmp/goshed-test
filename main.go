package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/pprof"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

const (
	goroutinesCount    = 2000
	iterationTimeout   = time.Minute
	counterDotInterval = 10
	statInterval       = counterDotInterval * 100
	lockCount          = 1000
	testDuration       = time.Minute * 10
)

var (
	fixedStat = map[int]bool{
		1:                  true,
		50:                 true,
		100:                true,
		counterDotInterval: true,
	}
)

func iteration() error {
	var wg sync.WaitGroup
	var m sync.Mutex
	var s int

	for i := 0; i < goroutinesCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for k := 0; k < lockCount; k++ {
				m.Lock()
				s++
				m.Unlock()
				runtime.Gosched()
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
		_, _ = io.Discard.Write([]byte(fmt.Sprintf("%d", s)))
		return nil
	case <-timer.C:
		return fmt.Errorf("timeout")
	}

}

func printStat(counter int, d time.Duration) {
	fmt.Println()

	MB := uint64(1024 * 1024)
	swap, _ := mem.SwapMemory()
	fmt.Printf("Counter: %d, Duration: %v\n", counter, d)
	fmt.Printf("swap free: %d, used: %d\n", swap.Free/MB, swap.Used/MB)
	virt, _ := mem.VirtualMemory()
	fmt.Printf("mem free: %d, used: %d\n", virt.Free/MB, virt.Used/MB)
	fmt.Println()
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
		iterationStart := time.Now()
		if iterationStart.Sub(start) > testDuration {
			fmt.Println()
			fmt.Println("OK")
			return
		}
		if err := iteration(); err != nil {
			fmt.Println()
			fmt.Printf("\ncounter: %d\n", counter)
			printStat(counter, time.Since(iterationStart))
			panic(err)
		}

		counter++
		if counter%counterDotInterval == 0 {
			fmt.Print(".")
		}
		if counter%statInterval == 0 || fixedStat[counter] {
			printStat(counter, time.Since(iterationStart))
		}
	}
}
