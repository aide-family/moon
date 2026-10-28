package mailer

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/pkg/merr"
)

// Sender 是一封纯文本邮件的发送器。未配置 SMTP 时 Send 返回服务端错误。
type Sender struct {
	host     string
	port     int
	username string
	password string
	from     string
	log      *slog.Logger
}

func NewSender(auth *conf.Auth, logger *slog.Logger) *Sender {
	sender := &Sender{log: logger}
	if auth != nil && auth.GetSmtp() != nil {
		sender.host = auth.GetSmtp().GetHost()
		sender.port = int(auth.GetSmtp().GetPort())
		sender.username = auth.GetSmtp().GetUsername()
		sender.password = auth.GetSmtp().GetPassword()
		sender.from = auth.GetSmtp().GetFrom()
	}
	return sender
}

func (s *Sender) Configured() bool {
	return s != nil && s.host != "" && s.port > 0 && s.from != ""
}

// Send 发送一封 UTF-8 纯文本邮件。主题和正文用 base64 编码，避免中文被 SMTP 网关截断。
// 信封发件地址和 From 头分开处理：163 等邮箱要求信封地址就是登录账号本身，
// 显示名（例如“Alioth 瑶光”）只能放在 From 头里，直接塞进 MAIL FROM 会被服务端拒绝。
func (s *Sender) Send(to, subject, body string) error {
	if !s.Configured() {
		return merr.Server("smtp is not configured")
	}
	if strings.ContainsAny(subject, "\r\n") {
		return merr.Client("invalid email header")
	}
	envelopeFrom, headerFrom, err := s.senderIdentity()
	if err != nil {
		return err
	}
	envelopeTo, err := recipientAddress(to)
	if err != nil {
		return err
	}
	message := buildMessage(headerFrom, envelopeTo, subject, body)
	addr := net.JoinHostPort(s.host, strconv.Itoa(s.port))
	var auth smtp.Auth
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}
	if err := smtp.SendMail(addr, auth, envelopeFrom, []string{envelopeTo}, message); err != nil {
		s.log.Error(fmt.Sprintf("smtp: %v", err))
		return merr.Server("failed to send email")
	}
	return nil
}

// senderIdentity 解析配置里的发件人。from 支持裸地址和带显示名两种写法：
// “aidedevops@163.com” 或 “Alioth 瑶光 <aidedevops@163.com>”。
// 历史配置可能只写了显示名（例如“Alioth 瑶光”），这时退化为用 SMTP 用户名当信封发件地址。
// 原因是 MAIL FROM:<Alioth 瑶光> 里的空格和中文会被 SMTP 服务器以 500/550 拒绝。
func (s *Sender) senderIdentity() (envelope, header string, err error) {
	raw := strings.TrimSpace(s.from)
	if parsed, parseErr := mail.ParseAddress(raw); parseErr == nil {
		return parsed.Address, parsed.String(), nil
	}
	username, parseErr := mail.ParseAddress(strings.TrimSpace(s.username))
	if parseErr != nil {
		return "", "", merr.Server("smtp from address is invalid")
	}
	s.log.Warn("smtp from is not a mailbox address, using the smtp username as the envelope sender")
	return username.Address, (&mail.Address{Name: raw, Address: username.Address}).String(), nil
}

// recipientAddress 把收件人规范成裸地址。换行注入在这里就拒绝，不进入 SMTP 协议层。
func recipientAddress(to string) (string, error) {
	to = strings.TrimSpace(to)
	if strings.ContainsAny(to, "\r\n") {
		return "", merr.Client("invalid email header")
	}
	parsed, err := mail.ParseAddress(to)
	if err != nil {
		return "", merr.Client("invalid email address")
	}
	return parsed.Address, nil
}

// buildMessage 拼出 RFC 5322 纯文本邮件。显示名由 mail.Address 按 RFC 2047 编码，
// 正文统一 base64，避免 SMTP 网关截断中文。
func buildMessage(headerFrom, to, subject, body string) []byte {
	payload := base64.StdEncoding.EncodeToString([]byte(body))
	return []byte("From: " + headerFrom + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + mime.QEncoding.Encode("utf-8", subject) + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: base64\r\n\r\n" +
		payload + "\r\n")
}
