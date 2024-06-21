package stressor

import (
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

// Automatic selection of the b.N does not work in this case, so you need
// to manually specify the number of iterations using the option -benchtime=Nx.
func BenchmarkStressor(b *testing.B) {
	stressor, _, err := prepareStressor()
	require.NoError(b, err)

	payloadConstantQuantity(b.N)

	stressor.Stop()
}

func prepareStressor() (*Stressor, time.Duration, error) {
	type config struct {
		AllocFactor    int           `env:"STRESSOR_ALLOC_FACTOR"`
		AllocSize      int           `env:"STRESSOR_ALLOC_SIZE"`
		LockFactor     int           `env:"STRESSOR_LOCK_FACTOR"`
		ScheduleFactor int           `env:"STRESSOR_SCHED_FACTOR"`
		ScheduleSleep  time.Duration `env:"STRESSOR_SCHED_SLEEP"`
		TestDuration   time.Duration `env:"STRESSOR_TEST_DURATION"`
	}

	cfg := config{}

	err := env.Parse(&cfg)
	if err != nil {
		return nil, 0, err
	}

	opts := Opts{
		AllocFactor:    cfg.AllocFactor,
		AllocSize:      cfg.AllocSize,
		LockFactor:     cfg.LockFactor,
		ScheduleFactor: cfg.ScheduleFactor,
		ScheduleSleep:  cfg.ScheduleSleep,
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
