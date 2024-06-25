package json

import (
	"encoding/json"
	"fmt"
)

type MetaFactory interface {
	Interpret(data []byte) (*string, error)
}

var DefaultMetaFactory = SimpleMetaFactory{}

type SimpleMetaFactory struct {
}

// Interpret will return the APIVersion and Kind of the JSON wire-format
// encoding of an object, or an error.
func (SimpleMetaFactory) Interpret(data []byte) (*string, error) {
	findKind := struct {
		// +optional
		Kind string `json:"kind,omitempty"`
	}{}
	if err := json.Unmarshal(data, &findKind); err != nil {
		return nil, fmt.Errorf("couldn't get kind; json parse error: %v", err)
	}
	return (*string)(&findKind.Kind), nil
}
