package message

import (
	"encoding/json"

	"github.com/aide-family/magicbox/strutil"
	"go.yaml.in/yaml/v2"
)

var templateFuncMap = func() map[string]any {
	funcs := strutil.TextTemplateFuncMap()
	funcs["json"] = json.Marshal
	funcs["yaml"] = yaml.Marshal
	return funcs
}()
