package biz

import (
	"context"
	"time"
)

type IdentityRepo interface {
	CreateUser(ctx context.Context, in Account) (*Account, error)
	UserByEmail(ctx context.Context, email string) (*Account, error)
	UserByUID(ctx context.Context, uid string) (*Account, error)
	UpdateUserProfile(ctx context.Context, uid string, email, avatar, remark *string) (*Account, error)
	UpdateUserStatus(ctx context.Context, uid, status string) (*Account, error)
	ListUsers(ctx context.Context, page, pageSize int, name string) ([]Account, int, error)
	SelectUsers(ctx context.Context, limit int, lastUID string) ([]SelectItem, error)

	SaveCaptcha(ctx context.Context, id, answerHash string, expiresAt time.Time) error
	ConsumeCaptcha(ctx context.Context, id, answerHash string, now time.Time) (bool, error)
	SaveEmailCode(ctx context.Context, email, codeHash string, expiresAt time.Time) error
	ConsumeEmailCode(ctx context.Context, email, codeHash string, now time.Time) (bool, error)

	OAuthAccount(ctx context.Context, provider, openID string) (*Account, error)
	LinkOAuth(ctx context.Context, userUID, provider, openID, email string) error

	CreateNamespace(ctx context.Context, name, remark, ownerUID string) (*Namespace, error)
	ListUserNamespaces(ctx context.Context, userUID string) ([]Namespace, error)
	ListNamespaces(ctx context.Context, userUID string, page, pageSize int, name string) ([]Namespace, int, error)
	SelectNamespaces(ctx context.Context, userUID string, limit int, lastUID string) ([]SelectItem, error)
	GetNamespace(ctx context.Context, uid string) (*Namespace, error)
	UpdateNamespace(ctx context.Context, uid, name, remark, logo string) (*Namespace, error)
	UpdateNamespaceStatus(ctx context.Context, uid, status string) (*Namespace, error)
	DeleteNamespace(ctx context.Context, uid string) error
	TransferOwner(ctx context.Context, namespaceUID, targetUserUID string) (*Namespace, error)

	ListMembers(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]Member, int, error)
	GetMember(ctx context.Context, namespaceUID, uid string) (*Member, error)
	MemberByUser(ctx context.Context, namespaceUID, userUID string) (*Member, error)
	SelectMembers(ctx context.Context, namespaceUID string, limit int, lastUID string) ([]SelectItem, error)
	InviteMember(ctx context.Context, namespaceUID, email, roleUID string) (*Member, error)
	DeleteMember(ctx context.Context, namespaceUID, uid string) error
	UpdateMemberStatus(ctx context.Context, namespaceUID, uid, status string) (*Member, error)
	UpdateMemberRoles(ctx context.Context, namespaceUID, memberUID string, roleUIDs []string) (*Member, error)

	ListRoles(ctx context.Context, namespaceUID string, page, pageSize int, name string) ([]Role, int, error)
	GetRole(ctx context.Context, namespaceUID, uid string) (*Role, error)
	SelectRoles(ctx context.Context, namespaceUID string, limit int, lastUID string) ([]SelectItem, error)
	CreateRole(ctx context.Context, namespaceUID, name string) (*Role, error)
	UpdateRole(ctx context.Context, namespaceUID, uid, name string) (*Role, error)
	UpdateRoleStatus(ctx context.Context, namespaceUID, uid, status string) (*Role, error)
	DeleteRole(ctx context.Context, namespaceUID, uid string) error
	UpdateRolePermissions(ctx context.Context, namespaceUID, uid string, names []string) (*Role, error)
	ListPermissions(ctx context.Context) ([]Permission, error)
	Permissions(ctx context.Context, namespaceUID, userUID string) (map[string]struct{}, error)
}
