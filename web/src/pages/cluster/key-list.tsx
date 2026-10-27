import { DetailTagList } from "@/components/detail-values"

interface KubernetesKeyListProps {
  keys: string[]
}

/**
 * ConfigMap / Secret 的键名列表。键名可能很多，用可换行的标签展示，
 * 点击标签复制键名；永不渲染键值，避免把配置内容或密钥带进页面。
 */
export function KubernetesKeyList({ keys }: KubernetesKeyListProps) {
  return <DetailTagList items={keys} copyable />
}
