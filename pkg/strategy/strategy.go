package strategy

type (
	Groups struct {
		Groups []*Group `json:"groups"`
	}

	Group struct {
		Name  string  `json:"name"`
		Rules []*Rule `json:"rules"`
	}

	Rule struct {
		Alert       string            `json:"alert"`
		Expr        string            `json:"expr"`
		For         string            `json:"for"`
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
	}
)
