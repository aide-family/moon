package feishu

import "github.com/aide-family/rabbit/pkg/message"

type CardConfig struct {
	UpdateMulti bool             `json:"update_multi"`
	Style       *CardConfigStyle `json:"style"`
}

type CardConfigStyle struct {
	TextSize *CardConfigStyleTextSize `json:"text_size"`
}

type CardConfigStyleTextSize struct {
	NormalV2 *CardConfigStyleTextSizeNormalV2 `json:"normal_v2"`
}

type CardConfigStyleTextSizeNormalV2 struct {
	Default string `json:"default"`
	Pc      string `json:"pc"`
	Mobile  string `json:"mobile"`
}

type CardBodyDirection string

const (
	CardBodyDirectionHorizontal CardBodyDirection = "horizontal"
	CardBodyDirectionVertical   CardBodyDirection = "vertical"
)

type CardBody struct {
	Direction CardBodyDirection  `json:"direction"`
	Padding   string             `json:"padding"`
	Elements  []*CardBodyElement `json:"elements"`
}

type CardBodyElement struct {
	Tag       string                      `json:"tag"`
	Content   string                      `json:"content,omitempty"`
	TextAlign string                      `json:"text_align,omitempty"`
	TextSize  string                      `json:"text_size,omitempty"`
	Margin    string                      `json:"margin"`
	Text      *CardBodyElementText        `json:"text,omitempty"`
	Type      string                      `json:"type,omitempty"`
	Width     string                      `json:"width,omitempty"`
	Size      string                      `json:"size,omitempty"`
	Behaviors []*CardBodyElementBehaviors `json:"behaviors,omitempty"`
}

type CardBodyElementText struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

type CardBodyElementBehaviors struct {
	Type       string `json:"type"`
	DefaultURL string `json:"default_url"`
	PcURL      string `json:"pc_url"`
	IosURL     string `json:"ios_url"`
	AndroidURL string `json:"android_url"`
}

type CardHeader struct {
	Title    *CardHeaderTitle    `json:"title"`
	Subtitle *CardHeaderSubtitle `json:"subtitle"`
	Template string              `json:"template"`
	Padding  string              `json:"padding"`
}

type CardHeaderTitle struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

type CardHeaderSubtitle struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

type Card struct {
	Schema string      `json:"schema"`
	Config *CardConfig `json:"config"`
	Body   *CardBody   `json:"body"`
	Header *CardHeader `json:"header"`
}

func NewCardMessage() *Card {
	return &Card{}
}

func (c *Card) WithSchema(schema string) *Card {
	c.Schema = schema
	return c
}

func (c *Card) WithConfig(config *CardConfig) *Card {
	c.Config = config
	return c
}

func (c *Card) WithBody(body *CardBody) *Card {
	c.Body = body
	return c
}

func (c *Card) WithHeader(header *CardHeader) *Card {
	c.Header = header
	return c
}

func (c *Card) Message() message.Message {
	return &Message{
		MsgType: MessageTypeCard,
		Content: &Content{
			Card: c,
		},
	}
}
