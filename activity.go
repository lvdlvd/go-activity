// Package activity provides a decaying activity counter.
//
// A decaying counter is appropriate to estimate the best candidate to evict
// from, say, a cache, assuming the event stream is some Poisson process.
//
// TODO redo the correction factors.
package activity

import (
	"fmt"
	"math"
	"time"
)

// Counter contains a current Value and a Timestamp.
// To initialize, all you need to provide is Tau, eg
//
//   a := Counter{Tau: time.Hour}
//
// Good choices or Tau are larger than the period
// with which you expect events to happen and small enough that your
// application adapts to changing circumstances.
// Eg. if you have an event stream of several events per second, and
// you have to make decisions about evicting a cache every few minutes,
// then a tau of a few minutes makes sense.
type Counter struct {
	Tau       time.Duration // Characteristic (filter decay) time
	Value     float64       // Decaying value
	Timestamp time.Time     // Time of last Tick
}

// Hz returns the current estimated frequency (in Hertz)
func (a Counter) Hz() float64 {
	return a.Value * float64(time.Second) / float64(a.Tau)
}

// String produces a nicely readable presentation.
func (a Counter) String() string {
	if a.Value == 0 {
		return "\u221E (0 Hz)"
	}
	n, d := a.Value*float64(time.Second), float64(a.Tau)
	if n > d {
		return fmt.Sprintf("%v (%.2g Hz)", time.Duration(d/a.Value), n/d)
	}
	return fmt.Sprintf("%v", time.Duration(d/a.Value))
}

// Update the decaying activity counter by 1 event at now.
func (a *Counter) Tick() { a.TickN(time.Now(), 1) }

// Update the decaying activity counter by N
func (a *Counter) TickN(ts time.Time, N int) {
	delta := float64(ts.Sub(a.Timestamp)) / float64(a.Tau)
	if delta >= 0 {
		a.Value *= math.Exp(-delta)
		a.Value += float64(N)
		a.Timestamp = ts
	} else {
		// out of order event: discount and add but leave lastmod unchanged
		a.Value += math.Exp(delta) * float64(N)
	}
}

// NextExpected returns duration (relative to now) until the next expected event
// given the state of a, the fact that no new event happened until
// now, and assuming the event source is a Poisson process with a mean
// interval time much smaller than our Tau.
func (a Counter) NextExpected(now time.Time) time.Duration {
	if a.Value <= 0 {
		return time.Duration(math.MaxInt64)
	}
	delta := float64(now.Sub(a.Timestamp)) / float64(a.Tau)          // discount activity till now
	return time.Duration(math.Exp(delta) * float64(a.Tau) / a.Value) // tau/(value * exp(-d))
}

// Add returns the Activity that would be the result of the two event streams that lead to a and b combined.
// This is only really meaningful if a.Tau == b.Tau, but we will return an Activity with the min(Tau) of the two
// and apply a plausible correction factor.
func Add(a, b Counter) Counter {
	if a.Tau > b.Tau {
		a, b = b, a
	}
	delta := float64(b.Timestamp.Sub(a.Timestamp)) / float64(a.Tau)
	if delta >= 0 {
		a.Value *= math.Exp(-delta)
		a.Value += float64(a.Tau) * b.Value / float64(b.Tau)
		a.Timestamp = b.Timestamp
	} else {
		a.Value += math.Exp(delta) * float64(a.Tau) * b.Value / float64(b.Tau)
	}
	return a
}
