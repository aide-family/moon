module github.com/aide-family/moon

go 1.21

require (
	github.com/aide-cloud/universal v1.2.2
	github.com/casbin/casbin/v2 v2.79.0
	github.com/casbin/gorm-adapter/v3 v3.20.0
	github.com/confluentinc/confluent-kafka-go/v2 v2.3.0
	github.com/envoyproxy/protoc-gen-validate v1.0.2
	github.com/go-kratos/kratos/contrib/registry/etcd/v2 v2.0.0-20231204033633-efac50c3d4c7
	github.com/go-kratos/kratos/v2 v2.7.1
	github.com/golang-jwt/jwt/v4 v4.5.0
	github.com/google/uuid v1.4.0
	github.com/google/wire v0.6.0
	github.com/gorilla/handlers v1.5.2
	github.com/gorilla/websocket v1.5.1
	github.com/mojocn/base64Captcha v1.3.6
	github.com/nutsdb/nutsdb v1.0.3
	github.com/prometheus/client_golang v1.17.0
	github.com/redis/go-redis/v9 v9.3.0
	github.com/spf13/viper v1.17.0
	github.com/stretchr/testify v1.8.4
	go.etcd.io/etcd/client/pkg/v3 v3.5.10
	go.etcd.io/etcd/client/v3 v3.5.10
	go.opentelemetry.io/otel v1.21.0
	go.opentelemetry.io/otel/exporters/jaeger v1.17.0
	go.opentelemetry.io/otel/sdk v1.21.0
	go.uber.org/automaxprocs v1.5.3
	go.uber.org/zap v1.26.0
	golang.org/x/exp v0.0.0-20230905200255-921286631fa9
	golang.org/x/sync v0.6.0
	google.golang.org/genproto/googleapis/api v0.0.0-20231127180814-3a041ad873d4
	google.golang.org/grpc v1.59.0
	google.golang.org/protobuf v1.31.0
	gopkg.in/gomail.v2 v2.0.0-20160411212932-81ebce5c23df
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/driver/mysql v1.5.2
	gorm.io/driver/sqlite v1.4.3
	gorm.io/gorm v1.25.5
	gorm.io/plugin/soft_delete v1.2.1
)

require (
	github.com/antlabs/stl v0.0.1 // indirect
	github.com/antlabs/timer v0.0.11 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bwmarrin/snowflake v0.3.0 // indirect
	github.com/casbin/govaluate v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/felixge/httpsnoop v1.0.3 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/glebarez/go-sqlite v1.20.3 // indirect
	github.com/glebarez/sqlite v1.7.0 // indirect
	github.com/go-kratos/aegis v0.2.0 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/form/v4 v4.2.0 // indirect
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	github.com/gofrs/flock v0.8.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-sql/civil v0.0.0-20220223132316-b832511892a9 // indirect
	github.com/golang-sql/sqlexp v0.1.0 // indirect
	github.com/golang/freetype v0.0.0-20170609003504-e2365dfdc4a0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.13.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.1 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.12.0 // indirect
	github.com/jackc/pgx/v4 v4.17.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mattn/go-sqlite3 v1.14.15 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/microsoft/go-mssqldb v0.17.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pelletier/go-toml/v2 v2.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.4.1-0.20230718164431-9a2bf3000d16 // indirect
	github.com/prometheus/common v0.44.0 // indirect
	github.com/prometheus/procfs v0.11.1 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230126093431-47fa9a501578 // indirect
	github.com/sagikazarmark/locafero v0.3.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.10.0 // indirect
	github.com/spf13/cast v1.5.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tidwall/btree v1.6.0 // indirect
	github.com/xujiajun/mmap-go v1.0.1 // indirect
	github.com/xujiajun/utils v0.0.0-20220904132955-5f7c5b914235 // indirect
	go.etcd.io/etcd/api/v3 v3.5.10 // indirect
	go.opentelemetry.io/otel/metric v1.21.0 // indirect
	go.opentelemetry.io/otel/trace v1.21.0 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/image v0.13.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto v0.0.0-20231120223509-83a465c0220f // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231120223509-83a465c0220f // indirect
	gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gorm.io/driver/postgres v1.4.4 // indirect
	gorm.io/driver/sqlserver v1.4.1 // indirect
	gorm.io/plugin/dbresolver v1.3.0 // indirect
	modernc.org/libc v1.22.2 // indirect
	modernc.org/mathutil v1.5.0 // indirect
	modernc.org/memory v1.5.0 // indirect
	modernc.org/sqlite v1.20.3 // indirect
)
