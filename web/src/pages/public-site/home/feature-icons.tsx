import {
  Activity,
  Bell,
  Boxes,
  Database,
  Gauge,
  GitBranch,
  Globe,
  Layers,
  Server,
  Shield,
  Terminal,
  Users,
  type LucideIcon,
} from "lucide-react"

/**
 * 能力卡片图标：键名与后端 `siteFeatureIcons` 白名单一一对应。
 *
 * 后台只能选白名单里的键名；这里仍然给未知键名一个兜底图标，避免后端新增键名
 * 时首页整块渲染失败。
 */
const FEATURE_ICONS: Record<string, LucideIcon> = {
  activity: Activity,
  bell: Bell,
  shield: Shield,
  database: Database,
  globe: Globe,
  boxes: Boxes,
  terminal: Terminal,
  gauge: Gauge,
  server: Server,
  users: Users,
  "git-branch": GitBranch,
  layers: Layers,
}

export function FeatureIcon({ iconKey }: { iconKey: string }) {
  const Icon = FEATURE_ICONS[iconKey] ?? Layers
  return <Icon aria-hidden="true" className="size-5" />
}
