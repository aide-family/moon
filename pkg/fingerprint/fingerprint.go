// Package fingerprint 用规则和标签计算稳定的告警指纹。
package fingerprint

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
)

// Of 对规则标识和按键排序后的标签做摘要。同一规则、同一组标签得到同一个值。
func Of(ruleUID string, labels map[string]string) string {
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	sum := sha256.New()
	_, _ = sum.Write([]byte(ruleUID))
	_, _ = sum.Write([]byte{0})
	for _, key := range keys {
		_, _ = sum.Write([]byte(key))
		_, _ = sum.Write([]byte{0})
		_, _ = sum.Write([]byte(labels[key]))
		_, _ = sum.Write([]byte{0})
	}
	return hex.EncodeToString(sum.Sum(nil))
}
