package bo

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

type SaveNoticeMemberItem struct {
	MemberID     uint32
	UserID       uint32
	NoticeType   vobj.NoticeType
	DutyCycleIds []uint32
	dutyCycle    []do.TimeEngine
}

func (m *SaveNoticeMemberItem) GetDutyCycles() []do.TimeEngine {
	if m == nil {
		return nil
	}
	return m.dutyCycle
}

type SaveNoticeGroup interface {
	GetHooks() []do.NoticeHook
	GetNoticeMembers() []*SaveNoticeMemberItem
	GetID() uint32
	GetName() string
	GetRemark() string
	GetStatus() vobj.GlobalStatus
	GetEmailConfig() do.TeamEmailConfig
	GetSMSConfig() do.TeamSMSConfig
	GetTimeEngineIds() []uint32
}

type SaveNoticeGroupReq struct {
	group         do.NoticeGroup
	hooks         []do.NoticeHook
	members       map[uint32]do.TeamMember
	emailConfig   do.TeamEmailConfig
	smsConfig     do.TeamSMSConfig
	timeEngines   []do.TimeEngine
	GroupID       uint32
	Name          string
	Remark        string
	HookIds       []uint32
	NoticeMembers []*SaveNoticeMemberItem
	EmailConfigID uint32
	SMSConfigID   uint32
}

func (r *SaveNoticeGroupReq) GetTimeEngineIds() []uint32 {
	if r == nil {
		return nil
	}
	ids := make([]uint32, 0, len(r.NoticeMembers)*2)
	for _, member := range r.NoticeMembers {
		ids = append(ids, member.DutyCycleIds...)
	}
	return ids
}

func (r *SaveNoticeGroupReq) GetHooks() []do.NoticeHook {
	if r == nil {
		return nil
	}
	return r.hooks
}

func (r *SaveNoticeGroupReq) GetNoticeMembers() []*SaveNoticeMemberItem {
	if r == nil {
		return nil
	}
	noticeMembers := make([]*SaveNoticeMemberItem, 0, len(r.members))
	for _, noticeMember := range r.NoticeMembers {
		if validate.IsNil(noticeMember) || noticeMember.MemberID <= 0 {
			continue
		}
		member, ok := r.members[noticeMember.MemberID]
		if !ok {
			continue
		}
		noticeMembers = append(noticeMembers, &SaveNoticeMemberItem{
			MemberID:     noticeMember.MemberID,
			UserID:       member.GetUserID(),
			NoticeType:   noticeMember.NoticeType,
			DutyCycleIds: noticeMember.DutyCycleIds,
		})
	}
	return noticeMembers
}

func (r *SaveNoticeGroupReq) GetMemberIds() []uint32 {
	if r == nil {
		return nil
	}
	return slices.MapFilter(r.NoticeMembers, func(member *SaveNoticeMemberItem) (uint32, bool) {
		if validate.IsNil(member) || member.MemberID <= 0 {
			return 0, false
		}
		return member.MemberID, true
	})
}

func (r *SaveNoticeGroupReq) GetID() uint32 {
	if r == nil {
		return 0
	}
	return r.GroupID
}

func (r *SaveNoticeGroupReq) GetName() string {
	if r == nil {
		return ""
	}
	return r.Name
}

func (r *SaveNoticeGroupReq) GetRemark() string {
	if r == nil {
		return ""
	}
	return r.Remark
}

func (r *SaveNoticeGroupReq) GetStatus() vobj.GlobalStatus {
	if r == nil {
		return vobj.GlobalStatusUnknown
	}
	if validate.IsNil(r.group) {
		return vobj.GlobalStatusEnable
	}
	return r.group.GetStatus()
}

func (r *SaveNoticeGroupReq) GetEmailConfig() do.TeamEmailConfig {
	if r == nil {
		return nil
	}
	return r.emailConfig
}

func (r *SaveNoticeGroupReq) GetSMSConfig() do.TeamSMSConfig {
	if r == nil {
		return nil
	}
	return r.smsConfig
}

func (r *SaveNoticeGroupReq) WithNoticeGroup(group do.NoticeGroup) *SaveNoticeGroupReq {
	r.group = group
	return r
}

