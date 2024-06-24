// Imposes a load on the system and the runtime in order to provide the main
// code with as little processor time as possible.
package stressor

import (
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/akramarenkov/breaker"
	"github.com/akramarenkov/starter"
)

const (
	lockersGroupNumber = 2
)

// Just used the minimum values.
const (
	DefaultAllocFactor    = 1
	DefaultAllocSize      = int(unsafe.Sizeof(int(0)))
	DefaultLockFactor     = 1
	DefaultScheduleFactor = 1
	DefaultScheduleSleep  = time.Nanosecond
)

// Options of the created Stressor instance.
//
// With some parameters, the duration of the program execution may become
// indefinitely long.
type Opts struct {
	// Determines how many times there will be more goroutines performing memory
	// allocations than logical processors. Loads the garbage collector
	AllocFactor int
	// Size of memory allocated by goroutines
	AllocSize int
	// Determines how many times there will be more goroutines performing reads and
	// writes to the channels than logical processors. Loads with empty wait loops
	// and futex calls
	LockFactor int
	// Determines how many times there will be more goroutines that calls time.Sleep()
	// than logical processors. Loads the scheduler
	ScheduleFactor int
	// Goroutines sleep duration
	ScheduleSleep time.Duration
}

func (opts Opts) normalize() Opts {
	if opts.AllocFactor == 0 {
		opts.AllocFactor = DefaultAllocFactor
	}

	if opts.AllocSize == 0 {
		opts.AllocSize = DefaultAllocSize
	}

	if opts.LockFactor == 0 {
		opts.LockFactor = DefaultLockFactor
	}

	if opts.ScheduleFactor == 0 {
		opts.ScheduleFactor = DefaultScheduleFactor
	}

	if opts.ScheduleSleep == 0 {
		opts.ScheduleSleep = DefaultScheduleSleep
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
	wg := &sync.WaitGroup{}
	defer wg.Wait()

	starter := starter.New()

	allocators := strs.opts.AllocFactor * runtime.NumCPU()

	lockers := divideWithMin(
		strs.opts.LockFactor*runtime.NumCPU(),
		lockersGroupNumber,
		1,
	)

	planned := strs.opts.ScheduleFactor * runtime.NumCPU()

	for range allocators {
		wg.Add(1)
		starter.Ready()

		go strs.allocator(wg, starter)
	}

	for range lockers {
		wg.Add(lockersGroupNumber)
		starter.ReadyN(lockersGroupNumber)

		forward := make(chan int)
		backward := make(chan int)

		go strs.forwarder(wg, starter, forward, backward)
		go strs.backwarder(wg, starter, forward, backward)
	}

	for range planned {
		wg.Add(1)
		starter.Ready()

		go strs.planned(wg, starter)
	}

	starter.Go()

	close(strs.onLoad)
}

func (strs *Stressor) allocator(
	wg *sync.WaitGroup,
	starter *starter.Starter,
) {
	defer wg.Done()

	starter.Set()

	for !strs.breaker.IsStopped() {
		_ = make([]byte, strs.opts.AllocSize)
	}
}

func (strs *Stressor) forwarder(
	wg *sync.WaitGroup,
	starter *starter.Starter,
	forward chan int,
	backward chan int,
) {
	defer wg.Done()

	starter.Set()

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

func (strs *Stressor) backwarder(
	wg *sync.WaitGroup,
	starter *starter.Starter,
	forward chan int,
	backward chan int,
) {
	defer wg.Done()

	starter.Set()

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

func (strs *Stressor) planned(
	wg *sync.WaitGroup,
	starter *starter.Starter,
) {
	defer wg.Done()

	starter.Set()

	for !strs.breaker.IsStopped() {
		time.Sleep(strs.opts.ScheduleSleep)
	}
}
