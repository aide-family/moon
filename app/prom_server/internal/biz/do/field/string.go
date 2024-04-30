package field

import (
	"strings"
)

var _ TableStringField = (*StringField)(nil)

type StringField string

func (s StringField) String() string {
	return string(s)
}

func (s StringField) Desc() string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(s.String())
	build.WriteString("` DESC")
	return build.String()
}

func (s StringField) Asc() string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(s.String())
	build.WriteString("`")
	return build.String()
}

func (s StringField) Eq(value string) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(s.String())
	build.WriteString("` = ")
	build.WriteString("'")
	build.WriteString(value)
	build.WriteString("'")
	return build.String()
}

func (s StringField) Neq(value string) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(s.String())
	build.WriteString("` != ")
	build.WriteString("'")
	build.WriteString(value)
	build.WriteString("'")
	return build.String()
}

func (s StringField) In(value ...string) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(s.String())
	build.WriteString("` IN (")
	for i, v := range value {
		if i > 0 {
			build.WriteString(",")
		}
		build.WriteString("'")
		build.WriteString(v)
		build.WriteString("'")
	}
	build.WriteString(")")
	return build.String()
}

func (s StringField) NotIn(value ...string) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(s.String())
	build.WriteString("` NOT IN (")
	for i, v := range value {
		if i > 0 {
			build.WriteString(",")
		}
		build.WriteString("'")
		build.WriteString(v)
		build.WriteString("'")
	}
	build.WriteString(")")
	return build.String()
}

func (s StringField) Like(value string) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(s.String())
	build.WriteString("` LIKE '%")
	build.WriteString(value)
	build.WriteString("%'")
	return build.String()
}

func (s StringField) Prefix(value string) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(s.String())
	build.WriteString("` LIKE '")
	build.WriteString(value)
	build.WriteString("%'")
	return build.String()
}

func (s StringField) Suffix(value string) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(s.String())
	build.WriteString("` LIKE '%")
	build.WriteString(value)
	build.WriteString("'")
	return build.String()
}

func (s StringField) NotPrefix(value string) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(s.String())
	build.WriteString("` NOT LIKE '")
	build.WriteString(value)
	build.WriteString("%'")
	return build.String()
}

func (s StringField) NotSuffix(value string) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(s.String())
	build.WriteString("` NOT LIKE '%")
	build.WriteString(value)
	build.WriteString("'")
	return build.String()
}

func (s StringField) NotLike(value string) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(s.String())
	build.WriteString("` NOT LIKE '%")
	build.WriteString(value)
	build.WriteString("%'")
	return build.String()
}
