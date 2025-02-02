package stressor_test

import (
	"time"

	"github.com/akramarenkov/stressor"
)

func ExampleStressor() {
	opts := stressor.Opts{
		Allocators:     1,
		AllocationSize: 2,
		Lockers:        1,
		Scheduled:      1,
		SleepDuration:  10 * time.Nanosecond,
	}

	strain := stressor.New(opts)
	defer strain.Stop()

	// Main code
	time.Sleep(time.Second)
	// Output:
}
