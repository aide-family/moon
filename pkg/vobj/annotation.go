package vobj

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
)

var _ sql.Scanner = (*Annotations)(nil)
var _ driver.Valuer = (*Annotations)(nil)

type Annotations map[string]string

func (l Annotations) Value() (driver.Value, error) {
	return json.Marshal(l)
}

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
