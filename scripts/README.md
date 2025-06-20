# Script Naming Convention

## Overview

This system supports automatic identification and execution of multiple script languages through specific file naming conventions to configure script execution intervals and interpreter types.

## Naming Convention

### Format
```
{execution_interval}_{interpreter_type}_{script_name}.{extension}
```

### Components

1. **Execution Interval** (0th part)
   - Uses Go's `time.Duration` format
   - Supported units: `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`
   - Examples: `5s`, `10s`, `1m`, `30s`

2. **Interpreter Type** (1st part)
   - `python` - Python 2.x
   - `python3` - Python 3.x
   - `sh` - Shell script
   - `bash` - Bash script

3. **Script Name** (2nd part and beyond)
   - Descriptive name, separated by underscores
   - Examples: `hello_world`, `system_metrics`

4. **File Extension**
   - `.py` - Python scripts
   - `.sh` - Shell/Bash scripts

### Naming Examples

| Filename                      | Execution Interval | Interpreter | Script Name    | Description                               |
| ----------------------------- | ------------------ | ----------- | -------------- | ----------------------------------------- |
| `5s_bash_hello.sh`            | 5 seconds          | bash        | hello          | Bash script, executes every 5 seconds     |
| `10s_python3_hello3.py`       | 10 seconds         | python3     | hello3         | Python3 script, executes every 10 seconds |
| `30s_sh_system_metrics.sh`    | 30 seconds         | sh          | system_metrics | Shell script, executes every 30 seconds   |
| `1m_python_data_collector.py` | 1 minute           | python      | data_collector | Python script, executes every 1 minute    |

## Supported Script Types

### Python Scripts
- **Interpreter**: `python`, `python3`
- **Extension**: `.py`
- **Example**:
  ```python
  #!/usr/bin/env python3
  print('{"metric": "value"}')
  ```

### Shell Scripts
- **Interpreter**: `sh`, `bash`
- **Extension**: `.sh`
- **Example**:
  ```bash
  #!/bin/bash
  echo '{"metric": "value"}'
  ```

## Script Output Format

Scripts should output JSON format data for metric collection:

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

## Important Notes

1. **File naming must strictly follow the format**, otherwise the system cannot identify correctly
2. **Execution interval must use valid time.Duration format**
3. **Interpreter type must be a supported type by the system**
4. **Scripts must output valid JSON format data**
5. **Script files must have execution permissions**

## Error Handling

- If the filename format is incorrect, the script will be ignored
- If the specified interpreter doesn't exist, script execution will fail
- If the script output format is incorrect, data will be discarded

## Implementation Details

The system uses the following logic to parse script files:

1. **File Type Detection**: Extracts the interpreter type from the filename using underscore separation
2. **Interval Parsing**: Converts the interval string to Go's `time.Duration` format
3. **Content Processing**: Reads the script content and calculates MD5 hash for change detection
4. **Execution**: Uses the appropriate interpreter to execute the script and capture output

## Best Practices

1. Use descriptive script names that clearly indicate the purpose
2. Choose appropriate execution intervals based on the metric collection needs
3. Ensure scripts are idempotent and can be safely executed multiple times
4. Include proper error handling in scripts
5. Use consistent JSON output format for all scripts 