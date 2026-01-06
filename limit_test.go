package simultaneous_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/singlestore-labs/simultaneous"
)

const (
	threadCount = 1000
	max         = 10
	sleep       = time.Microsecond * 10
)

func TestLimitSilent(t *testing.T) {
	testLimit(t, false)
}

func TestLimitVerbose(t *testing.T) {
	testLimit(t, true)
}

func testLimit(t *testing.T, withStuck bool) {
	t.Parallel()

	var running int32
	var wg sync.WaitGroup
	limit := simultaneous.New[any](max)

	var stuckCalled atomic.Int32
	var unstuckCalled atomic.Int32
	someUnstuck := make(chan struct{})

	if withStuck {
		limit = limit.SetForeverMessaging(time.Millisecond,
			func(context.Context) {
				if stuckCalled.Add(1) == 1 {
					close(someUnstuck)
				}
			},
			func(context.Context) {
				unstuckCalled.Add(1)
			},
		)
	}

	var fail atomic.Int32
	var success atomic.Int32

	for i := 0; i < threadCount; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			var done simultaneous.Limited[any]
			switch i % 5 {
			case 0:
				done = limit.Forever(context.Background())
			case 1:
				var err error
				done, err = limit.Timeout(context.Background(), 0)
				if err != nil {
					fail.Add(1)
					return
				} else {
					success.Add(1)
				}
			case 2:
				var err error
				done, err = limit.Timeout(context.Background(), time.Second*2)
				if !assert.NoError(t, err) {
					return
				}
			case 3:
				var err error
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
				t.Cleanup(cancel)
				done, err = limit.Timeout(ctx, time.Hour)
				if !assert.NoError(t, err) {
					return
				}
			case 4:
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
				cancel()
				limit.Forever(ctx).Done()
				return
			}
			assert.LessOrEqual(t, atomic.AddInt32(&running, 1), int32(max))
			time.Sleep(sleep)
			assert.GreaterOrEqual(t, atomic.AddInt32(&running, -1), int32(0))
			if withStuck {
				<-someUnstuck
			}
			done.Done()
		}()
	}
	wg.Wait()

	if withStuck {
		t.Logf("stuck called %d unstuck called %d", stuckCalled.Load(), unstuckCalled.Load())
		assert.NotZero(t, stuckCalled.Load(), "stuck reported")
		assert.Equal(t, stuckCalled.Load(), unstuckCalled.Load(), "stuck == unstuck")
	}
	t.Logf("timeout 0 failed %d succeeded %d", fail.Load(), success.Load())
	assert.NotZero(t, fail.Load(), "fail")
	assert.NotZero(t, success.Load(), "succeed")
}
