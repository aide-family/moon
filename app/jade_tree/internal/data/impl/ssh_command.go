package impl

import (
	"context"

	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"

	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/biz/repository"
	"github.com/aide-family/jade_tree/internal/data"
	"github.com/aide-family/jade_tree/internal/data/impl/convert"
	"github.com/aide-family/jade_tree/internal/data/impl/do"
	"github.com/aide-family/jade_tree/internal/data/impl/query"
)

func NewSSHCommandRepository(d *data.Data) repository.SSHCommand {
	query.SetDefault(d.DB())
	return &sshCommandRepository{Data: d}
}

type sshCommandRepository struct {
	*data.Data
}

func (r *sshCommandRepository) Create(ctx context.Context, in *bo.SSHCommandCreateRepoBo) (*bo.SSHCommandItemBo, error) {
	if in == nil {
		return nil, merr.ErrorInvalidArgument("create input is required")
	}
	row := convert.ToSSHCommandDO(in.Creator, &in.Fields)
	sc := query.SSHCommand
	if err := sc.WithContext(ctx).Create(row); err != nil {
		return nil, err
	}
	return convert.ToSSHCommandItemBo(row), nil
}

func (r *sshCommandRepository) Update(ctx context.Context, in *bo.SSHCommandUpdateRepoBo) error {
	if in == nil {
		return merr.ErrorInvalidArgument("update input is required")
	}
	sc := query.SSHCommand
	f := &in.Fields
	info, err := sc.WithContext(ctx).Where(sc.ID.Eq(in.UID.Int64())).Updates(&do.SSHCommand{
		Name:        f.Name,
		Description: f.Description,
		Content:     f.Content,
		WorkDir:     f.WorkDir,
		Env:         convert.ToSafetyEnv(f.Env),
	})
	if err != nil {
		return err
	}
	if info.RowsAffected == 0 {
		return merr.ErrorNotFound("ssh command not found")
	}
	return nil
}

func (r *sshCommandRepository) Get(ctx context.Context, uid snowflake.ID) (*bo.SSHCommandItemBo, error) {
	sc := query.SSHCommand
	row, err := sc.WithContext(ctx).Where(sc.ID.Eq(uid.Int64())).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("ssh command not found")
		}
		return nil, err
	}
	return convert.ToSSHCommandItemBo(row), nil
}

func (r *sshCommandRepository) List(ctx context.Context, req *bo.ListSSHCommandsBo) (*bo.PageResponseBo[*bo.SSHCommandItemBo], error) {
	sc := query.SSHCommand
	w := sc.WithContext(ctx)
	if strutil.IsNotEmpty(req.Keyword) {
		w = w.Where(sc.Name.Like("%" + req.Keyword + "%"))
	}
	if req != nil && req.PageRequestBo != nil {
		total, err := w.Count()
		if err != nil {
			return nil, err
		}
		req.WithTotal(total)
		w = w.Limit(req.Limit()).Offset(req.Offset())
	}
	rows, err := w.Order(sc.CreatedAt.Desc()).Find()
	if err != nil {
		return nil, err
	}
	items := make([]*bo.SSHCommandItemBo, 0, len(rows))
	for _, row := range rows {
		items = append(items, convert.ToSSHCommandItemBo(row))
	}
	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}

func (r *sshCommandRepository) CountByName(ctx context.Context, in *bo.SSHCommandCountByNameBo) (int64, error) {
	if in == nil {
		return 0, merr.ErrorInvalidArgument("count input is required")
	}
	sc := query.SSHCommand
	w := sc.WithContext(ctx).Where(sc.Name.Eq(in.Name))
	if in.ExcludeUID > 0 {
		w = w.Where(sc.ID.Neq(in.ExcludeUID.Int64()))
	}
	return w.Count()
}
