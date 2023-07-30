package stringer

import (
	"encoding/json"
	"fmt"
)

type model struct {
	Model any
}

func (l *model) String() string {
	b, _ := json.Marshal(*l)
	return string(b)
}

func New(m any) fmt.Stringer {
	return &model{Model: m}
}

var _ fmt.Stringer = (*model)(nil)
