package agent

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/google/wire"
	"google.golang.org/grpc/metadata"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/access"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/agenthub"
	"github.com/aide-family/Alioth/app/alioth/internal/data/repository"
	"github.com/aide-family/Alioth/app/alioth/pkg/actor"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/app/alioth/pkg/page"
	"github.com/aide-family/Alioth/pkg/merr"
)

var ProviderSet = wire.NewSet(NewUsecase)

type shellTicket struct {
	agentUID string
	expires  time.Time
}

type Usecase struct {
	agents   repository.AgentRepo
	hub      *agenthub.Hub
	access   access.Checker
	ticketMu sync.Mutex
	tickets  map[string]shellTicket
}

func NewUsecase(agents repository.AgentRepo, hub *agenthub.Hub, checker access.Checker) *Usecase {
	return &Usecase{agents: agents, hub: hub, access: checker, tickets: map[string]shellTicket{}}
}

func (uc *Usecase) Create(ctx context.Context, in bo.Agent) (*bo.Agent, error) {
	ns, err := uc.access.Allow(ctx, "agent:write")
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	in.Name = strings.TrimSpace(in.Name)
	token := newAgentToken()
	item, err := uc.agents.CreateAgent(ctx, in, hashToken(token))
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	item.Token = token
	return item, nil
}

func (uc *Usecase) Update(ctx context.Context, in bo.Agent) (*bo.Agent, error) {
	ns, err := uc.access.Allow(ctx, "agent:write")
	if err != nil {
		return nil, err
	}
	in.NamespaceUID = ns
	item, err := uc.agents.UpdateAgent(ctx, in)
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) Rotate(ctx context.Context, uid string) (*bo.Agent, error) {
	ns, err := uc.access.Allow(ctx, "agent:write")
	if err != nil {
		return nil, err
	}
	token := newAgentToken()
	item, err := uc.agents.SetTokenHash(ctx, ns, uid, hashToken(token), "ACTIVE")
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	if session, ok := uc.hub.Get(uid); ok {
		uc.hub.Remove(uid, session)
	}
	item.Token = token
	return item, nil
}

func (uc *Usecase) Revoke(ctx context.Context, uid string) (*bo.Agent, error) {
	ns, err := uc.access.Allow(ctx, "agent:write")
	if err != nil {
		return nil, err
	}
	item, err := uc.agents.SetTokenHash(ctx, ns, uid, hashToken("revoked-"+uid), "REVOKED")
	if err != nil {
		return nil, uc.access.Wrap(err)
	}
	_ = uc.agents.SetOnline(ctx, uid, false, "", "", "", "", "")
	if session, ok := uc.hub.Get(uid); ok {
		select {
		case session.Send <- &v1.AgentFrame{Close: &v1.AgentClose{Reason: "revoked"}}:
		default:
		}
		uc.hub.Remove(uid, session)
	}
	return item, nil
}

func (uc *Usecase) Delete(ctx context.Context, uid string) error {
	ns, err := uc.access.Allow(ctx, "agent:write")
	if err != nil {
		return err
	}
	if session, ok := uc.hub.Get(uid); ok {
		uc.hub.Remove(uid, session)
	}
	return uc.access.Wrap(uc.agents.DeleteAgent(ctx, ns, uid))
}

func (uc *Usecase) Get(ctx context.Context, uid string) (*bo.Agent, error) {
	ns, err := uc.access.Allow(ctx, "agent:read")
	if err != nil {
		return nil, err
	}
	item, err := uc.agents.GetAgent(ctx, ns, uid)
	if item != nil {
		item.Online = uc.hub.Online(uid)
	}
	return item, uc.access.Wrap(err)
}

func (uc *Usecase) List(ctx context.Context, pageNo, pageSize int32, name string) ([]bo.Agent, int, int, int, error) {
	ns, err := uc.access.Allow(ctx, "agent:read")
	if err != nil {
		return nil, 0, 0, 0, err
	}
	p, s := page.Normalize(pageNo, pageSize)
	items, total, err := uc.agents.ListAgents(ctx, ns, p, s, strings.TrimSpace(name))
	for i := range items {
		items[i].Online = uc.hub.Online(items[i].UID)
	}
	return items, total, p, s, uc.access.Wrap(err)
}

