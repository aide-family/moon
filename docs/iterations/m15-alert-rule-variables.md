# M15 告警规则自定义变量

目标：规则在「摘要 / 描述」这一步可以声明若干自定义查询变量。告警触发时，评测器用主查询命中序列的标签展开变量查询语句，查到的值再填进摘要和描述模板。

依赖：M3（规则与指标评测）、M4（实时/历史告警与通知）。

## 使用场景

磁盘使用率规则只返回 `value = 91.5`，摘要只能写「磁盘使用率超过 90%，当前 91.5%」，看不到剩余容量。配置：

| 变量名 | 查询语句 | 说明 |
| --- | --- | --- |
| `disk_free` | `node_filesystem_avail_bytes{instance="{{ $labels.instance }}",mountpoint="{{ $labels.mountpoint }}"} / 1024 / 1024 / 1024` | 磁盘剩余容量（GiB） |

摘要写成：

```
磁盘使用率超过 90%，当前 {{ $value | printf "%.1f" }}%，剩余 {{ $disk_free | printf "%.1f" }} GiB
```

触发时 `{{ $labels.instance }}`、`{{ $labels.mountpoint }}` 来自主查询命中的那条序列，`{{ $disk_free }}` 是变量查询的结果。

## 数据模型

变量挂在规则上，和摘要、描述同一步编辑，不挂在级别条件上。

```proto
// 规则上的自定义查询变量。
message AlertRuleVariableItem {
  // 模板变量名，^[A-Za-z_][A-Za-z0-9_]{0,31}$。
  string name = 1;
  // 查询模板，可用 {{ $labels.xxx }} 继承主查询标签。
  string expr = 2;
  // 说明，只用于展示。
  string description = 3;
}
```

- `AlertRuleItem.variables`（`AlertRuleVariableItem`）、`CreateAlertRuleRequest.variables`、`UpdateAlertRuleRequest.variables`（`AlertRuleVariableInput`，校验规则相同）。
- 存储：`alert_rule.variables`，JSON 数组文本，默认 `[]`。格式 `[{"name":"disk_free","expr":"...","description":"..."}]`。
- 内部快照：`megrez.v1.RuleSnapshot.variables`（`RuleVariable`），字段与上表一致。
- 上限：每条规则 20 个变量；名字 32 字符；查询语句 4096 字符；说明 200 字符；名字在规则内唯一。
- 保留名（不能再定义）：`labels`、`value`、`externalURL`、`externalLabels`。变量名只能用于模板标识符，因此不允许中文或短横线；中文写进 `description`。

## 求值语义

1. 主查询命中的序列合并规则标签后得到 `labels`，作为命中的事件标签。
2. 只对**摘要或描述里实际引用**的变量求值（模板里出现 `$<name>`，或出现 `.Vars`）。没有引用就不查，避免每条规则每次评测多打若干次即时查询。
3. 变量查询语句先按告警模板展开（数据与摘要模板相同：`$labels`、`$value`、`.Labels` 等），得到具体查询语句。这是「从主查询继承 labels」的入口。
4. 用数据源的即时查询（评测期的 `query` 模板函数同一条通道）执行展开后的语句，取返回的第一条序列：
   - 数值写入模板数据 `Vars`：`map[string]any`，键是变量名；
   - 该序列的标签写入 `VarLabels`：`map[string]map[string]string`；
   - 查询失败、空结果或数据源不支持即时查询时不写这个键，模板里显示为 `<no value>`，不中断评测。
5. 摘要和描述用同一份数据展开，`$<name>` 是注入的快捷变量；也可以用 `{{ index .Vars "name" }}`。需要变量结果标签时用 `{{ index .VarLabels "name" }}`（返回 map），例如 `{{ index (index .VarLabels "disk_free") "mountpoint" }}`。
6. 变量只影响告警文案，不参与 fingerprint、级别判定或恢复。修改变量不会恢复旧事件。

### 查询语言跟随驱动

变量查询语句必须与规则的数据源使用同一种查询语言；界面的示例、占位和说明都按当前驱动切换，不能把一种数据源的语句套到另一种数据源上。

