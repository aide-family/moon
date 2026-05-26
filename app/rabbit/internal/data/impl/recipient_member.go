package impl

import (
	"context"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/merr"
	"gorm.io/gorm/clause"

	"github.com/aide-family/rabbit/internal/biz/bo"
	"github.com/aide-family/rabbit/internal/biz/repository"
	"github.com/aide-family/rabbit/internal/data"
	"github.com/aide-family/rabbit/internal/data/impl/convert"
	"github.com/aide-family/rabbit/internal/data/impl/do"
	"github.com/aide-family/rabbit/internal/data/impl/query"
)

func NewRecipientMemberRepository(d *data.Data) repository.RecipientMember {
	query.SetDefault(d.DB())
	return &recipientMemberRepository{Data: d}
}

type recipientMemberRepository struct {
	*data.Data
}

// CreateRecipientMember implements [repository.RecipientMember].
func (r *recipientMemberRepository) CreateRecipientMember(ctx context.Context, members []*bo.RecipientMemberItemBo) error {
	doMembers := make([]*do.RecipientMember, 0, len(members))
	for _, member := range members {
		doMembers = append(doMembers, convert.ToRecipientMemberDo(ctx, member))
	}
	rm := query.RecipientMember
	// Upsert by namespace + user: API member UID is stored as id; legacy rows may share user_uid with another id.
	conflict := clause.OnConflict{
		Columns: []clause.Column{
			{Name: rm.NamespaceUID.ColumnName().String()},
			{Name: rm.UserUID.ColumnName().String()},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			rm.ID.ColumnName().String(),
			rm.UpdatedAt.ColumnName().String(),
			rm.Creator.ColumnName().String(),
			rm.Email.ColumnName().String(),
			rm.Phone.ColumnName().String(),
			rm.Status.ColumnName().String(),
			rm.Name.ColumnName().String(),
			rm.Nickname.ColumnName().String(),
			rm.Avatar.ColumnName().String(),
			rm.Remark.ColumnName().String(),
		}),
	}
	if err := rm.WithContext(ctx).Clauses(conflict).CreateInBatches(doMembers, 100); err != nil {
		return err
	}
	return nil
}

// ResolveRecipientMemberIDs implements [repository.RecipientMember].
func (r *recipientMemberRepository) ResolveRecipientMemberIDs(ctx context.Context, members []*bo.RecipientMemberItemBo) ([]int64, error) {
	if len(members) == 0 {
		return nil, nil
	}
	ns := contextx.GetNamespace(ctx)
	rm := query.RecipientMember
	memberIDs := make([]int64, 0, len(members))
	userUIDs := make([]int64, 0, len(members))
	wantUser := make(map[int64]struct{}, len(members))
	for _, m := range members {
		memberIDs = append(memberIDs, m.UID.Int64())
		userUIDs = append(userUIDs, m.UserUID.Int64())
		wantUser[m.UserUID.Int64()] = struct{}{}
	}

	byID, err := rm.WithContext(ctx).Where(
		rm.NamespaceUID.Eq(ns.Int64()),
		rm.ID.In(memberIDs...),
	).Find()
	if err != nil {
		return nil, err
	}
	byUser, err := rm.WithContext(ctx).Where(
		rm.NamespaceUID.Eq(ns.Int64()),
		rm.UserUID.In(userUIDs...),
	).Find()
	if err != nil {
		return nil, err
	}

	merged := make(map[int64]*do.RecipientMember, len(members))
	for _, row := range byID {
		merged[row.UserUID.Int64()] = row
	}
	for _, row := range byUser {
		merged[row.UserUID.Int64()] = row
	}
	if len(merged) != len(wantUser) {
		return nil, merr.ErrorInvalidArgument("recipient member not found")
	}

	out := make([]int64, 0, len(members))
	for _, m := range members {
		row, ok := merged[m.UserUID.Int64()]
		if !ok {
			return nil, merr.ErrorInvalidArgument("recipient member not found")
		}
		out = append(out, row.ID.Int64())
	}
	return out, nil
}
