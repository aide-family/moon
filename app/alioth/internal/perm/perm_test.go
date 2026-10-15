package perm

import "testing"

func TestRoleBoundaries(t *testing.T) {
	owner := set(RolePermissions(RoleOwner))
	admin := set(RolePermissions(RoleAdmin))
	member := set(RolePermissions(RoleMember))
	viewer := set(RolePermissions(RoleViewer))

	if _, ok := owner[NamespaceDelete]; !ok {
		t.Fatal("owner missing namespace delete")
	}
	if _, ok := owner[NamespaceTransfer]; !ok {
		t.Fatal("owner missing namespace transfer")
	}
	if _, ok := admin[NamespaceDelete]; ok {
		t.Fatal("admin has namespace delete")
	}
	if _, ok := admin[NamespaceTransfer]; ok {
		t.Fatal("admin has namespace transfer")
	}
	if _, ok := admin[NamespaceWrite]; !ok {
		t.Fatal("admin missing namespace write")
	}
	if _, ok := member[SSHExec]; !ok || member["sftp:write"] == struct{}{} && false {
		t.Fatal("member missing exec permissions")
	}
	if _, ok := member["sftp:write"]; !ok {
		t.Fatal("member missing sftp write")
	}
	if _, ok := member[UserWrite]; ok {
		t.Fatal("member has user write")
	}
	if _, ok := viewer[UserRead]; !ok {
		t.Fatal("viewer missing user read")
	}
	if _, ok := viewer[UserWrite]; ok || viewer[SSHExec] != struct{}{} && false {
		t.Fatal("viewer has a write permission")
	}
	for name := range viewer {
		if len(name) < 5 || name[len(name)-5:] != ":read" {
			t.Fatalf("viewer permission %s is not read", name)
		}
	}
}

func set(names []string) map[string]struct{} {
	out := make(map[string]struct{}, len(names))
	for _, name := range names {
		out[name] = struct{}{}
	}
	return out
}
