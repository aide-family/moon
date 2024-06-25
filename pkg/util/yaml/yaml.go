package yaml

import (
	"bytes"
	"gopkg.in/yaml.v3"
	"reflect"
	"unicode"
)

// IsJSONBuffer scans the provided buffer, looking
// for an open brace indicating this is JSON.
func IsJSONBuffer(buf []byte) bool {
	return hasJSONPrefix(buf)
}

var jsonPrefix = []byte("{")

// hasJSONPrefix returns true if the provided buffer appears to start with
// a JSON open brace.
func hasJSONPrefix(buf []byte) bool {
	return hasPrefix(buf, jsonPrefix)
}

// Return true if the first non-whitespace bytes in buf is
// prefix.
func hasPrefix(buf []byte, prefix []byte) bool {
	trim := bytes.TrimLeftFunc(buf, unicode.IsSpace)
	return bytes.HasPrefix(trim, prefix)
}

// JSONToYAML Converts JSON to YAML.
func JSONToYAML(j []byte) ([]byte, error) {
	// Convert the JSON to an object.
	var jsonObj interface{}
	// We are using yaml.Unmarshal here (instead of json.Unmarshal) because the
	// Go JSON library doesn't try to pick the right number type (int, float,
	// etc.) when unmarshalling to interface{}, it just picks float64
	// universally. go-yaml does go through the effort of picking the right
	// number type, so we can preserve number type throughout this process.
	err := yaml.Unmarshal(j, &jsonObj)
	if err != nil {
		return nil, err
	}

	// Marshal this object into YAML.
	return yaml.Marshal(jsonObj)
}
func YAMLToJSON(y []byte) ([]byte, error) {
	return yamlToJSON(y, nil, yaml.Unmarshal)
}

func yamlToJSON(y []byte, jsonTarget *reflect.Value, yamlUnmarshal func([]byte, interface{}) error) ([]byte, error) {
	// Convert the YAML to an object.
	var yamlObj interface{}
	err := yamlUnmarshal(y, &yamlObj)
	if err != nil {
		return nil, err
	}

	// YAML objects are not completely compatible with JSON objects (e.g. you
	// can have non-string keys in YAML). So, convert the YAML-compatible object
	// to a JSON-compatible object, failing with an error if irrecoverable
	// incompatibilties happen along the way.
	jsonObj, err := convertToJSONableObject(yamlObj, jsonTarget)
	if err != nil {
		return nil, err
	}

	// Convert this object to JSON and return the data.
	return json.Marshal(jsonObj)
}
