package data

import (
	"context"
	"fmt"
	"strings"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/namespace"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

func (r *repoBase) fail(err error) error {
	if err == nil {
		return nil
	}
	if kerrors.Reason(err) != "" {
		return err
	}
	if ent.IsNotFound(err) {
		return merr.NotFound("not found")
	}
	if ent.IsConstraintError(err) {
		return merr.Conflict(constraintMessage(err))
	}
	r.log.Error(fmt.Sprintf("database: %v", err))
	return merr.Server("request failed")
}

// constraintMessage 按唯一约束名给出可读冲突文案；未知约束保留通用提示。
// ent 的错误文本里带索引名，例如 notificationoncallgroup_oncall_uid_group_uid。
func constraintMessage(err error) string {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "notificationoncallgroup") &&
		strings.Contains(message, "oncall_uid") &&
		strings.Contains(message, "group_uid"):
		return "notification group is already bound to this on-call"
	case strings.Contains(message, "users_email"):
		return "email already exists"
	case strings.Contains(message, "member"):
		return "member already exists"
	default:
		return "name already exists"
	}
}

func (r *repoBase) GetNamespace(ctx context.Context, uid string) (*bo.Namespace, error) {
	row, err := r.client.Namespace.Query().Where(namespace.UIDEQ(uid), namespace.DeletedAtIsNil()).Only(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	item := toNamespace(row)
	return &item, nil
}
