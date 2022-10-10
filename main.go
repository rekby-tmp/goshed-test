package main

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
)

const (
	goroutinesCount    = 1000
	iterationTimeout   = time.Second * 10
	counterDotInterval = 10
	firstStat          = counterDotInterval
	statInterval       = counterDotInterval * 100
	schedCount         = 100
	testDuration       = time.Minute * 10
)

func iteration() error {
	var wg sync.WaitGroup

	for i := 0; i < goroutinesCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for k := 0; k < schedCount; k++ {
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
		return nil
	case <-timer.C:
		return fmt.Errorf("timeout")
	}

}

func printStat(d time.Duration) {
	fmt.Println()

	MB := uint64(1024 * 1024)
	swap, _ := mem.SwapMemory()
	fmt.Printf("Duration: %v\n", d)
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
			printStat(time.Since(iterationStart))
			panic(err)
		}

		counter++
		if counter%counterDotInterval == 0 {
			fmt.Print(".")
		}
		if counter%statInterval == 0 || counter == firstStat {
			printStat(time.Since(iterationStart))
		}
	}
}
