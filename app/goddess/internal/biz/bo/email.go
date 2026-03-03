package bo

import "net/http"

type SendEmailBo struct {
	To          []string
	Subject     string
	Body        string
	ContentType string
	Cc          []string
	Headers     http.Header
}
