package dobo

import (
	"time"

	query "github.com/aide-cloud/gorm-normalize"
	"prometheus-manager/api"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/valueobj"
	"prometheus-manager/pkg/util/slices"
)

type (
	NotifyMemberDO struct {
		Id          uint                 `json:"id"`
		Status      valueobj.Status      `json:"status"`
		CreatedAt   time.Time            `json:"createdAt"`
		UpdatedAt   time.Time            `json:"updatedAt"`
		DeletedAt   int64                `json:"deletedAt"`
		MemberId    uint                 `json:"memberId"`
		Member      *UserDO              `json:"member"`
		NotifyTypes valueobj.NotifyTypes `json:"notifyTypes"`
	}

	NotifyMemberBO struct {
		Id          uint                 `json:"id"`
		Status      valueobj.Status      `json:"status"`
		CreatedAt   int64                `json:"createdAt"`
		UpdatedAt   int64                `json:"updatedAt"`
		DeletedAt   int64                `json:"deletedAt"`
		MemberId    uint                 `json:"memberId"`
		Member      *UserBO              `json:"member"`
		NotifyTypes valueobj.NotifyTypes `json:"notifyTypes"`
	}
)

func NewNotifyMemberDO(values ...*NotifyMemberDO) IDO[*NotifyMemberBO, *NotifyMemberDO] {
	return NewDO[*NotifyMemberBO, *NotifyMemberDO](
		DOWithValues[*NotifyMemberBO, *NotifyMemberDO](values...),
		DOWithBToD[*NotifyMemberBO, *NotifyMemberDO](notifyMemberBoToDo),
		DOWithDToB[*NotifyMemberBO, *NotifyMemberDO](notifyMemberDoToBo),
	)
}

func NewNotifyMemberBO(values ...*NotifyMemberBO) IBO[*NotifyMemberBO, *NotifyMemberDO] {
	return NewBO[*NotifyMemberBO, *NotifyMemberDO](
		BOWithValues[*NotifyMemberBO, *NotifyMemberDO](values...),
		BOWithDToB[*NotifyMemberBO, *NotifyMemberDO](notifyMemberDoToBo),
		BOWithBToD[*NotifyMemberBO, *NotifyMemberDO](notifyMemberBoToDo),
	)
}

func notifyMemberDoToBo(d *NotifyMemberDO) *NotifyMemberBO {
	if d == nil {
		return nil
	}
	return &NotifyMemberBO{
		Id:        d.Id,
		Status:    d.Status,
		CreatedAt: d.CreatedAt.Unix(),
		UpdatedAt: d.UpdatedAt.Unix(),
		DeletedAt: d.DeletedAt,
		MemberId:  d.MemberId,
		Member:    userDoToBo(d.Member),
	}
}

func notifyMemberBoToDo(b *NotifyMemberBO) *NotifyMemberDO {
	if b == nil {
		return nil
	}
	return &NotifyMemberDO{
		Id:        b.Id,
		Status:    b.Status,
		CreatedAt: time.Unix(b.CreatedAt, 0),
		UpdatedAt: time.Unix(b.UpdatedAt, 0),
		DeletedAt: b.DeletedAt,
		MemberId:  b.MemberId,
		Member:    userBoToDo(b.Member),
	}
}

// ToModel ...
func (d *NotifyMemberDO) ToModel() *model.PromAlarmNotifyMember {
	if d == nil {
		return nil
	}
	return &model.PromAlarmNotifyMember{
		BaseModel:   query.BaseModel{ID: d.Id},
		Status:      d.Status,
		NotifyTypes: d.NotifyTypes,
		MemberId:    d.MemberId,
		Member:      d.Member.ToModel(),
	}
}

func (b *NotifyMemberBO) ToModel() *model.PromAlarmNotifyMember {
	if b == nil {
		return nil
	}
	return &model.PromAlarmNotifyMember{
		BaseModel:   query.BaseModel{ID: b.Id},
		Status:      b.Status,
		NotifyTypes: b.NotifyTypes,
		MemberId:    b.MemberId,
		Member:      b.Member.ToModel(),
	}
}

// ToApi ...
func (b *NotifyMemberBO) ToApi() *api.BeNotifyMemberDetail {
	if b == nil {
		return nil
	}

	return &api.BeNotifyMemberDetail{
		MemberId:    uint32(b.MemberId),
		NotifyTypes: slices.To(b.NotifyTypes, func(i valueobj.NotifyType) int32 { return i.Value() }),
		User:        b.Member.ToApiSelectV1(),
		Status:      b.Status.Value(),
		Id:          uint32(b.Id),
	}
}

// NotifyMemberModelToDO ...
func NotifyMemberModelToDO(m *model.PromAlarmNotifyMember) *NotifyMemberDO {
	if m == nil {
		return nil
	}
	return &NotifyMemberDO{
		Id:          m.ID,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   int64(m.DeletedAt),
		MemberId:    m.MemberId,
		Member:      UserModelToDO(m.Member),
		NotifyTypes: m.NotifyTypes,
	}
}

// NotifyMemberApiToBO ...
func NotifyMemberApiToBO(a *api.BeNotifyMember) *NotifyMemberBO {
	if a == nil {
		return nil
	}
	return &NotifyMemberBO{
		Id:          uint(a.Id),
		Status:      0,
		CreatedAt:   0,
		UpdatedAt:   0,
		DeletedAt:   0,
		MemberId:    uint(a.MemberId),
		Member:      &UserBO{Id: uint(a.MemberId)},
		NotifyTypes: slices.To(a.NotifyTypes, func(i api.NotifyType) valueobj.NotifyType { return valueobj.NotifyType(i) }),
	}
}

// NotifyMemberModelToBO ...
func NotifyMemberModelToBO(m *model.PromAlarmNotifyMember) *NotifyMemberBO {
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
		Member:      UserModelToBO(m.Member),
		NotifyTypes: m.NotifyTypes,
	}
}
