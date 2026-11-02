package data

import (
	"context"
	"strings"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/member"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/oncallrotation"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/oncallrotationmember"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/user"
	"github.com/aide-family/moon/app/alioth/pkg/oncall"
)

// rotationReceiverContacts 返回 OnCall 通知组当前当班成员的联系方式。
// 轮换停用、已删除、没有成员或成员没有邮箱/手机号时返回空，不报错；
// 只有数据库错误才向上抛出。field 取 email 或 phone。
func rotationReceiverContacts(ctx context.Context, tx *ent.Tx, namespaceUID, rotationUID, field string) ([]string, error) {
	row, err := tx.OnCallRotation.Query().Where(
		oncallrotation.UIDEQ(rotationUID),
		oncallrotation.NamespaceUIDEQ(namespaceUID),
		oncallrotation.StatusEQ(oncallrotation.StatusACTIVE),
		oncallrotation.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	members, err := tx.OnCallRotationMember.Query().Where(
		oncallrotationmember.RotationUIDEQ(row.UID),
	).Order(ent.Asc(oncallrotationmember.FieldPosition)).All(ctx)
	if err != nil {
		return nil, err
	}
	if len(members) == 0 {
		return nil, nil
	}
	rotation := oncall.Rotation{
		Timezone: row.Timezone, Period: string(row.Period),
		StartsAt: row.StartsAt, HandoffHour: row.HandoffHour,
	}
	rotation.Members = make([]oncall.RotationMember, 0, len(members))
	for _, item := range members {
		rotation.Members = append(rotation.Members, oncall.RotationMember{MemberUID: item.MemberUID})
	}
	current, ok := rotation.MemberAt(time.Now().UTC())
	if !ok {
		return nil, nil
	}
	memberRow, err := tx.Member.Query().Where(
		member.UIDEQ(current.MemberUID),
		member.NamespaceUIDEQ(namespaceUID),
		member.StatusEQ(member.StatusJOINED),
		member.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	account, err := tx.User.Query().Where(
		user.UIDEQ(memberRow.UserUID),
		user.StatusEQ(user.StatusACTIVE),
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	value := account.Email
	if field == "phone" {
		value = account.Phone
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	return []string{value}, nil
}
