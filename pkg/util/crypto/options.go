package crypto

import (
	"github.com/aide-family/moon/pkg/config"
)

// WithIV sets the IV for the AES cipher.
func WithIV(iv []byte) AesOption {
	return func(a *aesImpl) {
		a.iv = iv
	}
}

// WithKey sets the key for the AES cipher.
func WithKey(key []byte) AesOption {
	return func(a *aesImpl) {
		a.key = key
	}
}

// WithMod sets the AES mode.
func WithMod(mode config.Crypto_AesConfig_MODE) AesOption {
	return func(a *aesImpl) {
		a.mode = mode
	}
}
