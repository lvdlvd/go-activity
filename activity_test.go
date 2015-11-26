package activity

import (
	"testing"
	"time"
)

func TestThatItWorks(t *testing.T) {
	a := Counter{Tau: time.Minute}

	ts := time.Now()
	for i := 0; i < 1000; i++ {
		t.Logf("%d hz: %v  a:%q  next:%v", i, a.Hz(), a, a.NextExpected(ts))
		a.TickN(ts, 1)
		ts = ts.Add(time.Second)
	}

}
