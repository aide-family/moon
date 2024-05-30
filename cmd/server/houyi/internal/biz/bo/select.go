package bo

type (
	SelectOptionBo struct {
		Value    any    `json:"value"`
		Label    string `json:"label"`
		Disabled bool   `json:"disabled"`
	}
)
