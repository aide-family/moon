package bo

import (
	"github.com/bwmarrin/snowflake"

	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

// AlertStatisticsBo holds alert statistics for the dashboard.
type AlertStatisticsBo struct {
	TotalActiveCount    int64
	CountByLevel        []LevelCountBo
	TodayRecoveredCount int64
	CountByAlertPage    []AlertPageCountBo
}

// LevelCountBo is count of active alerts for one level.
type LevelCountBo struct {
	LevelUID  snowflake.ID
	LevelName string
	Count     int64
}

// AlertPageCountBo is count of active alerts for one alert page.
type AlertPageCountBo struct {
	AlertPageUID  snowflake.ID
	AlertPageName string
	Count         int64
}

// ToAPIV1GetAlertStatisticsReply converts AlertStatisticsBo to proto reply.
func ToAPIV1GetAlertStatisticsReply(b *AlertStatisticsBo) *apiv1.GetAlertStatisticsReply {
	if b == nil {
		return &apiv1.GetAlertStatisticsReply{}
	}
	byLevel := make([]*apiv1.LevelCount, 0, len(b.CountByLevel))
	for _, c := range b.CountByLevel {
		byLevel = append(byLevel, &apiv1.LevelCount{
			LevelUid:  c.LevelUID.Int64(),
			LevelName: c.LevelName,
			Count:     c.Count,
		})
	}
	byPage := make([]*apiv1.AlertPageCount, 0, len(b.CountByAlertPage))
	for _, c := range b.CountByAlertPage {
		byPage = append(byPage, &apiv1.AlertPageCount{
			AlertPageUid:  c.AlertPageUID.Int64(),
			AlertPageName: c.AlertPageName,
			Count:         c.Count,
		})
	}
	return &apiv1.GetAlertStatisticsReply{
		TotalActiveCount:    b.TotalActiveCount,
		CountByLevel:        byLevel,
		TodayRecoveredCount: b.TodayRecoveredCount,
		CountByAlertPage:    byPage,
	}
}
