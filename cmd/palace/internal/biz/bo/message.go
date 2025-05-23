package bo

import (
	"context"
	"fmt"
	"time"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/pkg/merr"
)

type SendEmailFun func(ctx context.Context, params *SendEmailParams) error

type CreateSendMessageLogParams struct {
	TeamID      uint32
	SendAt      time.Time
	MessageType vobj.MessageType
	Message     fmt.Stringer
	RequestID   string
}

type UpdateSendMessageLogStatusParams struct {
	TeamID    uint32
	RequestID string
	SendAt    time.Time
	Status    vobj.SendMessageStatus
	Error     string
}

type GetSendMessageLogParams struct {
	TeamID    uint32
	RequestID string
	SendAt    time.Time
}

func (p *GetSendMessageLogParams) WithTeamID(ctx context.Context) (*GetSendMessageLogParams, error) {
	teamId, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return nil, merr.ErrorUnauthorized("team id is not found")
	}
	p.TeamID = teamId
	return p, nil
}

type RetrySendMessageParams struct {
	TeamID    uint32
	RequestID string
	SendAt    time.Time
}

func (p *RetrySendMessageParams) WithTeamID(ctx context.Context) (*RetrySendMessageParams, error) {
	teamId, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return nil, merr.ErrorUnauthorized("team id is not found")
	}
	p.TeamID = teamId
	return p, nil
}

type ListSendMessageLogParams struct {
	*PaginationRequest
	TeamID      uint32
	RequestID   string
	Status      vobj.SendMessageStatus
	Keyword     string
	TimeRange   []time.Time
	MessageType vobj.MessageType
}

func (p *ListSendMessageLogParams) WithTeamID(ctx context.Context) (*ListSendMessageLogParams, error) {
	teamId, ok := permission.GetTeamIDByContext(ctx)
	if !ok {
		return nil, merr.ErrorUnauthorized("team id is not found")
	}
	p.TeamID = teamId
	return p, nil
}

func (p *ListSendMessageLogParams) ToListReply(logs []do.SendMessageLog) *ListSendMessageLogReply {
	return &ListSendMessageLogReply{
		PaginationReply: p.ToReply(),
		Items:           logs,
	}
}

type ListSendMessageLogReply = ListReply[do.SendMessageLog]
