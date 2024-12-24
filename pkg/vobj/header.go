package vobj

type (
	// Header http请求头
	Header struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
)

// NewHeader 创建一个Header对象
func NewHeader(name, value string) *Header {
	return &Header{
		Name:  name,
		Value: value,
	}
}
