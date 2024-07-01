package json

import (
	"encoding/json"
	"fmt"
	"github.com/aide-family/moon/pkg/runtime"
	"github.com/aide-family/moon/pkg/runtime/serializer/recognizer"
	"github.com/aide-family/moon/pkg/util/yaml"
	"io"
	"k8s.io/apimachinery/pkg/util/framer"
)

func NewSerializer(meta MetaFactory, creater runtime.ObjectCreator, typer runtime.ObjectTyper, pretty bool) *Serializer {
	return NewSerializerWithOptions(meta, creater, typer)
}

func NewYAMLSerializer(meta MetaFactory, creater runtime.ObjectCreator, typer runtime.ObjectTyper) *Serializer {
	return NewSerializerWithOptions(meta, creater, typer)
}

func NewSerializerWithOptions(meta MetaFactory, creater runtime.ObjectCreator, typer runtime.ObjectTyper) *Serializer {
	return &Serializer{
		meta:    meta,
		creator: creater,
		typer:   typer,
	}
}

// Serializer handles encoding versioned objects into the proper JSON form
type Serializer struct {
	meta    MetaFactory
	creator runtime.ObjectCreator
	typer   runtime.ObjectTyper
}

// Serializer implements Serializer
var _ runtime.Serializer = &Serializer{}
var _ recognizer.RecognizingDecoder = &Serializer{}

func gvkWithDefaults(actual, defaults string) string {
	if len(actual) == 0 {
		actual = defaults
	}
	return actual
}

func (s *Serializer) Decode(originalData []byte, gvk *string, into runtime.Object) (runtime.Object, *string, error) {
	data := originalData

	actual, err := s.meta.Interpret(data)
	if err != nil {
		return nil, nil, err
	}

	if gvk != nil {
		*actual = gvkWithDefaults(*actual, *gvk)
	}

	if into != nil {
		types, err := s.typer.ObjectKind(into)
		switch {
		case err != nil:
			return nil, actual, err
		default:
			*actual = gvkWithDefaults(*actual, types)
		}
	}

	if len(*actual) == 0 {
		return nil, actual, fmt.Errorf("missing kind %s", *actual)
	}

	// use the target if necessary
	obj, err := runtime.UseOrCreateObject(s.typer, s.creator, *actual, into)
	if err != nil {
		return nil, actual, err
	}

	err = s.unmarshal(obj, data, originalData)
	if err != nil {
		return nil, actual, err
	}
	return obj, actual, nil
}

// Encode serializes the provided object to the given writer.
func (s *Serializer) Encode(obj runtime.Object, w io.Writer) error {
	return s.doEncode(obj, w)
}

func (s *Serializer) doEncode(obj runtime.Object, w io.Writer) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(obj)
}

func (s *Serializer) unmarshal(into runtime.Object, data, originalData []byte) error {
	return json.Unmarshal(data, into)

}

// RecognizesData implements the RecognizingDecoder interface.
func (s *Serializer) RecognizesData(data []byte) (ok, unknown bool, err error) {
	return yaml.IsJSONBuffer(data), false, nil
}

// Framer is the default JSON framing behavior, with newlines delimiting individual objects.
var Framer = jsonFramer{}

type jsonFramer struct{}

// NewFrameWriter implements stream framing for this serializer
func (jsonFramer) NewFrameWriter(w io.Writer) io.Writer {
	// we can write JSON objects directly to the writer, because they are self-framing
	return w
}

// NewFrameReader implements stream framing for this serializer
func (jsonFramer) NewFrameReader(r io.ReadCloser) io.ReadCloser {
	// we need to extract the JSON chunks of data to pass to Decode()
	return framer.NewJSONFramedReader(r)
}
