// Library that provides to imposes a load on the system and the runtime in order to
// provide the main code with as little processor time as possible.
package stressor

import (
	"runtime"
	"sync"
	"time"

	"github.com/akramarenkov/breaker"
	"github.com/akramarenkov/starter"
)

const (
	DefaultAllocationSize = 1
	DefaultSleepDuration  = time.Nanosecond
)

// Options of the created Stressor instance.
//
// With some parameters, the duration of the program execution may become
// indefinitely long.
type Opts struct {
	// Number of goroutines performing memory allocation. When specifying a negative or
	// zero value, the value returned by [runtime.NumCPU] will be used. Loads the
	// garbage collector
	Allocators int

	// Size of memory allocated by goroutines. When specifying a negative or zero value,
	// the value of [DefaultAllocSize] will be used
	AllocationSize int

	// Number of goroutines pairs performing reads and writes to the channels. When
	// specifying a negative or zero value, the value returned by [runtime.NumCPU] will
	// be used. Loads by empty wait loops and futex calls
	Lockers int

	// Number of goroutines that calls [time.Sleep]. When specifying a negative or
	// zero value, the value returned by [runtime.NumCPU] will be used. Loads the
	// scheduler
	Scheduled int

	// Sleep duration of scheduled goroutines. When specifying a negative or
	// zero value, the value of [DefaultSleepDuration] will be used
	SleepDuration time.Duration
}

func (opts Opts) normalize() Opts {
	if opts.Allocators <= 0 {
		opts.Allocators = runtime.NumCPU()
	}

	if opts.AllocationSize <= 0 {
		opts.AllocationSize = DefaultAllocationSize
	}

	if opts.Lockers <= 0 {
		opts.Lockers = runtime.NumCPU()
	}

	if opts.Scheduled <= 0 {
		opts.Scheduled = runtime.NumCPU()
	}

	if opts.SleepDuration <= 0 {
		opts.SleepDuration = DefaultSleepDuration
	}

	return opts
}

// Imposes a load on the system and the runtime in order to provide the main
// code with as little processor time as possible.
//
// This is a very simple implementation that does not adapt to performance and
// the features of the system and runtime.
type Stressor struct {
	opts Opts

	breaker *breaker.Breaker
	onLoad  chan struct{}
	starter *starter.Starter
	wg      sync.WaitGroup
}

// Creates and run Stressor instance.
//
// Completion of the function means that the load is already present. To be completely
// sure of this, you can take a short pause before running the main code.
func New(opts Opts) *Stressor {
	opts = opts.normalize()

	strs := &Stressor{
		opts: opts,

		breaker: breaker.New(),
		onLoad:  make(chan struct{}),
		starter: starter.New(),
	}

	go strs.main()

	strs.waitLoad()

	return strs
}

func (strs *Stressor) waitLoad() {
	<-strs.onLoad
}

// Terminates work of the Stressor.
//
// The load may persist for some time due to garbage collection.
func (strs *Stressor) Stop() {
	strs.breaker.Break()
}

func (strs *Stressor) main() {
	defer strs.breaker.Complete()

	strs.loop()
}

func (strs *Stressor) loop() {
	for range strs.opts.Allocators {
		strs.wg.Add(1)
		strs.starter.Ready()

		go strs.allocator()
	}

	for range strs.opts.Lockers {
		forward := make(chan int)
		backward := make(chan int)

		strs.wg.Add(1)
		strs.wg.Add(1)

		strs.starter.Ready()
		strs.starter.Ready()

		go strs.forwarder(forward, backward)
		go strs.backwarder(forward, backward)
	}

	for range strs.opts.Scheduled {
		strs.wg.Add(1)
		strs.starter.Ready()

		go strs.scheduled()
	}

	strs.starter.Go()

	close(strs.onLoad)
}

func (strs *Stressor) allocator() {
	defer strs.wg.Done()

	strs.starter.Set()

	for !strs.breaker.IsStopped() {
		_ = make([]byte, strs.opts.AllocationSize)
	}
}

func (strs *Stressor) forwarder(forward, backward chan int) {
	defer strs.wg.Done()

	strs.starter.Set()

	select {
	case <-strs.breaker.IsBreaked():
		return
	case forward <- 0:
	}

	for {
		select {
		case <-strs.breaker.IsBreaked():
			return
		case item := <-backward:
			select {
			case <-strs.breaker.IsBreaked():
				return
			case forward <- item:
			}
		}
	}
}

func (strs *Stressor) backwarder(forward, backward chan int) {
	defer strs.wg.Done()

	strs.starter.Set()

	for {
		select {
		case <-strs.breaker.IsBreaked():
			return
		case item := <-forward:
			select {
			case <-strs.breaker.IsBreaked():
				return
			case backward <- item:
			}
		}
	}
}

func (strs *Stressor) scheduled() {
	defer strs.wg.Done()

	strs.starter.Set()

	for !strs.breaker.IsStopped() {
		time.Sleep(strs.opts.SleepDuration)
	}
}
