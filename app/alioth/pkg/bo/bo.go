package bo

// Account 是全局用户。PasswordHash 只在登录校验时使用，离开用例前要清空。
type Account struct {
	UID          string
	Name         string
	Nickname     string
	Email        string
	Avatar       string
	Remark       string
	Status       string
	PasswordHash string
}

// Namespace 是租户。OwnerUID 指向当前负责人。
type Namespace struct {
	UID      string
	Name     string
	OwnerUID string
	Logo     string
	Remark   string
	Status   string
}

// Member 是用户在某个租户中的成员身份。
type Member struct {
	UID          string
	NamespaceUID string
	UserUID      string
	Email        string
	Name         string
	Status       string
	RoleUIDs     []string
	RoleNames    []string
}

// Role 是租户内角色。IsSystem 为真时不能改名、删除或改授权。
type Role struct {
	UID          string
	NamespaceUID string
	Name         string
	Status       string
	IsSystem     bool
	Permissions  []string
}

// Permission 是全局权限定义。
type Permission struct {
	UID      string
	Name     string
	Resource string
	Action   string
}

// SelectItem 是下拉列表中的一项。
type SelectItem struct {
	UID  string
	Name string
}

// Session 是登录成功后交给调用方的令牌对。User 不含密码哈希。
type Session struct {
	AccessToken  string
	RefreshToken string
	User         Account
}
