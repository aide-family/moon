import { useRef } from "react"

import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"

interface DetailField {
  label: string
  value: string
}

interface DetailSheetProps {
  open: boolean
  title: string
  description?: string
  fields: DetailField[]
  onOpenChange: (open: boolean) => void
}

export function DetailSheet({
  open,
  title,
  description,
  fields,
  onOpenChange,
}: DetailSheetProps) {
  const shown = useRef({ title, description, fields })
  if (open) {
    shown.current = { title, description, fields }
  }
  const current = shown.current

  return (
    <Sheet open={open} onOpenChange={onOpenChange}>
      <SheetContent>
        <SheetHeader>
          <SheetTitle>{current.title}</SheetTitle>
          {current.description ? (
            <SheetDescription>{current.description}</SheetDescription>
          ) : null}
        </SheetHeader>
        <dl className="grid gap-4 px-4">
          {current.fields.map((field) => (
            <div key={field.label} className="grid gap-1">
              <dt className="text-sm text-muted-foreground">{field.label}</dt>
              <dd className="break-all">{field.value || "—"}</dd>
            </div>
          ))}
        </dl>
      </SheetContent>
    </Sheet>
  )
}
