package bo

type SelectItem interface {
	GetLabel() string
	GetValue() uint32
	GetDisabled() bool
	GetExtra() SelectItemExtra
}

type SelectItemExtra interface {
	GetRemark() string
	GetIcon() string
	GetColor() string
}

var _ SelectItem = (*selectItem)(nil)
var _ SelectItemExtra = (*selectItemExtra)(nil)

type selectItem struct {
	Value    uint32          `json:"value"`
	Label    string          `json:"label"`
	Disabled bool            `json:"disabled"`
	Extra    SelectItemExtra `json:"extra"`
}

// GetDisabled implements SelectItem.
func (s *selectItem) GetDisabled() bool {
	return s.Disabled
}

// GetExtra implements SelectItem.
func (s *selectItem) GetExtra() SelectItemExtra {
	return s.Extra
}

// GetLabel implements SelectItem.
func (s *selectItem) GetLabel() string {
	return s.Label
}

// GetValue implements SelectItem.
func (s *selectItem) GetValue() uint32 {
	return s.Value
}

type selectItemExtra struct {
	Remark string `json:"remark"`
	Icon   string `json:"icon"`
	Color  string `json:"color"`
}

// GetColor implements SelectItemExtra.
func (s *selectItemExtra) GetColor() string {
	return s.Color
}

// GetIcon implements SelectItemExtra.
func (s *selectItemExtra) GetIcon() string {
	return s.Icon
}

// GetRemark implements SelectItemExtra.
func (s *selectItemExtra) GetRemark() string {
	return s.Remark
}
