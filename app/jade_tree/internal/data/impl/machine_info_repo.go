package impl

import (
	"context"
	"strings"

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

	listQuery := query.MachineInfo.WithContext(ctx)
	listQuery = applyMachineInfoSearchFilters(listQuery, req)

	total, err := listQuery.Count()
	if err != nil {
		return nil, err
	}
	req.WithTotal(total)

	rows, err := listQuery.
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

func (m *machineInfoRepository) CountDispatchTargets(ctx context.Context, filter *bo.DispatchSSHCommandFilterBo) (int64, error) {
	q := query.MachineInfo.WithContext(ctx)
	q = applyDispatchFilters(q, filter)
	return q.Count()
}

func (m *machineInfoRepository) ListDispatchTargets(ctx context.Context, filter *bo.DispatchSSHCommandFilterBo) ([]*machine.MachineInfo, error) {
	q := query.MachineInfo.WithContext(ctx)
	q = applyDispatchFilters(q, filter)
	rows, err := q.Order(query.MachineInfo.UpdatedAt.Desc()).Find()
	if err != nil {
		return nil, err
	}
	items := make([]*machine.MachineInfo, 0, len(rows))
	for _, row := range rows {
		item, convErr := convert.ToMachineInfoItemBo(row)
		if convErr != nil {
			return nil, convErr
		}
		items = append(items, item)
	}
	return items, nil
}

func applyMachineInfoSearchFilters(q query.IMachineInfoDo, req *bo.ListMachineInfosBo) query.IMachineInfoDo {
	if req == nil {
		return q
	}
	if req.Keywords != "" {
		kw := "%" + escapeSQLLike(req.Keywords) + "%"
		q = q.Where(
			q.Where(query.MachineInfo.MachineUUID.Like(kw)).
				Or(query.MachineInfo.HostName.Like(kw)).
				Or(query.MachineInfo.LocalIP.Like(kw)),
		)
	}
	if req.IP != "" {
		ip := "%" + escapeSQLLike(req.IP) + "%"
		q = q.Where(query.MachineInfo.LocalIP.Like(ip))
	}
	if req.Hostname != "" {
		hostname := "%" + escapeSQLLike(req.Hostname) + "%"
		q = q.Where(query.MachineInfo.HostName.Like(hostname))
	}
	return q
}

func escapeSQLLike(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return replacer.Replace(value)
}

func applyDispatchFilters(q query.IMachineInfoDo, filter *bo.DispatchSSHCommandFilterBo) query.IMachineInfoDo {
	if filter == nil {
		return q
	}

	excludeIDSet := make(map[int64]struct{}, len(filter.ExcludeMachineUIDs))
	excludeIDs := make([]int64, 0, len(filter.ExcludeMachineUIDs))
	for _, id := range filter.ExcludeMachineUIDs {
		if id <= 0 {
			continue
		}
		idInt64 := id.Int64()
		if _, ok := excludeIDSet[idInt64]; ok {
			continue
		}
		excludeIDSet[idInt64] = struct{}{}
		excludeIDs = append(excludeIDs, idInt64)
	}

	if len(filter.IncludeMachineUIDs) > 0 {
		includeIDSet := make(map[int64]struct{}, len(filter.IncludeMachineUIDs))
		includeIDs := make([]int64, 0, len(filter.IncludeMachineUIDs))
		for _, id := range filter.IncludeMachineUIDs {
			if id <= 0 {
				continue
			}
			idInt64 := id.Int64()
			if _, ok := includeIDSet[idInt64]; ok {
				continue
			}
			includeIDSet[idInt64] = struct{}{}
			includeIDs = append(includeIDs, idInt64)
		}
		if len(includeIDs) == 0 {
			return q.Where(query.MachineInfo.ID.Eq(0))
		}
		q = q.Where(query.MachineInfo.ID.In(includeIDs...))

		// Explicit include has higher priority than exclude to preserve caller intent.
		if len(excludeIDs) > 0 {
			filteredExclude := make([]int64, 0, len(excludeIDs))
			for _, id := range excludeIDs {
				if _, included := includeIDSet[id]; included {
					continue
				}
				filteredExclude = append(filteredExclude, id)
			}
			excludeIDs = filteredExclude
		}
	}
	if len(filter.IncludeSystemTypes) > 0 {
		q = q.Where(query.MachineInfo.OSType.In(filter.IncludeSystemTypes...))
	}
	if len(filter.IncludeAgentVersions) > 0 {
		q = q.Where(query.MachineInfo.AgentVersion.In(filter.IncludeAgentVersions...))
	}

	if len(excludeIDs) > 0 {
		q = q.Where(query.MachineInfo.ID.NotIn(excludeIDs...))
	}
	if len(filter.ExcludeSystemTypes) > 0 {
		q = q.Where(query.MachineInfo.OSType.NotIn(filter.ExcludeSystemTypes...))
	}
	if len(filter.ExcludeAgentVersions) > 0 {
		q = q.Where(query.MachineInfo.AgentVersion.NotIn(filter.ExcludeAgentVersions...))
	}
	return q
}
