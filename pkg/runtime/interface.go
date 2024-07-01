package runtime

import (
	"github.com/aide-family/moon/pkg/runtime/schema"
	"io"
)

type Object interface {
	GetObjectKind() schema.ObjectKind
	DeepCopyObject() Object
}

type Encoder interface {
	Encode(obj Object, w io.Writer) error
}

type Decoder interface {
	Decode(data []byte, defaults *string, into Object) (Object, *string, error)
}

type Serializer interface {
	Encoder
	Decoder
}

type Framer interface {
	NewFrameReader(r io.ReadCloser) io.ReadCloser
	NewFrameWriter(w io.Writer) io.Writer
}

type ObjectTyper interface {
	ObjectKind(Object) (string, error)
	Recognizes(kind string) bool
}

type ObjectCreator interface {
	New(kind string) (out Object, err error)
}
