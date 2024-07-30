package template

import "io"

// GenericTemplateParser is a generic template parser.
type GenericTemplateParser struct {
	Template string
}

// Parse parses the template and writes the result to the writer.
func (g *GenericTemplateParser) Parse(in any, out io.Writer) error {
	return Parser(g.Template, in, out)
}
