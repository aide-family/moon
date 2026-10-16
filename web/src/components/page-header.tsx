import type { ReactNode } from "react"

interface PageHeaderProps {
  title: string
  description?: string
  children?: ReactNode
}

export function PageHeader({ title, description, children }: PageHeaderProps) {
  return (
    <div className="flex items-start justify-between gap-4">
      <div>
        <h1 className="font-heading text-2xl font-semibold tracking-tight">{title}</h1>
        {description ? (
          <p className="text-sm text-muted-foreground">{description}</p>
        ) : null}
      </div>
      {children}
    </div>
  )
}
