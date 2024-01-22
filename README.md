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
# 服务端
make local app=app/prom_server
# 代理端
make local app=app/prom_agent
```

## 运行效果

![策略列表](doc/img/runtime/strategy-list.png)

![策略编辑](doc/img/runtime/update-strategy.png)

![指标编辑](doc/img/runtime/metric-update.png)

![指标列表](doc/img/runtime/metric-list.png)

![指标图表](doc/img/runtime/metric-chart.png)

![实时告警页面](doc/img/runtime/realtime-alarm.png)




