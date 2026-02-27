// Package compress provides a gzip compression and decompression.
package compress

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
)

// Gzip compresses the data.
func Gzip(data []byte) (io.Reader, error) {
	buf := bytes.NewBuffer(make([]byte, 0, len(data)/3))
	gz := gzip.NewWriter(buf)
	if _, err := gz.Write(data); err != nil {
		return nil, err
	}
	if err := gz.Flush(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf, nil
}

// GzipBytes compresses the bytes.
func GzipBytes(data []byte) ([]byte, error) {
	reader, err := Gzip(data)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(reader)
}

// GzipJSON compresses the JSON data.
func GzipJSON(data any) (io.Reader, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return Gzip(jsonData)
}

// GzipJSONBytes compresses the JSON data.
func GzipJSONBytes(data any) ([]byte, error) {
	reader, err := GzipJSON(data)
	if err != nil {
		return nil, err
	}
	return io.ReadAll(reader)
}

// UnGzip decompresses the data.
func UnGzip(data io.Reader) ([]byte, error) {
	gz, err := gzip.NewReader(data)
	if err != nil {
		return nil, err
	}
	defer gz.Close()
	return io.ReadAll(gz)
}

// UnGzipBytes decompresses the bytes.
func UnGzipBytes(data []byte) ([]byte, error) {
	return UnGzip(bytes.NewReader(data))
}

// UnGzipJSONUnmarshal decompresses the JSON data.
func UnGzipJSONUnmarshal(data io.Reader, v any) error {
	uncompressed, err := UnGzip(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(uncompressed, v)
}

// UnGzipJSONUnmarshalBytes decompresses the JSON data.
func UnGzipJSONUnmarshalBytes(data []byte, v any) error {
	return UnGzipJSONUnmarshal(bytes.NewReader(data), v)
}
