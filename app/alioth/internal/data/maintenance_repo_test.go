package data

import (
	"testing"

	entmigrate "github.com/aide-family/moon/app/alioth/internal/data/ent/migrate"
)

// TestMigratedTablesCoverEntSchema 保证迁移包覆盖的表清单与 ent schema 对齐。
// 新增 ent 表时必须二选一：进迁移包，或明确列为运行数据；否则这里会失败，
// 避免升级到新版本后有一张配置表既没导出也没人发现。
func TestMigratedTablesCoverEntSchema(t *testing.T) {
	known := make(map[string]string, len(archivedTables)+len(runtimeTables))
	for _, name := range archivedTables {
		if _, ok := known[name]; ok {
			t.Fatalf("table %q is declared twice", name)
		}
		known[name] = "archived"
	}
	for _, name := range runtimeTables {
		if _, ok := known[name]; ok {
			t.Fatalf("table %q is declared twice", name)
		}
		known[name] = "runtime"
	}
	seen := make(map[string]struct{}, len(entmigrate.Tables))
	for _, table := range entmigrate.Tables {
		if _, ok := known[table.Name]; !ok {
			t.Fatalf("ent table %q is missing from the migration table list", table.Name)
		}
		seen[table.Name] = struct{}{}
	}
	for name := range known {
		if _, ok := seen[name]; !ok {
			t.Fatalf("migration table %q does not exist in the ent schema", name)
		}
	}
}
