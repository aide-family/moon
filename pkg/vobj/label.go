package vobj

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
)

var _ sql.Scanner = (*Labels)(nil)
var _ driver.Valuer = (*Labels)(nil)
var ErrUnsupportedType = errors.New("unsupported type")

type Labels map[string]string

func (l Labels) Value() (driver.Value, error) {
	return json.Marshal(l)
}

func (l *Labels) Scan(src any) error {
	switch src.(type) {
	case []byte:
		return json.Unmarshal(src.([]byte), l)
	case string:
		return json.Unmarshal([]byte(src.(string)), l)
	default:
		return ErrUnsupportedType
	}
}
