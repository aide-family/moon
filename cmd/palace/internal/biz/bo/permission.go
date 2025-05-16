package bo

type ReplaceUserRoleReq struct {
	UserID uint32
	Roles  []uint32
}

type ReplaceMemberRoleReq struct {
	MemberID uint32
	Roles    []uint32
}
