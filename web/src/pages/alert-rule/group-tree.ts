import type { AlertRuleGroupItem } from "@/api/alert-rule/types"

export interface AlertRuleGroupNode extends AlertRuleGroupItem {
  depth: number
  children: AlertRuleGroupNode[]
}

/** 把平铺的规则组按 parentUid 组装成树；孤儿节点按顶级处理。 */
export function buildAlertRuleGroupTree(
  items: AlertRuleGroupItem[]
): AlertRuleGroupNode[] {
  const nodes = new Map<string, AlertRuleGroupNode>()
  for (const item of items) {
    nodes.set(item.uid, { ...item, depth: 1, children: [] })
  }
  const roots: AlertRuleGroupNode[] = []
  for (const item of items) {
    const node = nodes.get(item.uid)
    if (!node) {
      continue
    }
    const parent = item.parentUid ? nodes.get(item.parentUid) : undefined
    if (parent && parent.uid !== node.uid) {
      parent.children.push(node)
    } else {
      roots.push(node)
    }
  }
  const attachDepth = (list: AlertRuleGroupNode[], depth: number) => {
    for (const node of list) {
      node.depth = depth
      node.children.sort((a, b) => a.name.localeCompare(b.name))
      attachDepth(node.children, depth + 1)
    }
  }
  roots.sort((a, b) => a.name.localeCompare(b.name))
  attachDepth(roots, 1)
  return roots
}

/** 深度优先展开树，父节点排在子节点前。 */
export function flattenAlertRuleGroupTree(
  nodes: AlertRuleGroupNode[]
): AlertRuleGroupNode[] {
  const flattened: AlertRuleGroupNode[] = []
  const walk = (list: AlertRuleGroupNode[]) => {
    for (const node of list) {
      flattened.push(node)
      walk(node.children)
    }
  }
  walk(nodes)
  return flattened
}

/** 返回需要展示的节点：命中节点本身 + 它的所有上级。 */
export function visibleAlertRuleGroupUids(
  nodes: AlertRuleGroupNode[],
  keyword: string
): Set<string> {
  const normalized = keyword.trim().toLowerCase()
  const visible = new Set<string>()
  if (!normalized) {
    for (const node of flattenAlertRuleGroupTree(nodes)) {
      visible.add(node.uid)
    }
    return visible
  }
  const walk = (node: AlertRuleGroupNode): boolean => {
    let matched = node.name.toLowerCase().includes(normalized)
    for (const child of node.children) {
      if (walk(child)) {
        matched = true
      }
    }
    if (matched) {
      visible.add(node.uid)
    }
    return matched
  }
  for (const node of nodes) {
    walk(node)
  }
  return visible
}

/** 某节点及其全部后代的 uid，用于编辑上级时排除会形成环的选项。 */
export function alertRuleGroupDescendantUids(
  nodes: AlertRuleGroupNode[],
  uid: string
): Set<string> {
  const descendants = new Set<string>()
  const find = (list: AlertRuleGroupNode[]): AlertRuleGroupNode | null => {
    for (const node of list) {
      if (node.uid === uid) {
        return node
      }
      const found = find(node.children)
      if (found) {
        return found
      }
    }
    return null
  }
  const walk = (node: AlertRuleGroupNode) => {
    descendants.add(node.uid)
    for (const child of node.children) {
      walk(child)
    }
  }
  const target = find(nodes)
  if (target) {
    walk(target)
  }
  return descendants
}

/** 子树高度（叶子为 1），用于判断移动后是否超过 3 级。 */
export function alertRuleGroupSubtreeHeight(node: AlertRuleGroupNode): number {
  let height = 0
  for (const child of node.children) {
    height = Math.max(height, alertRuleGroupSubtreeHeight(child))
  }
  return height + 1
}

/** 当前节点所在层级；找不到时按顶级处理。 */
export function alertRuleGroupDepth(
  nodes: AlertRuleGroupNode[],
  uid: string
): number {
  const match = flattenAlertRuleGroupTree(nodes).find((node) => node.uid === uid)
  return match?.depth ?? 1
}
