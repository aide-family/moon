package schema

type ObjectKind interface {
	SetKind(kind string)
	GetKind() string
}

type emptyObjectKind struct{}

func (emptyObjectKind) SetKind(kind string) {}

func (emptyObjectKind) GetKind() string { return "" }
