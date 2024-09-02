package ptr

import (
	"reflect"
	"testing"
)

type TestStruct struct {
	Field1 string
	Field2 int
}

func TestEnforcePtr(t *testing.T) {
	obj := &TestStruct{Field1: "test", Field2: 123}

	// Valid case
	val, err := EnforcePtr(obj)
	if err != nil {
		t.Errorf("EnforcePtr returned an unexpected error: %v", err)
	}
	if val.Kind() != reflect.Struct {
		t.Errorf("EnforcePtr returned wrong value: got %v want struct", val.Kind())
	}

	// Invalid case: not a pointer
	_, err = EnforcePtr(TestStruct{})
	if err == nil {
		t.Errorf("EnforcePtr should return error for non-pointer type")
	}

	// Invalid case: nil pointer
	var nilObj *TestStruct
	_, err = EnforcePtr(nilObj)
	if err == nil {
		t.Errorf("EnforcePtr should return error for nil pointer")
	}
}

func TestGenerateElementBySlice(t *testing.T) {
	// Test case 1: Valid slice input
	slice := []int{1, 2, 3}
	element, err := GenerateElementBySlice(slice)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if reflect.TypeOf(element) != reflect.TypeOf(slice[0]) {
		t.Errorf("Expected element type %v, got %v", reflect.TypeOf(&slice[0]), reflect.TypeOf(element))
	}

	// Test case 2: Invalid input (not a slice)
	nonSlice := 123
	_, err = GenerateElementBySlice(nonSlice)
	if err == nil {
		t.Errorf("Expected error for non-slice input, got none")
	}

	expectedErrMsg := "input must be a slice or pointer to a slice"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error message %q, got %q", expectedErrMsg, err.Error())
	}

	// Test case 3: Empty slice input
	emptySlice := []string{}
	element, err = GenerateElementBySlice(emptySlice)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if reflect.TypeOf(element).Kind() != reflect.String {
		t.Errorf("Expected element type %v, got %v", reflect.TypeOf(reflect.String), reflect.TypeOf(element))
	}

	element, err = GenerateElementBySlice(&emptySlice)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if reflect.TypeOf(element).Kind() != reflect.String {
		t.Errorf("Expected element type %v, got %v", reflect.TypeOf(reflect.String), reflect.TypeOf(element))
	}
}

func TestGenerateElementPtrBySlice(t *testing.T) {
	// Test case 1: Valid slice input
	slice := []int{1, 2, 3}
	element, err := GenerateElementPtrBySlice(slice)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if reflect.TypeOf(element) != reflect.TypeOf(&slice[0]) {
		t.Errorf("Expected element type %v, got %v", reflect.TypeOf(&slice[0]), reflect.TypeOf(element))
	}

	element, err = GenerateElementPtrBySlice(&slice)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if reflect.TypeOf(element) != reflect.TypeOf(&slice[0]) {
		t.Errorf("Expected element type %v, got %v", reflect.TypeOf(&slice[0]), reflect.TypeOf(element))
	}

	var slice2 []int
	element, err = GenerateElementPtrBySlice(&slice2)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test case 2: Invalid input (not a slice)
	nonSlice := 123
	_, err = GenerateElementPtrBySlice(nonSlice)
	if err == nil {
		t.Errorf("Expected error for non-slice input, got none")
	}

	expectedErrMsg := "input must be a slice or pointer to a slice"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error message %q, got %q", expectedErrMsg, err.Error())
	}

	// Test case 3: Empty slice input
	emptySlice := []string{}
	element, err = GenerateElementPtrBySlice(emptySlice)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if reflect.TypeOf(element) != reflect.TypeOf(new(string)) {
		t.Errorf("Expected element type %v, got %v", reflect.TypeOf(new(string)), reflect.TypeOf(element))
	}
}
