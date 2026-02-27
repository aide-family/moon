package oauth

import (
	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/safety"
)

var globalRegistry = NewRegistry()

// NewRegistry creates a new OAuth2 login function registry.
func NewRegistry() Register {
	return &registry{
		oauth2LoginFuns: safety.NewSyncMap(make(map[config.OAuth2_APP]OAuth2LoginFun)),
	}
}

// RegisterOAuth2LoginFun registers a new OAuth2 login function.
// If the OAuth2 login function is not valid, it will return an error.
func RegisterOAuth2LoginFun(app config.OAuth2_APP, loginFun OAuth2LoginFun) {
	globalRegistry.RegisterOAuth2LoginFun(app, loginFun)
}

// GetOAuth2LoginFun gets the OAuth2 login function from the registry.
// If the OAuth2 login function is not valid, it will return an error.
func GetOAuth2LoginFun(app config.OAuth2_APP) (OAuth2LoginFun, bool) {
	return globalRegistry.GetOAuth2LoginFun(app)
}

// Register is the interface for the OAuth2 login function registry.
type Register interface {
	RegisterOAuth2LoginFun(app config.OAuth2_APP, loginFun OAuth2LoginFun)
	GetOAuth2LoginFun(app config.OAuth2_APP) (OAuth2LoginFun, bool)
}

type registry struct {
	oauth2LoginFuns *safety.SyncMap[config.OAuth2_APP, OAuth2LoginFun]
}

func (r *registry) RegisterOAuth2LoginFun(app config.OAuth2_APP, loginFun OAuth2LoginFun) {
	r.oauth2LoginFuns.Set(app, loginFun)
}

func (r *registry) GetOAuth2LoginFun(app config.OAuth2_APP) (OAuth2LoginFun, bool) {
	return r.oauth2LoginFuns.Get(app)
}
