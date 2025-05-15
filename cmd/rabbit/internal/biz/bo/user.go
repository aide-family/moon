package bo

type NoticeUser interface {
	GetName() string
	GetEmail() string
	GetPhone() string
}

type GetNoticeUserConfigParams struct {
	TeamID            string
	Name              *string
	DefaultNoticeUser NoticeUser
}

type SetNoticeUserConfigParams struct {
	TeamID  string
	Configs []NoticeUser
}
