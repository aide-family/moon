package do

import (
	"encoding/json"

	"github.com/moon-monitor/moon/pkg/plugin/cache"
)

var _ cache.Object = (*NoticeUserConfig)(nil)

type NoticeUserConfig struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// GetEmail implements bo.NoticeUser.
func (n *NoticeUserConfig) GetEmail() string {
	if n == nil {
		return ""
	}
	return n.Email
}

// GetName implements bo.NoticeUser.
func (n *NoticeUserConfig) GetName() string {
	if n == nil {
		return ""
	}
	return n.Name
}

// GetPhone implements bo.NoticeUser.
func (n *NoticeUserConfig) GetPhone() string {
	if n == nil {
		return ""
	}
	return n.Phone
}

// MarshalBinary implements cache.Object.
func (n *NoticeUserConfig) MarshalBinary() (data []byte, err error) {
	return json.Marshal(n)
}

// UnmarshalBinary implements cache.Object.
func (n *NoticeUserConfig) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, n)
}

func (n *NoticeUserConfig) UniqueKey() string {
	return n.Name
}
