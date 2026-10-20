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

func TestSendSFTPAndK8sRegisterResults(t *testing.T) {
	hub := NewHub()
	session := NewSession("agent-1", "ns-1")
	hub.Put(session)

	sftpCh, err := hub.SendSFTP("agent-1", &v1.AgentFrame{SftpRequest: &v1.AgentSFTPRequest{RequestUid: "sftp-1"}})
	if err != nil {
		t.Fatalf("send sftp: %v", err)
	}
	session.DeliverSFTP(&v1.AgentSFTPResult{RequestUid: "sftp-1", Ok: true})
	if result := <-sftpCh; !result.GetOk() {
		t.Fatal("sftp result was not delivered")
	}

	k8sCh, err := hub.SendK8s("agent-1", &v1.AgentFrame{K8SRequest: &v1.AgentK8SRequest{RequestUid: "k8s-1"}})
	if err != nil {
		t.Fatalf("send k8s: %v", err)
	}
	session.DeliverK8s(&v1.AgentK8SResult{RequestUid: "k8s-1", StatusCode: 200})
	if result := <-k8sCh; result.GetStatusCode() != 200 {
		t.Fatal("k8s result was not delivered")
	}
}

func TestCloseAllClosesEverySession(t *testing.T) {
	hub := NewHub()
	first := NewSession("agent-1", "ns-1")
	second := NewSession("agent-2", "ns-2")
	hub.Put(first)
	hub.Put(second)
	hub.CloseAll()
	select {
	case <-first.Closed():
	default:
		t.Fatal("first session was not closed")
	}
	select {
	case <-second.Closed():
	default:
		t.Fatal("second session was not closed")
	}
	if len(hub.sessions) != 0 {
		t.Fatal("close all left sessions behind")
	}
}

func TestCloseFailsPendingResults(t *testing.T) {
	hub := NewHub()
	session := NewSession("agent-1", "ns-1")
	hub.Put(session)
	execCh, err := hub.SendExec("agent-1", &v1.AgentFrame{ExecRequest: &v1.AgentExecRequest{RequestUid: "exec-1"}})
	if err != nil {
		t.Fatal(err)
	}
	sftpCh, err := hub.SendSFTP("agent-1", &v1.AgentFrame{SftpRequest: &v1.AgentSFTPRequest{RequestUid: "sftp-1"}})
	if err != nil {
		t.Fatal(err)
	}
	k8sCh, err := hub.SendK8s("agent-1", &v1.AgentFrame{K8SRequest: &v1.AgentK8SRequest{RequestUid: "k8s-1"}})
	if err != nil {
		t.Fatal(err)
	}
	hub.Remove("agent-1", session)

	if result := <-execCh; result.GetError() == "" {
		t.Fatal("pending exec result was not failed")
	}
	if result := <-sftpCh; result.GetError() == "" {
		t.Fatal("pending sftp result was not failed")
	}
	if result := <-k8sCh; result.GetError() == "" {
		t.Fatal("pending k8s result was not failed")
	}
}
