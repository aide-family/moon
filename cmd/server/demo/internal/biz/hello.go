package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/demo/internal/biz/repository"
)

// NewHelloBiz 实例化HelloBiz
func NewHelloBiz(helloRepository repository.Hello) *HelloBiz {
	return &HelloBiz{
		helloRepository: helloRepository,
	}
}

// HelloBiz .
type HelloBiz struct {
	helloRepository repository.Hello
}

// SayHello  输出Hello {name}
func (b *HelloBiz) SayHello(ctx context.Context, name string) (string, error) {
	return b.helloRepository.SayHello(ctx, name)
}
