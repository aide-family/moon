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
	shellData    chan []byte
	shellClosed  chan string
	shellUID     string
	mu           sync.Mutex
	closed       chan struct{}
	closeOnce    sync.Once
}

// NewSession 创建一条会话。closed 用来通知等待方和发送方，Send 本身从不关闭。
func NewSession(agentUID, namespaceUID string) *Session {
	return &Session{
		AgentUID: agentUID, NamespaceUID: namespaceUID,
		Send:        make(chan *v1.AgentFrame, 128),
		Results:     map[string]chan *v1.AgentExecResult{},
		SFTPResults: map[string]chan *v1.AgentSFTPResult{},
		K8sResults:  map[string]chan *v1.AgentK8SResult{},
		closed:      make(chan struct{}),
	}
}

// Closed 在会话被替换或断开时关闭，供流发送协程退出。
func (s *Session) Closed() <-chan struct{} {
	return s.closed
}

func (s *Session) close() {
	s.closeOnce.Do(func() {
		close(s.closed)
		s.failPending("agent is offline")
	})
}

// failPending 让还在等待结果的请求立即失败。
// 进程退出时不能只关闭连接，否则 HTTP/SFTP 处理协程会等到自己的超时。
func (s *Session) failPending(reason string) {
	s.mu.Lock()
	execResults := s.Results
	sftpResults := s.SFTPResults
	k8sResults := s.K8sResults
	s.Results = map[string]chan *v1.AgentExecResult{}
	s.SFTPResults = map[string]chan *v1.AgentSFTPResult{}
	s.K8sResults = map[string]chan *v1.AgentK8SResult{}
	s.mu.Unlock()
	for uid, ch := range execResults {
		select {
		case ch <- &v1.AgentExecResult{RequestUid: uid, Error: reason}:
		default:
		}
		close(ch)
	}
	for uid, ch := range sftpResults {
		select {
		case ch <- &v1.AgentSFTPResult{RequestUid: uid, Error: reason}:
		default:
		}
		close(ch)
	}
	for uid, ch := range k8sResults {
		select {
		case ch <- &v1.AgentK8SResult{RequestUid: uid, Error: reason}:
		default:
		}
		close(ch)
	}
}

func (s *Session) send(frame *v1.AgentFrame) bool {
	select {
	case <-s.closed:
		return false
	default:
	}
	select {
	case s.Send <- frame:
		return true
	case <-s.closed:
		return false
	default:
		// 发送缓冲满时快速失败，避免请求线程一直等到 HTTP 超时。
		return false
	}
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
		old.finishShell("replaced")
		old.close()
	}
	h.sessions[session.AgentUID] = session
}

func (h *Hub) Remove(agentUID string, session *Session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if cur, ok := h.sessions[agentUID]; ok && cur == session {
		delete(h.sessions, agentUID)
		session.finishShell("disconnected")
		session.close()
	}
}

func (h *Hub) Get(agentUID string) (*Session, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	session, ok := h.sessions[agentUID]
	return session, ok
}

func (h *Hub) Online(agentUID string) bool {
	session, ok := h.Get(agentUID)
	if !ok {
		return false
	}
	select {
	case <-session.closed:
		return false
	default:
		return true
	}
}

// Drop 清理等待超时或取消后遗留的结果通道，避免会话内 map 无限增长。
func (h *Hub) Drop(agentUID, requestUID string) {
	session, ok := h.Get(agentUID)
	if !ok {
		return
	}
	session.mu.Lock()
	delete(session.Results, requestUID)
	delete(session.SFTPResults, requestUID)
	delete(session.K8sResults, requestUID)
	session.mu.Unlock()
}

