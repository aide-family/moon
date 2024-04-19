package agent

import (
	"context"
)

type Eval interface {
	Eval(ctx context.Context) (*Alarm, error)
}
