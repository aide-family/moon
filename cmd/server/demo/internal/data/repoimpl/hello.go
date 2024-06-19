package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/demo/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/demo/internal/data"
)

func NewHelloRepository(data *data.Data) repository.Hello {
	return &helloRepositoryImpl{data: data}
}

type helloRepositoryImpl struct {
	data *data.Data
}

func (h *helloRepositoryImpl) SayHello(ctx context.Context, name string) (string, error) {
	return "hello " + name, nil
}