func (r *SaveNoticeGroupReq) WithHooks(hooks []do.NoticeHook) *SaveNoticeGroupReq {
	r.hooks = slices.MapFilter(hooks, func(hook do.NoticeHook) (do.NoticeHook, bool) {
		if validate.IsNil(hook) || hook.GetID() <= 0 {
			return nil, false
		}
		return hook, true
	})
	return r
}

func (r *SaveNoticeGroupReq) WithNoticeMembers(members []do.TeamMember) *SaveNoticeGroupReq {
	r.members = make(map[uint32]do.TeamMember, len(members))
	for _, member := range members {
		if validate.IsNil(member) || member.GetID() <= 0 {
			continue
		}
		r.members[member.GetID()] = member
	}
	return r
}

func (r *SaveNoticeGroupReq) WithEmailConfig(config do.TeamEmailConfig) *SaveNoticeGroupReq {
	r.emailConfig = config
	return r
}

func (r *SaveNoticeGroupReq) WithSMSConfig(config do.TeamSMSConfig) *SaveNoticeGroupReq {
	r.smsConfig = config
	return r
}

func (r *SaveNoticeGroupReq) WithTimeEngines(engines []do.TimeEngine) *SaveNoticeGroupReq {
	r.timeEngines = slices.MapFilter(engines, func(engine do.TimeEngine) (do.TimeEngine, bool) {
		if validate.IsNil(engine) || engine.GetID() <= 0 {
			return nil, false
		}
		return engine, true
	})
	timeEnginesMap := slices.ToMap(r.timeEngines, func(engine do.TimeEngine) uint32 {
		return engine.GetID()
	})
	for _, member := range r.NoticeMembers {
		member.dutyCycle = slices.MapFilter(member.DutyCycleIds, func(id uint32) (do.TimeEngine, bool) {
			engine, ok := timeEnginesMap[id]
			return engine, ok
		})
	}
	return r
}

func (r *SaveNoticeGroupReq) Validate() error {
	if validate.IsNil(r.group) {
		return merr.ErrorParams("invalid notice group")
	}
	if len(r.timeEngines) != len(r.GetTimeEngineIds()) {
		return merr.ErrorParams("invalid time engines")
	}
	if len(r.hooks) != len(r.HookIds) {
		return merr.ErrorParams("invalid hooks")
	}
	if len(r.members) != len(r.GetMemberIds()) {
		return merr.ErrorParams("invalid members")
	}
	if r.EmailConfigID > 0 && validate.IsNil(r.emailConfig) {
		return merr.ErrorParams("invalid email config")
	}
	if r.SMSConfigID > 0 && validate.IsNil(r.smsConfig) {
		return merr.ErrorParams("invalid sms config")
	}
	return nil
}

type UpdateTeamNoticeGroupStatusRequest struct {
	GroupIds []uint32
	Status   vobj.GlobalStatus
}

type ListNoticeGroupReq struct {
	*PaginationRequest
	MemberIds []uint32
	Status    vobj.GlobalStatus
	Keyword   string
}

func (r *ListNoticeGroupReq) ToListReply(groups []do.NoticeGroup) *ListNoticeGroupReply {
	return &ListNoticeGroupReply{
		PaginationReply: r.ToReply(),
		Items:           groups,
	}
}

type ListNoticeGroupReply = ListReply[do.NoticeGroup]

type TeamNoticeGroupSelectRequest struct {
	*PaginationRequest
	Keyword string
	Status  vobj.GlobalStatus
}

func (r *TeamNoticeGroupSelectRequest) ToSelectReply(groups []do.NoticeGroup) *TeamNoticeGroupSelectReply {
	return &TeamNoticeGroupSelectReply{
		PaginationReply: r.ToReply(),
		Items: slices.Map(groups, func(group do.NoticeGroup) SelectItem {
			return &selectItem{
				Value:    group.GetID(),
				Label:    group.GetName(),
				Disabled: !group.GetStatus().IsEnable() || group.GetDeletedAt() != 0,
				Extra: &selectItemExtra{
					Remark: group.GetRemark(),
				},
			}
		}),
	}
}

type TeamNoticeGroupSelectReply = ListReply[SelectItem]
