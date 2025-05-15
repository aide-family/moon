package hook

import (
	"context"

	"github.com/aide-family/moon/pkg/api/rabbit/common"
)

type Sender interface {
	Type() common.HookAPP
	Send(ctx context.Context, message Message) error
}

type BasicAuth struct {
	Username string
	Password string
}
