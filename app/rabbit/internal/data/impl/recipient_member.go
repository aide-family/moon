package impl

import (
	"context"

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
	conflict := clause.OnConflict{
		UpdateAll: true,
	}
	if err := query.RecipientMember.WithContext(ctx).Clauses(conflict).CreateInBatches(doMembers, 100); err != nil {
		return err
	}
	return nil
}
