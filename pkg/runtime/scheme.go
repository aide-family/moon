package runtime

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/conversion"
)

type Scheme struct {
	kindToType map[string]reflect.Type

	typeToKind map[reflect.Type]string
}

func NewScheme() *Scheme {
	return &Scheme{
		kindToType: map[string]reflect.Type{},
		typeToKind: map[reflect.Type]string{},
	}
}

func (s *Scheme) AddKnownTypeWithName(kind string, obj Object) {
	t := reflect.TypeOf(obj)
	if len(kind) == 0 {
		panic(fmt.Sprintf("kind is required on all types: %s %v", kind, t))
	}
	if t.Kind() != reflect.Ptr {
		panic("All types must be pointers to structs.")
	}
	t = t.Elem()
	if t.Kind() != reflect.Struct {
		panic("All types must be pointers to structs.")
	}

	if oldT, found := s.kindToType[kind]; found && oldT != t {
		panic(fmt.Sprintf("Double registration of different types for %v: old=%v.%v, new=%v.%v in scheme", kind, oldT.PkgPath(), oldT.Name(), t.PkgPath(), t.Name()))
	}

	s.kindToType[kind] = t

	if s.typeToKind[t] == kind {
		return
	}
	s.typeToKind[t] = kind
}

func (s *Scheme) ObjectKind(obj Object) (string, error) {
	v, err := conversion.EnforcePtr(obj)
	if err != nil {
		return "", err
	}
	t := v.Type()

	kind, ok := s.typeToKind[t]
	if !ok {
		return "", fmt.Errorf("kind %s not registered", kind)
	}
	return kind, nil
}

func (s *Scheme) Recognizes(kind string) bool {
	_, exists := s.kindToType[kind]
	return exists
}

// AllKnownTypes returns the all known types.
func (s *Scheme) AllKnownTypes() map[string]reflect.Type {
	return s.kindToType
}

func (s *Scheme) New(kind string) (Object, error) {
	if t, exists := s.kindToType[kind]; exists {
		return reflect.New(t).Interface().(Object), nil
	}
	return nil, fmt.Errorf("kind %s not registered", kind)
}

func UseOrCreateObject(t ObjectTyper, c ObjectCreator, kind string, obj Object) (Object, error) {
	if obj != nil {
		_kind, err := t.ObjectKind(obj)
		if err != nil {
			return nil, err
		}
		if kind == _kind {
			return obj, nil
		}
	}
	return c.New(kind)
}
