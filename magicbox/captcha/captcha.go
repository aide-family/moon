package captcha

import (
	captcha "github.com/mojocn/base64Captcha"
)

// NewCaptcha creates a new captcha instance.
func NewCaptcha(opts ...Option) *captcha.Captcha {
	c := &config{
		driver: captcha.NewDriverDigit(80, 240, 6, 0.7, 80),
		store:  captcha.DefaultMemStore,
	}
	for _, opt := range opts {
		opt(c)
	}
	return captcha.NewCaptcha(c.GetDriver(), c.GetStore())
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
