---
name: alioth-update-api
description: >-
  Syncs Alioth web API clients from the alioth OpenAPI spec. Use when the user
  says 帮我更新 API, 更新 api, or asks to align web/src/api with the backend.
---

# 更新 API

从玉衡的 OpenAPI 同步 `web/src/api`。上游只有 `alioth`，文档为 `{alioth base URL}/doc/swagger/openapi.yaml`。不按 Moon 的 account、marksman、rabbit 拆目录。

## 步骤

1. 读取 OpenAPI 的 `paths` 和 `components.schemas`。
2. 请求 base URL 是 `/v1`。OpenAPI 路径已含 `/v1` 时，函数里的路径去掉这个前缀。
3. 按资源放到 `web/src/api/<资源>/`。每个目录有 `index.ts` 和 `types.ts`。
4. 在 `index.ts` 用统一的 `http.get`、`http.post`、`http.put`、`http.patch`、`http.delete` 导出函数。路径参数用模板字符串，query 和 body 作为后续参数。每个函数上方写一行 JSDoc，包含方法和路径。
5. `types.ts` 用 `interface` 描述对象，枚举与后端一致时导出枚举或字符串联合。命名沿用 `XxxItem`、`ListXxxParams`、`CreateXxxParams`、`UpdateXxxParams`。
6. 在 `web/src/api/index.ts` 导出新模块。
7. 搜索 `web/src/pages` 和 `web/src/components` 里对已变更函数、类型、枚举的引用，一并修改。被删除或重命名的符号不能留下旧引用。

不改 `web/src/api` 的 HTTP 封装签名，除非这一步就是在建立该封装。不在页面文件里补请求函数。

## 示例

```ts
import { http } from '../http'
import type { DatasourceItem, ListDatasourcesParams } from './types'

/** 获取数据源列表 GET /v1/datasources */
export function listDatasources(params?: ListDatasourcesParams) {
  return http.get<DatasourceItem[]>('/datasources', params)
}

export type { DatasourceItem, ListDatasourcesParams } from './types'
```

完成后运行项目的 TypeScript 检查。路径、方法和字段必须与 OpenAPI 一致。
