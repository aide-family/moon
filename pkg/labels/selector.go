package labels

import "strings"

// Selector is a label selector.
type Selector interface {
	// Matches returns true if the specified Labels match this Selector.
	Matches(labels Labels) bool

	// AddRequirement adds new Requirement to this Selector.
	AddRequirement(r ...Requirement) Selector
}

var _ Selector = &emptySelector{}

type emptySelector struct{}

func (e *emptySelector) Matches(labels Labels) bool {
	return false
}

func (e *emptySelector) AddRequirement(r ...Requirement) Selector {
	return nil
}

// Requirements is a single element selector.
type Requirements []Requirement

// NewSelector creates an empty Selector.
func NewSelector() Selector {
	return Requirements(nil)
}

// Matches returns true if all requirements matches the given Labels.
func (x Requirements) Matches(labels Labels) bool {
	for i := range x {
		if !x[i].Matches(labels) {
			return false
		}
	}
	return true
}

// AddRequirement adds new Requirement to this Selector.
func (x Requirements) AddRequirement(r ...Requirement) Selector {
	req := make(Requirements, 0, len(x)+len(r))
	req = append(req, x...)
	req = append(req, r...)
	return req
}

// String converts this Selector to a string.
func (x Requirements) String() string {
	strs := make([]string, 0, len(x))
	for _, requirement := range x {
		strs = append(strs, requirement.String())
	}
	return strings.Join(strs, ", ")
}
