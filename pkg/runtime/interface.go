package runtime

import (
	"io"

	"github.com/aide-family/moon/pkg/runtime/schema"
)

// Object is an object that has an ObjectKind.
type Object interface {
	// GetObjectKind returns the kind of object, as represented by the ObjectKind
	GetObjectKind() schema.ObjectKind
	// DeepCopyObject returns a copy of the object.
	DeepCopyObject() Object
}

// Encoder knows how to encode an object into a stream.
type Encoder interface {
	// Encode writes a set of objects to a stream.
	Encode(obj Object, w io.Writer) error
}

// Decoder knows how to decode an object from a stream.
type Decoder interface {
	// Decode attempts to deserialize the provided bytes into the provided object.
	Decode(data []byte, defaults *string, into Object) (Object, *string, error)
}

// Serializer knows how to serialize and deserialize objects.
type Serializer interface {
	// Encoder knows how to serialize an object.
	Encoder
	// Decoder knows how to deserialize an object.
	Decoder
}

// Framer knows how to wrap and unwrap a stream of data.
type Framer interface {
	// NewFrameReader returns a Reader that wraps the provided Reader,
	NewFrameReader(r io.ReadCloser) io.ReadCloser
	// NewFrameWriter returns a Writer that wraps the provided Writer.
	NewFrameWriter(w io.Writer) io.Writer
}

// ObjectTyper knows how to get the kind of an object.
type ObjectTyper interface {
	// ObjectKind returns the kind of the provided object.
	ObjectKind(Object) (string, error)
	// Recognizes returns true if the provided kind is known to this typer.
	Recognizes(kind string) bool
}

// ObjectCreator knows how to create an object.
type ObjectCreator interface {
	// New returns a new object of the provided kind.
	New(kind string) (out Object, err error)
}
