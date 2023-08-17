package helper

import "encoding/json"

func BuildLabels(labelsStr string) map[string]string {
	result := make(map[string]string)
	if labelsStr != "" {
		_ = json.Unmarshal([]byte(labelsStr), &result)
	}
	return result
}

func BuildAnnotations(annotationsStr string) map[string]string {
	result := make(map[string]string)
	if annotationsStr != "" {
		_ = json.Unmarshal([]byte(annotationsStr), &result)
	}
	return result
}
