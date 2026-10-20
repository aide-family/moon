package agenthub

import (
	"testing"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
)

func TestPutReplacesAndClosesOldSession(t *testing.T) {
	hub := NewHub()
	first := NewSession("agent-1", "ns-1")
	hub.Put(first)
	second := NewSession("agent-1", "ns-1")
	hub.Put(second)

	select {
	case <-first.Closed():
	default:
		t.Fatal("replaced session was not closed")
	}
	if !hub.Online("agent-1") {
		t.Fatal("new session should be online")
	}
	hub.Remove("agent-1", second)
	if hub.Online("agent-1") {
		t.Fatal("removed session is still online")
	}
}

func TestSendExecAfterRemoveFailsWithoutPanic(t *testing.T) {
	hub := NewHub()
	session := NewSession("agent-1", "ns-1")
	hub.Put(session)
	hub.Remove("agent-1", session)

	_, err := hub.SendExec("agent-1", &v1.AgentFrame{ExecRequest: &v1.AgentExecRequest{RequestUid: "req-1"}})
	if err == nil {
		t.Fatal("send to removed agent succeeded")
	}
}

func TestDropRemovesPendingResult(t *testing.T) {
	hub := NewHub()
	session := NewSession("agent-1", "ns-1")
	hub.Put(session)
	request := &v1.AgentExecRequest{RequestUid: "req-1"}
	if _, err := hub.SendExec("agent-1", &v1.AgentFrame{ExecRequest: request}); err != nil {
		t.Fatal(err)
	}
	hub.Drop("agent-1", "req-1")
	session.DeliverResult(&v1.AgentExecResult{RequestUid: "req-1"})

	session.mu.Lock()
	_, exists := session.Results["req-1"]
	session.mu.Unlock()
	if exists {
		t.Fatal("dropped request is still registered")
	}
}
