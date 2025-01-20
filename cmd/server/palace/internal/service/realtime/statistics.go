package realtime

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sort"
	"strconv"
	"time"

	"github.com/aide-family/moon/api/admin"
	pb "github.com/aide-family/moon/api/admin/realtime"
	"github.com/aide-family/moon/pkg/util/types"
)

type StatisticsService struct {
	pb.UnimplementedStatisticsServer
}

func NewStatisticsService() *StatisticsService {
	return &StatisticsService{}
}

func (s *StatisticsService) SummaryAlarm(ctx context.Context, req *pb.SummaryAlarmRequest) (*pb.SummaryAlarmReply, error) {
	randomChartData := make([]float64, 24)
	for i := 0; i < 24; i++ {
		randomChartData[i] = rand.Float64() * 1000
	}
	rand.Shuffle(len(randomChartData), func(i, j int) {
		randomChartData[i], randomChartData[j] = randomChartData[j], randomChartData[i]
	})
	return &pb.SummaryAlarmReply{
		Total:                     rand.Int64N(1000),
		Ongoing:                   rand.Int64N(500),
		Recovered:                 rand.Int64N(1000),
		HighestPriority:           rand.Int64N(100),
		ChartData:                 randomChartData,
		TotalComparison:           fmt.Sprintf("%.2f%%", rand.Float64()*100),
		OngoingComparison:         fmt.Sprintf("%.2f%%", rand.Float64()*100),
		RecoveredComparison:       fmt.Sprintf("%.2f%%", rand.Float64()*100),
		HighestPriorityComparison: fmt.Sprintf("%.2f%%", rand.Float64()*100),
	}, nil
}

func (s *StatisticsService) SummaryNotice(ctx context.Context, req *pb.SummaryNoticeRequest) (*pb.SummaryNoticeReply, error) {
	randomChartData := make([]float64, 24)
	for i := 0; i < 24; i++ {
		randomChartData[i] = rand.Float64() * 1000
	}
	rand.Shuffle(len(randomChartData), func(i, j int) {
		randomChartData[i], randomChartData[j] = randomChartData[j], randomChartData[i]
	})
	return &pb.SummaryNoticeReply{
		Total:            rand.Int64N(1000),
		Failed:           rand.Int64N(500),
		ChartData:        randomChartData,
		TotalComparison:  fmt.Sprintf("%.2f%%", rand.Float64()*100),
		FailedComparison: fmt.Sprintf("%.2f%%", rand.Float64()*100),
		NotifyTypes:      []*pb.SummaryNoticeReply_NotifyType{},
	}, nil
}

func (s *StatisticsService) TopStrategyAlarm(ctx context.Context, req *pb.TopStrategyAlarmRequest) (*pb.TopStrategyAlarmReply, error) {
	topN := make([]*pb.TopStrategyAlarmReply_StrategyAlarmTopN, 0, req.GetLimit())
	for i := 0; i < int(req.GetLimit()); i++ {
		topN = append(topN, &pb.TopStrategyAlarmReply_StrategyAlarmTopN{
			StrategyId:   rand.Uint64(),
			Total:        rand.Int64N(1000),
			StrategyName: "策略" + strconv.Itoa(i),
		})
	}
	sort.Slice(topN, func(i, j int) bool {
		return topN[i].Total > topN[j].Total
	})
	return &pb.TopStrategyAlarmReply{
		TopN: topN,
	}, nil
}

func (s *StatisticsService) LatestAlarmEvent(ctx context.Context, req *pb.LatestAlarmEventRequest) (*pb.LatestAlarmEventReply, error) {
	events := make([]*pb.LatestAlarmEventReply_LatestAlarmEvent, 0, req.GetLimit())
	for i := 0; i < int(req.GetLimit()); i++ {
		events = append(events, &pb.LatestAlarmEventReply_LatestAlarmEvent{
			Fingerprint: types.MD5(time.Now().Format(time.RFC3339)),
			Level:       "P0",
			EventTime:   time.Now().Add(-time.Second).Format(time.RFC3339),
			Summary:     "事件" + strconv.Itoa(i),
		})
	}
	return &pb.LatestAlarmEventReply{
		Events: events,
	}, nil
}

func (s *StatisticsService) LatestInterventionEvent(ctx context.Context, req *pb.LatestInterventionEventRequest) (*pb.LatestInterventionEventReply, error) {
	events := make([]*pb.LatestInterventionEventReply_LatestInterventionEvent, 0, req.GetLimit())
	for i := 0; i < int(req.GetLimit()); i++ {
		events = append(events, &pb.LatestInterventionEventReply_LatestInterventionEvent{
			Fingerprint: types.MD5(time.Now().Format(time.RFC3339)),
			EventTime:   time.Now().Add(-time.Second).Format(time.RFC3339),
			Summary:     "事件" + strconv.Itoa(i),
			Level:       "P0",
			Status:      "处理中",
			Handler: &admin.UserItem{
				Id:   uint32(rand.Uint64()),
				Name: "处理人" + strconv.Itoa(i),
			},
			HandledAt: time.Now().Add(-time.Second).Format(time.RFC3339),
		})
	}
	return &pb.LatestInterventionEventReply{
		Events: events,
	}, nil
}
