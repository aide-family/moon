package bo

import (
	"testing"

	"github.com/bwmarrin/snowflake"
)

func TestDatasourceFilterBo_MatchesDatasource(t *testing.T) {
	filter := &DatasourceFilterBo{
		DatasourceUIDs:        []int64{1, 2},
		ExcludeDatasourceUIDs: []int64{3},
		DatasourceLabels:      map[string]string{"env": "prod"},
		ExcludeDatasourceLabels: map[string]string{"tier": "test"},
	}
	if filter.MatchesDatasource(3, map[string]string{"env": "prod"}) {
		t.Fatal("expected excluded uid to not match")
	}
	if filter.MatchesDatasource(1, map[string]string{"tier": "test"}) {
		t.Fatal("expected excluded labels to not match")
	}
	if !filter.MatchesDatasource(1, map[string]string{"env": "dev"}) {
		t.Fatal("expected included uid to match")
	}
	if !filter.MatchesDatasource(99, map[string]string{"env": "prod"}) {
		t.Fatal("expected included labels to match")
	}
}

func TestDatasourceFilterBo_MatchesDatasource_emptyFilter(t *testing.T) {
	var filter *DatasourceFilterBo
	if !filter.MatchesDatasource(1, nil) {
		t.Fatal("nil filter should match all")
	}
	if !(&DatasourceFilterBo{}).MatchesDatasource(1, nil) {
		t.Fatal("empty filter should match all")
	}
}

func TestOrderDatasourceItemsByUIDs(t *testing.T) {
	items := []*DatasourceItemBo{
		{UID: snowflake.ParseInt64(2), Name: "b"},
		{UID: snowflake.ParseInt64(1), Name: "a"},
	}
	ordered := OrderDatasourceItemsByUIDs([]int64{1, 2, 3}, items)
	if len(ordered) != 2 {
		t.Fatalf("expected 2 items, got %d", len(ordered))
	}
	if ordered[0].Name != "a" || ordered[1].Name != "b" {
		t.Fatalf("unexpected order: %+v", ordered)
	}
}

func TestDatasourceFilterBo_IncludeExcludeDatasourceUIDs(t *testing.T) {
	filter := &DatasourceFilterBo{
		DatasourceUIDs:        []int64{1, 2},
		ExcludeDatasourceUIDs: []int64{3},
	}
	include := filter.IncludeFilterDatasourceUIDs()
	if len(include) != 2 || include[0] != 1 || include[1] != 2 {
		t.Fatalf("unexpected include uids: %v", include)
	}
	exclude := filter.ExcludeFilterDatasourceUIDs()
	if len(exclude) != 1 || exclude[0] != 3 {
		t.Fatalf("unexpected exclude uids: %v", exclude)
	}
	ref := filter.FilterReferencedDatasourceUIDs()
	if len(ref) != 3 || ref[0] != 1 || ref[1] != 2 || ref[2] != 3 {
		t.Fatalf("unexpected referenced uids: %v", ref)
	}
	if (&DatasourceFilterBo{}).FilterReferencedDatasourceUIDs() != nil {
		t.Fatal("expected nil for empty filter")
	}
}
