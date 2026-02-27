package feishu

import "github.com/aide-family/rabbit/pkg/message"

type Post struct {
	ZhCn *Markdown `json:"zh_cn,omitempty"`
	EnUs *Markdown `json:"en_us,omitempty"`
}

type Markdown struct {
	Title   string         `json:"title"`
	Content [][]*Paragraph `json:"content"`
}

type Paragraph struct {
	Tag      string `json:"tag"`
	Text     string `json:"text"`
	UnEscape string `json:"un_escape"`
	Href     string `json:"href"`
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	ImageKey string `json:"image_key"`
}

func NewPostMessage() *Post {
	return &Post{}
}

func (p *Post) WithZhCn(title string, content [][]*Paragraph) *Post {
	p.ZhCn = &Markdown{
		Title:   title,
		Content: content,
	}
	return p
}

func (p *Post) WithEnUs(title string, content [][]*Paragraph) *Post {
	p.EnUs = &Markdown{
		Title:   title,
		Content: content,
	}
	return p
}

func (p *Post) Message() message.Message {
	return &Message{
		MsgType: MessageTypePost,
		Content: &Content{
			Post: p,
		},
	}
}
