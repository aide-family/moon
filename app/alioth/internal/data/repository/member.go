package repository

import (
	"context"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

// MemberRepo 读写租户成员及其角色绑定。
type MemberRepo interface {
	ListMembers(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]bo.Member, int, error)
	ListMembersByUIDs(ctx context.Context, namespaceUID string, memberUIDs []string) ([]bo.Member, error)
	GetMember(ctx context.Context, namespaceUID, uid string) (*bo.Member, error)
	MemberByUser(ctx context.Context, namespaceUID, userUID string) (*bo.Member, error)
	SelectMembers(ctx context.Context, namespaceUID string, limit int, lastUID string) ([]bo.SelectItem, error)
	InviteMember(ctx context.Context, namespaceUID, email, roleUID string) (*bo.Member, error)
	DeleteMember(ctx context.Context, namespaceUID, uid string) error
	UpdateMemberStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.Member, error)
	UpdateMemberRoles(ctx context.Context, namespaceUID, memberUID string, roleUIDs []string) (*bo.Member, error)
}
