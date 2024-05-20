package bo

type (
	SelectExtend struct {
		Icon, Color, Remark, Image string
	}
	SelectOptionBo struct {
		Value    uint32            `json:"value"`
		Label    string            `json:"label"`
		Disabled bool              `json:"disabled"`
		Children []*SelectOptionBo `json:"children"`
		Extend   *SelectExtend     `json:"extend"`
	}
)
