package runtime

import (
	"reflect"
	"testing"
)

var _ SchemeObject = &TestStruct{}

type TestStruct struct {
	Field1 string
	Field2 int
}

func (t *TestStruct) KeyFunc(obj interface{}) (string, error) {
	return obj.(*TestStruct).Field1, nil
}

type AnotherStruct struct {
	FieldA bool
}

func (t *AnotherStruct) KeyFunc(obj interface{}) (string, error) {
	return "", nil
}

var ts = &TestStruct{}
var as = &AnotherStruct{}

func TestNewScheme(t *testing.T) {
	s := NewScheme()
	if s == nil {
		t.Fatal("NewScheme should not return nil")
	}
	if len(s.AllKnownTypes()) != 0 {
		t.Fatal("New Scheme should have no registered types")
	}
}

func TestAddKnownTypeWithName(t *testing.T) {
	s := NewScheme()

	err := s.AddKnownTypeWithName("TestStruct", &TestStruct{}, ts.KeyFunc)
	if err != nil {
		t.Errorf("AddKnownTypeWithName returned an unexpected error: %v", err)
	}

	// Double registration with the same type should not err
	err = s.AddKnownTypeWithName("TestStruct", &TestStruct{}, as.KeyFunc)
	if err != nil {
		t.Errorf("AddKnownTypeWithName should not err for valid input: %v", err)
	}

	// Double registration with different type should err
	err = s.AddKnownTypeWithName("TestStruct", &AnotherStruct{}, as.KeyFunc)
	if err == nil {
		t.Errorf("AddKnownTypeWithName should err for different type registration with the same kind")
	}
}

func TestObjectKind(t *testing.T) {
	s := NewScheme()
	err := s.AddKnownTypeWithName("TestStruct", &TestStruct{}, ts.KeyFunc)
	if err != nil {
		t.Errorf("AddKnownTypeWithName returned an unexpected error: %v", err)
	}

	// Valid case
	obj := &TestStruct{}
	kind, err := s.ObjectKind(obj)
	if err != nil {
		t.Errorf("ObjectKind returned an unexpected error: %v", err)
	}
	if kind != "TestStruct" {
		t.Errorf("ObjectKind returned wrong kind: got %v want %v", kind, "TestStruct")
	}

	// Unregistered type
	_, err = s.ObjectKind(&AnotherStruct{})
	if err == nil {
		t.Errorf("ObjectKind should return error for unregistered type")
	} else {
		t.Logf("ObjectKind returned error: %v", err)
	}
}

func TestObjectsKind(t *testing.T) {
	s := NewScheme()
	err := s.AddKnownTypeWithName("TestStruct", &TestStruct{}, ts.KeyFunc)
	if err != nil {
		t.Errorf("AddKnownTypeWithName returned an unexpected error: %v", err)
	}

	// Valid case
	objs := []TestStruct{}
	kind, err := s.ObjectsKind(objs)
	if err != nil {
		t.Errorf("ObjectKind returned an unexpected error: %v", err)
	}
	if kind != "TestStruct" {
		t.Errorf("ObjectKind returned wrong kind: got %v want %v", kind, "TestStruct")
	}

	// Unregistered type
	_, err = s.ObjectsKind([]AnotherStruct{})
	if err == nil {
		t.Errorf("ObjectKind should return error for unregistered type")
	} else {
		t.Logf("ObjectKind returned error: %v", err)
	}
}