// AuthenticateToken 用令牌哈希解析 Agent。
func (uc *Usecase) AuthenticateToken(ctx context.Context, token string) (*bo.Agent, error) {
	if !strings.HasPrefix(token, "ag-") {
		return nil, merr.Unauthenticated("invalid token")
	}
	item, err := uc.agents.GetAgentByTokenHash(ctx, hashToken(token))
	if err != nil {
		return nil, merr.Unauthenticated("invalid token")
	}
	return item, nil
}

// Connect 处理出站双向流。
// 双向流不会走 unary 中间件写入的上下文，所以这里从 metadata 再取一次令牌。
func (uc *Usecase) Connect(ctx context.Context, stream v1.AgentService_ConnectServer) error {
	token, _ := actor.AgentToken(ctx)
	if token == "" {
		token = bearerFromMetadata(ctx)
	}
	agent, err := uc.AuthenticateToken(ctx, token)
	if err != nil {
		return err
	}
	session := agenthub.NewSession(agent.UID, agent.NamespaceUID)
	uc.hub.Put(session)
	defer func() {
		uc.hub.Remove(agent.UID, session)
		_ = uc.agents.SetOnline(context.Background(), agent.UID, false, "", "", "", "", "")
	}()

	errCh := make(chan error, 2)
	go func() {
		for {
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			case <-session.Closed():
				errCh <- nil
				return
			case frame := <-session.Send:
				if err := stream.Send(frame); err != nil {
					errCh <- err
					return
				}
			}
		}
	}()
	go func() {
		for {
			frame, err := stream.Recv()
			if err == io.EOF {
				errCh <- nil
				return
			}
			if err != nil {
				errCh <- err
				return
			}
			if err := uc.handleFrame(ctx, agent, session, frame); err != nil {
				errCh <- err
				return
			}
		}
	}()
	return <-errCh
}

func (uc *Usecase) handleFrame(ctx context.Context, agent *bo.Agent, session *agenthub.Session, frame *v1.AgentFrame) error {
	current, err := uc.agents.GetAgentByUID(ctx, agent.UID)
	if err != nil {
		return err
	}
	if current.Status == "REVOKED" {
		_ = streamSendClose(session, "revoked")
		return merr.Forbidden("agent token revoked")
	}
	switch {
	case frame.GetRegister() != nil:
		reg := frame.GetRegister()
		return uc.agents.SetOnline(ctx, agent.UID, true, reg.GetHostname(), reg.GetMachineId(), reg.GetLocalIp(), reg.GetVersion(), "")
	case frame.GetHeartbeat() != nil:
		return uc.agents.SetOnline(ctx, agent.UID, true, "", "", "", "", "")
	case frame.GetMachineInfo() != nil:
		info := frame.GetMachineInfo()
		return uc.agents.SetOnline(ctx, agent.UID, true, info.GetHostname(), info.GetMachineId(), info.GetLocalIp(), info.GetVersion(), info.GetCollectJson())
	case frame.GetExecResult() != nil:
		session.DeliverResult(frame.GetExecResult())
		return nil
	case frame.GetShellData() != nil:
		session.DeliverShell(frame.GetShellData().GetSessionUid(), frame.GetShellData().GetData())
		return nil
	case frame.GetShellClose() != nil:
		session.FinishShell(frame.GetShellClose().GetSessionUid(), frame.GetShellClose().GetReason())
		return nil
	default:
		return nil
	}
}

func streamSendClose(session *agenthub.Session, reason string) error {
	select {
	case session.Send <- &v1.AgentFrame{Close: &v1.AgentClose{Reason: reason}}:
	default:
	}
	return nil
}

func bearerFromMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	for _, header := range md.Get("authorization") {
		header = strings.TrimSpace(header)
		if len(header) >= 7 && strings.EqualFold(header[:7], "bearer ") {
			return strings.TrimSpace(header[7:])
		}
	}
	return ""
}

func newAgentToken() string {
	buf := make([]byte, 24)
	_, _ = rand.Read(buf)
	return "ag-" + hex.EncodeToString(buf)
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
