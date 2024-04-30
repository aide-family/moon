package field

import (
	"strconv"
	"strings"
)

var _ TableNumberField = (*NumberField)(nil)

type NumberField string

func (n NumberField) String() string {
	return string(n)
}

func (n NumberField) Desc() string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(n.String())
	build.WriteString("` DESC")
	return build.String()
}

func (n NumberField) Asc() string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(n.String())
	build.WriteString("`")
	return build.String()
}

func (n NumberField) Eq(value int) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(n.String())
	build.WriteString("` = ")
	build.WriteString(strconv.Itoa(value))
	return build.String()
}

func (n NumberField) Neq(value int) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(n.String())
	build.WriteString("` != ")
	build.WriteString(strconv.Itoa(value))
	return build.String()
}

func (n NumberField) In(value ...int) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(n.String())
	build.WriteString("` IN (")
	for i, v := range value {
		if i > 0 {
			build.WriteString(",")
		}
		build.WriteString(strconv.Itoa(v))
	}
	build.WriteString(")")
	return build.String()
}

func (n NumberField) NotIn(value ...int) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(n.String())
	build.WriteString("` NOT IN (")
	for i, v := range value {
		if i > 0 {
			build.WriteString(",")
		}
		build.WriteString(strconv.Itoa(v))
	}
	build.WriteString(")")
	return build.String()
}

func (n NumberField) Gt(value int) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(n.String())
	build.WriteString("` > ")
	build.WriteString(strconv.Itoa(value))
	return build.String()
}

func (n NumberField) Gte(value int) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(n.String())
	build.WriteString("` >= ")
	build.WriteString(strconv.Itoa(value))
	return build.String()
}

func (n NumberField) Lt(value int) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(n.String())
	build.WriteString("` < ")
	build.WriteString(strconv.Itoa(value))
	return build.String()
}

func (n NumberField) Lte(value int) string {
	build := strings.Builder{}
	build.WriteString("`")
	build.WriteString(n.String())
	build.WriteString("` <= ")
	build.WriteString(strconv.Itoa(value))
	return build.String()
}
