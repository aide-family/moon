package watch_test

import (
	"context"
	"testing"

	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

func TestMessage_IsHandledPath(t *testing.T) {
	msg := watch.NewMessage(&MyMsg{Data: 1}, vobj.TopicUnknown)
	h1 := func(ctx context.Context, msg *watch.Message) error {
		return nil
	}
	msg.WithHandledPath(0, h1)

	t.Log(msg.IsHandled(0))
	t.Log(msg.IsHandled(1))
	t.Log(msg.IsHandled(0))
}
