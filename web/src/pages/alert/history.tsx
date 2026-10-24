import { Columns3 } from "lucide-react"
import { useTranslation } from "react-i18next"

import { Button } from "@/components/ui/button"
import { AlertColumnDialog } from "@/pages/alert/column-dialog"
import {
  DEFAULT_HISTORY_COLUMNS,
  HISTORY_COLUMNS_SCOPE,
  HISTORY_FIELD_COLUMNS,
} from "@/pages/alert/columns"
import { AlertEventList } from "@/pages/alert/list"
import { useAlertColumns } from "@/pages/alert/use-alert-columns"

export function AlertHistoryPage() {
  const { t } = useTranslation()
  const { columns, open, setOpen, pending, error, setError, save } =
    useAlertColumns(HISTORY_COLUMNS_SCOPE, DEFAULT_HISTORY_COLUMNS)

  return (
    <>
      <AlertEventList
        view="HISTORY"
        columns={columns}
        columnsAction={
          <Button
            type="button"
            variant="outline"
            size="icon"
            className="size-8"
            aria-label={t("alertEvent.columns")}
            title={t("alertEvent.columns")}
            onClick={() => {
              setError("")
              setOpen(true)
            }}
          >
            <Columns3 aria-hidden />
          </Button>
        }
      />
      <AlertColumnDialog
        key={open ? "history-columns" : "history-columns-closed"}
        open={open}
        columns={columns}
        fieldColumns={HISTORY_FIELD_COLUMNS}
        pending={pending}
        error={error}
        onSave={(next) => void save(next)}
        onOpenChange={(value) => {
          if (!value) {
            setOpen(false)
            setError("")
          }
        }}
      />
    </>
  )
}
