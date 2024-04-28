package conf

import (
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/wire"

	"github.com/aide-family/moon/pkg/helper/plog"
)

// ProviderSetConf is conf providers.
var ProviderSetConf = wire.NewSet(
	wire.FieldsOf(new(*Bootstrap), "Server"),
	wire.FieldsOf(new(*Bootstrap), "Data"),
	wire.FieldsOf(new(*Bootstrap), "Env"),
	wire.FieldsOf(new(*Bootstrap), "Log"),
	wire.FieldsOf(new(*Bootstrap), "ApiWhite"),
	wire.FieldsOf(new(*Bootstrap), "Interflow"),
	wire.FieldsOf(new(*Bootstrap), "JWT"),
	wire.Bind(new(plog.Config), new(*Log)),
	LoadConfig,
)

type Before func(bc *Bootstrap) error

func LoadConfig(flagConf *string, before Before) (*Bootstrap, error) {
	if flagConf == nil || *flagConf == "" {
		return nil, errors.NotFound("FLAG_CONFIGS", "config path not found")
	}
	c := config.New(
		config.WithSource(
			file.NewSource(*flagConf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		return nil, err
	}

	var bc Bootstrap
	if err := c.Scan(&bc); err != nil {
		return nil, err
	}

	return &bc, before(&bc)
}
