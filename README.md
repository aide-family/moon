# Prometheus-manager

> prometheus unified rules and alarms management platform

<h1 style="display: flex; align-items: center; justify-content: center; gap: 10px; width: 100%; text-align: center;">
    <img alt="Prometheus" src="doc/img/logo.svg">
    <img alt="Prometheus" src="doc/img/prometheus-logo.svg">
</h1>

## Architecture overview

![Architecture overview](doc/img/Prometheus-manager.png)

## Init

```bash
# init
make init
```

## dev

```bash
kratos run
```

## add api

```bash
 kratos proto add api/<module-name>/<version>/<api-name>.proto
```

## generate code

```bash
# generate api pb
make api

# generate service
kratos proto server api/<module-name>/<version>/<api-name>.proto -t apps/<server-app-name>/internal/service

# generate config
make config
```

## Docker

```bash
# build
docker build -t <your-docker-image-name> .

# run
docker run --rm -p 8000:8000 -p 9000:9000 -v </path/to/your/configs>:/data/conf <your-docker-image-name>
```

## Catalog

```bash
.
├── Dockerfile
├── LICENSE
├── Makefile
├── README.md
├── api
│   ├── base.pb.go
│   ├── base.pb.validate.go
│   ├── base.proto
│   ├── helloworld
│   │   └── v1
│   │       ├── error_reason.pb.go
│   │       ├── error_reason.pb.validate.go
│   │       ├── error_reason.proto
│   │       ├── error_reason_errors.pb.go
│   │       ├── greeter.pb.go
│   │       ├── greeter.pb.validate.go
│   │       ├── greeter.proto
│   │       ├── greeter_grpc.pb.go
│   │       └── greeter_http.pb.go
│   ├── perrors
│   │   ├── error.pb.go
│   │   ├── error.pb.validate.go
│   │   ├── error.proto
│   │   └── error_errors.pb.go
│   ├── ping.pb.go
│   ├── ping.pb.validate.go
│   ├── ping.proto
│   ├── ping_grpc.pb.go
│   ├── ping_http.pb.go
│   ├── prom
│   │   ├── node.pb.go
│   │   ├── node.pb.validate.go
│   │   ├── node.proto
│   │   └── v1
│   │       ├── dir.pb.go
│   │       ├── dir.pb.validate.go
│   │       ├── dir.proto
│   │       ├── dir_grpc.pb.go
│   │       ├── dir_http.pb.go
│   │       ├── file.pb.go
│   │       ├── file.pb.validate.go
│   │       ├── file.proto
│   │       ├── file_grpc.pb.go
│   │       ├── file_http.pb.go
│   │       ├── group.pb.go
│   │       ├── group.pb.validate.go
│   │       ├── group.proto
│   │       ├── group_grpc.pb.go
│   │       ├── group_http.pb.go
│   │       ├── node.pb.go
│   │       ├── node.pb.validate.go
│   │       ├── node.proto
│   │       ├── node_grpc.pb.go
│   │       ├── node_http.pb.go
│   │       ├── rule.pb.go
│   │       ├── rule.pb.validate.go
│   │       ├── rule.proto
│   │       ├── rule_grpc.pb.go
│   │       └── rule_http.pb.go
│   └── strategy
│       ├── strategy.pb.go
│       ├── strategy.pb.validate.go
│       ├── strategy.proto
│       └── v1
│           ├── crud.pb.go
│           ├── crud.pb.validate.go
│           ├── crud.proto
│           ├── crud_grpc.pb.go
│           ├── crud_http.pb.go
│           ├── load
│           │   ├── load.pb.go
│           │   ├── load.pb.validate.go
│           │   ├── load.proto
│           │   ├── load_grpc.pb.go
│           │   └── load_http.pb.go
│           ├── pull
│           │   ├── pull.pb.go
│           │   ├── pull.pb.validate.go
│           │   ├── pull.proto
│           │   ├── pull_grpc.pb.go
│           │   └── pull_http.pb.go
│           └── push
│               ├── push.pb.go
│               ├── push.pb.validate.go
│               ├── push.proto
│               ├── push_grpc.pb.go
│               └── push_http.pb.go
├── apps
│   ├── master
│   │   ├── Dockerfile
│   │   ├── Makefile
│   │   ├── cmd
│   │   │   └── master
│   │   │       ├── main.go
│   │   │       ├── wire.go
│   │   │       └── wire_gen.go
│   │   ├── configs
│   │   │   ├── config.yaml
│   │   │   └── local.yaml
│   │   └── internal
│   │       ├── biz
│   │       │   ├── biz.go
│   │       │   ├── crud.go
│   │       │   ├── ping.go
│   │       │   └── prom
│   │       │       └── v1
│   │       │           ├── dir.go
│   │       │           ├── file.go
│   │       │           ├── group.go
│   │       │           ├── node.go
│   │       │           ├── rule.go
│   │       │           └── v1.go
│   │       ├── conf
│   │       │   ├── conf.go
│   │       │   ├── conf.pb.go
│   │       │   └── conf.proto
│   │       ├── data
│   │       │   ├── crud.go
│   │       │   ├── data.go
│   │       │   ├── dir.go
│   │       │   ├── file.go
│   │       │   ├── group.go
│   │       │   ├── node.go
│   │       │   ├── ping.go
│   │       │   └── rule.go
│   │       ├── server
│   │       │   ├── grpc.go
│   │       │   ├── http.go
│   │       │   └── server.go
│   │       └── service
│   │           ├── crud.go
│   │           ├── ping.go
│   │           ├── prom
│   │           │   └── v1
│   │           │       ├── dir.go
│   │           │       ├── file.go
│   │           │       ├── group.go
│   │           │       ├── node.go
│   │           │       └── rule.go
│   │           └── service.go
│   └── node
│       ├── Dockerfile
│       ├── Makefile
│       ├── cmd
│       │   └── node
│       │       ├── main.go
│       │       ├── wire.go
│       │       └── wire_gen.go
│       ├── configs
│       │   ├── config.yaml
│       │   ├── local.yaml
│       │   └── rule_temp
│       │       ├── apache
│       │       │   └── apache_rules.yaml
│       │       ├── core_dns
│       │       │   ├── core_dns_rules.yaml
│       │       │   ├── etcd_rules.yaml
│       │       │   ├── jvm_rules.yaml
│       │       │   ├── kafka_rules.yaml
│       │       │   ├── kubernetes_rules.yaml
│       │       │   ├── mysql_rules.yaml
│       │       │   ├── nginx_rules.yaml
│       │       │   ├── php_rules.yaml
│       │       │   ├── prometheus_rules.yaml
│       │       │   ├── redis_rules.yaml
│       │       │   ├── ssl_rules.yaml
│       │       │   ├── test_rules.yaml
│       │       │   ├── thanos_rules.yaml
│       │       │   ├── traefik_rules.yaml
│       │       │   └── xx.yaml
│       │       ├── xx.yaml
│       │       └── xx_rules.yaml
│       └── internal
│           ├── biz
│           │   ├── biz.go
│           │   ├── load.go
│           │   ├── ping.go
│           │   ├── pull.go
│           │   ├── push.go
│           │   └── v1.go
│           ├── conf
│           │   ├── conf.go
│           │   ├── conf.pb.go
│           │   └── conf.proto
│           ├── data
│           │   ├── data.go
│           │   ├── load.go
│           │   ├── ping.go
│           │   ├── pull.go
│           │   └── push.go
│           ├── server
│           │   ├── grpc.go
│           │   ├── http.go
│           │   ├── server.go
│           │   └── timer.go
│           └── service
│               ├── load.go
│               ├── ping.go
│               ├── pull.go
│               ├── push.go
│               └── service.go
├── bin
│   ├── master
│   └── node
├── dal
│   ├── model
│   │   ├── prom_combo_strategies.gen.go
│   │   ├── prom_combos.gen.go
│   │   ├── prom_node_dir_file_group_strategies.gen.go
│   │   ├── prom_node_dir_file_groups.gen.go
│   │   ├── prom_node_dir_files.gen.go
│   │   ├── prom_node_dirs.gen.go
│   │   ├── prom_nodes.gen.go
│   │   └── prom_rules.gen.go
│   └── query
│       ├── gen.go
│       ├── gen_test.db
│       ├── gen_test.go
│       ├── prom_combos.gen.go
│       ├── prom_combos.gen_test.go
│       ├── prom_node_dir_file_group_strategies.gen.go
│       ├── prom_node_dir_file_group_strategies.gen_test.go
│       ├── prom_node_dir_file_groups.gen.go
│       ├── prom_node_dir_file_groups.gen_test.go
│       ├── prom_node_dir_files.gen.go
│       ├── prom_node_dir_files.gen_test.go
│       ├── prom_node_dirs.gen.go
│       ├── prom_node_dirs.gen_test.go
│       ├── prom_nodes.gen.go
│       ├── prom_nodes.gen_test.go
│       ├── prom_rules.gen.go
│       └── prom_rules.gen_test.go
├── deploy
│   └── sql
│       └── init.sql
├── doc
│   └── img
│       ├── Prometheus-manager.png
│       ├── aide-cloud-logo.png
│       ├── logo.svg
│       └── prometheus-logo.svg
├── gen
│   ├── dsn.yml
│   ├── main.go
│   └── strategy
│       └── strategy.go
├── go.mod
├── go.sum
├── openapi.yaml
├── pkg
│   ├── conn
│   │   ├── mysql.go
│   │   ├── redis.go
│   │   └── trace.go
│   ├── curl
│   │   └── curl.go
│   ├── hello
│   │   └── init.go
│   ├── middler
│   │   ├── cors.go
│   │   └── ip.go
│   ├── models
│   ├── prom
│   │   └── metrics.go
│   ├── runtimehelper
│   │   └── recover.go
│   ├── servers
│   │   └── timer.go
│   └── util
│       ├── dir
│       │   └── dir.go
│       ├── strategyload
│       │   └── load.go
│       └── strategystore
│           └── store.go
└── third_party
    ├── README.md
    ├── errors
    │   └── errors.proto
    ├── google
    │   ├── api
    │   │   ├── annotations.proto
    │   │   ├── client.proto
    │   │   ├── field_behavior.proto
    │   │   ├── http.proto
    │   │   └── httpbody.proto
    │   └── protobuf
    │       ├── any.proto
    │       ├── api.proto
    │       ├── compiler
    │       │   └── plugin.proto
    │       ├── descriptor.proto
    │       ├── duration.proto
    │       ├── empty.proto
    │       ├── field_mask.proto
    │       ├── source_context.proto
    │       ├── struct.proto
    │       ├── timestamp.proto
    │       ├── type.proto
    │       └── wrappers.proto
    ├── openapi
    │   └── v3
    │       ├── annotations.proto
    │       └── openapi.proto
    └── validate
        ├── README.md
        └── validate.proto

```

