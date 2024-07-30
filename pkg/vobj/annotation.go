package vobj

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

var _ sql.Scanner = (*Annotations)(nil)
var _ driver.Valuer = (*Annotations)(nil)

// Annotations 告警文案
type Annotations map[string]string

// Value implements the driver.Valuer interface.
func (l Annotations) Value() (driver.Value, error) {
	return json.Marshal(l)
}

// Scan implements the sql.Scanner interface.
func (l *Annotations) Scan(src any) error {
	switch src.(type) {
	case []byte:
		return json.Unmarshal(src.([]byte), l)
	case string:
		return json.Unmarshal([]byte(src.(string)), l)
	default:
		return ErrUnsupportedType
	}
}
