package notify

import (
	"testing"
	"time"
)

func TestAdvanceBackoff(t *testing.T) {
	wait, next := advanceBackoff(2*time.Second, false)
	if wait != 0 || next != pollEvery {
		t.Fatalf("reset = %s %s", wait, next)
	}
	wait, next = advanceBackoff(2*time.Second, true)
	if wait != 2*time.Second || next != 4*time.Second {
		t.Fatalf("first stretch = %s %s", wait, next)
	}
	wait, next = advanceBackoff(16*time.Second, true)
	if wait != 16*time.Second || next != backoffMax {
		t.Fatalf("cap = %s %s", wait, next)
	}
	wait, next = advanceBackoff(backoffMax, true)
	if wait != backoffMax || next != backoffMax {
		t.Fatalf("stay capped = %s %s", wait, next)
	}
}
