package victoria

type Names struct {
	Status string   `json:"status"`
	Data   []string `json:"data"`
}

type Series struct {
	Status string              `json:"status"`
	Data   []map[string]string `json:"data"`
}
