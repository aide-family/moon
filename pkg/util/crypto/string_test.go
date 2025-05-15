package crypto_test

import (
	"strings"
	"testing"

	"github.com/moon-monitor/moon/pkg/util/crypto"
)

func TestString_Scan_Success(t *testing.T) {
	var s crypto.String = "1058165620@qq.com"
	val, err := s.Value()
	if err != nil {
		t.Fatalf("Expected no error, got %v\n", err)
	}

	t.Logf("val: %v, %d", string(val.([]byte)), len(val.([]byte)))
	var got crypto.String
	if err := got.Scan(val); err != nil {
		t.Fatalf("Expected no error, got %v\n", err)
	}
	if !strings.EqualFold(string(s), string(got)) {
		t.Errorf("Expected '%v', got '%v'", s, got)
	}
}

func TestString_Scan_NULL_Success(t *testing.T) {
	var s crypto.String = ""
	val, err := s.Value()
	if err != nil {
		t.Fatalf("Expected no error, got %v\n", err)
	}

	t.Logf("val: %v, %d", string(val.([]byte)), len(val.([]byte)))
	var got crypto.String
	if err := got.Scan(val); err != nil {
		t.Fatalf("Expected no error, got %v\n", err)
	}
	if !strings.EqualFold(string(s), string(got)) {
		t.Errorf("Expected '%v', got '%v'", s, got)
	}
}
