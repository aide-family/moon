package recognizer

import (
	"fmt"
	"github.com/aide-family/moon/pkg/runtime"
)

type RecognizingDecoder interface {
	runtime.Decoder
	RecognizesData(peek []byte) (ok, unknown bool, err error)
}

func NewDecoder(decoders ...runtime.Decoder) runtime.Decoder {
	return &decoder{
		decoders: decoders,
	}
}

type decoder struct {
	decoders []runtime.Decoder
}

var _ RecognizingDecoder = &decoder{}

func (d *decoder) RecognizesData(data []byte) (bool, bool, error) {
	var (
		lastErr    error
		anyUnknown bool
	)
	for _, r := range d.decoders {
		switch t := r.(type) {
		case RecognizingDecoder:
			ok, unknown, err := t.RecognizesData(data)
			if err != nil {
				lastErr = err
				continue
			}
			anyUnknown = anyUnknown || unknown
			if !ok {
				continue
			}
			return true, false, nil
		}
	}
	return false, anyUnknown, lastErr
}

func (d *decoder) Decode(data []byte, gvk *string, into runtime.Object) (runtime.Object, *string, error) {
	var (
		lastErr error
		skipped []runtime.Decoder
	)

	// try recognizers, record any decoders we need to give a chance later
	for _, r := range d.decoders {
		switch t := r.(type) {
		case RecognizingDecoder:
			ok, unknown, err := t.RecognizesData(data)
			if err != nil {
				lastErr = err
				continue
			}
			if unknown {
				skipped = append(skipped, t)
				continue
			}
			if !ok {
				continue
			}
			return r.Decode(data, gvk, into)
		default:
			skipped = append(skipped, t)
		}
	}

	// try recognizers that returned unknown or didn't recognize their data
	for _, r := range skipped {
		out, actual, err := r.Decode(data, gvk, into)
		if err != nil {
			if out == nil {
				lastErr = err
				continue
			}
		}
		return out, actual, err
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("no serialization format matched the provided data")
	}
	return nil, nil, lastErr
}
