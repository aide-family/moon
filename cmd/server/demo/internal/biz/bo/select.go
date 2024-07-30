package bo

type (
	// SelectOptionBo select option
	SelectOptionBo struct {
		Value    any    `json:"value"`
		Label    string `json:"label"`
		Disabled bool   `json:"disabled"`
	}
)
