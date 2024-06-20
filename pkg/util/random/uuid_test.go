package random

import (
	"testing"
)

func TestUUID(t *testing.T) {
	t.Log(UUID())
	t.Log(UUID())
	t.Log(UUID(true))
	t.Log(UUID(true))
}

func TestUUIDToUpperCase(t *testing.T) {
	t.Log(UUIDToUpperCase())
	t.Log(UUIDToUpperCase())
	t.Log(UUIDToUpperCase(true))
	t.Log(UUIDToUpperCase(true))
}
