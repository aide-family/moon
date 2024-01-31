# Prometheus Manager 技术文档

## 1. 概述

* 项目简介

  Prometheus Manager 是由Aide-Cloud团队开发并维护的一款管理工具，专为简化Prometheus监控系统的运维工作而设计。该工具提供了一站式的解决方案，能够集中管理和配置多个Prometheus实例及其相关的服务发现、规则集和警报通知。

* 主要功能

## 2. 快速开始

### 2.1 系统要求

* 操作系统：Linux、macOS、Windows
* Go语言环境：Go 1.20+
* Docker (可选，用于快速部署)
* K8s (暂时未尝试)
* 环境依赖：
  * mysql数据库：8.0+
  * redis数据库（可选）
  * kafka消息队列（可选）

### 2.2 安装部署

#### 本地开发方式启动

* 准备如下配置文件

```yaml
# app/prom_server/configs_local/config.yaml
env:
  name: prometheus-manager_prom_server
  version: 0.0.1
  # local dev两种模式会自动migrate数据库
  env: local
  metadata:
    description: Prometheus Manager Server APP
    version: 0.0.1
    author: 梧桐
    license: MIT
    email: aidecloud@163.com
    url: https://github.com/aide-cloud/prometheus-manager
    
server:
  http:
    addr: 0.0.0.0:8000
    timeout: 1s
  grpc:
    addr: 0.0.0.0:9000
    timeout: 1s

data:
  database:
    driver: mysql
    # mysql数据库地址，替换为自己的数据库实际连接，并创建prometheus-manager数据库
    source: root:123456@tcp(localhost:3306)/prometheus-manager?charset=utf8mb4&parseTime=True&loc=Local
    debug: true
# 开启redis配置，则使用redis作为缓存组件
#  redis:
#    addr: localhost:6379
#    password: redis#single#test
#    read_timeout: 0.2s
#    write_timeout: 0.2s

apiWhite:
  all:
    - /api.auth.Auth/Login
    - /api.auth.Auth/Captcha
    - /api.interflows.HookInterflow/Receive

  jwtApi:

  rbacApi:
    - /api.auth.Auth/Logout
    - /api.auth.Auth/RefreshToken

log:
  filename: ./log/prometheus-server.log
  level: debug
  encoder: json
  maxSize: 2
  compress: true
  
# 添加mq配置，则会使用mq通信
#mq:
#  kafka:
#    groupId: http://localhost:8001/api/v1/interflows/receive
#    endpoints:
#      - localhost:9092
```

```yaml
# app/prom_agent/configs_local/config.yaml
env:
  name: prometheus-manager_prom_agent
  version: 0.0.1
  env: local
  metadata:
    description: Prometheus Manager Agent APP
    version: 0.0.1
    author: 梧桐
    license: MIT
    email: aidecloud@163.com
    url: https://github.com/aide-cloud/prometheus-manager
server:
  http:
    addr: 0.0.0.0:8001
    timeout: 1s
  grpc:
    addr: 0.0.0.0:9001
    timeout: 1s

# 开启redis配置，则使用redis作为缓存组件
#data:
#  redis:
#    addr: localhost:6379
#    password: redis#single#test
#    read_timeout: 0.2s
#    write_timeout: 0.2s

watchProm:
  interval: 10s

# mq配置
#mq:
#  kafka:
#    groupId: prometheus-agent
#    endpoints:
#      - localhost:9092

# mq替代配置， 二选一
interflow:
  server: http://localhost:8000/api/v1/interflows/receive
  agent: http://localhost:8001/api/v1/interflows/receive
```

* 按顺序执行启动命令

```shell
# 1. 服务端启动
make local app=app/prom_server
# 2. agent启动
make local app=app/prom_agent
# 3. web端启动
make web
```

##### docker方式启动

* 准备上述类似配置

* 执行打包命令

```shell
# 打包服务端
make docker-build app=app/prom_server
# 打包agent
make docker-build app=app/prom_agent
# 打包web
make docker-build-web
```

* 执行启动命令

> 如果配置文件目录不一致， 请对应修改根目录下docker-compose.yaml文件

```shell
# 启动
docker-compose up -d
# 停止
docker-compose down
# 重启
docker-compose restart
```

## 3. 功能详解

### 3.1 系统管理

#### 用户管理

* 功能说明：

主要管理用户信息，包括新增、修改、删除等操作。该系统不提供用户注册功能，用户信息由管理员添加。

* 注意事项：

无

#### 角色管理

* 功能说明：

主要管理角色信息，包括新增、修改、删除等操作。通过权限和角色绑定，实现权限控制，精确到接口粒度。
采用RBAC模式实现，具体请参见[RBAC](../..)。

* 注意事项：

无

#### 权限管理

* 功能说明：

主要管理权限信息，包括新增、修改、删除等操作。这里维护系统全部需要权限控制的接口。系统新增接口， 需要再次录入到权限管理中。

#### 字典管理

* 功能说明：

主要管理字典信息，包括新增、修改、删除等操作。字典信息主要用于系统中需要使用到的枚举值，比如状态、类型、告警等级等。

### 3.2 告警配置

* 功能说明：

主要维护告警规则组、告警规则，通过表单方式维护prometheus规则信息，并支持多数据源场景，我们在配置规则时候，可以选择不同数据源，编写不同的报警规则，完成告警规则配置。
同时，还支持配置报警页面，告警事件发生后，能够把相同类型的各种规则事件归类到同一个报警页面，帮助我们运维同学集中处理告警。

### 3.3 实时告警

* 功能说明：

主要是展示产生的告警数据，并按照不同报警页面分类展示。每一条告警数据除了展示基本信息外， 还可以展示持续时常，支持告警静默、强制删除、告警升级等操作。

### 3.4 历史告警（开发中）

* 功能说明：

主要用于查询历史告警，提供统计数据大盘，为复盘提供数据支撑

### 3.5 告警通知

* 功能说明：

主要提供报警组、报警hook等通信方式维护，为报警策略提供通知对象数据。

## 4. 功能TODO

## 5. 常见问题解答