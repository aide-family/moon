package captcha

import (
	captcha "github.com/mojocn/base64Captcha"
)

// NewCaptcha creates a new captcha instance.
func NewCaptcha(opts ...Option) *captcha.Captcha {
	c := &Config{
		driver: captcha.NewDriverDigit(80, 240, 6, 0.7, 80),
		store:  captcha.DefaultMemStore,
	}
	for _, opt := range opts {
		opt(c)
	}
	return captcha.NewCaptcha(c.GetDriver(), c.GetStore())
}

type Config struct {
	driver captcha.Driver
	store  captcha.Store
}

type Option func(c *Config)

// GetDriver returns the driver of the captcha.
func (c *Config) GetDriver() captcha.Driver {
	return c.driver
}

func (c *Config) GetStore() captcha.Store {
	return c.store
}

// WithDriver sets the driver of the captcha.
func WithDriver(driver captcha.Driver) Option {
	return func(c *Config) {
		c.driver = driver
	}
}

// WithStore sets the store of the captcha.
func WithStore(store captcha.Store) Option {
	return func(c *Config) {
		c.store = store
	}
}
