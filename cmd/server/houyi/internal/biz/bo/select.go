package bo

type (
	// SelectOptionBo 选项构造器明细
	SelectOptionBo struct {
		Value    any    `json:"value"`
		Label    string `json:"label"`
		Disabled bool   `json:"disabled"`
	}
)
