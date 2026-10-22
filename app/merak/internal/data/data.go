package data

import (
	"time"

	"github.com/google/wire"

	"github.com/aide-family/Alioth/app/merak/internal/data/repository"
	"github.com/aide-family/Alioth/pkg/healthcheck"
)

var ProviderSet = wire.NewSet(NewLocker, NewGateway, NewReadinessChecker)

// NewReadinessChecker 只探测 Redis：merak 通过 gRPC 访问控制面，不直连 PostgreSQL。
func NewReadinessChecker(locker repository.Locker) healthcheck.Checker {
	checks := healthcheck.NewChecks(2 * time.Second)
	if locker != nil {
		checks.Add("redis", locker.Available)
	}
	return checks
}
