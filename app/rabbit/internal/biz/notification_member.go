package biz

import (
	"context"

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
	memberUIDs := collectNotificationMemberUIDs(members)
	if len(memberUIDs) == 0 {
		return nil
	}
	membersResp, err := memberRepo.ListMember(ctx, &goddessv1.ListMemberRequest{
		Page:     1,
		PageSize: int32(len(memberUIDs)),
		Uids:     memberUIDs,
	})
	if err != nil {
		if helper != nil {
			helper.WithContext(ctx).Warnw("msg", "notification member list failed", "error", err)
		}
		return nil
	}
	memberMap := make(map[int64]*goddessv1.MemberItem, len(membersResp.GetItems()))
	for _, item := range membersResp.GetItems() {
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
	return memberUIDs
}
