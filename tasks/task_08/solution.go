package main

import (
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
	l.mu.Lock()
	defer l.mu.Unlock()

	if limiterOFF := n <= 0 || l.burst <= 0 || n > l.burst || l.clock == nil; limiterOFF {
		return false
	}

	now := l.clock.Now()
	if now.After(l.last) {
		timePassed := now.Sub(l.last)
		tokensToAdd := l.rate * timePassed.Seconds()
		l.last = now

		l.tokens += tokensToAdd
		l.tokens = min(l.tokens, float64(l.burst))
	}

	nFloat := float64(n)
	if l.tokens >= nFloat {
		l.tokens -= nFloat
		return true
	}

	return false
}
