# Prometheus-manager

> prometheus 规则和告警统一管理平台

<h1 align="center" style="border-bottom: none">
    <div>
        <div style="display: flex; align-items: center; justify-content: center; gap: 10px;">
            <img alt="Prometheus" src="doc/img/aide-cloud-logo.png" style="height: 114px; width: 114px; border-radius: 50%;">
            <div>+</div>
            <img alt="Prometheus" src="doc/img/prometheus-logo.svg">
        </div>
        <br>
        Prometheus-manager
    </div>
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

