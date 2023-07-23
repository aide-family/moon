# Prometheus-manager
> prometheus 规则和告警统一管理平台

## Init
```bash
# init
make init
```

## dev
```bash
kratos run
# 然后选择运行的服务回车即可
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

