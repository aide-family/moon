package random

import (
	"strings"

	"github.com/google/uuid"
)

// UUID 生成UUID
func UUID(isRemoveInvalidChar ...bool) string {
	uid := uuid.New().String()
	if len(isRemoveInvalidChar) > 0 && isRemoveInvalidChar[0] {
		return strings.ReplaceAll(uid, "-", "")
	}
	return uid
}

// UUIDToUpperCase 生成UUID并转换为大写
func UUIDToUpperCase(isRemoveInvalidChar ...bool) string {
	return strings.ToUpper(UUID(isRemoveInvalidChar...))
}
