import type { ReactNode } from "react"

import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"

interface DetailField {
  label: string
  value: ReactNode
  // 占满整行的字段（权限分组这类长内容），默认半行。
  span?: 1 | 2
}

interface DetailSheetProps {
  open: boolean
  title: string
  description?: ReactNode
  fields: DetailField[]
  wide?: boolean
  onOpenChange: (open: boolean) => void
}

export function DetailSheet({
  open,
  title,
  description,
  fields,
  wide = false,
  onOpenChange,
}: DetailSheetProps) {
  const current = { title, description, fields, wide }

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent
        className={
          current.wide
            ? "data-[side=right]:w-full data-[side=right]:sm:max-w-3xl!"
            : undefined
        }
      >
        <SheetHeader>
          <SheetTitle>{current.title}</SheetTitle>
          {current.description ? (
            <SheetDescription>{current.description}</SheetDescription>
          ) : null}
        </SheetHeader>
        <dl
          className={
            current.wide
              ? "grid min-h-0 flex-1 content-start gap-4 overflow-y-auto px-4 pb-4 sm:grid-cols-2"
              : "grid min-h-0 flex-1 content-start gap-4 overflow-y-auto px-4 pb-4"
          }
        >
          {current.fields.map((field) => (
            <div
              key={field.label}
              className={field.span === 2 ? "grid gap-1 sm:col-span-2" : "grid gap-1"}
            >
              <dt className="text-sm text-muted-foreground">{field.label}</dt>
              <dd className="break-all">{field.value || "—"}</dd>
            </div>
          ))}
        </dl>
      </SheetContent>
    </Sheet>
  )
}
