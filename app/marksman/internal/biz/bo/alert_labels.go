package bo

import (
	"github.com/aide-family/magicbox/safety"
	"github.com/aide-family/magicbox/strutil/cnst"
)

var (
	reservedAlertSystemLabelKeys = safety.NewSyncMap(newReservedAlertSystemLabelKeys())
	volatileAlertLabelKeys       = safety.NewSyncMap(newVolatileAlertLabelKeys())
)

func newReservedAlertSystemLabelKeys() map[string]struct{} {
	keys := []string{
		cnst.LabelAlertName,
		cnst.LabelSeverity,
		cnst.LabelAlertState,
		cnst.LabelNamespaceUID,
		cnst.LabelStrategyGroupUID,
		cnst.LabelStrategyGroupName,
		cnst.LabelStrategyUID,
		cnst.LabelStrategyName,
		cnst.LabelLevelUID,
		cnst.LabelLevelName,
		cnst.LabelDatasourceUID,
		cnst.LabelDatasourceName,
		cnst.LabelDatasourceLevelName,
		cnst.LabelAlertEventUID,
		cnst.LabelFingerprint,
	}
	return alertLabelKeySet(keys)
}

func newVolatileAlertLabelKeys() map[string]struct{} {
	keys := []string{
		cnst.LabelAlertName,
		cnst.LabelSeverity,
		cnst.LabelAlertState,
		cnst.LabelStrategyGroupName,
		cnst.LabelStrategyName,
		cnst.LabelLevelName,
		cnst.LabelDatasourceName,
		cnst.LabelDatasourceLevelName,
		cnst.LabelAlertEventUID,
		cnst.LabelFingerprint,
	}
	return alertLabelKeySet(keys)
}

func alertLabelKeySet(keys []string) map[string]struct{} {
	set := make(map[string]struct{}, len(keys))
	for _, key := range keys {
		set[key] = struct{}{}
	}
	return set
}

// AlertNamespaceUIDLabelKeys returns label keys that may carry namespace UID on alert payloads.
func AlertNamespaceUIDLabelKeys() []string {
	return []string{cnst.LabelNamespaceUID}
}

// IsReservedAlertSystemLabelKey reports whether a label key is owned by alert system fields
// and must not be supplied from strategy or prometheus series labels.
func IsReservedAlertSystemLabelKey(key string) bool {
	_, ok := reservedAlertSystemLabelKeys.Get(key)
	return ok
}

// IsVolatileAlertLabelKey reports whether a label key must be excluded from alert fingerprint calculation.
func IsVolatileAlertLabelKey(key string) bool {
	_, ok := volatileAlertLabelKeys.Get(key)
	return ok
}

// FilterLabelsForAlertFingerprint returns a copy of labels with volatile/name/meta keys removed.
func FilterLabelsForAlertFingerprint(labels map[string]string) map[string]string {
	if len(labels) == 0 {
		return map[string]string{}
	}
	filtered := make(map[string]string, len(labels))
	for key, value := range labels {
		if IsVolatileAlertLabelKey(key) {
			continue
		}
		filtered[key] = value
	}
	return filtered
}
