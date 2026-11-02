package service

import (
	"context"
	"time"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/probesla"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

// ProbeSlaService 把 SLA 查询转成 biz 入参。
type ProbeSlaService struct {
	v1.UnimplementedProbeSlaServiceServer
	gateway *probesla.Gateway
}

func NewProbeSlaService(gateway *probesla.Gateway) *ProbeSlaService {
	return &ProbeSlaService{gateway: gateway}
}

func (s *ProbeSlaService) GetProbeSla(ctx context.Context, req *v1.GetProbeSlaRequest) (*v1.GetProbeSlaResponse, error) {
	from, err := parseSlaTime(req.GetFrom())
	if err != nil {
		return nil, err
	}
	to, err := parseSlaTime(req.GetTo())
	if err != nil {
		return nil, err
	}
	report, err := s.gateway.Report(ctx, req.GetGroupUids(), from, to)
	if err != nil {
		return nil, err
	}
	groups := make([]*v1.ProbeSlaGroup, 0, len(report.Groups))
	for _, group := range report.Groups {
		groups = append(groups, probeSlaGroup(group))
	}
	return &v1.GetProbeSlaResponse{
		From: report.From.UTC().Format(time.RFC3339), To: report.To.UTC().Format(time.RFC3339),
		Groups: groups,
	}, nil
}

func parseSlaTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		parsed, err = time.Parse(time.RFC3339Nano, value)
	}
	if err != nil {
		return time.Time{}, merr.Invalid("invalid time")
	}
	return parsed.UTC(), nil
}

func probeSlaGroup(group bo.ProbeSlaGroup) *v1.ProbeSlaGroup {
	days := make([]*v1.ProbeSlaDay, 0, len(group.Days))
	for _, day := range group.Days {
		days = append(days, &v1.ProbeSlaDay{
			Day: day.Day, Checks: int32(day.Checks), OkChecks: int32(day.OkChecks),
			FailedChecks: int32(day.FailedChecks), Availability: day.Availability,
			AvgDurationMs: int32(day.AvgDurationMS),
		})
	}
	return &v1.ProbeSlaGroup{
		GroupUid: group.GroupUID, GroupName: group.GroupName,
		Checks: int32(group.Checks), OkChecks: int32(group.OkChecks),
		FailedChecks: int32(group.FailedChecks), Availability: group.Availability,
		AvgDurationMs: int32(group.AvgDurationMS), P95DurationMs: int32(group.P95DurationMS),
		P95Approximate: group.P95Approximate, EventCount: int32(group.EventCount), Days: days,
	}
}
