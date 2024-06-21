package watch

type (
	Schemer interface {
		// Decode 解码
		Decode(in Indexer, out any) error
		// Encode 编码
		Encode(in Indexer, out any) error
	}
)
