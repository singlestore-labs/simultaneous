package simultaneous_test

import (
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

func TestLimit(t *testing.T) {
	t.Parallel()

	var running int32
	var wg sync.WaitGroup
	limit := simultaneous.New[any](max)

	for i := 0; i < threadCount; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			var done simultaneous.Limited[any]
			if i%2 == 0 {
				done = limit.Forever()
			} else {
				var err error
				done, err = limit.Timeout(time.Second * 2)
				if !assert.NoError(t, err) {
					return
				}
			}
			assert.LessOrEqual(t, atomic.AddInt32(&running, 1), int32(max))
			time.Sleep(sleep)
			done.Done()
			assert.GreaterOrEqual(t, atomic.AddInt32(&running, -1), int32(0))
		}()
	}
	wg.Wait()
}
