/*
Package simultaneous exists to place a limit on simultaneous actions that
need a limit.
*/
package simultaneous

import (
	"time"

	"github.com/memsql/errors"
)

// Limited is a type to take as a parameter so that the type system enforces
// that a reservation has been taken and limits are obeyed.
type Limited[T any] interface {
	Enforced[T]
	Done()
}

// Enforced is a type that exists just to signal that a simultaneous limit
// is being enforced. When passing a Limited as a argument, have the receiver
// take an Enforced instead.
type Enforced[T any] interface {
	privateMethod()
}

// Limit implements Enforced so it can be used to fulfill the Enforced
// contract.
type Limit[T any] struct {
	queue           chan struct{}
	stuckCallback   func()
	unstuckCallback func()
	stuckTimeout    time.Duration
}

// New takes both a type and a count. The type is so that if the limit is passed
// around it can be done so with type safety so that a limit of one kind of thing
// cannot be used as limit of another kind of thing. If you're not passing the
// resulting limit around, then the type argument can be anything. Like "string".
func New[T any](limit int) *Limit[T] {
	return &Limit[T]{
		queue: make(chan struct{}, limit),
	}
}

// Unlimited provides a way to bypass enforcement
func Unlimited[T any]() Enforced[T] {
	return &unlimited[T]{}
}

// Forever waits until there is space in the Limit for another
// simultaneous runner. It will wait forever. The Done() method
// must be called to release the space.
//
//	defer limit.Forever().Done()
func (l *Limit[T]) Forever() Limited[T] {
	if l.stuckTimeout == 0 {
		l.queue <- struct{}{}
	} else {
		timer := time.NewTimer(l.stuckTimeout)
		select {
		case l.queue <- struct{}{}:
			timer.Stop()
		case <-timer.C:
			if l.stuckCallback != nil {
				l.stuckCallback()
			}
			l.queue <- struct{}{}
			if l.unstuckCallback != nil {
				l.unstuckCallback()
			}
		}
	}
	return limited[T](func() {
		<-l.queue
	})
}

var ErrTimeout errors.String = "could not get permission to run before timeout"

// Timeout waits for a limited time for there to be space for another
// simultaneous runner. In the case of a timeout, ErrTimeout is returned
// and the Done method is a no-op. If there is room, the Done method must
// be invoked to make room for another runner.
func (l *Limit[T]) Timeout(timeout time.Duration) (Limited[T], error) {
	if timeout <= 0 {
		select {
		case l.queue <- struct{}{}:
			return limited[T](func() {
				<-l.queue
			}), nil
		default:
			return limited[T](nil), ErrTimeout.Errorf("timeout (%s) expired before any simultaneous runner (of %d) became available", timeout, cap(l.queue))
		}
	}
	timer := time.NewTimer(timeout)
	select {
	case l.queue <- struct{}{}:
		timer.Stop()
		return limited[T](func() {
			<-l.queue
		}), nil
	case <-timer.C:
		return limited[T](nil), ErrTimeout.Errorf("timeout (%s) expired before any simultaneous runner (of %d) became available", timeout, cap(l.queue))
	}
}

// SetForeverMessaging returns a modified Limit that changes the behavior of Forever() so that
// it will call stuckCallback() (if set) after waiting for stuckTimeout duration. If past that duration,
// and it will call unstuckCallback() (if set) when it finally gets a limit.
func (l Limit[T]) SetForeverMessaging(stuckTimeout time.Duration, stuckCallback func(), unstuckCallback func()) *Limit[T] {
	l.stuckTimeout = stuckTimeout
	l.stuckCallback = stuckCallback
	l.unstuckCallback = unstuckCallback
	return &l
}

var (
	_ Limited[any]  = limited[any](nil)
	_ Enforced[any] = limited[any](nil)
	_ Enforced[any] = unlimited[any]{}
)

type limited[T any] func()

func (l limited[T]) privateMethod() {}
func (l limited[T]) Done() {
	if l != nil {
		l()
	}
}

type unlimited[T any] struct{}

func (u unlimited[T]) privateMethod() {}
