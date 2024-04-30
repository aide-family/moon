package field

import (
	"strings"
	"time"
)

var _ TableTimeField = (*TimeField)(nil)

type TimeField string

func (n TimeField) Desc() string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(n.String())
	build.WriteString("` DESC")
	return build.String()
}

func (n TimeField) Asc() string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(n.String())
	build.WriteString("`")
	return build.String()
}

func (n TimeField) Eq(value time.Time) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(n.String())
	build.WriteString("` = ")
	build.WriteString("'")
	build.WriteString(value.Format(time.DateTime))
	build.WriteString("'")
	return build.String()
}

func (n TimeField) Neq(value time.Time) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(n.String())
	build.WriteString("` != ")
	build.WriteString("'")
	build.WriteString(value.Format(time.DateTime))
	build.WriteString("'")
	return build.String()
}

func (n TimeField) In(value ...time.Time) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(n.String())
	build.WriteString("` IN (")
	for i, v := range value {
		build.WriteString("'")
		build.WriteString(v.Format(time.DateTime))
		build.WriteString("'")
		if i != len(value)-1 {
			build.WriteString(",")
		}
	}
	build.WriteString(")")
	return build.String()
}

func (n TimeField) NotIn(value ...time.Time) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(n.String())
	build.WriteString("` NOT IN (")
	for i, v := range value {
		build.WriteString("'")
		build.WriteString(v.Format(time.DateTime))
		build.WriteString("'")
		if i != len(value)-1 {
			build.WriteString(",")
		}
	}
	build.WriteString(")")
	return build.String()
}

func (n TimeField) Gt(value time.Time) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(n.String())
	build.WriteString("` > ")
	build.WriteString("'")
	build.WriteString(value.Format(time.DateTime))
	build.WriteString("'")
	return build.String()
}

func (n TimeField) Gte(value time.Time) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(n.String())
	build.WriteString("` >= ")
	build.WriteString("'")
	build.WriteString(value.Format(time.DateTime))
	build.WriteString("'")
	return build.String()
}

func (n TimeField) Lt(value time.Time) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(n.String())
	build.WriteString("` < ")
	build.WriteString("'")
	build.WriteString(value.Format(time.DateTime))
	build.WriteString("'")
	return build.String()
}

func (n TimeField) Lte(value time.Time) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(n.String())
	build.WriteString("` <= ")
	build.WriteString("'")
	build.WriteString(value.Format(time.DateTime))
	build.WriteString("'")
	return build.String()
}

func (n TimeField) String() string {
	return string(n)
}
