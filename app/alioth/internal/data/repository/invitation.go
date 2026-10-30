package repository

import (
	"context"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// InviteRepo 读写租户邀请消息。
// 邀请只存在于平台消息里：用户确认前不会创建 Member，确认后才成为租户成员。
type InviteRepo interface {
	// UpsertInvite 创建邀请；同一租户同一邮箱已有待确认邀请时刷新令牌和有效期。
	UpsertInvite(ctx context.Context, in bo.MemberInvite, tokenHash string) (*bo.MemberInvite, error)
	// InviteByToken 按邀请令牌查找邀请。令牌只保存摘要。
	InviteByToken(ctx context.Context, tokenHash string) (*bo.MemberInvite, error)
	// ListInvitesForUser 返回发给这个用户的平台邀请消息，未注册邮箱按 email 匹配。
	ListInvitesForUser(ctx context.Context, userUID, email string) ([]bo.MemberInvite, error)
	// ListInvitesForNamespace 返回当前租户还在等待确认的邀请。
	ListInvitesForNamespace(ctx context.Context, namespaceUID string) ([]bo.MemberInvite, error)
	// AcceptInvite 校验身份和有效期后把用户变成 JOINED 成员并授予角色。
	// 租户套餐的人数上限在事务内锁定租户后重新统计 JOINED 成员再判断，
	// 避免并发确认把成员数顶过上限；超限返回 member limit reached。
	AcceptInvite(ctx context.Context, tokenHash, inviteUID, userUID, email string) (*bo.Member, error)
	// DeclineInvite 记录用户拒绝，不创建成员。
	DeclineInvite(ctx context.Context, tokenHash, inviteUID, userUID, email string) error
	// RevokeInvite 由租户管理员撤销一条待确认邀请。
	RevokeInvite(ctx context.Context, namespaceUID, uid string) error
}
