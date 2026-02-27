package message

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/aide-family/magicbox/strutil"
	"go.yaml.in/yaml/v2"
)

var templateFuncMap = map[string]any{
	"now":          time.Now,
	"hasPrefix":    strings.HasPrefix,
	"hasSuffix":    strings.HasSuffix,
	"contains":     strings.Contains,
	"trimSpace":    strings.TrimSpace,
	"trimPrefix":   strings.TrimPrefix,
	"trimSuffix":   strings.TrimSuffix,
	"toUpper":      strings.ToUpper,
	"toLower":      strings.ToLower,
	"replace":      strings.Replace,
	"split":        strings.Split,
	"mask":         strutil.MaskString,
	"maskEmail":    strutil.MaskEmail,
	"maskPhone":    strutil.MaskPhone,
	"maskBankCard": strutil.MaskBankCard,
	"title":        strutil.Title,
	"json":         json.Marshal,
	"yaml":         yaml.Marshal,
}
