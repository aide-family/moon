import type { QueryDialect, QuerySnapshot } from "@/lib/query-store"

import { DEFAULT_ES_QUERY } from "@/pages/datasource/query-dialect"

export interface BuiltinTemplate {
  id: string
  /** 文案走 i18n；语句本身是字面量，必须可直接执行，不能随语言变化。 */
  nameKey: string
  snapshot: Partial<QuerySnapshot>
}

// 内置模板按方言共享：同一种方言的所有数据源看到同一份目录。
export const BUILTIN_QUERY_TEMPLATES: Record<QueryDialect, BuiltinTemplate[]> = {
  promql: [
    {
      id: "promql-cpu",
      nameKey: "datasource.template.cpu",
      snapshot: {
        expr: '100 - avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100',
      },
    },
    {
      id: "promql-memory",
      nameKey: "datasource.template.memory",
      snapshot: {
        expr: "100 * (1 - node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)",
      },
    },
    {
      id: "promql-disk",
      nameKey: "datasource.template.disk",
      snapshot: {
        expr: '100 * (1 - node_filesystem_avail_bytes{fstype!~"tmpfs|overlay"} / node_filesystem_size_bytes{fstype!~"tmpfs|overlay"})',
      },
    },
    {
      id: "promql-error-rate",
      nameKey: "datasource.template.errorRate",
      snapshot: {
        expr: 'sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))',
      },
    },
    {
      id: "promql-up",
      nameKey: "datasource.template.up",
      snapshot: { expr: "up" },
    },
  ],
  logql: [
    {
      id: "logql-all",
      nameKey: "datasource.template.all",
      snapshot: { expr: '{job=~".+"}' },
    },
    {
      id: "logql-error",
      nameKey: "datasource.template.error",
      snapshot: { expr: '{job=~".+"} |= "error"' },
    },
    {
      id: "logql-timeout",
      nameKey: "datasource.template.timeout",
      snapshot: { expr: '{job=~".+"} |= "timeout"' },
    },
    {
      id: "logql-count",
      nameKey: "datasource.template.count",
      snapshot: {
        expr: 'sum by (app) (count_over_time({job=~".+"} [5m]))',
      },
    },
  ],
  logsql: [
    {
      id: "logsql-all",
      nameKey: "datasource.template.all",
      snapshot: { expr: "*" },
    },
    {
      id: "logsql-error",
      nameKey: "datasource.template.error",
      snapshot: { expr: "level:error" },
    },
    {
      id: "logsql-warn",
      nameKey: "datasource.template.warn",
      snapshot: { expr: "level:warn" },
    },
    {
      id: "logsql-count",
      nameKey: "datasource.template.count",
      snapshot: { expr: "* | stats by (app) count() as value" },
    },
  ],
  dsl: [
    {
      id: "dsl-all",
      nameKey: "datasource.template.dslAll",
      snapshot: { expr: DEFAULT_ES_QUERY },
    },
    {
      id: "dsl-error",
      nameKey: "datasource.template.dslError",
      snapshot: { expr: '{\n  "match": {\n    "level": "error"\n  }\n}' },
    },
  ],
  http: [
    {
      id: "http-root",
      nameKey: "datasource.template.root",
      snapshot: { method: "GET", path: "/" },
    },
    {
      id: "http-health",
      nameKey: "datasource.template.health",
      snapshot: { method: "GET", path: "/health" },
    },
    {
      id: "http-metrics",
      nameKey: "datasource.template.metrics",
      snapshot: { method: "GET", path: "/metrics" },
    },
  ],
}
