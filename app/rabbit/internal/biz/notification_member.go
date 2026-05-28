package biz

import (
	"context"
	"slices"

	klog "github.com/go-kratos/kratos/v2/log"

	goddessv1 "github.com/aide-family/goddess/pkg/api/v1"
	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/biz/repository"
)

func fillNotificationMemberDetails(
	ctx context.Context,
	memberRepo repository.Member,
	helper *klog.Helper,
	members []*bo.NotificationMemberBo,
) error {
	memberUIDSet := make(map[int64]struct{})
	for _, member := range members {
		if member != nil && member.MemberUID > 0 {
			memberUIDSet[member.MemberUID] = struct{}{}
		}
	}
	if len(memberUIDSet) == 0 {
		return nil
	}
	memberUIDs := make([]int64, 0, len(memberUIDSet))
	for uid := range memberUIDSet {
		memberUIDs = append(memberUIDs, uid)
	}
	slices.Sort(memberUIDs)
	memberMap := make(map[int64]*goddessv1.MemberItem, len(memberUIDs))
	for _, uid := range memberUIDs {
		item, err := memberRepo.GetMember(ctx, &goddessv1.GetMemberRequest{Uid: uid})
		if err != nil {
			if helper != nil {
				helper.WithContext(ctx).Warnw("msg", "notification member lookup failed", "memberUID", uid, "error", err)
			}
			continue
		}
		if item != nil {
			memberMap[item.GetUid()] = item
		}
	}
	for _, member := range members {
		if member == nil {
			continue
		}
		item := memberMap[member.MemberUID]
		if item == nil {
			continue
		}
		member.MemberName = item.GetName()
		member.MemberAvatar = item.GetAvatar()
		member.MemberEmail = item.GetEmail()
		member.MemberPhone = item.GetPhone()
	}
	return nil
}

func collectNotificationMemberUIDs(members []*bo.NotificationMemberBo) []int64 {
	memberUIDSet := make(map[int64]struct{})
	for _, member := range members {
		if member != nil && member.MemberUID > 0 {
			memberUIDSet[member.MemberUID] = struct{}{}
		}
	}
	if len(memberUIDSet) == 0 {
		return nil
	}
	memberUIDs := make([]int64, 0, len(memberUIDSet))
	for uid := range memberUIDSet {
		memberUIDs = append(memberUIDs, uid)
	}
	slices.Sort(memberUIDs)
	return memberUIDs
}
