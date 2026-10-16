# 控制面多副本部署

命名空间 `alioth`。应用清单顺序：

```bash
kubectl apply -f namespace.yaml
kubectl apply -f postgres.yaml
kubectl apply -f redis.yaml
kubectl apply -f alioth.yaml
kubectl apply -f megrez.yaml
kubectl apply -f merak.yaml
kubectl apply -f web.yaml
```

## 副本

| 组件 | replicas | 对外 |
| --- | --- | --- |
| web | 2 | LoadBalancer :8080 |
| alioth | 2 | 仅集群内 8000/9000 |
| megrez | 2 | 仅集群内 8001 |
| merak | 2 | 仅集群内 8002 |

PostgreSQL 与 Redis 独立 Deployment；Redis 关闭 RDB/AOF 业务持久化。连接串与密钥放在 Secret，不写进镜像。

## nginx

`web` 只把 `/v1` 转到 `alioth:8000`。配置中不出现 megrez、merak、alkaid。

## 验收要点

- 就绪副本均 ≥ 2。
- 停一个 alioth / megrez / merak 副本后，经 `http://<web>:8080` 的登录、评测、通知仍可用。
- 集群外只能访问 web:8080；8000、8001、8002、8022、9000、9022 不可达。
- 集群内可访问 `megrez:8001/health` 与 `merak:8002/health`。
- 摇光安装见 [../alkaid/README.md](../alkaid/README.md)。
