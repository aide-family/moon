package bo

type ReplaceUserRoleReq struct {
	UserID uint32   `json:"userID"`
	Roles  []uint32 `json:"roles"`
}

type ReplaceMemberRoleReq struct {
	MemberID uint32   `json:"memberID"`
	Roles    []uint32 `json:"roles"`
}
