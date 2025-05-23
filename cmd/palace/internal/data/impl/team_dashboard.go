package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

// NewDashboardRepo creates a new dashboard repository
func NewDashboardRepo(data *data.Data, logger log.Logger) repository.Dashboard {
	return &dashboardRepoImpl{
		Data:   data,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.dashboard")),
	}
}

type dashboardRepoImpl struct {
	*data.Data

	helper *log.Helper
}

func (r *dashboardRepoImpl) CreateDashboard(ctx context.Context, dashboard bo.Dashboard) error {
	query := getTeamBizQuery(ctx, r)
	dashboardDo := &team.Dashboard{
		Title:    dashboard.GetTitle(),
		Remark:   dashboard.GetRemark(),
		Status:   vobj.GlobalStatusEnable,
		ColorHex: dashboard.GetColorHex(),
	}
	dashboardDo.WithContext(ctx)
	return query.Dashboard.WithContext(ctx).Create(dashboardDo)
}

func (r *dashboardRepoImpl) UpdateDashboard(ctx context.Context, dashboard bo.Dashboard) error {
	query, teamID := getTeamBizQueryWithTeamID(ctx, r)
	mutation := query.Dashboard
	wrappers := []gen.Condition{
		mutation.TeamID.Eq(teamID),
		mutation.ID.Eq(dashboard.GetID()),
	}
	mutations := []field.AssignExpr{
		mutation.Title.Value(dashboard.GetTitle()),
		mutation.Remark.Value(dashboard.GetRemark()),
		mutation.ColorHex.Value(dashboard.GetColorHex()),
	}
	_, err := mutation.WithContext(ctx).Where(wrappers...).UpdateColumnSimple(mutations...)
	return err
}

// DeleteDashboard delete dashboard by id
func (r *dashboardRepoImpl) DeleteDashboard(ctx context.Context, id uint32) error {
	query, teamID := getTeamBizQueryWithTeamID(ctx, r)
	mutation := query.Dashboard
	wrappers := []gen.Condition{
		mutation.TeamID.Eq(teamID),
		mutation.ID.Eq(id),
	}
	_, err := mutation.WithContext(ctx).Where(wrappers...).Delete()
	return err
}

// GetDashboard get dashboard by id
func (r *dashboardRepoImpl) GetDashboard(ctx context.Context, id uint32) (do.Dashboard, error) {
	query, teamID := getTeamBizQueryWithTeamID(ctx, r)
	mutation := query.Dashboard
	wrappers := []gen.Condition{
		mutation.TeamID.Eq(teamID),
		mutation.ID.Eq(id),
	}
	dashboardDo, err := mutation.WithContext(ctx).Where(wrappers...).First()
	if err != nil {
		return nil, teamDashboardNotFound(err)
	}
	return dashboardDo, nil
}

// ListDashboards list dashboards with filter
func (r *dashboardRepoImpl) ListDashboards(ctx context.Context, req *bo.ListDashboardReq) (*bo.ListDashboardReply, error) {
	query, teamID := getTeamBizQueryWithTeamID(ctx, r)
	mutation := query.Dashboard
	wrapper := mutation.WithContext(ctx).Where(mutation.TeamID.Eq(teamID))

	if !req.Status.IsUnknown() {
		wrapper = wrapper.Where(mutation.Status.Eq(req.Status.GetValue()))
	}

	if validate.IsNotNil(req.PaginationRequest) {
		total, err := wrapper.Count()
		if err != nil {
			return nil, err
		}
		req.WithTotal(total)
		wrapper = wrapper.Offset(req.Offset()).Limit(int(req.Limit))
	}

	dashboards, err := wrapper.Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(dashboards, func(dashboard *team.Dashboard) do.Dashboard { return dashboard })
	return req.ToListReply(rows), nil
}

func (r *dashboardRepoImpl) SelectTeamDashboard(ctx context.Context, req *bo.SelectTeamDashboardReq) (*bo.SelectTeamDashboardReply, error) {
	tx, teamID := getTeamBizQueryWithTeamID(ctx, r)
	mutation := tx.Dashboard
	query := mutation.WithContext(ctx).Where(mutation.TeamID.Eq(teamID))
	if validate.TextIsNotNull(req.Keyword) {
		query = query.Where(mutation.Title.Like(req.Keyword))
	}
	if !req.Status.IsUnknown() {
		query = query.Where(mutation.Status.Eq(req.Status.GetValue()))
	}
	if validate.IsNotNil(req.PaginationRequest) {
		total, err := query.Count()
		if err != nil {
			return nil, err
		}
		req.WithTotal(total)
		query = query.Limit(int(req.Limit)).Offset(req.Offset())
	}

	selectColumns := []field.Expr{
		mutation.ID,
		mutation.Title,
		mutation.Remark,
		mutation.Status,
		mutation.DeletedAt,
	}
	dashboards, err := query.WithContext(ctx).Select(selectColumns...).Order(mutation.ID.Desc()).Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(dashboards, func(dashboard *team.Dashboard) do.Dashboard { return dashboard })
	return req.ToSelectReply(rows), nil
}

// BatchUpdateDashboardStatus update multiple dashboards status
func (r *dashboardRepoImpl) BatchUpdateDashboardStatus(ctx context.Context, req *bo.BatchUpdateDashboardStatusReq) error {
	if len(req.Ids) == 0 {
		return nil
	}
	query, teamID := getTeamBizQueryWithTeamID(ctx, r)
	mutation := query.Dashboard
	wrappers := []gen.Condition{
		mutation.TeamID.Eq(teamID),
		mutation.ID.In(req.Ids...),
	}
	_, err := mutation.WithContext(ctx).Where(wrappers...).UpdateColumnSimple(mutation.Status.Value(req.Status.GetValue()))
	return err
}
