package ratelimit

import (
	"time"
)

type Limiter struct {
	rate      int
	last      time.Time
	allowance float64
}

// Create new rate limiter that limits at rate/sec
func New(rate int) *Limiter {
	ret := &Limiter{
		rate:      rate,
		last:      time.Now(),
		allowance: float64(rate),
	}
	return ret
}

// Return true if the current call exceeds the set rate, false otherwise
func (r *Limiter) Limit() bool {
	if r.rate == 0 {
		return false
	}

	rate := float64(r.rate)
	now := time.Now()
	elapsed := now.Sub(r.last)
	r.last = now
	r.allowance += float64(elapsed) * rate

	// Clamp number of tokens in the bucket. Don't let it get
	// unboundedly large
	if r.allowance > rate {
		r.allowance = rate
	}

	if r.allowance < 1.0 {
		return true
	}

	r.allowance -= 1.0
	return false
}
