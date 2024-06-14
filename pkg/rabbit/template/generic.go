package template

import "io"

type GenericTemplateParser struct {
	Template string
}

func (g *GenericTemplateParser) Parse(in any, out io.Writer) error {
	return Parser(g.Template, in, out)
}
