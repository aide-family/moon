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
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
)

// onCallRotationRepo 管理值班轮换与成员顺序。
type onCallRotationRepo struct{ *repoBase }

// replaceRotationMembers 校验成员都是当前租户的 JOINED 成员后，整体替换成员列表。
// position 由数组下标决定，重复成员直接拒绝。
func replaceRotationMembers(ctx context.Context, tx *ent.Tx, namespaceUID, rotationUID string, members []bo.OnCallRotationMember) error {
	memberUIDs := make([]string, 0, len(members))
	seen := make(map[string]struct{}, len(members))
	for _, item := range members {
		memberUID := strings.TrimSpace(item.MemberUID)
		if memberUID == "" {
			return merr.Invalid("rotation member is required")
		}
		if _, ok := seen[memberUID]; ok {
			return merr.Invalid("duplicate rotation member")
		}
		seen[memberUID] = struct{}{}
		memberUIDs = append(memberUIDs, memberUID)
	}
	if len(memberUIDs) == 0 {
		return merr.Invalid("rotation member is required")
	}
	joined, err := tx.Member.Query().Where(
		member.UIDIn(memberUIDs...),
		member.NamespaceUIDEQ(namespaceUID),
		member.StatusEQ(member.StatusJOINED),
		member.DeletedAtIsNil(),
	).Count(ctx)
	if err != nil {
		return err
	}
	if joined != len(memberUIDs) {
		return merr.Invalid("invalid rotation member")
	}
	if _, err := tx.OnCallRotationMember.Delete().Where(
		oncallrotationmember.RotationUIDEQ(rotationUID),
	).Exec(ctx); err != nil {
		return err
	}
	for position, memberUID := range memberUIDs {
		if _, err := tx.OnCallRotationMember.Create().
			SetUID(id.New()).
			SetNamespaceUID(namespaceUID).
			SetRotationUID(rotationUID).
			SetMemberUID(memberUID).
			SetPosition(position).
			Save(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (r *onCallRotationRepo) CreateRotation(ctx context.Context, in bo.OnCallRotation) (*bo.OnCallRotation, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	row, err := tx.OnCallRotation.Create().
		SetUID(id.New()).
		SetNamespaceUID(in.NamespaceUID).
		SetName(in.Name).
		SetRemark(in.Remark).
		SetTimezone(in.Timezone).
		SetPeriod(oncallrotation.Period(in.Period)).
		SetStartsAt(in.StartsAt).
		SetHandoffHour(in.HandoffHour).
		SetStatus(oncallrotation.StatusACTIVE).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if err := replaceRotationMembers(ctx, tx, in.NamespaceUID, row.UID, in.Members); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetRotation(ctx, in.NamespaceUID, row.UID)
}

func (r *onCallRotationRepo) UpdateRotation(ctx context.Context, in bo.OnCallRotation) (*bo.OnCallRotation, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	n, err := tx.OnCallRotation.Update().Where(
		oncallrotation.UIDEQ(in.UID),
		oncallrotation.NamespaceUIDEQ(in.NamespaceUID),
		oncallrotation.DeletedAtIsNil(),
	).
		SetName(in.Name).
		SetRemark(in.Remark).
		SetTimezone(in.Timezone).
		SetPeriod(oncallrotation.Period(in.Period)).
		SetStartsAt(in.StartsAt).
		SetHandoffHour(in.HandoffHour).
		Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("rotation not found")
	}
	if err := replaceRotationMembers(ctx, tx, in.NamespaceUID, in.UID, in.Members); err != nil {
		return nil, r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return nil, r.fail(err)
	}
	return r.GetRotation(ctx, in.NamespaceUID, in.UID)
}

func (r *onCallRotationRepo) UpdateRotationStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.OnCallRotation, error) {
	value := oncallrotation.StatusACTIVE
	if status == "DISABLED" {
		value = oncallrotation.StatusDISABLED
	} else if status != "ACTIVE" {
		return nil, merr.Invalid("invalid request")
	}
	n, err := r.client.OnCallRotation.Update().Where(
		oncallrotation.UIDEQ(uid),
		oncallrotation.NamespaceUIDEQ(namespaceUID),
		oncallrotation.DeletedAtIsNil(),
	).SetStatus(value).Save(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if n == 0 {
		return nil, merr.NotFound("rotation not found")
	}
	return r.GetRotation(ctx, namespaceUID, uid)
}

func (r *onCallRotationRepo) DeleteRotation(ctx context.Context, namespaceUID, uid string) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return r.fail(err)
	}
	defer func() { _ = tx.Rollback() }()
	row, err := tx.OnCallRotation.Query().Where(
		oncallrotation.UIDEQ(uid),
		oncallrotation.NamespaceUIDEQ(namespaceUID),
		oncallrotation.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return merr.NotFound("rotation not found")
		}
		return r.fail(err)
	}
	if _, err := tx.OnCallRotationMember.Delete().Where(
		oncallrotationmember.RotationUIDEQ(row.UID),
	).Exec(ctx); err != nil {
		return r.fail(err)
	}
	if _, err := tx.OnCallRotation.Update().Where(oncallrotation.IDEQ(row.ID)).
		SetDeletedAt(time.Now().UTC()).Save(ctx); err != nil {
		return r.fail(err)
	}
	if err := tx.Commit(); err != nil {
		return r.fail(err)
	}
	return nil
}

