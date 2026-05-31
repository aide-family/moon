package do

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// DatasourceFilter is the JSON shape stored in strategy_metrics.datasource_filter.
type DatasourceFilter struct {
	DatasourceUIDs         []int64           `json:"datasource_uids,omitempty"`
	ExcludeDatasourceUIDs  []int64           `json:"exclude_datasource_uids,omitempty"`
	DatasourceLabels       map[string]string `json:"datasource_labels,omitempty"`
	ExcludeDatasourceLabels map[string]string `json:"exclude_datasource_labels,omitempty"`
}

// Value implements driver.Valuer for JSON storage.
func (c *DatasourceFilter) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

// Scan implements sql.Scanner for JSON storage.
func (c *DatasourceFilter) Scan(value any) error {
	if value == nil {
		*c = DatasourceFilter{}
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("datasource_filter: expected []byte or string")
	}
	if len(b) == 0 {
		*c = DatasourceFilter{}
		return nil
	}
	// Backward compatibility: legacy column stored a JSON array of datasource UIDs.
	if b[0] == '[' {
		var uids []int64
		if err := json.Unmarshal(b, &uids); err != nil {
			return err
		}
		*c = DatasourceFilter{DatasourceUIDs: uids}
		return nil
	}
	return json.Unmarshal(b, c)
}
