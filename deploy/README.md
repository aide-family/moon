# Moon

Moon 是一款集成prometheus系列的监控管理工具，专为简化Prometheus监控系统的运维工作而设计。该工具提供了一站式的解决方案，能够集中管理和配置多个Prometheus实例及其相关的服务发现、规则集和警报通知。

本 chart 使用 [Helm](https://helm.sh) 包管理器实现在[Kubernetes](https://kubernetes.io) (k8s)集群上部署moon

## 安装Chart

注意: 暂时只支持helm3

- 通过源码安装(本示例默认源码安装)

  ```
  kubectl create ns moon
  helm install [RELEASE_NAME] moon/ -n moon 
  ```

- 修改默认参数后, 打包上传到repos后, 例如https://artifacthub.io/, 再行安装

  ```
  helm repo add moon https://xxxx.xxx.xx/moon
  helm repo update
  helm install [RELEASE_NAME] ./moon -n moon
  ```

## 卸载Chart

通过以下命令卸载:

```console
helm delete moon -n moon
```

## 更新Chart

```
helm upgrade [RELEASE_NAME] [CHART] -n moon --install
```

## 配置

- 详情请参考: [values.yaml](./values.yaml)

- 自定义参数:

  ```
  helm install moon moon/ --set=service.port=31008,resources.limits.cpu=300m
  ```