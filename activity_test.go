package activity

import (
	"testing"
	"time"
)

func TestThatItWorks(t *testing.T) {
	a := Counter{Tau: time.Minute}

	ts := time.Now()
	for i := 0; i < 1000; i++ {
		t.Log(a.Hz(), a)
		a.TickN(ts, 1)
		ts = ts.Add(time.Second)
	}

}
