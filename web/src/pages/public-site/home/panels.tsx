import { Fragment, type ReactNode } from "react"

import { cn } from "@/lib/class-names"

/**
 * 官网装饰性面板的固定外框。
 *
 * 与旧截图槽位同尺寸（16:10 + 22px 圆角 + 1px 描边 + bg-card 衬底 + 8px 内衬），
 * 这样 Hero 与三个解释区块换掉截图后版式不跳动。面板只承载装饰内容，全部
 * `aria-hidden`：信息由旁边的文案承担，屏幕阅读器不会读到重复内容。
 */
export function PanelFrame({
  children,
  className,
}: {
  children: ReactNode
  className?: string
}) {
  return (
    <div
      aria-hidden="true"
      className={cn(
        "min-w-0 overflow-hidden rounded-[22px] border border-border bg-card p-1.5 sm:p-2",
        className
      )}
    >
      <div className="aspect-[16/10] w-full overflow-hidden rounded-[16px] border border-border/60 bg-muted/20">
        {children}
      </div>
    </div>
  )
}

/** 浅色/深色都成立的抽象柱条：一部分走品牌淡色，一部分走中性面。 */
const CHART_BARS = [38, 62, 44, 78, 52, 68, 30, 58, 72]

/** Hero 兜底主视觉：抽象控制台预览，不出现任何数字或真实业务指标。 */
export function ConsolePanel() {
  return (
    <div className="flex h-full flex-col">
      {/* 窗口条：三个圆点 + 假地址栏 */}
      <div className="flex h-6 shrink-0 items-center gap-2 border-b border-border/60 bg-card/60 px-2 sm:h-8 sm:px-3">
        <span className="flex items-center gap-1">
          <span className="size-1.5 rounded-full bg-border sm:size-2" />
          <span className="size-1.5 rounded-full bg-border sm:size-2" />
          <span className="size-1.5 rounded-full bg-border sm:size-2" />
        </span>
        <span className="h-2.5 w-24 max-w-[40%] rounded-[4px] border border-border/60 bg-muted sm:h-3.5" />
        <span className="ml-auto h-2 w-10 rounded-full bg-border/70" />
      </div>

      <div className="flex min-h-0 flex-1">
        {/* 左侧细导航条 */}
        <div className="hidden w-[18%] shrink-0 flex-col gap-2 border-r border-border/60 bg-muted/20 p-2 sm:flex sm:p-2.5">
          <div className="flex items-center gap-1.5">
            <span className="size-4 shrink-0 rounded-[5px] bg-primary-soft" />
            <span className="h-2 w-9 rounded-full bg-border" />
          </div>
          <div className="mt-1 grid gap-1.5">
            <span className="h-2 w-full rounded-full bg-primary-soft" />
            <span className="h-2 w-4/5 rounded-full bg-border/70" />
            <span className="h-2 w-3/5 rounded-full bg-border/70" />
            <span className="h-2 w-4/5 rounded-full bg-border/70" />
            <span className="h-2 w-2/3 rounded-full bg-border/70" />
          </div>
        </div>

        {/* 主区：统计块 + 柱状图 + 列表片段 */}
        <div className="grid min-h-0 flex-1 grid-rows-[auto_minmax(0,1fr)] gap-1.5 p-2 sm:gap-3 sm:p-3">
          <div className="grid grid-cols-3 gap-1.5 sm:gap-2">
            {[0, 1, 2].map((index) => (
              <div
                key={index}
                className="flex min-w-0 flex-col gap-1 rounded-[6px] border border-border/60 bg-card/50 p-1.5 sm:gap-1.5 sm:rounded-[8px] sm:p-2"
              >
                <div className="flex items-center justify-between gap-1">
                  <span className="h-1.5 w-7 rounded-full bg-border sm:h-2 sm:w-9" />
                  <span className="size-2.5 shrink-0 rounded-[3px] bg-primary-soft sm:size-3" />
                </div>
                <span className="h-2.5 w-8 rounded-[3px] bg-border sm:h-3.5 sm:w-10" />
              </div>
            ))}
          </div>

          <div className="grid min-h-0 grid-cols-5 gap-1.5 sm:gap-3">
            <div className="col-span-3 flex min-h-0 flex-col rounded-[6px] border border-border/60 bg-card/50 p-1.5 sm:rounded-[8px] sm:p-2.5">
              <div className="flex items-center justify-between">
                <span className="h-1.5 w-10 rounded-full bg-border sm:h-2 sm:w-14" />
                <span className="h-1.5 w-6 rounded-full bg-muted sm:h-2 sm:w-8" />
              </div>
              <div className="mt-auto flex h-[74%] items-end gap-1 sm:gap-1.5">
                {CHART_BARS.map((height, index) => (
                  <span
                    key={index}
                    className={cn(
                      "w-full rounded-t-[2px]",
                      index % 3 === 1 ? "bg-primary-soft" : "bg-border"
                    )}
                    style={{ height: `${height}%` }}
                  />
                ))}
              </div>
            </div>

            <div className="col-span-2 flex min-h-0 flex-col justify-center gap-1 rounded-[6px] border border-border/60 bg-card/50 p-1.5 sm:gap-1.5 sm:rounded-[8px] sm:p-2.5">
              {[0, 1, 2, 3, 4].map((index) => (
                <div key={index} className="flex min-w-0 items-center gap-1.5">
                  <span className="size-1.5 shrink-0 rounded-full bg-border" />
                  <span className="h-1.5 flex-1 rounded-full bg-border/70" />
                  <span className="h-1.5 w-4 shrink-0 rounded-full bg-border/60" />
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function PanelArrow() {
  return (
    <span className="flex shrink-0 items-center" aria-hidden="true">
      <span className="h-px w-3 bg-border sm:w-5" />
      <span className="size-1.5 rotate-45 border-t border-r border-border" />
    </span>
  )
}

/** 工作原理：三步流程卡 + 连接线，标签复用旁边步骤文案。 */
export function FlowPanel({ steps }: { steps: string[] }) {
  return (
    <div className="flex h-full items-center justify-center gap-1.5 p-2.5 sm:gap-4 sm:p-5">
      {steps.map((step, index) => (
        <Fragment key={step}>
          <div className="flex h-[68%] min-w-0 flex-1 flex-col items-center justify-center gap-2 rounded-[10px] border border-border/60 bg-card/50 px-1.5 text-center sm:gap-3">
            <span className="grid size-6 place-items-center rounded-[7px] bg-primary-soft sm:size-9 sm:rounded-[9px]">
              <span className="size-2 rounded-[2px] bg-primary-ink sm:size-2.5" />
            </span>
            <span className="text-[10px] font-medium text-muted-foreground sm:text-xs">
              {step}
            </span>
            <span className="h-1 w-8 rounded-full bg-border sm:w-10" />
          </div>
          {index < steps.length - 1 ? <PanelArrow /> : null}
        </Fragment>
      ))}
    </div>
  )
}

/** 权限模型：两列小表，列标题复用旁边文案，行内容只用抽象条块。 */
export function PermissionPanel({ labels }: { labels: string[] }) {
  return (
    <div className="grid h-full grid-cols-2 gap-2 p-2.5 sm:gap-4 sm:p-5">
      {labels.map((label) => (
        <div
          key={label}
          className="flex min-h-0 min-w-0 flex-col gap-2 rounded-[10px] border border-border/60 bg-card/50 p-2 sm:gap-3 sm:p-3"
        >
          <span className="truncate text-[10px] font-medium text-muted-foreground sm:text-xs">
            {label}
          </span>
          <div className="grid min-h-0 flex-1 grid-rows-3 gap-1.5 sm:gap-2">
            {[0, 1, 2].map((index) => (
              <div
                key={index}
                className="flex min-w-0 items-center gap-1.5 rounded-[6px] border border-border/50 bg-muted/30 px-1.5 py-1.5 sm:gap-2 sm:px-2 sm:py-2"
              >
                <span className="size-1.5 shrink-0 rounded-full bg-primary-soft" />
                <span className="h-1.5 flex-1 rounded-full bg-border/70" />
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  )
}

/** 私有部署：控制面 / 三个服务 / PostgreSQL 的抽象拓扑，连线只画层级关系。 */
export function DeployPanel() {
  return (
    <div className="relative h-full w-full">
      {/* 层级连线：控制面 → 服务总线 → PostgreSQL */}
      <span className="absolute top-[22%] left-1/2 h-[9%] w-px -translate-x-1/2 bg-border" />
      <span className="absolute top-[31%] right-[19%] left-[19%] h-px bg-border" />
      <span className="absolute top-[31%] left-[19%] h-[5%] w-px bg-border" />
      <span className="absolute top-[31%] left-1/2 h-[5%] w-px -translate-x-1/2 bg-border" />
      <span className="absolute top-[31%] right-[19%] h-[5%] w-px bg-border" />
      <span className="absolute top-[52%] left-[19%] h-[9%] w-px bg-border" />
      <span className="absolute top-[52%] left-1/2 h-[9%] w-px -translate-x-1/2 bg-border" />
      <span className="absolute top-[52%] right-[19%] h-[9%] w-px bg-border" />
      <span className="absolute top-[61%] right-[19%] left-[19%] h-px bg-border" />
      <span className="absolute top-[61%] left-1/2 h-[11%] w-px -translate-x-1/2 bg-border" />

      <Node className="top-[6%] left-[38%] w-[24%]">alioth</Node>
      <Node className="top-[36%] left-[9%] w-[20%]">megrez</Node>
      <Node className="top-[36%] left-[40%] w-[20%]">merak</Node>
      <Node className="top-[36%] left-[71%] w-[20%]">alkaid</Node>
      <Node className="top-[72%] left-[35%] w-[30%]" muted>
        PostgreSQL
      </Node>
    </div>
  )
}

function Node({
  children,
  className,
  muted = false,
}: {
  children: ReactNode
  className?: string
  muted?: boolean
}) {
  return (
    <div
      className={cn(
        "absolute flex h-[16%] min-w-0 items-center justify-center gap-1.5 rounded-[8px] border border-border/60 bg-card/60 px-1",
        className
      )}
    >
      <span
        className={cn(
          "size-1.5 shrink-0 rounded-[2px]",
          muted ? "bg-border" : "bg-primary-soft"
        )}
      />
      <span className="truncate font-mono text-[9px] text-muted-foreground sm:text-[11px]">
        {children}
      </span>
    </div>
  )
}