func (r *onCallRotationRepo) GetRotation(ctx context.Context, namespaceUID, uid string) (*bo.OnCallRotation, error) {
	row, err := r.client.OnCallRotation.Query().Where(
		oncallrotation.UIDEQ(uid),
		oncallrotation.NamespaceUIDEQ(namespaceUID),
		oncallrotation.DeletedAtIsNil(),
	).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, merr.NotFound("rotation not found")
		}
		return nil, r.fail(err)
	}
	members, err := r.rotationMembers(ctx, []string{row.UID})
	if err != nil {
		return nil, err
	}
	item := toOnCallRotation(row, members[row.UID])
	return &item, nil
}

func (r *onCallRotationRepo) ListRotations(ctx context.Context, namespaceUID string, name string, page, pageSize int) ([]bo.OnCallRotation, int, error) {
	query := r.client.OnCallRotation.Query().Where(
		oncallrotation.NamespaceUIDEQ(namespaceUID),
		oncallrotation.DeletedAtIsNil(),
	)
	if name != "" {
		query = query.Where(oncallrotation.NameContainsFold(name))
	}
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	rows, err := query.Order(ent.Desc(oncallrotation.FieldUpdatedAt)).
		Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, 0, r.fail(err)
	}
	uids := make([]string, 0, len(rows))
	for _, row := range rows {
		uids = append(uids, row.UID)
	}
	members, err := r.rotationMembers(ctx, uids)
	if err != nil {
		return nil, 0, err
	}
	items := make([]bo.OnCallRotation, 0, len(rows))
	for _, row := range rows {
		items = append(items, toOnCallRotation(row, members[row.UID]))
	}
	return items, total, nil
}

// rotationMembers 批量读取轮换成员并按 position 排序，附带成员展示信息。
func (r *onCallRotationRepo) rotationMembers(ctx context.Context, rotationUIDs []string) (map[string][]bo.OnCallRotationMember, error) {
	result := make(map[string][]bo.OnCallRotationMember, len(rotationUIDs))
	if len(rotationUIDs) == 0 {
		return result, nil
	}
	rows, err := r.client.OnCallRotationMember.Query().Where(
		oncallrotationmember.RotationUIDIn(rotationUIDs...),
	).Order(ent.Asc(oncallrotationmember.FieldPosition)).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	if len(rows) == 0 {
		return result, nil
	}
	memberUIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		memberUIDs = append(memberUIDs, row.MemberUID)
	}
	members, err := r.memberAccounts(ctx, memberUIDs)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		account := members[row.MemberUID]
		result[row.RotationUID] = append(result[row.RotationUID], bo.OnCallRotationMember{
			UID: row.UID, NamespaceUID: row.NamespaceUID, RotationUID: row.RotationUID,
			MemberUID: row.MemberUID, Position: row.Position,
			Name: account.Name, Email: account.Email, Phone: account.Phone,
		})
	}
	return result, nil
}

// memberAccounts 按成员 UID 返回展示用的名称与联系方式。成员或用户缺失时返回零值，
// 不阻断轮换读取：历史成员的邮箱为空时只是不追加接收方。
func (r *onCallRotationRepo) memberAccounts(ctx context.Context, memberUIDs []string) (map[string]bo.Member, error) {
	result := make(map[string]bo.Member, len(memberUIDs))
	if len(memberUIDs) == 0 {
		return result, nil
	}
	rows, err := r.client.Member.Query().Where(
		member.UIDIn(uniqueStrings(memberUIDs)...),
		member.DeletedAtIsNil(),
	).All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	userUIDs := make([]string, 0, len(rows))
	for _, row := range rows {
		userUIDs = append(userUIDs, row.UserUID)
	}
	byUser := map[string]*ent.User{}
	if len(userUIDs) > 0 {
		users, err := r.client.User.Query().Where(user.UIDIn(userUIDs...)).All(ctx)
		if err != nil {
			return nil, r.fail(err)
		}
		for _, account := range users {
			byUser[account.UID] = account
		}
	}
	for _, row := range rows {
		item := bo.Member{UID: row.UID, NamespaceUID: row.NamespaceUID, UserUID: row.UserUID, Status: string(row.Status)}
		if account := byUser[row.UserUID]; account != nil {
			item.Name = account.Name
			item.Email = account.Email
			item.Phone = account.Phone
		}
		result[row.UID] = item
	}
	return result, nil
}

func toOnCallRotation(row *ent.OnCallRotation, members []bo.OnCallRotationMember) bo.OnCallRotation {
	return bo.OnCallRotation{
		UID: row.UID, NamespaceUID: row.NamespaceUID, Name: row.Name, Remark: row.Remark,
		Timezone: row.Timezone, Period: string(row.Period), StartsAt: row.StartsAt,
		HandoffHour: row.HandoffHour, Status: string(row.Status),
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt, Members: members,
	}
}
