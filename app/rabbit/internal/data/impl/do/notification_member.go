package do

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// NotificationMember stores a namespace member with per-channel delivery preferences.
type NotificationMember struct {
	MemberUID int64 `json:"member_uid"`
	IsEmail   bool  `json:"is_email"`
	IsSMS     bool  `json:"is_sms"`
	IsPhone   bool  `json:"is_phone"`
}

func (m *NotificationMember) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if v, ok := raw["member_uid"]; ok {
		if err := json.Unmarshal(v, &m.MemberUID); err != nil {
			return err
		}
	} else if v, ok := raw["memberUid"]; ok {
		if err := json.Unmarshal(v, &m.MemberUID); err != nil {
			return err
		}
	}
	if v, ok := raw["is_email"]; ok {
		if err := json.Unmarshal(v, &m.IsEmail); err != nil {
			return err
		}
	} else if v, ok := raw["isEmail"]; ok {
		if err := json.Unmarshal(v, &m.IsEmail); err != nil {
			return err
		}
	}
	if v, ok := raw["is_sms"]; ok {
		if err := json.Unmarshal(v, &m.IsSMS); err != nil {
			return err
		}
	} else if v, ok := raw["isSms"]; ok {
		if err := json.Unmarshal(v, &m.IsSMS); err != nil {
			return err
		}
	}
	if v, ok := raw["is_phone"]; ok {
		if err := json.Unmarshal(v, &m.IsPhone); err != nil {
			return err
		}
	} else if v, ok := raw["isPhone"]; ok {
		if err := json.Unmarshal(v, &m.IsPhone); err != nil {
			return err
		}
	}
	return nil
}

type NotificationMembers []NotificationMember

func (c NotificationMembers) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

func (c *NotificationMembers) Scan(value any) error {
	if value == nil {
		*c = nil
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return errors.New("notification_members: expected []byte or string")
	}
	if len(b) == 0 {
		*c = nil
		return nil
	}
	return json.Unmarshal(b, c)
}
