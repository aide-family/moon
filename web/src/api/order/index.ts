// 租户购买流程：公开套餐、下单、虚拟支付和订单查询。
//
// 接口契约（identity.proto → HTTP JSON，后端统一输出 camelCase / json_name）：
// - GET  /v1/packages                        登录即可，返回公开且启用的套餐
// - POST /v1/orders                          创建待支付订单
// - POST /v1/orders/{uid}/pay                虚拟支付：创建租户并绑定套餐
// - POST /v1/orders/{uid}/cancel             取消待支付订单
// - GET  /v1/orders?page&pageSize&status     当前用户自己的订单
// - GET  /v1/system/namespace/{uid}/orders   平台按租户查订单
import { get, post, type RequestOptions } from "@/api/http"

import type {
  CreateOrderParams,
  ListOrdersParams,
  ListOrdersResult,
  OrderItem,
  PayOrderResult,
  PublicPackageItem,
} from "./types"

/** GET /v1/packages。创建租户时可选的功能套餐，只包含公开且启用的套餐。 */
export function listPublicPackages(): Promise<{ items?: PublicPackageItem[] }> {
  return get("/packages")
}

/** POST /v1/orders。下单只记录套餐快照，不代表已付款。 */
export function createOrder(
  body: CreateOrderParams
): Promise<{ order: OrderItem }> {
  return post("/orders", body)
}

/** POST /v1/orders/{uid}/pay。虚拟支付：成功后创建租户、绑定套餐并返回租户。 */
export function payOrder(uid: string): Promise<PayOrderResult> {
  return post(`/orders/${uid}/pay`)
}

/** POST /v1/orders/{uid}/cancel。只能取消自己的待支付订单。 */
export function cancelOrder(uid: string): Promise<{ order: OrderItem }> {
  return post(`/orders/${uid}/cancel`)
}

/** GET /v1/orders。当前用户自己的订单，按创建时间倒序。 */
export function listOrders(
  params: ListOrdersParams = {},
  options?: RequestOptions
): Promise<ListOrdersResult> {
  return get("/orders", params, undefined, options)
}

/** GET /v1/system/namespace/{uid}/orders。需要 system:namespace:read。 */
export function listSystemNamespaceOrders(
  namespaceUid: string,
  params: ListOrdersParams = {},
  options?: RequestOptions
): Promise<ListOrdersResult> {
  return get(
    `/system/namespace/${namespaceUid}/orders`,
    params,
    undefined,
    options
  )
}
