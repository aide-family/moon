import { cn } from "@/lib/class-names"

/**
 * 连接/运行状态圆点。
 *
 * 圆点自身 aria-hidden，必须与旁边的文字（状态文案）成对使用——
 * 只靠颜色表意在色弱和高对比模式下都不可读。
 * 颜色来自状态语义 token，不跟品牌主题走，换主题时“成功/失败”的含义不变。
 */
export function StatusDot({ active }: { active: boolean }) {
  return (
    <span
      className={cn("size-2 shrink-0 rounded-full", active ? "bg-success" : "bg-danger")}
      aria-hidden
    />
  )
}
