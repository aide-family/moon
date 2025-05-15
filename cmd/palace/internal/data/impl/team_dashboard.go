package impl

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do/team"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/repository"
	"github.com/moon-monitor/moon/cmd/palace/internal/data"
)

// NewDashboardRepo creates a new dashboard repository
func NewDashboardRepo(data *data.Data, logger log.Logger) repository.Dashboard {
	return &dashboardImpl{
		Data:   data,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.dashboard")),
	}
}

type dashboardImpl struct {
	*data.Data

	helper *log.Helper
}

func (r *dashboardImpl) CreateDashboard(ctx context.Context, dashboard bo.Dashboard) error {
	query := getTeamBizQuery(ctx, r)
	dashboardDo := &team.Dashboard{
		Title:    dashboard.GetTitle(),
		Remark:   dashboard.GetRemark(),
		Status:   dashboard.GetStatus(),
		ColorHex: dashboard.GetColorHex(),
	}
	dashboardDo.WithContext(ctx)
	return query.Dashboard.WithContext(ctx).Create(dashboardDo)
}

func (r *dashboardImpl) UpdateDashboard(ctx context.Context, dashboard bo.Dashboard) error {
	query, teamID := getTeamBizQueryWithTeamID(ctx, r)
	mutation := query.Dashboard
	wrappers := []gen.Condition{
		mutation.TeamID.Eq(teamID),
		mutation.ID.Eq(dashboard.GetID()),
	}
	mutations := []field.AssignExpr{
		mutation.Title.Value(dashboard.GetTitle()),
		mutation.Remark.Value(dashboard.GetRemark()),
		mutation.Status.Value(dashboard.GetStatus().GetValue()),
		mutation.ColorHex.Value(dashboard.GetColorHex()),
	}
	_, err := mutation.WithContext(ctx).Where(wrappers...).UpdateColumnSimple(mutations...)
	return err
}

// DeleteDashboard delete dashboard by id
func (r *dashboardImpl) DeleteDashboard(ctx context.Context, id uint32) error {
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
func (r *dashboardImpl) GetDashboard(ctx context.Context, id uint32) (do.Dashboard, error) {
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
func (r *dashboardImpl) ListDashboards(ctx context.Context, req *bo.ListDashboardReq) (*bo.ListDashboardReply, error) {
	query, teamID := getTeamBizQueryWithTeamID(ctx, r)
	mutation := query.Dashboard
	wrapper := mutation.WithContext(ctx).Where(mutation.TeamID.Eq(teamID))

	if !req.Status.IsUnknown() {
		wrapper = wrapper.Where(mutation.Status.Eq(req.Status.GetValue()))
	}

	if req.PaginationRequest != nil {
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
	return req.ToListDashboardReply(dashboards), nil
}

// BatchUpdateDashboardStatus update multiple dashboards status
func (r *dashboardImpl) BatchUpdateDashboardStatus(ctx context.Context, req *bo.BatchUpdateDashboardStatusReq) error {
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
