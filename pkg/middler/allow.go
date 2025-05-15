package middler

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware/selector"
)

// AllowListMatcher new allow operation list matcher
func AllowListMatcher(list ...string) selector.MatchFunc {
	whiteList := make(map[string]struct{})
	for _, v := range list {
		whiteList[v] = struct{}{}
	}
	return func(ctx context.Context, operation string) bool {
		_, ok := whiteList[operation]
		return !ok
	}
}
