package agent

import (
	"context"
	"time"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/pkg/id"
	"github.com/aide-family/Alioth/pkg/merr"
)

const shellTicketTTL = 60 * time.Second

// ShellLink 是浏览器和某台机器之间的一条终端。
type ShellLink struct {
	SessionUID string
	Data       <-chan []byte
	Closed     <-chan string
	write      func([]byte) error
	resize     func(cols, rows int32) error
	stop       func()
}

func (l *ShellLink) Write(data []byte) error {
	if l == nil || l.write == nil {
		return merr.Invalid("agent is offline")
	}
	return l.write(data)
}

func (l *ShellLink) Resize(cols, rows int32) error {
	if l == nil || l.resize == nil {
		return merr.Invalid("agent is offline")
	}
	return l.resize(cols, rows)
}

func (l *ShellLink) Stop() {
	if l != nil && l.stop != nil {
		l.stop()
	}
}

// IssueShellTicket 给在线机器签发一次性终端票据。
func (uc *Usecase) IssueShellTicket(ctx context.Context, uid string) (string, error) {
	ns, err := uc.access.Allow(ctx, "ssh:exec")
	if err != nil {
		return "", err
	}
	machine, err := uc.agents.GetAgent(ctx, ns, uid)
	if err != nil {
		return "", uc.access.Wrap(err)
	}
	if machine.Status != "ACTIVE" {
		return "", merr.Forbidden("agent is revoked")
	}
	if !uc.hub.Online(uid) {
		return "", merr.Invalid("agent is offline")
	}
	ticket := id.New()
	uc.ticketMu.Lock()
	uc.tickets[ticket] = shellTicket{agentUID: uid, expires: time.Now().Add(shellTicketTTL)}
	uc.ticketMu.Unlock()
	return ticket, nil
}

// OpenShell 消费票据并打开终端。同一台机器只保留最新的一条。
func (uc *Usecase) OpenShell(ctx context.Context, agentUID, ticket string, cols, rows int32) (*ShellLink, error) {
	if err := uc.takeTicket(agentUID, ticket); err != nil {
		return nil, err
	}
	machine, err := uc.agents.GetAgentByUID(ctx, agentUID)
	if err != nil || machine.Status != "ACTIVE" || !uc.hub.Online(agentUID) {
		return nil, merr.Invalid("agent is offline")
	}
	password, privateKey, err := uc.agents.GetSSHSecrets(ctx, agentUID)
	if err != nil {
		return nil, merr.Invalid("ssh credentials missing")
	}
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}
	sessionUID := id.New()
	data, closed, err := uc.hub.ClaimShell(agentUID, sessionUID)
	if err != nil {
		return nil, err
	}
	open := &v1.AgentFrame{ShellOpen: &v1.AgentShellOpen{
		SessionUid: sessionUID, SshHost: machine.SSHHost, SshPort: int32(machine.SSHPort),
		SshUser: machine.SSHUser, SshPassword: password, SshPrivateKey: privateKey,
		Cols: cols, Rows: rows,
	}}
	if err := uc.hub.PushFrame(agentUID, open); err != nil {
		if session, ok := uc.hub.Get(agentUID); ok {
			session.FinishShell(sessionUID, "offline")
		}
		return nil, err
	}
	link := &ShellLink{SessionUID: sessionUID, Data: data, Closed: closed}
	link.write = func(payload []byte) error {
		return uc.hub.PushFrame(agentUID, &v1.AgentFrame{ShellData: &v1.AgentShellData{SessionUid: sessionUID, Data: payload}})
	}
	link.resize = func(nextCols, nextRows int32) error {
		return uc.hub.PushFrame(agentUID, &v1.AgentFrame{ShellResize: &v1.AgentShellResize{SessionUid: sessionUID, Cols: nextCols, Rows: nextRows}})
	}
	link.stop = func() {
		_ = uc.hub.PushFrame(agentUID, &v1.AgentFrame{ShellClose: &v1.AgentShellClose{SessionUid: sessionUID, Reason: "closed"}})
		if session, ok := uc.hub.Get(agentUID); ok {
			session.FinishShell(sessionUID, "closed")
		}
	}
	return link, nil
}

func (uc *Usecase) takeTicket(agentUID, ticket string) error {
	if ticket == "" {
		return merr.Unauthenticated("invalid token")
	}
	uc.ticketMu.Lock()
	defer uc.ticketMu.Unlock()
	item, ok := uc.tickets[ticket]
	if ok {
		delete(uc.tickets, ticket)
	}
	if !ok || item.agentUID != agentUID || time.Now().After(item.expires) {
		return merr.Unauthenticated("invalid token")
	}
	return nil
}
