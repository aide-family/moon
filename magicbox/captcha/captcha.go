// Package captcha provides a captcha service.
package captcha

import (
	"sync"

	captcha "github.com/mojocn/base64Captcha"
)

var (
	global *captcha.Captcha
	once   sync.Once
)

// New creates a new captcha instance.
func New(opts ...Option) *captcha.Captcha {
	c := &config{
		driver: captcha.NewDriverDigit(80, 240, 6, 0.7, 80),
		store:  captcha.DefaultMemStore,
	}
	for _, opt := range opts {
		opt(c)
	}
	return captcha.NewCaptcha(c.GetDriver(), c.GetStore())
}

func Global() *captcha.Captcha {
	once.Do(func() {
		global = New()
	})
	return global
}

type config struct {
	driver captcha.Driver
	store  captcha.Store
}

type Option func(c *config)

// GetDriver returns the driver of the captcha.
func (c *config) GetDriver() captcha.Driver {
	return c.driver
}

func (c *config) GetStore() captcha.Store {
	return c.store
}

// WithDriver sets the driver of the captcha.
func WithDriver(driver captcha.Driver) Option {
	return func(c *config) {
		c.driver = driver
	}
}

// WithStore sets the store of the captcha.
func WithStore(store captcha.Store) Option {
	return func(c *config) {
		c.store = store
	}
}
