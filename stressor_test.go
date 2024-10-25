package stressor

import (
	"runtime"
	"testing"
	"time"

	"github.com/caarlos0/env/v10"
	"github.com/stretchr/testify/require"
)

const (
	defaultTestDuration = 5 * time.Second
)

func TestStressor(t *testing.T) {
	stressor, duration, err := prepareStressor()
	require.NoError(t, err)

	defer stressor.Stop()

	time.Sleep(duration)
}

func TestStressorComparative(t *testing.T) {
	stressor, duration, err := prepareStressor()
	require.NoError(t, err)

	withLoad := payloadConstantTime(duration)

	stressor.Stop()

	withoutLoad := payloadConstantTime(duration)

	require.Greater(t, withoutLoad, withLoad)
}

func TestStressorQuickCompletion(t *testing.T) {
	stressor, _, err := prepareStressor()
	require.NoError(t, err)

	stressor.Stop()
}

func TestOptsNormalize(t *testing.T) {
	defaulted := Opts{
		Allocators:     runtime.NumCPU(),
		AllocationSize: DefaultAllocSize,
		Lockers:        runtime.NumCPU(),
		Scheduled:      runtime.NumCPU(),
		SleepDuration:  DefaultSleepDuration,
	}

	zero := Opts{}
	require.Equal(t, defaulted, zero.normalize())

	negative := Opts{
		Allocators:     -1,
		AllocationSize: -1,
		Lockers:        -1,
		Scheduled:      -1,
		SleepDuration:  -time.Nanosecond,
	}
	require.Equal(t, defaulted, negative.normalize())

	custom := Opts{
		Allocators:     10 * runtime.NumCPU(),
		AllocationSize: 2 * DefaultAllocSize,
		Lockers:        10 * runtime.NumCPU(),
		Scheduled:      10 * runtime.NumCPU(),
		SleepDuration:  2 * DefaultSleepDuration,
	}
	require.Equal(t, custom, custom.normalize())
}

func BenchmarkStressor(b *testing.B) {
	stressor, _, err := prepareStressor()
	require.NoError(b, err)

	payloadConstantQuantity(b.N)

	stressor.Stop()
}

func prepareStressor() (*Stressor, time.Duration, error) {
	type config struct {
		Allocators     int           `env:"STRESSOR_ALLOCATORS"`
		AllocationSize int           `env:"STRESSOR_ALLOCATION_SIZE"`
		Lockers        int           `env:"STRESSOR_LOCKERS"`
		Scheduled      int           `env:"STRESSOR_SCHEDULED"`
		SleepDuration  time.Duration `env:"STRESSOR_SLEEP_DURATION"`
		TestDuration   time.Duration `env:"STRESSOR_TEST_DURATION"`
	}

	cfg := config{}

	err := env.Parse(&cfg)
	if err != nil {
		return nil, 0, err
	}

	opts := Opts{
		Allocators:     cfg.Allocators,
		AllocationSize: cfg.AllocationSize,
		Lockers:        cfg.Lockers,
		Scheduled:      cfg.Scheduled,
		SleepDuration:  cfg.SleepDuration,
	}

	stressor := New(opts)

	if cfg.TestDuration == 0 {
		cfg.TestDuration = defaultTestDuration
	}

	return stressor, cfg.TestDuration, nil
}

func payloadConstantTime(duration time.Duration) int {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	counter := 0

	for {
		select {
		case <-timer.C:
			return counter
		default:
		}

		counter++
	}
}

func payloadConstantQuantity(quantity int) {
	counter := 0

	for range quantity {
		counter++
	}
}
