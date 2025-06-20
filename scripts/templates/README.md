# 脚本模板

本目录包含各种脚本语言的模板文件，用于快速创建符合命名规范的脚本。

## 模板文件

### Python 脚本模板

#### `python_template.py`
```python
#!/usr/bin/env python3
"""
Python脚本模板
文件名格式: {interval}_python_{name}.py
示例: 10s_python_my_script.py
"""

import json
import sys

def main():
    """主函数"""
    try:
        # 在这里添加你的脚本逻辑
        result = {
            "vec": {
                "metricType": "METRIC_TYPE_COUNTER",
                "namespace": "moon",
                "subSystem": "laurel",
                "name": "example_metric",
                "labels": ["env"],
                "help": "This is an example metric"
            },
            "data": {
                "metricType": 1,
                "namespace": "moon",
                "subSystem": "laurel",
                "name": "example_metric",
                "labels": {
                    "env": "python_script"
                },
                "value": 1
            }
        }
        
        # 输出JSON格式数据
        print(json.dumps(result))
        return 0
        
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        return 1

if __name__ == "__main__":
    sys.exit(main())
```

#### `python3_template.py`
```python
#!/usr/bin/env python3
"""
Python3脚本模板
文件名格式: {interval}_python3_{name}.py
示例: 10s_python3_my_script.py
"""

import json
import sys

def main():
    """主函数"""
    try:
        # 在这里添加你的脚本逻辑
        result = {
            "vec": {
                "metricType": "METRIC_TYPE_COUNTER",
                "namespace": "moon",
                "subSystem": "laurel",
                "name": "example_metric",
                "labels": ["env"],
                "help": "This is an example metric"
            },
            "data": {
                "metricType": 1,
                "namespace": "moon",
                "subSystem": "laurel",
                "name": "example_metric",
                "labels": {
                    "env": "python3_script"
                },
                "value": 1
            }
        }
        
        # 输出JSON格式数据
        print(json.dumps(result))
        return 0
        
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        return 1

if __name__ == "__main__":
    sys.exit(main())
```

### Shell 脚本模板

#### `bash_template.sh`
```bash
#!/bin/bash
"""
Bash脚本模板
文件名格式: {interval}_bash_{name}.sh
示例: 5s_bash_my_script.sh
"""

# 设置错误处理
set -e

main() {
    # 在这里添加你的脚本逻辑
    
    # 输出JSON格式数据
    cat << EOF
{
  "vec": {
    "metricType": "METRIC_TYPE_COUNTER",
    "namespace": "moon",
    "subSystem": "laurel",
    "name": "example_metric",
    "labels": ["env"],
    "help": "This is an example metric"
  },
  "data": {
    "metricType": 1,
    "namespace": "moon",
    "subSystem": "laurel",
    "name": "example_metric",
    "labels": {
      "env": "bash_script"
    },
    "value": 1
  }
}
EOF
}

# 执行主函数
main "$@"
```

#### `sh_template.sh`
```bash
#!/bin/sh
"""
Shell脚本模板
文件名格式: {interval}_sh_{name}.sh
示例: 5s_sh_my_script.sh
"""

# 设置错误处理
set -e

main() {
    # 在这里添加你的脚本逻辑
    
    # 输出JSON格式数据
    cat << EOF
{
  "vec": {
    "metricType": "METRIC_TYPE_COUNTER",
    "namespace": "moon",
    "subSystem": "laurel",
    "name": "example_metric",
    "labels": ["env"],
    "help": "This is an example metric"
  },
  "data": {
    "metricType": 1,
    "namespace": "moon",
    "subSystem": "laurel",
    "name": "example_metric",
    "labels": {
      "env": "sh_script"
    },
    "value": 1
  }
}
EOF
}

# 执行主函数
main "$@"
```

## 使用方法

1. 复制相应的模板文件
2. 重命名为符合规范的格式：`{interval}_{interpreter}_{name}.{ext}`
3. 修改脚本内容，添加你的业务逻辑
4. 确保输出JSON格式的数据
5. 给脚本文件添加执行权限：`chmod +x script_name`

## 示例

```bash
# 复制Python3模板
cp templates/python3_template.py 30s_python3_system_metrics.py

# 复制Bash模板
cp templates/bash_template.sh 5s_bash_network_check.sh

# 添加执行权限
chmod +x 30s_python3_system_metrics.py
chmod +x 5s_bash_network_check.sh
``` 