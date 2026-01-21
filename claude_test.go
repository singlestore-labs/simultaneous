package simultaneous_test

// This generated with Claude 3.7 Sonnet

import (
	"context"
	"testing"
	"time"

	"github.com/singlestore-labs/simultaneous"
	"github.com/stretchr/testify/assert"
)

// TestTimeoutRejection verifies that Timeout correctly rejects when the limit is full
func TestTimeoutRejection(t *testing.T) {
	t.Parallel()
	limit := simultaneous.New[any](1)

	// Take the only available slot
	done := limit.Forever(context.Background())
	defer done.Done()

	// This should time out immediately
	_, err := limit.Timeout(context.Background(), 0)
	assert.Error(t, err)
	t.Log("Timeout(0) correctly rejected when limit is full")

	// This should time out after a short wait
	start := time.Now()
	_, err = limit.Timeout(context.Background(), 50*time.Millisecond)
	duration := time.Since(start)

	assert.Error(t, err)
	assert.GreaterOrEqual(t, duration, 50*time.Millisecond)
	t.Log("Timeout(50ms) correctly waited then rejected when limit remained full")
}

// TestUnlimited verifies that Unlimited provides a way to bypass enforcement
func TestUnlimited(t *testing.T) {
	t.Parallel()

	// Set up a function that requires proof of limit
	called := false
	requireLimit := func(proof simultaneous.Enforced[string]) {
		called = true
	}

	// Call it with an unlimited proof
	unlimited := simultaneous.Unlimited[string]()
	requireLimit(unlimited)

	assert.True(t, called)
	t.Log("Unlimited token successfully passed type enforcement")
}
