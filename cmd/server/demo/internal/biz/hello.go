package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/demo/internal/biz/repository"
)

func NewHelloBiz(helloRepository repository.Hello) *HelloBiz {
	return &HelloBiz{
		helloRepository: helloRepository,
	}
}

// HelloBiz .
type HelloBiz struct {
	helloRepository repository.Hello
}

func (b *HelloBiz) SayHello(ctx context.Context, name string) (string, error) {
	return b.helloRepository.SayHello(ctx, name)
}
