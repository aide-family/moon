import type { TFunction } from "i18next"

export function enumLabel(t: TFunction, value: string): string {
  if (!value) {
    return ""
  }
  const key = `enum.${value}`
  const label = t(key)
  return label === key ? value : label
}
