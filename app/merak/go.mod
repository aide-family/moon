module github.com/aide-family/Alioth/app/merak

go 1.26.0

require (
	github.com/aide-family/Alioth/api v0.0.0-00010101000000-000000000000
	github.com/aide-family/Alioth/pkg v0.0.0
	github.com/go-kratos/kratos/v2 v2.9.2
	github.com/google/wire v0.7.0
	github.com/redis/go-redis/v9 v9.23.0
	github.com/spf13/cobra v1.10.2
	google.golang.org/grpc v1.84.0
	google.golang.org/protobuf v1.36.12
)

require (
	dario.cat/mergo v1.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/go-kratos/aegis v0.2.0 // indirect
	github.com/go-playground/form/v4 v4.2.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	go.uber.org/atomic v1.12.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.48.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20261012220534-4341edf3f941 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261005182115-fad411399dd8 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/aide-family/Alioth/pkg => ../../pkg

replace github.com/aide-family/Alioth/api => ../../api
