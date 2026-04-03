package machine

import "strings"

// resolveMachineEndpoints returns target API endpoints: positional args, then config endpoints list,
// then a single fallback (e.g. --endpoint) or cfg.Endpoint when still empty.
func resolveMachineEndpoints(args []string, cfg *clientConfig, singleFallback string) []string {
	endpoints := append([]string(nil), args...)
	if len(endpoints) == 0 {
		endpoints = append(endpoints, cfg.Endpoints...)
	}
	if len(endpoints) == 0 {
		ep := strings.TrimSpace(singleFallback)
		if ep == "" {
			ep = strings.TrimSpace(cfg.Endpoint)
		}
		if ep != "" {
			endpoints = []string{ep}
		}
	}
	return endpoints
}
