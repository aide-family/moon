package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
)

type Invite interface {
	TeamInviteUser(ctx context.Context, req bo.InviteMember) error
}
