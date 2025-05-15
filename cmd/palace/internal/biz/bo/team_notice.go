package bo

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/team"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
)

type SaveNoticeMemberItem struct {
	MemberID   uint32          `json:"memberID"`
	UserID     uint32          `json:"userID"`
	NoticeType vobj.NoticeType `json:"noticeType"`
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
}

type SaveNoticeGroupReq struct {
	group         do.NoticeGroup
	hooks         []do.NoticeHook
	members       map[uint32]do.TeamMember
	emailConfig   do.TeamEmailConfig
	smsConfig     do.TeamSMSConfig
	GroupID       uint32                  `json:"groupId"`
	Name          string                  `json:"name"`
	Remark        string                  `json:"remark"`
	HookIds       []uint32                `json:"hookIds"`
	NoticeMembers []*SaveNoticeMemberItem `json:"noticeMembers"`
	EmailConfigID uint32                  `json:"emailConfigId"`
	SMSConfigID   uint32                  `json:"smsConfigId"`
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
			MemberID:   noticeMember.MemberID,
			UserID:     member.GetUserID(),
			NoticeType: noticeMember.NoticeType,
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

func (r *SaveNoticeGroupReq) Validate() error {
	if validate.IsNil(r.group) {
		return merr.ErrorParamsError("invalid notice group")
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

func (r *ListNoticeGroupReq) ToListNoticeGroupReply(groups []*team.NoticeGroup) *ListNoticeGroupReply {
	return &ListNoticeGroupReply{
		PaginationReply: r.ToReply(),
		Items:           slices.Map(groups, func(group *team.NoticeGroup) do.NoticeGroup { return group }),
	}
}

type ListNoticeGroupReply = ListReply[do.NoticeGroup]
