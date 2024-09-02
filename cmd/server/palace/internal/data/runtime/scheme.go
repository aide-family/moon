package runtime

import (
	"fmt"
	"reflect"

	"github.com/aide-family/moon/pkg/util/ptr"

	"k8s.io/client-go/tools/cache"
)

// Scheme is a structure used for managing type registration and lookup by their kind names.
// It maintains a bidirectional mapping between kind names (strings) and their corresponding Go types (reflect.Type).
type Scheme struct {
	// kindToType maps a kind name (string) to its corresponding Go type (reflect.Type).
	// This allows looking up a type by its registered kind name.
	kindToType map[string]reflect.Type

	// kindToFunc maps a kind name (string) to a function that generates a unique key for that kind.
	kindToKeyFunc map[string]cache.KeyFunc

	// typeToKind maps a Go type (reflect.Type) to its corresponding kind name (string).
	// This allows looking up the registered kind name for a given type.
	typeToKind map[reflect.Type]string

	// typeToKeyFunc maps a Go type (reflect.Type) to a function that generates a unique key for that type.
	typeToKeyFunc map[reflect.Type]cache.KeyFunc
}

type SchemeObject interface {
	KeyFunc(obj interface{}) (string, error)
}

func NewScheme() *Scheme {
	return &Scheme{
		kindToType:    map[string]reflect.Type{},
		kindToKeyFunc: map[string]cache.KeyFunc{},
		typeToKind:    map[reflect.Type]string{},
		typeToKeyFunc: map[reflect.Type]cache.KeyFunc{},
	}
}

func (s *Scheme) AddKnownTypes(types ...any) error {
	for _, obj := range types {
		object, ok := obj.(SchemeObject)
		if !ok {
			return fmt.Errorf("all types must implement SchemeObject")
		}
		t := reflect.TypeOf(obj)
		if t.Kind() != reflect.Ptr {
			return fmt.Errorf("all types must be pointers to structs")
		}
		t = t.Elem()
		err := s.AddKnownTypeWithName(t.Name(), obj, object.KeyFunc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Scheme) AddKnownTypeWithName(kind string, obj any, keyFunc cache.KeyFunc) error {
	if len(kind) == 0 {
		return fmt.Errorf("kind is required and cannot be empty")
	}

	t := reflect.TypeOf(obj)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("all types must be pointers to structs")
	}

	t = t.Elem()

	if oldT, found := s.kindToType[kind]; found && oldT != t {
		return fmt.Errorf("double registration of different types for %v: old=%v.%v, new=%v.%v in scheme", kind, oldT.PkgPath(), oldT.Name(), t.PkgPath(), t.Name())
	}

	if keyFunc == nil {
		return fmt.Errorf("keyFunc is required for object %s, in scheme", kind)
	}

	s.kindToType[kind] = t
	s.kindToKeyFunc[kind] = keyFunc
	s.typeToKind[t] = kind
	s.typeToKeyFunc[t] = keyFunc
	return nil
}

func (s *Scheme) ObjectKind(obj any) (string, error) {
	v, err := ptr.EnforcePtr(obj)
	if err != nil {
		return "", err
	}

	if kind, ok := s.typeToKind[v.Type()]; ok {
		return kind, nil
	}
	return "", fmt.Errorf("kind not registered for type %v", v.Type())
}

func (s *Scheme) ObjectsKind(objs any) (string, error) {
	obj, err := ptr.GenerateElementPtrBySlice(objs)
	if err != nil {
		return "", err
	}
	return s.ObjectKind(obj)
}

func (s *Scheme) ObjectKeyFunc(obj any) (cache.KeyFunc, error) {
	v, err := ptr.EnforcePtr(obj)
	if err != nil {
		return nil, err
	}

	if keyFunc, ok := s.typeToKeyFunc[v.Type()]; ok {
		return keyFunc, nil
	}
	return nil, fmt.Errorf("keyFunc not registered for type %v", v.Type())
}

func (s *Scheme) ObjectsKeyFunc(objs any) (cache.KeyFunc, error) {
	obj, err := ptr.GenerateElementPtrBySlice(objs)
	if err != nil {
		return nil, err
	}
	return s.ObjectKeyFunc(obj)
}

func (s *Scheme) KindKeyFunc(kind string) (cache.KeyFunc, error) {

	if keyFunc, ok := s.kindToKeyFunc[kind]; ok {
		return keyFunc, nil
	}
	return nil, fmt.Errorf("keyFunc not registered for kind %v", kind)
}

func (s *Scheme) Recognizes(kind string) bool {
	_, exists := s.kindToType[kind]
	return exists
}

// AllKnownTypes returns the all known types.
func (s *Scheme) AllKnownTypes() map[string]reflect.Type {
	return s.kindToType
}

func (s *Scheme) New(kind string) (any, error) {
	if t, exists := s.kindToType[kind]; exists {
		return reflect.New(t).Interface(), nil
	}
	return nil, fmt.Errorf("kind %s not registered", kind)
}