| 驱动 | 变量查询语言 | 标签继承 |
| --- | --- | --- |
| `METRICS_PROMETHEUS` / `METRICS_VICTORIA_METRICS` | PromQL / MetricsQL | 可用 `{{ $labels.xxx }}` |
| `LOGS_LOKI` | LogQL 数值查询 | 可用 `{{ $labels.xxx }}` |
| `LOGS_VLOGS` | LogsQL stats 数值查询 | 可用 `{{ $labels.xxx }}` |
| `LOGS_ELASTICSEARCH` | 查询 DSL 或普通检索语句，按时间桶计数 | 主查询不返回标签，不能继承 |
| `TRACE_JAEGER` | 尚未参与评测 | 变量不会执行 |

### 模板数据

| 写法 | 含义 |
| --- | --- |
| `{{ $disk_free }}` | 变量 `disk_free` 的数值 |
| `{{ $disk_free | humanize }}` | 按 Prometheus 模板函数格式化 |
| `{{ index .Vars "disk_free" }}` | 同上，不使用快捷变量 |
| `{{ index (index .VarLabels "disk_free") "instance" }}` | 变量查询结果的标签 |
| `{{ $labels.instance }}` | 主查询命中的标签（变量查询语句里同样可用） |

保存时的模板校验必须把已声明的变量名注入模板，否则 `{{ $disk_free }}` 会被判为未定义变量。

## 评测实现

- `pkg/alerttmpl`：新增 `RenderWithVars`、`ValidateWithVars`，在固定的 `Defs` 之后追加 `{{$name := index .Vars "name"}}`。原 `Render`/`Validate` 保持不变，通知投递继续用原函数。
- `app/megrez/internal/biz/eval/judge.go`：命中后解析引用变量，渲染变量语句、执行即时查询、组装修正后的模板数据，再展开摘要和描述。变量求值错误只丢弃该变量。
- `app/megrez/internal/data/gateway.go`：把 `RuleVariable` 映射进 `bo.RuleSnapshot.Variables`。
- `app/alioth`：proto/ent/仓储/服务转换保存变量；`alertrule.prepare` 校验名字、数量、长度、唯一性和模板语法；`service.snapshotItem` 把变量下发给天权。

## Web

- 规则编辑第 2 步（摘要、描述）增加「自定义变量」区域：变量名、查询语句（复用模板编辑器，`scope=alert`）、说明，可增删；保存时前端先做同样的名字/数量校验。
- 模板编辑器补全清单增加 `$labels`、`$value` 之外的已声明变量；摘要/描述编辑框把变量名作为 `$name` 提供补全。
- 摘要、描述的提示文案说明变量写法；变量的示例、占位和说明按规则驱动切换（PromQL/MetricsQL、LogQL、LogsQL、Elasticsearch 查询 DSL），链路规则不展示查询示例。
- 规则详情展示变量表（名字、说明、查询语句）。
- 中文与英文语言包同时补齐。

## 验收标准

- [ ] 规则保存后重新打开，变量、摘要、描述的写法和顺序都保持不变；非法名字、重名、空语句、超长、超过 20 个都被拒绝。
- [ ] `{{ $disk_free }}` 不配置变量时保存报「未定义变量」，配置后保存成功。
- [ ] 主查询命中 `instance=a`、`mountpoint=/data` 时，变量查询语句里的 `{{ $labels.instance }}`、`{{ $labels.mountpoint }}` 被替换成 `a`、`/data` 再查询。
- [ ] 变量查询返回 42.5 时，摘要渲染出 `42.5`；`printf "%.1f"`、`humanize` 可用。
- [ ] 变量查询失败、空结果或未引用时不产生告警内容中的评测错误，摘要/描述仍能展开。
- [ ] 未配置变量的旧规则行为完全不变：摘要、描述、fingerprint、级别判定和恢复路径都不受影响。
- [ ] 通知投递使用已经渲染好的摘要/描述，不再展开变量。
- [ ] 变量示例、占位和说明跟随驱动切换：Prometheus/VictoriaMetrics 显示 PromQL/MetricsQL，Loki 显示 LogQL，VictoriaLogs 显示 LogsQL，Elasticsearch 显示查询 DSL；链路规则不显示其他数据源的查询示例。
- [ ] `go test ./...`（`pkg`、`api`、`app/alioth`、`app/megrez`、`app/merak`）、`make lint`、Web 的 typecheck/lint/build 全部通过。
