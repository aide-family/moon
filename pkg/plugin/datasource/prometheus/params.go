package prometheus

type (
	PromMetricSeriesResponse struct {
		Status    string              `json:"status"`
		Data      []map[string]string `json:"data"`
		Err       string              `json:"error"`
		ErrorType string              `json:"errorType"`
	}

	PromMetricInfo struct {
		Type string `json:"type"`
		Help string `json:"help"`
		Unit string `json:"unit"`
	}

	PromMetadataResponse struct {
		Status string                      `json:"status"`
		Data   map[string][]PromMetricInfo `json:"data"`
	}
)

// IsSuccessResponse is response success
func (p *PromMetricSeriesResponse) IsSuccessResponse() bool {
	return p.Status == "success"
}

// Error is response error
func (p *PromMetricSeriesResponse) Error() string {
	if !p.IsSuccessResponse() {
		return p.Err
	}
	return ""
}