func TestSchemeObjectKeyFunc(t *testing.T) {
	s := NewScheme()
	s.AddKnownTypes(&TestStruct{})

	// Test get KeyFunc
	keyFunc, err := s.ObjectKeyFunc(&TestStruct{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	obj := &TestStruct{
		Field1: "mock",
	}
	// Test whether KeyFunc can correctly generate key values
	key, err := keyFunc(obj)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if key != "mock" {
		t.Errorf("Expected key to be mock, got %v", key)
	}
}

func TestSchemeObjectsKeyFunc(t *testing.T) {
	s := NewScheme()
	s.AddKnownTypes(&TestStruct{})

	// Test get KeyFunc
	keyFunc, err := s.ObjectsKeyFunc([]TestStruct{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	obj := &TestStruct{
		Field1: "mock",
	}
	// Test whether KeyFunc can correctly generate key values
	key, err := keyFunc(obj)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if key != "mock" {
		t.Errorf("Expected key to be mock, got %v", key)
	}
}

func TestRecognizes(t *testing.T) {
	s := NewScheme()
	s.AddKnownTypeWithName("TestStruct", &TestStruct{}, ts.KeyFunc)

	if !s.Recognizes("TestStruct") {
		t.Errorf("Recognizes should return true for registered kind")
	}
	if s.Recognizes("UnknownKind") {
		t.Errorf("Recognizes should return false for unregistered kind")
	}
}

func TestAllKnownTypes(t *testing.T) {
	s := NewScheme()
	err := s.AddKnownTypeWithName("TestStruct", &TestStruct{}, ts.KeyFunc)
	if err != nil {
		t.Errorf("AddKnownTypeWithName returned an unexpected error: %v", err)
	}
	s.AddKnownTypeWithName("AnotherStruct", &AnotherStruct{}, as.KeyFunc)

	knownTypes := s.AllKnownTypes()
	if len(knownTypes) != 2 {
		t.Errorf("AllKnownTypes should return all registered types, got %v want %v", len(knownTypes), 2)
	}
	if knownTypes["TestStruct"] != reflect.TypeOf(TestStruct{}) {
		t.Errorf("AllKnownTypes returned wrong type for TestStruct")
	}
	if knownTypes["AnotherStruct"] != reflect.TypeOf(AnotherStruct{}) {
		t.Errorf("AllKnownTypes returned wrong type for AnotherStruct")
	}
}

func TestNew(t *testing.T) {
	s := NewScheme()
	s.AddKnownTypeWithName("TestStruct", &TestStruct{}, ts.KeyFunc)

	// Valid case
	newObj, err := s.New("TestStruct")
	if err != nil {
		t.Errorf("New returned an unexpected error: %v", err)
	}
	if reflect.TypeOf(newObj) != reflect.TypeOf(&TestStruct{}) {
		t.Errorf("New returned wrong type: got %v want %v", reflect.TypeOf(newObj), reflect.TypeOf(&TestStruct{}))
	}

	// Unregistered kind
	_, err = s.New("UnknownKind")
	if err == nil {
		t.Errorf("New should return error for unregistered kind")
	}
}

func TestAddKnownTypes(t *testing.T) {
	s := NewScheme()

	err := s.AddKnownTypes(&TestStruct{}, &AnotherStruct{})
	if err != nil {
		t.Errorf("AddKnownTypes returned an unexpected error: %v", err)
	}

	if !s.Recognizes("TestStruct") {
		t.Errorf("AddKnownTypes failed to register TestStruct")
	}
	if !s.Recognizes("AnotherStruct") {
		t.Errorf("AddKnownTypes failed to register AnotherStruct")
	}

	// Invalid case: non-pointer type
	err = s.AddKnownTypes(TestStruct{})
	if err == nil {
		t.Errorf("AddKnownTypes should err for non-pointer input")
	} else {
		t.Logf("AddKnownTypes returned error: %v", err)
	}

	// Mixed case: pointer and non-pointer types
	err = s.AddKnownTypes(&TestStruct{}, AnotherStruct{})
	if err == nil {
		t.Errorf("AddKnownTypes should err when passed a mix of pointer and non-pointer types")
	} else {
		t.Logf("AddKnownTypes returned error: %v", err)
	}
}

func TestAddKnownTypesEmpty(t *testing.T) {
	s := NewScheme()

	// Valid case: no types provided
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("AddKnownTypes should not panic when no types are provided")
		}
	}()
	s.AddKnownTypes()

	if len(s.AllKnownTypes()) != 0 {
		t.Errorf("AddKnownTypes should not register any types when none are provided")
	}
}
