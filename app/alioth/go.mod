module github.com/aide-family/Alioth/app/alioth

go 1.26.0

require (
	entgo.io/ent v0.14.6
	github.com/aide-family/Alioth/api v0.0.0
	github.com/aide-family/Alioth/pkg v0.0.0
	github.com/go-kratos/kratos/v2 v2.9.2
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/google/wire v0.7.0
	github.com/jackc/pgx/v5 v5.11.0
	github.com/spf13/cobra v1.10.2
	golang.org/x/crypto v0.54.0
	golang.org/x/image v0.47.0
	google.golang.org/protobuf v1.36.12
)

require (
	ariga.io/atlas v0.36.2-0.20250730182955-2c6300d0a3e1 // indirect
	buf.build/gen/go/bufbuild/protovalidate/protocolbuffers/go v1.36.12-20260825204119-511051f7f437.2 // indirect
	dario.cat/mergo v1.0.0 // indirect
	github.com/agext/levenshtein v1.2.3 // indirect
	github.com/apparentlymart/go-textseg/v15 v15.0.0 // indirect
	github.com/bmatcuk/doublestar v1.3.4 // indirect
	github.com/go-kratos/aegis v0.2.0 // indirect
	github.com/go-openapi/inflect v0.19.0 // indirect
	github.com/go-playground/form/v4 v4.2.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/gorilla/mux v1.8.1 // indirect
	github.com/hashicorp/hcl/v2 v2.18.1 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	github.com/zclconf/go-cty v1.14.4 // indirect
	github.com/zclconf/go-cty-yaml v1.1.0 // indirect
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sync v0.24.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/text v0.43.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20261012220534-4341edf3f941 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20261005182115-fad411399dd8 // indirect
	google.golang.org/grpc v1.84.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/aide-family/Alioth/api => ../../api

replace github.com/aide-family/Alioth/pkg => ../../pkg
