package safety

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCopyValueCtx_ValuePreserved(t *testing.T) {
	ctx := context.WithValue(context.Background(), "key", "value")
	copied := CopyValueCtx(ctx)

	assert.Equal(t, "value", copied.Value("key"))
	assert.Nil(t, copied.Value("not-exist"))
}

func TestCopyValueCtx_DeadlineDoneErr(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	copied := CopyValueCtx(ctx)

	// valueOnlyContext 的 Deadline、Done、Err 都是零值
	deadline, ok := copied.Deadline()
	assert.False(t, ok)
	assert.True(t, deadline.IsZero())
	assert.Nil(t, copied.Done())
	assert.Nil(t, copied.Err())
}

func TestCopyValueCtx_OriginalCancelNotAffectCopied(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = context.WithValue(ctx, "k", 123)
	copied := CopyValueCtx(ctx)
	cancel()

	// copied 不受 cancel 影响
	assert.Nil(t, copied.Done())
	assert.Nil(t, copied.Err())
	assert.Equal(t, 123, copied.Value("k"))
}

func TestValueOnlyContext_ImplementsContext(t *testing.T) {
	var _ context.Context = valueOnlyContext{context.Background()}
}
