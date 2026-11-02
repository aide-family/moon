package repository

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// ProbeSlaRepo 负责探测结果的日聚合并出具 SLA 报表。
type ProbeSlaRepo interface {
	// Rollup 重算 from 之后各自然日的日聚合，供 SLA 报表读取。
	Rollup(ctx context.Context, from time.Time) error
	// Report 返回按目标组汇总的 SLA 报表。
	Report(ctx context.Context, namespaceUID string, groupUIDs []string, from, to time.Time) (bo.ProbeSlaReport, error)
}
