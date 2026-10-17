import { cn } from "@/lib/utils"

export function StatusDot({ active }: { active: boolean }) {
  return (
    <span
      className={cn("size-2 shrink-0 rounded-full", active ? "bg-emerald-500" : "bg-red-500")}
      aria-hidden
    />
  )
}
