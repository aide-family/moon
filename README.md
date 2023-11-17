# Prometheus-manager

> prometheus unified rules and alarms management platform

<h1 style="display: flex; align-items: center; justify-content: center; gap: 10px; width: 100%; text-align: center;">
    <img alt="Prometheus" src="doc/img/logo.svg">
    <img alt="Prometheus" src="doc/img/prometheus-logo.svg">
</h1>

## Architecture overview

![Architecture overview](doc/img/Prometheus-manager.png)

## 开发

```bash
# 克隆代码
git clone https://github.com/aide-cloud/prometheus-manager.git

# 进入项目目录
cd prometheus-manager

# 安装依赖
make init

# 启动服务
kratos run
```

## 创建 api

```bash
 kratos proto add api/<module-name>/<version>/<api-name>.proto
```

## 生成 api 文件

```bash
# 生成 api pb
make api

# 生成 service
kratos proto server api/<module-name>/<version>/<api-name>.proto -t apps/<server-app-name>/internal/service

# 生成 config
make config
```

## Docker

```bash
# build
docker build -t <your-docker-image-name> .

# run
docker run --rm -p 8000:8000 -p 9000:9000 -v </path/to/your/configs>:/data/conf <your-docker-image-name>
```


