package do

import "testing"

func TestDatasourceFilter_Scan_legacyArray(t *testing.T) {
	var filter DatasourceFilter
	if err := filter.Scan([]byte(`[1,2,3]`)); err != nil {
		t.Fatalf("scan legacy array: %v", err)
	}
	if len(filter.DatasourceUIDs) != 3 || filter.DatasourceUIDs[0] != 1 {
		t.Fatalf("unexpected uids: %v", filter.DatasourceUIDs)
	}
}

func TestDatasourceFilter_Scan_object(t *testing.T) {
	var filter DatasourceFilter
	if err := filter.Scan([]byte(`{"datasource_uids":[1],"exclude_datasource_uids":[2],"datasource_labels":{"env":"prod"}}`)); err != nil {
		t.Fatalf("scan object: %v", err)
	}
	if len(filter.DatasourceUIDs) != 1 || filter.ExcludeDatasourceUIDs[0] != 2 {
		t.Fatalf("unexpected filter: %+v", filter)
	}
	if filter.DatasourceLabels["env"] != "prod" {
		t.Fatalf("unexpected labels: %v", filter.DatasourceLabels)
	}
}
