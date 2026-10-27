package main

import (
	// "fmt"
	"sync"
	"time"
)

type Clock interface{ Now() time.Time }
type Limiter struct {
	mu     sync.Mutex
	clock  Clock
	rate   float64
	burst  int
	tokens float64
	last   time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	l := &Limiter{
		clock: clock,
		rate:  max(0, ratePerSec),
		burst: burst,
	}
	if clock != nil {
		l.tokens = float64(burst)
		l.last = clock.Now()
	}

	return l
}
func (l *Limiter) AllowN(n int) bool {
	if limiterOFF := n <= 0 || l.burst <= 0 || n > l.burst || l.clock == nil; limiterOFF {
		return false
	}

	timePassed := l.clock.Now().Sub(l.last)
	tokensToAdd := l.rate * timePassed.Seconds()
	// fmt.Printf("tokens: %v, tokensToAdd: %v\n", l.tokens, tokensToAdd)
	if l.tokens+tokensToAdd < float64(l.burst) {
		l.tokens += tokensToAdd
	} else {
		l.tokens = float64(l.burst)
	}

	l.last = l.clock.Now()

	nFloat := float64(n)
	// fmt.Printf("tokens to substract: %v\n", nFloat)
	if l.tokens >= nFloat {
		l.tokens -= nFloat
		// fmt.Printf("tokens: %v\n", l.tokens)
		return true
	}

	return false
}
