package label

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/util/cnst"
	"github.com/aide-family/moon/pkg/util/kv"
	"github.com/aide-family/moon/pkg/util/template"
)

var _ json.Marshaler = (*Annotation)(nil)
var _ json.Unmarshaler = (*Annotation)(nil)

func NewAnnotation(summary, description string) *Annotation {
	return &Annotation{
		kvMap: kv.NewStringMap(map[string]string{
			cnst.AnnotationKeySummary:     summary,
			cnst.AnnotationKeyDescription: description,
		}),
	}
}

func NewAnnotationFromMap(annotations map[string]string) *Annotation {
	return &Annotation{
		kvMap: annotations,
	}
}

type Annotation struct {
	kvMap kv.StringMap
}

func (a *Annotation) UnmarshalJSON(bytes []byte) error {
	return json.Unmarshal(bytes, &a.kvMap)
}

func (a *Annotation) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.kvMap)
}

func (a *Annotation) String() string {
	bs, _ := a.MarshalBinary()
	return string(bs)
}

func (a *Annotation) MarshalBinary() (data []byte, err error) {
	return a.kvMap.MarshalBinary()
}

func (a *Annotation) UnmarshalBinary(data []byte) error {
	return a.kvMap.UnmarshalBinary(data)
}

func (a *Annotation) Copy() *Annotation {
	return &Annotation{
		kvMap: a.kvMap.Copy(),
	}
}

func (a *Annotation) GetSummary() string {
	summary, ok := a.kvMap.Get(cnst.AnnotationKeySummary)
	if !ok {
		return ""
	}
	return summary
}

func (a *Annotation) SetSummary(summary string) {
	a.kvMap.Set(cnst.AnnotationKeySummary, summary)
}

func (a *Annotation) GetDescription() string {
	description, ok := a.kvMap.Get(cnst.AnnotationKeyDescription)
	if !ok {
		return ""
	}
	return description
}

func (a *Annotation) SetDescription(description string) {
	a.kvMap.Set(cnst.AnnotationKeyDescription, description)
}

func (a *Annotation) Format(data interface{}) *Annotation {
	for k, v := range a.kvMap.ToMap() {
		a.kvMap.Set(k, template.TextFormatterX(v, data))
	}
	return a
}
