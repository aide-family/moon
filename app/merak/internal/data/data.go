package data

import (
	"time"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/merak/internal/data/repository"
	"github.com/aide-family/Alioth/pkg/healthcheck"
)

var ProviderSet = wire.NewSet(NewData, NewLocker, NewGateway, NewReadiness)

type Data struct{}

func NewData() (*Data, func(), error) {
	return &Data{}, func() {}, nil
}

// NewReadiness 只探测 Redis：merak 通过 gRPC 访问控制面，不直连 PostgreSQL。
func NewReadiness(locker repository.Locker) healthcheck.Checker {
	checks := healthcheck.New(2 * time.Second)
	if locker != nil {
		checks.Add("redis", locker.Available)
	}
	return checks
}
