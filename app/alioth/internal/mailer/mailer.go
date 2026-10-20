package mailer

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"mime"
	"net"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/aide-family/Alioth/app/alioth/internal/conf"
	"github.com/aide-family/Alioth/pkg/merr"
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

func New(auth *conf.Auth, logger *slog.Logger) *Sender {
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
func (s *Sender) Send(to, subject, body string) error {
	if !s.Configured() {
		return merr.Server("smtp is not configured")
	}
	if strings.ContainsAny(to, "\r\n") || strings.ContainsAny(subject, "\r\n") {
		return merr.Client("invalid email header")
	}
	subject = mime.QEncoding.Encode("utf-8", subject)
	payload := base64.StdEncoding.EncodeToString([]byte(body))
	message := "From: " + s.from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"Content-Transfer-Encoding: base64\r\n\r\n" +
		payload + "\r\n"
	addr := net.JoinHostPort(s.host, strconv.Itoa(s.port))
	var auth smtp.Auth
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}
	if err := smtp.SendMail(addr, auth, s.from, []string{to}, []byte(message)); err != nil {
		s.log.Error(fmt.Sprintf("smtp: %v", err))
		return merr.Server("failed to send email")
	}
	return nil
}
