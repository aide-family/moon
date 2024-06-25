package api

func (o *ObjectMeta) SetName(name string) {
	o.Name = name
}

func (o *TypeMeta) SetKind(kind string) {
	o.Kind = kind
}
