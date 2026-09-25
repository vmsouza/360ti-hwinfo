package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

type stepRec struct {
	label string
	d     time.Duration
}

var (
	stepsMu          sync.Mutex
	steps            []stepRec
	timingLogEnabled bool
)

func recordStep(label string, d time.Duration) {
	stepsMu.Lock()
	steps = append(steps, stepRec{label, d})
	stepsMu.Unlock()
}

func timedStep(label string, fn func()) {
	start := time.Now()
	fn()
	recordStep(label, time.Since(start))
}

// writeTimingLog appends the per-step timings to "360ti-hwinfo-timing.log"
// in the current working directory (only when -timing is enabled).
func writeTimingLog() {
	if !timingLogEnabled {
		return
	}
	stepsMu.Lock()
	defer stepsMu.Unlock()

	f, err := os.OpenFile("360ti-hwinfo-timing.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "--- %s ---\n", time.Now().Format("2006-01-02 15:04:05"))
	total := time.Duration(0)
	for _, s := range steps {
		total += s.d
		fmt.Fprintf(f, "%-30s %s\n", s.label, s.d.Round(time.Millisecond))
	}
	fmt.Fprintf(f, "TOTAL                     %s\n", total.Round(time.Millisecond))
	steps = nil
}