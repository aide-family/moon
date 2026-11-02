package service

import (
	"context"
	"time"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/oncallrotation"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

// OnCallRotationService 把值班轮换接口转成 biz 入参；排班计算与权限都在 biz。
type OnCallRotationService struct {
	v1.UnimplementedOnCallRotationServiceServer
	gateway *oncallrotation.Gateway
}

func NewOnCallRotationService(gateway *oncallrotation.Gateway) *OnCallRotationService {
	return &OnCallRotationService{gateway: gateway}
}

func (s *OnCallRotationService) CreateOnCallRotation(ctx context.Context, req *v1.CreateOnCallRotationRequest) (*v1.CreateOnCallRotationResponse, error) {
	in, err := onCallRotationInput("", req.GetName(), req.GetRemark(), req.GetTimezone(), req.GetPeriod(),
		req.GetStartsAt(), req.GetHandoffHour(), req.GetMembers())
	if err != nil {
		return nil, err
	}
	item, err := s.gateway.Create(ctx, in)
	if err != nil {
		return nil, err
	}
	return &v1.CreateOnCallRotationResponse{Rotation: onCallRotationItem(*item)}, nil
}

func (s *OnCallRotationService) UpdateOnCallRotation(ctx context.Context, req *v1.UpdateOnCallRotationRequest) (*v1.UpdateOnCallRotationResponse, error) {
	in, err := onCallRotationInput(req.GetUid(), req.GetName(), req.GetRemark(), req.GetTimezone(), req.GetPeriod(),
		req.GetStartsAt(), req.GetHandoffHour(), req.GetMembers())
	if err != nil {
		return nil, err
	}
	item, err := s.gateway.Update(ctx, in)
	if err != nil {
		return nil, err
	}
	return &v1.UpdateOnCallRotationResponse{Rotation: onCallRotationItem(*item)}, nil
}

func (s *OnCallRotationService) UpdateOnCallRotationStatus(ctx context.Context, req *v1.UpdateOnCallRotationStatusRequest) (*v1.UpdateOnCallRotationStatusResponse, error) {
	item, err := s.gateway.UpdateStatus(ctx, req.GetUid(), req.GetStatus())
	if err != nil {
		return nil, err
	}
	return &v1.UpdateOnCallRotationStatusResponse{Rotation: onCallRotationItem(*item)}, nil
}

func (s *OnCallRotationService) DeleteOnCallRotation(ctx context.Context, req *v1.DeleteOnCallRotationRequest) (*v1.DeleteOnCallRotationResponse, error) {
	if err := s.gateway.Delete(ctx, req.GetUid()); err != nil {
		return nil, err
	}
	return &v1.DeleteOnCallRotationResponse{}, nil
}

func (s *OnCallRotationService) GetOnCallRotation(ctx context.Context, req *v1.GetOnCallRotationRequest) (*v1.GetOnCallRotationResponse, error) {
	item, err := s.gateway.Get(ctx, req.GetUid())
	if err != nil {
		return nil, err
	}
	return &v1.GetOnCallRotationResponse{Rotation: onCallRotationItem(*item)}, nil
}

func (s *OnCallRotationService) ListOnCallRotations(ctx context.Context, req *v1.ListOnCallRotationsRequest) (*v1.ListOnCallRotationsResponse, error) {
	items, total, pageNo, pageSize, err := s.gateway.List(ctx, req.GetPage(), req.GetPageSize(), req.GetName())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.OnCallRotationItem, 0, len(items))
	for i := range items {
		out = append(out, onCallRotationItem(items[i]))
	}
	return &v1.ListOnCallRotationsResponse{
		Items: out, Total: int32(total), Page: int32(pageNo), PageSize: int32(pageSize),
	}, nil
}

func (s *OnCallRotationService) GetOnCallRotationSchedule(ctx context.Context, req *v1.GetOnCallRotationScheduleRequest) (*v1.GetOnCallRotationScheduleResponse, error) {
	shifts, err := s.gateway.Schedule(ctx, req.GetUid(), req.GetDays())
	if err != nil {
		return nil, err
	}
	out := make([]*v1.OnCallRotationShiftItem, 0, len(shifts))
	for _, shift := range shifts {
		out = append(out, &v1.OnCallRotationShiftItem{
			StartAt: formatTime(shift.StartAt), EndAt: formatTime(shift.EndAt),
			MemberUid: shift.Member.MemberUID, MemberName: shift.Member.Name,
			Email: shift.Member.Email, Current: shift.Current,
		})
	}
	return &v1.GetOnCallRotationScheduleResponse{Shifts: out}, nil
}

// onCallRotationInput 解析轮换请求。starts_at 由客户端按 RFC3339 提交。
func onCallRotationInput(uid, name, remark, timezone, period, startsAt string, handoffHour int32, members []*v1.OnCallRotationMemberInput) (bo.OnCallRotation, error) {
	items := make([]bo.OnCallRotationMember, 0, len(members))
	for _, member := range members {
		items = append(items, bo.OnCallRotationMember{MemberUID: member.GetMemberUid()})
	}
	start, err := parseRotationTime(startsAt)
	if err != nil {
		return bo.OnCallRotation{}, err
	}
	return bo.OnCallRotation{
		UID: uid, Name: name, Remark: remark, Timezone: timezone, Period: period,
		StartsAt: start, HandoffHour: int(handoffHour), Members: items,
	}, nil
}

// parseRotationTime 把 RFC3339 时间解析成 UTC；空值返回零值由 biz 拒绝。
func parseRotationTime(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, merr.Invalid("invalid rotation starts at")
	}
	return parsed.UTC(), nil
}

func onCallRotationItem(in bo.OnCallRotation) *v1.OnCallRotationItem {
	members := make([]*v1.OnCallRotationMemberItem, 0, len(in.Members))
	for _, member := range in.Members {
		members = append(members, &v1.OnCallRotationMemberItem{
			MemberUid: member.MemberUID, MemberName: member.Name, Email: member.Email,
			Phone: member.Phone, Position: int32(member.Position),
		})
	}
	return &v1.OnCallRotationItem{
		Uid: in.UID, Name: in.Name, Remark: in.Remark, Timezone: in.Timezone,
		Period: in.Period, StartsAt: formatTime(in.StartsAt), HandoffHour: int32(in.HandoffHour),
		Status: in.Status, CreatedAt: formatTime(in.CreatedAt), UpdatedAt: formatTime(in.UpdatedAt),
		Members: members,
	}
}
