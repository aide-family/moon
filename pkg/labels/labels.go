package labels

import "strings"

type Labels interface {
	Get(key string) string
	Has(key string) bool
}

type Set map[string]string

func (s Set) Get(key string) string {
	return s[key]
}
func (s Set) Has(key string) bool {
	_, ok := s[key]
	return ok
}

// String returns all labels listed as a human-readable string.
// The returned operator defaults to Equals,
// Used ',' join between multiple key-value pairs.
func (s Set) String() string {
	strs := make([]string, 0, len(s))
	for k, v := range s {
		strs = append(strs, k+"="+v)
	}
	return strings.Join(strs, ",")
}

// AsSelector converts labels into a selectors. It does not
// perform any validation, which means the server will reject
// the request if the Set contains invalid values.
func (s Set) AsSelector() Selector {
	x := make(Requirements, 0, len(s))
	for k, v := range s {
		r := Requirement{
			key:      k,
			operator: Equals,
			values:   []string{v},
		}
		x = append(x, r)
	}
	return x
}

// AsValidatedSelector converts labels into a selectors.
// The Set is validated client-side, which allows to catch errors early.
func (s Set) AsValidatedSelector() (Selector, error) {
	x := NewSelector()
	for k, v := range s {
		r, err := NewRequirement(k, Equals, []string{v})
		if err != nil {
			return nil, err
		}
		x = x.AddRequirement(*r)
	}
	return x, nil
}
