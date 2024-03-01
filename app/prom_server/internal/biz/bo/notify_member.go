package bo

import (
	"encoding/json"

	"prometheus-manager/api"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/util/slices"
)

type (
	NotifyMemberBO struct {
		Id          uint32         `json:"id"`
		Status      vo.Status      `json:"status"`
		CreatedAt   int64          `json:"createdAt"`
		UpdatedAt   int64          `json:"updatedAt"`
		DeletedAt   int64          `json:"deletedAt"`
		MemberId    uint32         `json:"memberId"`
		Member      *UserBO        `json:"member"`
		NotifyTypes vo.NotifyTypes `json:"notifyTypes"`
	}
)

// String json string
func (b *NotifyMemberBO) String() string {
	if b == nil {
		return "{}"
	}
	marshal, err := json.Marshal(b)
	if err != nil {
		return "{}"
	}
	return string(marshal)
}

// GetMember 获取用户详情
func (b *NotifyMemberBO) GetMember() *UserBO {
	if b == nil {
		return nil
	}
	return b.Member
}

// GetNotifyTypes 获取通知类型
func (b *NotifyMemberBO) GetNotifyTypes() vo.NotifyTypes {
	if b == nil {
		return nil
	}
	return b.NotifyTypes
}

func (b *NotifyMemberBO) ToModel() *do.PromAlarmNotifyMember {
	if b == nil {
		return nil
	}
	return &do.PromAlarmNotifyMember{
		BaseModel:   do.BaseModel{ID: b.Id},
		Status:      b.Status,
		NotifyTypes: b.GetNotifyTypes(),
		MemberId:    b.MemberId,
		Member:      b.GetMember().ToModel(),
	}
}

// ToApi ...
func (b *NotifyMemberBO) ToApi() *api.BeNotifyMemberDetail {
	if b == nil {
		return nil
	}

	return &api.BeNotifyMemberDetail{
		MemberId:    b.MemberId,
		NotifyTypes: slices.To(b.GetNotifyTypes(), func(i vo.NotifyType) int32 { return i.Value() }),
		User:        b.GetMember().ToApiSelectV1(),
		Status:      b.Status.Value(),
		Id:          b.Id,
	}
}

// NotifyMemberApiToBO ...
func NotifyMemberApiToBO(a *api.BeNotifyMember) *NotifyMemberBO {
	if a == nil {
		return nil
	}
	return &NotifyMemberBO{
		Id:          a.Id,
		MemberId:    a.GetMemberId(),
		Member:      &UserBO{Id: a.GetMemberId()},
		NotifyTypes: slices.To(a.GetNotifyTypes(), func(i api.NotifyType) vo.NotifyType { return vo.NotifyType(i) }),
	}
}

// NotifyMemberModelToBO ...
func NotifyMemberModelToBO(m *do.PromAlarmNotifyMember) *NotifyMemberBO {
	if m == nil {
		return nil
	}
	return &NotifyMemberBO{
		Id:          m.ID,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt.Unix(),
		UpdatedAt:   m.UpdatedAt.Unix(),
		DeletedAt:   int64(m.DeletedAt),
		MemberId:    m.MemberId,
		Member:      UserModelToBO(m.GetMember()),
		NotifyTypes: m.GetNotifyTypes(),
	}
}
