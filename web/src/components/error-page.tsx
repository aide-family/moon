import type { ReactNode } from "react"

interface ErrorPageProps {
  status: number
  label: string
  title: string
  description: string
  image: string
  detail?: string
  /** 整页错误（502/503）不进入应用外壳；403/404 在侧栏内容区展示。 */
  fullPage?: boolean
  children?: ReactNode
}

/** 403/404/502/503 共用的错误页版式：插画 + 状态码 + 文案 + 操作。 */
export function ErrorPage({
  status,
  label,
  title,
  description,
  image,
  detail,
  fullPage = false,
  children,
}: ErrorPageProps) {
  const content = (
    <div className="grid w-full items-center gap-8 md:grid-cols-[minmax(0,1fr)_minmax(0,1.1fr)]">
      <img src={image} alt="" className="mx-auto w-full max-w-xs md:max-w-sm" />
      <div className="text-center md:text-left">
        <p className="text-sm font-semibold tracking-[0.18em] text-primary-ink">
          {status}
          <span className="ml-2 font-normal tracking-normal text-muted-foreground">
            {label}
          </span>
        </p>
        <h1 className="mt-2 font-heading text-2xl font-semibold tracking-tight sm:text-3xl">
          {title}
        </h1>
        <p className="mt-3 text-sm leading-6 text-muted-foreground">
          {description}
        </p>
        {detail ? (
          <code className="mt-4 block max-w-xl rounded-md border bg-muted/50 px-3 py-2 text-left font-mono text-xs break-all text-muted-foreground">
            {detail}
          </code>
        ) : null}
        {children ? (
          <div className="mt-6 flex flex-wrap justify-center gap-2 md:justify-start">
            {children}
          </div>
        ) : null}
      </div>
    </div>
  )

  if (fullPage) {
    return (
      <main className="flex min-h-svh flex-col bg-background px-6 py-16">
        <div className="m-auto w-full max-w-4xl">{content}</div>
      </main>
    )
  }

  return (
    <section className="flex min-h-0 flex-1 flex-col overflow-y-auto p-4">
      <div className="m-auto w-full max-w-4xl">{content}</div>
    </section>
  )
}
