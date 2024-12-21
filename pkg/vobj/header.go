package vobj

type (
	// Header http请求头
	Header struct {
		Name  string
		Value string
	}
)

// NewHeader 创建一个Header对象
func NewHeader(name, value string) *Header {
	return &Header{
		Name:  name,
		Value: value,
	}
}
