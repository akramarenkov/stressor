package stressor_test

import (
	"time"

	"github.com/akramarenkov/stressor"
)

func ExampleStressor() {
	opts := stressor.Opts{
		AllocFactor:    1,
		AllocSize:      1,
		LockFactor:     1,
		ScheduleFactor: 1,
		ScheduleSleep:  time.Nanosecond,
	}

	stressor := stressor.New(opts)
	defer stressor.Stop()

	// Main code
	time.Sleep(time.Second)

	// Output:
}
