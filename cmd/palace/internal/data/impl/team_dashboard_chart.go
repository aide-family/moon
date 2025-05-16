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
	"github.com/aide-family/moon/cmd/palace/internal/data"
	"github.com/aide-family/moon/pkg/util/slices"
)

// NewDashboardChartRepo creates a new dashboard chart repository
func NewDashboardChartRepo(data *data.Data, logger log.Logger) repository.DashboardChart {
	return &dashboardChartImpl{
		Data:   data,
		helper: log.NewHelper(log.With(logger, "module", "data.repo.dashboard_chart")),
	}
}

type dashboardChartImpl struct {
	*data.Data

	helper *log.Helper
}

func (r *dashboardChartImpl) DeleteDashboardChartByDashboardID(ctx context.Context, dashboardID uint32) error {
	tx, teamID := getTeamBizQueryWithTeamID(ctx, r)
	mutation := tx.DashboardChart
	wrapper := []gen.Condition{
		mutation.TeamID.Eq(teamID),
		mutation.DashboardID.Eq(dashboardID),
	}
	_, err := mutation.WithContext(ctx).Where(wrapper...).Delete()
	return err
}

func (r *dashboardChartImpl) CreateDashboardChart(ctx context.Context, chart bo.DashboardChart) error {
	dashboardChartDo := &team.DashboardChart{
		DashboardID: chart.GetDashboardID(),
		Title:       chart.GetTitle(),
		Remark:      chart.GetRemark(),
		Status:      chart.GetStatus(),
		Url:         chart.GetUrl(),
		Width:       chart.GetWidth(),
		Height:      chart.GetHeight(),
	}
	dashboardChartDo.WithContext(ctx)
	tx := getTeamBizQuery(ctx, r)
	return tx.DashboardChart.WithContext(ctx).Create(dashboardChartDo)
}

func (r *dashboardChartImpl) UpdateDashboardChart(ctx context.Context, chart bo.DashboardChart) error {
	tx, teamID := getTeamBizQueryWithTeamID(ctx, r)
	mutation := tx.DashboardChart
	wrapper := []gen.Condition{
		mutation.TeamID.Eq(teamID),
		mutation.ID.Eq(chart.GetID()),
	}
	updates := []field.AssignExpr{
		mutation.Title.Value(chart.GetTitle()),
		mutation.Remark.Value(chart.GetRemark()),
		mutation.Status.Value(chart.GetStatus().GetValue()),
		mutation.Url.Value(chart.GetUrl()),
		mutation.Width.Value(chart.GetWidth()),
		mutation.Height.Value(chart.GetHeight()),
	}
	_, err := mutation.WithContext(ctx).Where(wrapper...).UpdateColumnSimple(updates...)
	return err
}

// DeleteDashboardChart delete dashboard chart by id
func (r *dashboardChartImpl) DeleteDashboardChart(ctx context.Context, req *bo.OperateOneDashboardChartReq) error {
	tx, teamID := getTeamBizQueryWithTeamID(ctx, r)
	mutation := tx.DashboardChart
	wrapper := []gen.Condition{
		mutation.TeamID.Eq(teamID),
		mutation.ID.Eq(req.ID),
		mutation.DashboardID.Eq(req.DashboardID),
	}
	_, err := mutation.WithContext(ctx).Where(wrapper...).Delete()
	return err
}

// GetDashboardChart get dashboard chart by id
func (r *dashboardChartImpl) GetDashboardChart(ctx context.Context, req *bo.OperateOneDashboardChartReq) (do.DashboardChart, error) {
	tx, teamID := getTeamBizQueryWithTeamID(ctx, r)
	mutation := tx.DashboardChart
	wrapper := []gen.Condition{
		mutation.TeamID.Eq(teamID),
		mutation.ID.Eq(req.ID),
		mutation.DashboardID.Eq(req.DashboardID),
	}
	dashboardChartDo, err := mutation.WithContext(ctx).Where(wrapper...).First()
	if err != nil {
		return nil, teamDashboardChartNotFound(err)
	}
	return dashboardChartDo, nil
}

// ListDashboardCharts list dashboard charts with filter
func (r *dashboardChartImpl) ListDashboardCharts(ctx context.Context, req *bo.ListDashboardChartReq) (*bo.ListDashboardChartReply, error) {
	tx, teamID := getTeamBizQueryWithTeamID(ctx, r)
	mutation := tx.DashboardChart
	query := mutation.WithContext(ctx).Where(mutation.TeamID.Eq(teamID), mutation.DashboardID.Eq(req.DashboardID))

	if !req.Status.IsUnknown() {
		query = query.Where(mutation.Status.Eq(req.Status.GetValue()))
	}

	if req.PaginationRequest != nil {
		total, err := query.Count()
		if err != nil {
			return nil, err
		}
		req.WithTotal(total)
		query = query.Offset(req.Offset()).Limit(int(req.Limit))
	}

	charts, err := query.Find()
	if err != nil {
		return nil, err
	}
	rows := slices.Map(charts, func(chart *team.DashboardChart) do.DashboardChart { return chart })
	return req.ToListReply(rows), nil
}

// BatchUpdateDashboardChartStatus update multiple dashboard charts status
func (r *dashboardChartImpl) BatchUpdateDashboardChartStatus(ctx context.Context, req *bo.BatchUpdateDashboardChartStatusReq) error {
	tx, teamID := getTeamBizQueryWithTeamID(ctx, r)
	mutation := tx.DashboardChart
	wrapper := []gen.Condition{
		mutation.TeamID.Eq(teamID),
		mutation.ID.In(req.Ids...),
	}
	_, err := mutation.WithContext(ctx).Where(wrapper...).UpdateColumnSimple(mutation.Status.Value(req.Status.GetValue()))
	return err
}
