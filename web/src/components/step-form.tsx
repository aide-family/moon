import type { FormEvent, ReactNode } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { DialogFooter } from "@/components/ui/dialog"
import { cn } from "@/lib/utils"

interface StepFormProps {
  step: number
  stepCount: number
  pending?: boolean
  error?: string
  fieldsClassName?: string
  onBack: () => void
  onNext: () => void
  onSubmit: (event: FormEvent) => void
  children: ReactNode
}

export function StepForm({
  step,
  stepCount,
  pending = false,
  error,
  fieldsClassName,
  onBack,
  onNext,
  onSubmit,
  children,
}: StepFormProps) {
  const { t } = useTranslation()
  const last = step >= stepCount - 1

  function onFormSubmit(event: FormEvent) {
    event.preventDefault()
    if (!last) {
      onNext()
      return
    }
    onSubmit(event)
  }

  return (
    <form className="flex min-h-0 flex-1 flex-col gap-4" onSubmit={onFormSubmit}>
      <p className="text-sm text-muted-foreground">
        {t("form.step", { current: step + 1, total: stepCount })}
      </p>
      <div className={cn("grid min-h-0 flex-1 content-start gap-4 overflow-y-auto", fieldsClassName)}>
        {children}
      </div>
      {error ? <p className="text-sm text-destructive">{error}</p> : null}
      <DialogFooter>
        {step > 0 ? (
          <Button key="back" type="button" variant="outline" onClick={onBack}>
            {t("form.prev")}
          </Button>
        ) : null}
        {last ? (
          <Button key="save" type="submit" disabled={pending}>
            {t("action.save")}
          </Button>
        ) : (
          <Button key="next" type="button" onClick={onNext}>
            {t("form.next")}
          </Button>
        )}
      </DialogFooter>
    </form>
  )
}
