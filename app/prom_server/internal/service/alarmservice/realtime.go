package alarmservice

import (
	"context"
	"time"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	pb "prometheus-manager/api/alarm/realtime"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/pkg/helper/middler"
	"prometheus-manager/pkg/helper/model/alarmscopes"
	"prometheus-manager/pkg/util/slices"
)

type RealtimeService struct {
	pb.UnimplementedRealtimeServer

	log           *log.Helper
	alarmRealtime *biz.AlarmRealtimeBiz
}

// NewRealtimeService 实时告警服务
func NewRealtimeService(alarmRealtime *biz.AlarmRealtimeBiz, logger log.Logger) *RealtimeService {
	return &RealtimeService{
		log:           log.NewHelper(log.With(logger, "module", "service.alarm.realtime")),
		alarmRealtime: alarmRealtime,
	}
}

// GetRealtime 实时告警详情
func (l *RealtimeService) GetRealtime(ctx context.Context, req *pb.GetRealtimeRequest) (*pb.GetRealtimeReply, error) {
	realtimeAlarmDetail, err := l.alarmRealtime.GetRealtimeDetailById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetRealtimeReply{
		Detail: realtimeAlarmDetail.ToApi(),
	}, nil
}

// ListRealtime 实时告警列表
func (l *RealtimeService) ListRealtime(ctx context.Context, req *pb.ListRealtimeRequest) (*pb.ListRealtimeReply, error) {
	pgReq := req.GetPage()
	pgInfo := query.NewPage(int(pgReq.GetCurr()), int(pgReq.GetSize()))
	wheres := []query.ScopeMethod{
		alarmscopes.RealtimeLike(req.GetKeyword()),
		alarmscopes.RealtimeEventAtDesc(),
	}

	realtimeAlarmList, err := l.alarmRealtime.GetRealtimeList(ctx, pgInfo, wheres...)
	if err != nil {
		return nil, err
	}
	return &pb.ListRealtimeReply{
		Page: &api.PageReply{
			Curr:  pgReq.GetCurr(),
			Size:  pgReq.GetSize(),
			Total: pgInfo.GetTotal(),
		},
		List: slices.To(realtimeAlarmList, func(t *bo.AlarmRealtimeBO) *api.RealtimeAlarmData {
			return t.ToApi()
		}),
	}, nil
}

// Intervene 告警干预
func (l *RealtimeService) Intervene(ctx context.Context, req *pb.InterveneRequest) (*pb.InterveneReply, error) {
	authClaims, ok := middler.GetAuthClaims(ctx)
	if !ok {
		return nil, middler.ErrTokenInvalid
	}

	alarmInterveneBO := &bo.AlarmInterveneBO{
		RealtimeAlarmID: uint(req.GetId()),
		UserID:          authClaims.ID,
		IntervenedAt:    time.Now().Unix(),
		Remark:          req.GetRemark(),
	}
	if err := l.alarmRealtime.AlarmIntervene(ctx, req.GetId(), alarmInterveneBO); err != nil {
		return nil, err
	}

	return &pb.InterveneReply{}, nil
}

// Upgrade 告警升级
func (l *RealtimeService) Upgrade(ctx context.Context, req *pb.UpgradeRequest) (*pb.UpgradeReply, error) {
	authClaims, ok := middler.GetAuthClaims(ctx)
	if !ok {
		return nil, middler.ErrTokenInvalid
	}

	alarmUpgradeBO := &bo.AlarmUpgradeBO{
		RealtimeAlarmID: uint(req.GetId()),
		UserID:          authClaims.ID,
		UpgradedAt:      time.Now().Unix(),
		Remark:          req.GetRemark(),
	}
	if err := l.alarmRealtime.AlarmUpgrade(ctx, req.GetId(), alarmUpgradeBO); err != nil {
		return nil, err
	}

	return &pb.UpgradeReply{}, nil
}

// Suppress 告警抑制
func (l *RealtimeService) Suppress(ctx context.Context, req *pb.SuppressRequest) (*pb.SuppressReply, error) {
	authClaims, ok := middler.GetAuthClaims(ctx)
	if !ok {
		return nil, middler.ErrTokenInvalid
	}

	alarmSuppressBO := &bo.AlarmSuppressBO{
		RealtimeAlarmID: uint(req.GetId()),
		UserID:          authClaims.ID,
		SuppressedAt:    time.Now().Unix(),
		Remark:          req.GetRemark(),
		Duration:        req.GetDuration(),
	}
	if err := l.alarmRealtime.AlarmSuppress(ctx, req.GetId(), alarmSuppressBO); err != nil {
		return nil, err
	}
	return &pb.SuppressReply{}, nil
}