// CloseAll 在进程退出前关闭全部存活流，避免 gRPC GracefulStop 一直等待。
func (h *Hub) CloseAll() {
	h.mu.Lock()
	sessions := make([]*Session, 0, len(h.sessions))
	for _, session := range h.sessions {
		sessions = append(sessions, session)
	}
	h.sessions = map[string]*Session{}
	h.mu.Unlock()
	// 先释放 Hub 锁，再逐个结束会话。发送通道和结果通道都由 Session 自己加锁，
	// 不能在持有 Hub 锁时等待它们，否则任一会话异常都会拖住整个退出流程。
	for _, session := range sessions {
		session.finishShell("shutdown")
		session.send(&v1.AgentFrame{Close: &v1.AgentClose{Reason: "shutdown"}})
		session.close()
	}
}

// SendExec 向在线 Agent 下发执行帧，并等待结果通道。
func (h *Hub) SendExec(agentUID string, frame *v1.AgentFrame) (<-chan *v1.AgentExecResult, error) {
	session, ok := h.Get(agentUID)
	if !ok || frame.GetExecRequest() == nil {
		return nil, merr.Invalid("agent is offline")
	}
	reqUID := frame.GetExecRequest().GetRequestUid()
	ch := make(chan *v1.AgentExecResult, 1)
	session.addExecResult(reqUID, ch)
	if !session.send(frame) {
		h.Drop(agentUID, reqUID)
		return nil, merr.Invalid("agent is offline")
	}
	return ch, nil
}

func (s *Session) addExecResult(requestUID string, ch chan *v1.AgentExecResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Results[requestUID] = ch
}

func (s *Session) addSFTPResult(requestUID string, ch chan *v1.AgentSFTPResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SFTPResults[requestUID] = ch
}

func (s *Session) addK8sResult(requestUID string, ch chan *v1.AgentK8SResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.K8sResults[requestUID] = ch
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
	session.addSFTPResult(reqUID, ch)
	if !session.send(frame) {
		h.Drop(agentUID, reqUID)
		return nil, merr.Invalid("agent is offline")
	}
	return ch, nil
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
	session.addK8sResult(reqUID, ch)
	if !session.send(frame) {
		h.Drop(agentUID, reqUID)
		return nil, merr.Invalid("agent is offline")
	}
	return ch, nil
}

// ClaimShell 占用这台机器上唯一的终端。新的登录会结束上一份终端。
func (h *Hub) ClaimShell(agentUID, sessionUID string) (<-chan []byte, <-chan string, error) {
	session, ok := h.Get(agentUID)
	if !ok {
		return nil, nil, merr.Invalid("agent is offline")
	}
	session.mu.Lock()
	defer session.mu.Unlock()
	if session.shellClosed != nil {
		select {
		case session.shellClosed <- "replaced":
		default:
		}
	}
	session.shellUID = sessionUID
	session.shellData = make(chan []byte, 64)
	session.shellClosed = make(chan string, 1)
	return session.shellData, session.shellClosed, nil
}

// PushFrame 把一帧写入存活流。
func (h *Hub) PushFrame(agentUID string, frame *v1.AgentFrame) error {
	session, ok := h.Get(agentUID)
	if !ok {
		return merr.Invalid("agent is offline")
	}
	select {
	case session.Send <- frame:
		return nil
	default:
		return merr.Invalid("agent is offline")
	}
}

func (s *Session) DeliverShell(sessionUID string, data []byte) {
	s.mu.Lock()
	ch := s.shellData
	match := s.shellUID == sessionUID
	s.mu.Unlock()
	if !match || ch == nil {
		return
	}
	select {
	case ch <- append([]byte(nil), data...):
	default:
	}
}

func (s *Session) FinishShell(sessionUID, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.shellUID != sessionUID || s.shellClosed == nil {
		return
	}
	select {
	case s.shellClosed <- reason:
	default:
	}
	s.shellData = nil
	s.shellClosed = nil
	s.shellUID = ""
}

func (s *Session) finishShell(reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.shellClosed == nil {
		return
	}
	select {
	case s.shellClosed <- reason:
	default:
	}
	s.shellData = nil
	s.shellClosed = nil
	s.shellUID = ""
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
