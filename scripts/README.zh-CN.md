# 脚本命名规则说明

## 概述

本系统支持多种脚本语言的自动识别和执行，通过特定的文件命名规则来配置脚本的执行间隔和解释器类型。

## 命名规则

### 格式
```
{执行间隔}_{解释器类型}_{脚本名称}.{扩展名}
```

### 组成部分

1. **执行间隔** (第0个部分)
   - 使用Go的 `time.Duration` 格式
   - 支持的单位：`ns`, `us` (或 `µs`), `ms`, `s`, `m`, `h`
   - 示例：`5s`, `10s`, `1m`, `30s`

2. **解释器类型** (第1个部分)
   - `python` - Python 2.x
   - `python3` - Python 3.x
   - `sh` - Shell脚本
   - `bash` - Bash脚本

3. **脚本名称** (第2个部分及之后)
   - 描述性名称，使用下划线分隔
   - 示例：`hello_world`, `system_metrics`

4. **文件扩展名**
   - `.py` - Python脚本
   - `.sh` - Shell/Bash脚本

### 命名示例

| 文件名                        | 执行间隔 | 解释器  | 脚本名称       | 说明                        |
| ----------------------------- | -------- | ------- | -------------- | --------------------------- |
| `5s_bash_hello.sh`            | 5秒      | bash    | hello          | Bash脚本，每5秒执行一次     |
| `10s_python3_hello3.py`       | 10秒     | python3 | hello3         | Python3脚本，每10秒执行一次 |
| `30s_sh_system_metrics.sh`    | 30秒     | sh      | system_metrics | Shell脚本，每30秒执行一次   |
| `1m_python_data_collector.py` | 1分钟    | python  | data_collector | Python脚本，每1分钟执行一次 |

## 支持的脚本类型

### Python脚本
- **解释器**: `python`, `python3`
- **扩展名**: `.py`
- **示例**:
  ```python
  #!/usr/bin/env python3
  print('{"metric": "value"}')
  ```

### Shell脚本
- **解释器**: `sh`, `bash`
- **扩展名**: `.sh`
- **示例**:
  ```bash
  #!/bin/bash
  echo '{"metric": "value"}'
  ```

## 脚本输出格式

脚本应输出JSON格式的数据，用于指标收集：

```json
{
  "vec": {
    "metricType": "METRIC_TYPE_COUNTER",
    "namespace": "moon",
    "subSystem": "laurel",
    "name": "test",
    "labels": ["env"],
    "help": "this is test metric"
  },
  "data": {
    "metricType": 1,
    "namespace": "moon",
    "subSystem": "laurel",
    "name": "test",
    "labels": {
      "env": "script_name"
    },
    "value": 1
  }
}
```

## 注意事项

1. **文件命名必须严格遵循格式**，否则系统无法正确识别
2. **执行间隔必须使用有效的time.Duration格式**
3. **解释器类型必须是系统支持的类型**
4. **脚本必须输出有效的JSON格式数据**
5. **脚本文件必须具有执行权限**

## 错误处理

- 如果文件名格式不正确，脚本将被忽略
- 如果指定的解释器不存在，脚本执行将失败
- 如果脚本输出格式不正确，数据将被丢弃

