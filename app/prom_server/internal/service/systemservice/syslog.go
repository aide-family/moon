package systemservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api"
	pb "prometheus-manager/api/server/system"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/util/slices"
)

type SyslogService struct {
	pb.UnimplementedSyslogServer
	log    *log.Helper
	logBiz *biz.SysLogBiz
}

func NewSyslogService(logBiz *biz.SysLogBiz, l log.Logger) *SyslogService {
	return &SyslogService{
		log:    log.NewHelper(log.With(l, "service", "SyslogService")),
		logBiz: logBiz,
	}
}

func (s *SyslogService) ListSyslog(ctx context.Context, req *pb.ListSyslogRequest) (*pb.ListSyslogReply, error) {
	pageReq := req.GetPage()
	pageInfo := bo.NewPage(pageReq.GetCurr(), pageReq.GetSize())
	logList, err := s.logBiz.ListSysLog(ctx, &bo.ListSyslogReq{
		Page:     pageInfo,
		Module:   vo.Module(req.GetModuleName()),
		ModuleId: req.GetModuleId(),
	})
	if err != nil {
		return nil, err
	}
	return &pb.ListSyslogReply{
		List: slices.To(logList, func(item *bo.SysLogBo) *api.SysLogV1Item { return item.ToApiV1() }),
		Page: &api.PageReply{
			Curr:  pageReq.GetCurr(),
			Size:  pageReq.GetSize(),
			Total: pageInfo.GetTotal(),
		},
	}, nil
}
