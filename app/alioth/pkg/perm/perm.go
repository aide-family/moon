package perm

const (
	MemberRead        = "member:read"
	MemberWrite       = "member:write"
	RoleRead          = "role:read"
	RoleWrite         = "role:write"
	NamespaceRead     = "namespace:read"
	NamespaceWrite    = "namespace:write"
	NamespaceDelete   = "namespace:delete"
	NamespaceTransfer = "namespace:transfer"

	DatasourceRead  = "datasource:read"
	DatasourceWrite = "datasource:write"
	AlertRuleRead   = "alert_rule:read"
	AlertRuleWrite  = "alert_rule:write"

	AlertRead      = "alert:read"
	AlertWrite     = "alert:write"
	AlertIntervene = "alert:intervene"

	NotificationRead  = "notification:read"
	NotificationWrite = "notification:write"

	SSHExec = "ssh:exec"
)

// 系统权限属于平台，不写入租户的 permission 表，也不参与租户角色授权。
const (
	SystemUserRead   = "system:user:read"
	SystemUserWrite  = "system:user:write"
	SystemRoleWrite  = "system:role:write"
	SystemRoleRead   = "system:role:read"
	SystemInviteRead = "system:invite:read"
)

const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
	RoleViewer = "viewer"
)

// 平台角色只有固定三档。第一个注册用户是超级管理员，其余由超级管理员授权。
const (
	SystemRoleSuperAdmin  = "SUPER_ADMIN"
	SystemRoleSystemAdmin = "SYSTEM_ADMIN"
	SystemRoleUser        = "USER"
)

// resources 只包含租户资源。user 是平台资源，不在租户权限目录里。
var resources = []string{
	"member", "role", "namespace", "datasource", "alert_rule",
	"alert", "notification", "agent", "ssh",
}

type Item struct {
	Name     string
	Resource string
	Action   string
}

// Catalog 返回租户权限目录。
func Catalog() []Item {
	items := make([]Item, 0, len(resources)*2+4)
	for _, resource := range resources {
		items = append(items,
			Item{Name: resource + ":read", Resource: resource, Action: "read"},
			Item{Name: resource + ":write", Resource: resource, Action: "write"},
		)
	}
	items = append(items,
		Item{Name: AlertIntervene, Resource: "alert", Action: "intervene"},
		Item{Name: SSHExec, Resource: "ssh", Action: "exec"},
		Item{Name: NamespaceDelete, Resource: "namespace", Action: "delete"},
		Item{Name: NamespaceTransfer, Resource: "namespace", Action: "transfer"},
	)
	return items
}

// SystemCatalog 返回平台权限目录，和租户权限目录互不混用。
func SystemCatalog() []Item {
	return []Item{
		{Name: SystemUserRead, Resource: "system", Action: "user:read"},
		{Name: SystemUserWrite, Resource: "system", Action: "user:write"},
		{Name: SystemRoleRead, Resource: "system", Action: "role:read"},
		{Name: SystemRoleWrite, Resource: "system", Action: "role:write"},
		{Name: SystemInviteRead, Resource: "system", Action: "invite:read"},
	}
}

func Names() []string {
	catalog := Catalog()
	names := make([]string, len(catalog))
	for i, item := range catalog {
		names[i] = item.Name
	}
	return names
}

func SystemNames() []string {
	catalog := SystemCatalog()
	names := make([]string, len(catalog))
	for i, item := range catalog {
		names[i] = item.Name
	}
	return names
}

func RolePermissions(role string) []string {
	switch role {
	case RoleOwner:
		return Names()
	case RoleAdmin:
		return without(Names(), NamespaceDelete, NamespaceTransfer)
	case RoleMember:
		return memberPermissions()
	case RoleViewer:
		return readPermissions()
	default:
		return nil
	}
}

// SystemPermissions 返回平台角色拥有的系统权限。
// 系统权限不叠加租户权限；超级管理员要使用租户能力仍需成为该租户成员。
func SystemPermissions(role string) []string {
	switch role {
	case SystemRoleSuperAdmin:
		return SystemNames()
	case SystemRoleSystemAdmin:
		return []string{SystemUserRead, SystemUserWrite, SystemRoleRead, SystemInviteRead}
	default:
		return nil
	}
}

// TenantRoles 返回每个租户预置的四个角色。
func TenantRoles() []string {
	return []string{RoleOwner, RoleAdmin, RoleMember, RoleViewer}
}

// SystemRoles 返回平台固定角色。
func SystemRoles() []string {
	return []string{SystemRoleSuperAdmin, SystemRoleSystemAdmin, SystemRoleUser}
}

// IsTenantPermission 只认租户权限目录。用于阻止把 user:* 或 system:* 授给租户角色。
func IsTenantPermission(name string) bool {
	for _, item := range Catalog() {
		if item.Name == name {
			return true
		}
	}
	return false
}

func IsSystemRole(role string) bool {
	switch role {
	case SystemRoleSuperAdmin, SystemRoleSystemAdmin, SystemRoleUser:
		return true
	default:
		return false
	}
}

func IsSystemAdmin(role string) bool {
	return role == SystemRoleSuperAdmin || role == SystemRoleSystemAdmin
}

func readPermissions() []string {
	names := make([]string, 0, len(resources))
	for _, resource := range resources {
		names = append(names, resource+":read")
	}
	return names
}

func memberPermissions() []string {
	names := readPermissions()
	return append(names, AlertIntervene, SSHExec)
}

func without(names []string, drop ...string) []string {
	blocked := map[string]struct{}{}
	for _, name := range drop {
		blocked[name] = struct{}{}
	}
	out := make([]string, 0, len(names))
	for _, name := range names {
		if _, ok := blocked[name]; ok {
			continue
		}
		out = append(out, name)
	}
	return out
}
