import type { FormEvent, ReactNode } from "react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { DialogFooter } from "@/components/ui/dialog"

interface StepFormProps {
  step: number
  stepCount: number
  pending?: boolean
  error?: string
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
  onBack,
  onNext,
  onSubmit,
  children,
}: StepFormProps) {
  const { t } = useTranslation()
  const last = step >= stepCount - 1

  return (
    <form className="flex min-h-0 flex-1 flex-col gap-4" onSubmit={onSubmit}>
      <p className="text-sm text-muted-foreground">
        {t("form.step", { current: step + 1, total: stepCount })}
      </p>
      <div className="grid min-h-0 flex-1 content-start gap-4 overflow-y-auto">
        {children}
      </div>
      {error ? <p className="text-sm text-destructive">{error}</p> : null}
      <DialogFooter>
        {step > 0 ? (
          <Button type="button" variant="outline" onClick={onBack}>
            {t("form.prev")}
          </Button>
        ) : null}
        {last ? (
          <Button type="submit" disabled={pending}>
            {t("action.save")}
          </Button>
        ) : (
          <Button type="button" onClick={onNext}>
            {t("form.next")}
          </Button>
        )}
      </DialogFooter>
    </form>
  )
}
