package kv

import (
	"database/sql/driver"
	"encoding/json"
	"sort"
	"strings"
)

type KV struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

type KeyValues []*KV

func (kvs KeyValues) ToMap() map[string]string {
	m := make(map[string]string, len(kvs))
	for _, kv := range kvs {
		m[kv.Key] = kv.Value
	}
	return m
}

func (kvs KeyValues) MarshalBinary() ([]byte, error) {
	return json.Marshal(kvs)
}

func (kvs *KeyValues) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, kvs)
}

func (kvs KeyValues) Value() (driver.Value, error) {
	return json.Marshal(kvs)
}

func (kvs *KeyValues) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), kvs)
}

func (kvs KeyValues) String() string {
	marshaled, err := json.Marshal(kvs)
	if err != nil {
		return "[]"
	}
	return string(marshaled)
}

func NewStringMap(ms ...map[string]string) StringMap {
	return New(ms...)
}

type StringMap = Map[string, string]

func SortString(m map[string]string) string {
	keys := StringMap(m).Keys()
	sort.Strings(keys)
	var buf strings.Builder
	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteString(": ")
		buf.WriteString(m[k])
	}
	return buf.String()
}
