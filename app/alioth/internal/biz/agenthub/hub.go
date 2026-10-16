package agenthub

import (
	"sync"

	"github.com/google/wire"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewHub)

// Session 是一条存活的 Agent 出站流。只存在于接受连接的进程里。
type Session struct {
	AgentUID     string
	NamespaceUID string
	Send         chan *v1.AgentFrame
	Results      map[string]chan *v1.AgentExecResult
	SFTPResults  map[string]chan *v1.AgentSFTPResult
	K8sResults   map[string]chan *v1.AgentK8SResult
	mu           sync.Mutex
}

// Hub 保存本进程上的存活流。
type Hub struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewHub() *Hub {
	return &Hub{sessions: map[string]*Session{}}
}

func (h *Hub) Put(session *Session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if old, ok := h.sessions[session.AgentUID]; ok {
		close(old.Send)
	}
	h.sessions[session.AgentUID] = session
}

func (h *Hub) Remove(agentUID string, session *Session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if cur, ok := h.sessions[agentUID]; ok && cur == session {
		delete(h.sessions, agentUID)
		close(session.Send)
	}
}

func (h *Hub) Get(agentUID string) (*Session, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	session, ok := h.sessions[agentUID]
	return session, ok
}

func (h *Hub) Online(agentUID string) bool {
	_, ok := h.Get(agentUID)
	return ok
}

// SendExec 向在线 Agent 下发执行帧，并等待结果通道。
func (h *Hub) SendExec(agentUID string, frame *v1.AgentFrame) (<-chan *v1.AgentExecResult, error) {
	session, ok := h.Get(agentUID)
	if !ok || frame.GetExecRequest() == nil {
		return nil, merr.Invalid("agent is offline")
	}
	reqUID := frame.GetExecRequest().GetRequestUid()
	ch := make(chan *v1.AgentExecResult, 1)
	session.mu.Lock()
	if session.Results == nil {
		session.Results = map[string]chan *v1.AgentExecResult{}
	}
	session.Results[reqUID] = ch
	session.mu.Unlock()
	select {
	case session.Send <- frame:
		return ch, nil
	default:
		session.mu.Lock()
		delete(session.Results, reqUID)
		session.mu.Unlock()
		return nil, merr.Invalid("agent is offline")
	}
}

func (s *Session) DeliverResult(result *v1.AgentExecResult) {
	s.mu.Lock()
	ch := s.Results[result.GetRequestUid()]
	delete(s.Results, result.GetRequestUid())
	s.mu.Unlock()
	if ch != nil {
		ch <- result
		close(ch)
	}
}

// SendSFTP 下发 SFTP 帧并等待结果。
func (h *Hub) SendSFTP(agentUID string, frame *v1.AgentFrame) (<-chan *v1.AgentSFTPResult, error) {
	session, ok := h.Get(agentUID)
	if !ok || frame.GetSftpRequest() == nil {
		return nil, merr.Invalid("agent is offline")
	}
	reqUID := frame.GetSftpRequest().GetRequestUid()
	ch := make(chan *v1.AgentSFTPResult, 1)
	session.mu.Lock()
	if session.SFTPResults == nil {
		session.SFTPResults = map[string]chan *v1.AgentSFTPResult{}
	}
	session.SFTPResults[reqUID] = ch
	session.mu.Unlock()
	select {
	case session.Send <- frame:
		return ch, nil
	default:
		session.mu.Lock()
		delete(session.SFTPResults, reqUID)
		session.mu.Unlock()
		return nil, merr.Invalid("agent is offline")
	}
}

func (s *Session) DeliverSFTP(result *v1.AgentSFTPResult) {
	s.mu.Lock()
	ch := s.SFTPResults[result.GetRequestUid()]
	delete(s.SFTPResults, result.GetRequestUid())
	s.mu.Unlock()
	if ch != nil {
		ch <- result
		close(ch)
	}
}

// SendK8s 下发 Kubernetes 代理帧并等待结果。
func (h *Hub) SendK8s(agentUID string, frame *v1.AgentFrame) (<-chan *v1.AgentK8SResult, error) {
	session, ok := h.Get(agentUID)
	if !ok || frame.GetK8SRequest() == nil {
		return nil, merr.Invalid("agent is offline")
	}
	reqUID := frame.GetK8SRequest().GetRequestUid()
	ch := make(chan *v1.AgentK8SResult, 1)
	session.mu.Lock()
	if session.K8sResults == nil {
		session.K8sResults = map[string]chan *v1.AgentK8SResult{}
	}
	session.K8sResults[reqUID] = ch
	session.mu.Unlock()
	select {
	case session.Send <- frame:
		return ch, nil
	default:
		session.mu.Lock()
		delete(session.K8sResults, reqUID)
		session.mu.Unlock()
		return nil, merr.Invalid("agent is offline")
	}
}

func (s *Session) DeliverK8s(result *v1.AgentK8SResult) {
	s.mu.Lock()
	ch := s.K8sResults[result.GetRequestUid()]
	delete(s.K8sResults, result.GetRequestUid())
	s.mu.Unlock()
	if ch != nil {
		ch <- result
		close(ch)
	}
}
