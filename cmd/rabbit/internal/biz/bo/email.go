package bo

import (
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/moon-monitor/moon/pkg/merr"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

func NewSendEmailParams(config EmailConfig, opts ...SendEmailParamsOption) (SendEmailParams, error) {
	if config == nil {
		return nil, merr.ErrorParamsError("No email configuration is available")
	}
	params := &sendEmailParams{
		config:      config,
		Subject:     "Moon Alert",
		ContentType: "text/html",
	}

	for _, opt := range opts {
		if err := opt(params); err != nil {
			return nil, err
		}
	}

	return params, nil
}

type EmailConfig interface {
	GetUser() string
	GetPass() string
	GetHost() string
	GetPort() uint32
	GetEnable() bool
	GetName() string
}

type GetEmailConfigParams struct {
	TeamID             string
	Name               *string
	DefaultEmailConfig EmailConfig
}

type SetEmailConfigParams struct {
	TeamID  string
	Configs []EmailConfig
}

type SendEmailParams interface {
	GetEmails() []string
	GetBody() string
	GetSubject() string
	GetContentType() string
	GetAttachment() string
	GetCc() []string
	GetConfig() EmailConfig
}

type sendEmailParams struct {
	Emails      []string
	Body        string
	Subject     string
	ContentType string
	Attachment  string
	Cc          []string

	config EmailConfig
}

func (s *sendEmailParams) GetEmails() []string {
	if s == nil {
		return nil
	}
	return slices.Unique(s.Emails)
}

func (s *sendEmailParams) GetBody() string {
	if s == nil {
		return ""
	}
	return s.Body
}

func (s *sendEmailParams) GetSubject() string {
	if s == nil {
		return ""
	}
	return s.Subject
}

func (s *sendEmailParams) GetContentType() string {
	if s == nil {
		return ""
	}
	return s.ContentType
}

func (s *sendEmailParams) GetAttachment() string {
	if s == nil {
		return ""
	}
	return s.Attachment
}

func (s *sendEmailParams) GetCc() []string {
	if s == nil {
		return nil
	}
	return slices.Unique(s.Cc)
}

func (s *sendEmailParams) GetConfig() EmailConfig {
	if s == nil {
		return nil
	}
	return s.config
}

type SendEmailParamsOption func(params *sendEmailParams) error

func WithSendEmailParamsOptionEmail(emails ...string) SendEmailParamsOption {
	return func(params *sendEmailParams) error {
		if len(emails) == 0 {
			return merr.ErrorParamsError("email is required").WithMetadata(map[string]string{
				"emails": "emails is required",
			})
		}
		for _, email := range emails {
			if err := validate.CheckEmail(email); err != nil {
				return errors.FromError(err).WithMetadata(map[string]string{
					"email": "email is invalid",
				})
			}
		}

		params.Emails = emails
		return nil
	}
}

func WithSendEmailParamsOptionBody(body string) SendEmailParamsOption {
	return func(params *sendEmailParams) error {
		if body == "" {
			return merr.ErrorParamsError("body is required").WithMetadata(map[string]string{
				"body": "body is required",
			})
		}

		params.Body = body
		return nil
	}
}

func WithSendEmailParamsOptionSubject(subject string) SendEmailParamsOption {
	return func(params *sendEmailParams) error {
		if subject == "" {
			subject = "Moon Alert"
		}

		params.Subject = subject
		return nil
	}
}

func WithSendEmailParamsOptionContentType(contentType string) SendEmailParamsOption {
	return func(params *sendEmailParams) error {
		if contentType == "" {
			contentType = "text/html"
		}

		params.ContentType = contentType
		return nil
	}
}

func WithSendEmailParamsOptionAttachment(attachment string) SendEmailParamsOption {
	return func(params *sendEmailParams) error {
		params.Attachment = attachment
		return nil
	}
}

func WithSendEmailParamsOptionCc(cc ...string) SendEmailParamsOption {
	return func(params *sendEmailParams) error {
		for _, email := range cc {
			if err := validate.CheckEmail(email); err != nil {
				return errors.FromError(err).WithMetadata(map[string]string{
					"cc": "cc email is invalid",
				})
			}
		}

		params.Cc = cc
		return nil
	}
}
