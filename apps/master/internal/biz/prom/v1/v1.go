package biz

import "context"

type V1Repo interface {
	V1(ctx context.Context) string
}
