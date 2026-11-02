package repository

import (
	"context"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// OnCallRotationRepo 读写值班轮换与成员顺序。
type OnCallRotationRepo interface {
	CreateRotation(ctx context.Context, in bo.OnCallRotation) (*bo.OnCallRotation, error)
	UpdateRotation(ctx context.Context, in bo.OnCallRotation) (*bo.OnCallRotation, error)
	UpdateRotationStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.OnCallRotation, error)
	DeleteRotation(ctx context.Context, namespaceUID, uid string) error
	GetRotation(ctx context.Context, namespaceUID, uid string) (*bo.OnCallRotation, error)
	ListRotations(ctx context.Context, namespaceUID string, name string, page, pageSize int) ([]bo.OnCallRotation, int, error)
}
