package perm

const (
	UserRead          = "user:read"
	UserWrite         = "user:write"
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

	AlertIntervene = "alert:intervene"
	SSHExec        = "ssh:exec"
	K8sProxy       = "k8s:proxy"
)

const (
	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
	RoleViewer = "viewer"
)

var resources = []string{
	"user", "member", "role", "namespace", "datasource", "alert_rule",
	"alert", "notification", "agent", "ssh", "sftp", "k8s",
}

type Item struct {
	Name     string
	Resource string
	Action   string
}

func Catalog() []Item {
	items := make([]Item, 0, len(resources)*2+5)
	for _, resource := range resources {
		items = append(items,
			Item{Name: resource + ":read", Resource: resource, Action: "read"},
			Item{Name: resource + ":write", Resource: resource, Action: "write"},
		)
	}
	items = append(items,
		Item{Name: AlertIntervene, Resource: "alert", Action: "intervene"},
		Item{Name: SSHExec, Resource: "ssh", Action: "exec"},
		Item{Name: K8sProxy, Resource: "k8s", Action: "proxy"},
		Item{Name: NamespaceDelete, Resource: "namespace", Action: "delete"},
		Item{Name: NamespaceTransfer, Resource: "namespace", Action: "transfer"},
	)
	return items
}

func Names() []string {
	catalog := Catalog()
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

func SystemRoles() []string {
	return []string{RoleOwner, RoleAdmin, RoleMember, RoleViewer}
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
	return append(names, AlertIntervene, SSHExec, "sftp:write", K8sProxy)
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
