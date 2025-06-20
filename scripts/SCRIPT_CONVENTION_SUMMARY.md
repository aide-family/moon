# 脚本命名规则总结

## 完成的工作

根据代码分析和现有脚本文件，我已经完成了以下工作：

### 1. 分析脚本识别规则

通过分析 `pkg/plugin/command/file.go` 中的代码，确定了脚本命名和识别的核心逻辑：

- **文件类型识别**: `GetFileTypeByPrefix(file, index)` - 从文件名中提取解释器类型
- **执行间隔解析**: `GetIntervalByPrefix(file, index)` - 从文件名中解析执行间隔
- **支持的解释器**: `python`, `python3`, `sh`, `bash`
- **时间格式**: 使用Go的 `time.Duration` 格式

### 2. 创建说明文档

#### 中文文档 (`README.md`)
- 详细的命名规则说明
- 组成部分解释
- 命名示例表格
- 支持的脚本类型
- 输出格式要求
- 注意事项和错误处理

#### 英文文档 (`README_EN.md`)
- 完整的英文版本说明
- 包含实现细节和最佳实践
- 与中文版本内容一致

### 3. 验证现有脚本

创建了验证脚本 `validate_naming.py`，验证结果显示所有现有脚本都符合命名规范：

```
✅ 10s_python_hello2.py: Valid
✅ 10s_python3_hello3.py: Valid
✅ 20s_python3.py: Valid
✅ 5s_bash_hello_bash.sh: Valid
✅ 5s_sh_hello_sh.sh: Valid
```

### 4. 创建模板文档

在 `templates/README.md` 中提供了各种脚本语言的模板：
- Python 脚本模板
- Python3 脚本模板
- Bash 脚本模板
- Shell 脚本模板

## 命名规则总结

### 格式
```
{执行间隔}_{解释器类型}_{脚本名称}.{扩展名}
```

### 示例
| 文件名                     | 执行间隔 | 解释器  | 脚本名称       |
| -------------------------- | -------- | ------- | -------------- |
| `5s_bash_hello.sh`         | 5秒      | bash    | hello          |
| `10s_python3_hello3.py`    | 10秒     | python3 | hello3         |
| `30s_sh_system_metrics.sh` | 30秒     | sh      | system_metrics |

### 支持的时间单位
- `ns` - 纳秒
- `us` 或 `µs` - 微秒
- `ms` - 毫秒
- `s` - 秒
- `m` - 分钟
- `h` - 小时

### 支持的解释器
- `python` - Python 2.x
- `python3` - Python 3.x
- `sh` - Shell脚本
- `bash` - Bash脚本

### 输出格式要求
脚本必须输出JSON格式的数据，用于指标收集：

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

## 文件结构

```
scripts/
├── README.md                    # 中文说明文档
├── README_EN.md                 # 英文说明文档
├── SCRIPT_CONVENTION_SUMMARY.md # 本总结文档
├── validate_naming.py           # 命名验证脚本
├── templates/
│   └── README.md               # 脚本模板说明
├── py/
│   └── 20s_python3.py          # 示例脚本
├── 10s_python3_hello3.py       # 示例脚本
├── 10s_python_hello2.py        # 示例脚本
├── 5s_bash_hello_bash.sh       # 示例脚本
└── 5s_sh_hello_sh.sh           # 示例脚本
```

## 使用建议

1. **严格遵循命名规范** - 确保文件名格式正确
2. **使用模板** - 参考 `templates/README.md` 中的模板
3. **验证脚本** - 使用 `validate_naming.py` 验证命名
4. **测试输出** - 确保脚本输出正确的JSON格式
5. **设置权限** - 给脚本文件添加执行权限

## 技术实现

系统通过以下步骤处理脚本：

1. **文件扫描** - 扫描指定目录中的脚本文件
2. **命名解析** - 使用下划线分割文件名，提取间隔和解释器
3. **内容读取** - 读取脚本内容并计算MD5哈希
4. **定时执行** - 根据解析的间隔定时执行脚本
5. **输出处理** - 捕获脚本输出并解析JSON数据 