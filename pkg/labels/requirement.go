package labels

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Requirement struct {
	key      string
	operator Operator
	values   []string
}

type Operator string

const (
	Exists         Operator = "exists"
	NotExist       Operator = "!"
	Equals         Operator = "="
	NotEquals      Operator = "!="
	In             Operator = "in"
	NotIn          Operator = "notin"
	GreaterThan    Operator = ">"
	LessThan       Operator = "<"
	GreaterOrEqual Operator = ">="
	LessOrEqual    Operator = "<="
)

// NewRequirement creates a new Requirement based on the provided key, operator and values.
// If any of these rules is violated, an error is returned when constructed.
// 1. The Operator must be one of the predefined constants.
// such as: Exists, NotExist, Equals, NotEquals, In, NotIn, GreaterThan, LessThan, GreaterOrEqual, LessOrEqual
// 2. For operators GreaterThan, LessThan, GreaterOrEqual, LessOrEqual, the values must be a number.
// 3. For operators In, NotIn, the values must not be empty.
// 4. For operators Exists, NotExist,  the values must not be empty.
// 5. For operators Equals, NotEquals, the values must have exactly one entry.
func NewRequirement(key string, op Operator, vals []string) (*Requirement, error) {
	if key == "" {
		return nil, fmt.Errorf("key must not be empty")
	}
	switch op {
	case Exists, NotExist:
		if len(vals) != 0 {
			return nil, fmt.Errorf("values must be empty for operator %q", op)
		}
	case Equals, NotEquals:
		if len(vals) != 1 {
			return nil, fmt.Errorf("values must have exactly one entry for operator %q", op)
		}
	case In, NotIn:
		if len(vals) == 0 {
			return nil, fmt.Errorf("values must not be empty for operator %q", op)
		}
	case GreaterThan, LessThan, GreaterOrEqual, LessOrEqual:
		if len(vals) != 1 {
			return nil, fmt.Errorf("values must have exactly one entry for operator %q", op)
		}
		if _, err := strconv.ParseInt(vals[0], 10, 64); err != nil {
			return nil, fmt.Errorf("invalid value for operator %q: %v", op, err)
		}
	default:
		return nil, fmt.Errorf("unknown operator %q", op)
	}
	return &Requirement{key: key, operator: op, values: vals}, nil
}

func (r *Requirement) hasValue(value string) bool {
	for _, s := range r.values {
		if s == value {
			return true
		}
	}
	return false
}

func (r *Requirement) Matches(labels Labels) bool {
	switch r.operator {
	case Equals, In:
		if !labels.Has(r.key) {
			return false
		}
		return r.hasValue(labels.Get(r.key))
	case NotEquals, NotIn:
		if !labels.Has(r.key) {
			return true
		}
		return !r.hasValue(labels.Get(r.key))
	case Exists:
		return labels.Has(r.key)
	case NotExist:
		return !labels.Has(r.key)
	case GreaterThan, LessThan, GreaterOrEqual, LessOrEqual:
		if !labels.Has(r.key) {
			return false
		}
		matchValue, err := strconv.ParseInt(labels.Get(r.key), 10, 64)
		if err != nil {
			return false
		}
		if len(r.values) != 1 {
			return false
		}
		requireValue, err := strconv.ParseInt(r.values[0], 10, 64)
		if err != nil {
			return false
		}
		return (r.operator == GreaterThan && matchValue > requireValue) || (r.operator == LessThan && matchValue < requireValue) ||
			(r.operator == GreaterOrEqual && matchValue >= requireValue) || (r.operator == LessOrEqual && matchValue <= requireValue)
	default:
		return false
	}
}

func (r *Requirement) String() string {
	var sb strings.Builder
	length := 0
	// length of r.key
	length += len(r.key)
	// length of r.operator
	length += len(r.operator)
	// the most bad case, need 2 whitespace on both sides of operator
	length += 2
	for i := range r.values {
		// length of per r.values
		length += len(r.values[i])
	}
	// with the case need ',' between values
	length += len(r.values) - 1
	// the most bad case, need '()' on both sides of r.values
	length += 2

	sb.Grow(length)
	if r.operator == NotExist {
		sb.WriteString(string(r.operator))
	}
	sb.WriteString(r.key)
	switch r.operator {
	case Exists, NotExist:
		return sb.String()
	default:
		sb.WriteString(" " + string(r.operator) + " ")
	}
	switch r.operator {
	case In, NotIn:
		sb.WriteString("(")
	}
	if len(r.values) == 1 {
		sb.WriteString(r.values[0])
	} else {
		sb.WriteString(strings.Join(r.values, ","))
	}
	switch r.operator {
	case In, NotIn:
		sb.WriteString(")")
	}
	return sb.String()
}

func Parse(in string) Selector {
	return parse(in)
}

// parse parses a string into a Requirements.
// this implementation is a demo.
func parse(in string) Requirements {
	var reqs Requirements

	// define the regex patterns
	operatorPatterns := map[Operator]*regexp.Regexp{
		In:             regexp.MustCompile(`(\w+)\s+in\s+\(([^)]+)\)`),
		NotIn:          regexp.MustCompile(`(\w+)\s+notin\s+\(([^)]+)\)`),
		NotExist:       regexp.MustCompile(`!\s*(\w+)`),
		Equals:         regexp.MustCompile(`(\w+)\s*=\s*(\w+)`),
		NotEquals:      regexp.MustCompile(`(\w+)\s*!=\s*(\w+)`),
		GreaterThan:    regexp.MustCompile(`(\w+)\s*>\s*(\w+)`),
		LessThan:       regexp.MustCompile(`(\w+)\s*<\s*(\w+)`),
		GreaterOrEqual: regexp.MustCompile(`(\w+)\s*>=\s*(\w+)`),
		LessOrEqual:    regexp.MustCompile(`(\w+)\s*<=\s*(\w+)`),
	}

	// iterate over patterns to find matches
	for op, pattern := range operatorPatterns {
		matches := pattern.FindAllStringSubmatch(in, -1)
		for _, match := range matches {
			key := match[1]
			values := []string{}
			if op == In || op == NotIn {
				values = strings.Split(match[2], ",")
				for i := range values {
					values[i] = strings.TrimSpace(values[i])
				}
			} else if len(match) > 2 {
				values = append(values, match[2])
			}
			reqs = append(reqs, Requirement{
				key:      key,
				operator: op,
				values:   values,
			})
			in = strings.Replace(in, match[0], "", 1)
		}
	}

	// handle 'exists' operator separately
	existsPattern := regexp.MustCompile(`\b(\w+)\b`)
	matches := existsPattern.FindAllStringSubmatch(in, -1)
	for _, match := range matches {
		key := strings.TrimSpace(match[1])
		if key != "" {
			reqs = append(reqs, Requirement{
				key:      key,
				operator: Exists,
				values:   []string{},
			})
		}
	}

	return reqs
}
