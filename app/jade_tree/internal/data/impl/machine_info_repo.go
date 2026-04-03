package impl

import (
	"context"

	"github.com/aide-family/magicbox/merr"
	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/biz/repository"
	"github.com/aide-family/jade_tree/internal/data/impl/convert"
	"github.com/aide-family/jade_tree/internal/data/impl/do"
	"github.com/aide-family/jade_tree/internal/data/impl/query"
	"github.com/aide-family/jade_tree/pkg/machine"
)

var _ repository.MachineInfoProvider = (*machineInfoRepository)(nil)

func (m *machineInfoRepository) GetMachineInfoByIdentity(ctx context.Context, id *bo.MachineInfoIdentityBo) (*machine.MachineInfo, error) {
	if !m.enabledCollectSelf {
		return nil, merr.ErrorParams("collect self is not enabled")
	}
	if id == nil || id.MachineUUID == "" {
		return nil, merr.ErrorInvalidArgument("machine identity is required")
	}
	row, err := query.MachineInfo.WithContext(ctx).
		Where(query.MachineInfo.MachineUUID.Eq(id.MachineUUID)).
		Where(query.MachineInfo.HostName.Eq(id.HostName)).
		Where(query.MachineInfo.LocalIP.Eq(id.LocalIP)).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorNotFound("machine info not found")
		}
		return nil, err
	}
	return convert.ToMachineInfoItemBo(row)
}

func (m *machineInfoRepository) UpsertMachineInfos(ctx context.Context, machines []*machine.MachineInfo) error {
	if len(machines) == 0 {
		return nil
	}

	rows := make([]*do.MachineInfo, 0, len(machines))
	for _, mi := range machines {
		if mi == nil {
			continue
		}
		row, err := convert.ToMachineInfoDO(mi)
		if err != nil || row == nil || row.MachineUUID == "" {
			continue
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		return nil
	}

	return m.upsertMachineInfosByIdentity(ctx, rows)
}

// upsertMachineInfosByIdentity upserts rows by composite natural key (machine_uuid, host_name, local_ip).
func (m *machineInfoRepository) upsertMachineInfosByIdentity(ctx context.Context, rows []*do.MachineInfo) error {
	if len(rows) == 0 {
		return nil
	}
	table := query.MachineInfo
	return m.DB().WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: table.MachineUUID.ColumnName().String()},
			{Name: table.HostName.ColumnName().String()},
			{Name: table.LocalIP.ColumnName().String()},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			table.Source.ColumnName().String(),
			table.Info.ColumnName().String(),
			table.UpdatedAt.ColumnName().String(),
			table.DeletedAt.ColumnName().String(),
		}),
	}).Create(&rows).Error
}

func (m *machineInfoRepository) UpdateLocalMachineInfo(ctx context.Context, machine *machine.MachineInfo) error {
	if !m.enabledCollectSelf {
		return merr.ErrorParams("collect self is not enabled")
	}
	if machine == nil || machine.MachineUUID == "" {
		return merr.ErrorInvalidArgument("machine is required")
	}
	row, err := convert.ToMachineInfoDO(machine)
	if err != nil {
		return err
	}
	return m.upsertMachineInfosByIdentity(ctx, []*do.MachineInfo{row})
}

func (m *machineInfoRepository) ListMachineInfos(ctx context.Context, req *bo.ListMachineInfosBo) (*bo.PageResponseBo[*machine.MachineInfo], error) {
	if req == nil {
		return nil, merr.ErrorInvalidArgument("list request is required")
	}
	if req.PageRequestBo == nil || req.Page == 0 || req.PageSize == 0 {
		// Keep behavior aligned with constructors: page/pageSize should be normalized.
		return nil, merr.ErrorInvalidArgument("page and pageSize are required")
	}

	total, err := query.MachineInfo.WithContext(ctx).Count()
	if err != nil {
		return nil, err
	}
	req.WithTotal(total)

	rows, err := query.MachineInfo.WithContext(ctx).
		Order(query.MachineInfo.UpdatedAt.Desc()).
		Limit(req.Limit()).
		Offset(req.Offset()).
		Find()
	if err != nil {
		return nil, err
	}

	items := make([]*machine.MachineInfo, 0, len(rows))
	for _, row := range rows {
		item, err := convert.ToMachineInfoItemBo(row)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return bo.NewPageResponseBo(req.PageRequestBo, items), nil
}
