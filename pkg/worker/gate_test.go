package worker

import (
	"testing"
	"time"
)

func TestAdvanceBackoff(t *testing.T) {
	const pollEvery = 2 * time.Second
	const max = 30 * time.Second
	wait, next := AdvanceBackoff(pollEvery, pollEvery, max, false)
	if wait != 0 || next != pollEvery {
		t.Fatalf("reset = %s %s", wait, next)
	}
	wait, next = AdvanceBackoff(pollEvery, pollEvery, max, true)
	if wait != pollEvery || next != 2*pollEvery {
		t.Fatalf("first stretch = %s %s", wait, next)
	}
	wait, next = AdvanceBackoff(16*time.Second, pollEvery, max, true)
	if wait != 16*time.Second || next != max {
		t.Fatalf("cap = %s %s", wait, next)
	}
	wait, next = AdvanceBackoff(max, pollEvery, max, true)
	if wait != max || next != max {
		t.Fatalf("stay capped = %s %s", wait, next)
	}
}

func TestGateSerializesRounds(t *testing.T) {
	gate := NewGate(2*time.Second, 30*time.Second)
	if !gate.Begin() {
		t.Fatal("first Begin rejected")
	}
	if gate.Begin() {
		t.Fatal("concurrent Begin accepted")
	}
	gate.Note(false)
	if !gate.Begin() {
		t.Fatal("Begin rejected after reset")
	}
	gate.Note(true)
	if gate.Begin() {
		t.Fatal("Begin accepted during backoff")
	}
}
