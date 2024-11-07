package vobj

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnnotations_String(t *testing.T) {
	a := NewAnnotations(map[string]string{
		SummaryKey:     "summary",
		DescriptionKey: "description",
	})

	t.Log(a.String())
}

func TestAnnotations_Get(t *testing.T) {
	a := NewAnnotations(map[string]string{
		SummaryKey:     "summary",
		DescriptionKey: "description",
	})

	assert.Equal(t, "summary", a.Get(SummaryKey))
	assert.Equal(t, "description", a.Get(DescriptionKey))
}

func TestAnnotations_Value(t *testing.T) {
	a := NewAnnotations(map[string]string{
		SummaryKey:     "summary",
		DescriptionKey: "description",
	})
	v, err := a.Value()
	assert.Nil(t, err)
	t.Log(v)
}

func TestAnnotations_Set(t *testing.T) {
	a := NewAnnotations(map[string]string{})

	a.Set(SummaryKey, "summary")
	a.Set(DescriptionKey, "description")
	assert.Equal(t, "summary", a.Get(SummaryKey))
	assert.Equal(t, "description", a.Get(DescriptionKey))
	t.Log(a)
}
