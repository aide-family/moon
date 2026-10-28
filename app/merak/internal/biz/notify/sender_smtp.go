package notify

import (
	"context"
	"crypto/tls"
	"fmt"
	"mime"
	"net"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/aide-family/moon/app/merak/pkg/bo"
	pkgnotify "github.com/aide-family/moon/pkg/notify"
)

type smtpSender struct{}

func (smtpSender) send(ctx context.Context, task bo.TaskSnapshot, subject, body string) error {
	return sendSMTP(ctx, task, subject, body)
}

func sendSMTP(ctx context.Context, task bo.TaskSnapshot, subject, body string) error {
	host, port, err := net.SplitHostPort(task.ChannelEndpoint)
	if err != nil {
		host = task.ChannelEndpoint
		port = "25"
	}
	addr := net.JoinHostPort(host, port)
	from := task.ChannelFrom
	if from == "" {
		from = task.ChannelUsername
	}
	fromAddr, err := mail.ParseAddress(headerValue(from))
	if err != nil {
		return fmt.Errorf("invalid from address")
	}
	receivers := pkgnotify.SplitReceivers(task.Receivers)
	if len(receivers) == 0 {
		return fmt.Errorf("no receivers")
	}
	toAddrs := make([]string, 0, len(receivers))
	toHeader := make([]string, 0, len(receivers))
	for _, item := range receivers {
		parsed, err := mail.ParseAddress(headerValue(item))
		if err != nil {
			return fmt.Errorf("invalid receiver address")
		}
		toAddrs = append(toAddrs, parsed.Address)
		toHeader = append(toHeader, parsed.String())
	}
	subject = headerValue(subject)
	if !isASCII(subject) {
		subject = mime.QEncoding.Encode("UTF-8", subject)
	}
	message := "From: " + fromAddr.String() + "\r\n" +
		"To: " + strings.Join(toHeader, ", ") + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" + body
	var auth smtp.Auth
	if task.ChannelUsername != "" {
		auth = smtp.PlainAuth("", task.ChannelUsername, task.ChannelPassword, host)
	}
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	var client *smtp.Client
	if port == "465" {
		// 465 是隐式 TLS，先握手再交给 smtp.Client。
		tlsConn := tls.Client(conn, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			_ = conn.Close()
			return err
		}
		client, err = smtp.NewClient(tlsConn, host)
	} else {
		client, err = smtp.NewClient(conn, host)
		if err == nil {
			// 587/25 通常先明文再 STARTTLS；不支持时带密码的 PlainAuth 会自行拒绝明文连接。
			if ok, _ := client.Extension("STARTTLS"); ok {
				err = client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
			}
		}
	}
	if err != nil {
		_ = conn.Close()
		return err
	}
	defer client.Close()
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(fromAddr.Address); err != nil {
		return err
	}
	for _, to := range toAddrs {
		if err := client.Rcpt(to); err != nil {
			return err
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write([]byte(message)); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

// headerValue 去掉邮件头里的换行，防止主题、地址或显示名注入额外头。
func headerValue(value string) string {
	return strings.Map(func(r rune) rune {
		if r == '\r' || r == '\n' {
			return ' '
		}
		return r
	}, strings.TrimSpace(value))
}

func isASCII(value string) bool {
	for _, r := range value {
		if r > 127 {
			return false
		}
	}
	return true
}
