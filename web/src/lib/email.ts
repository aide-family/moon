const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export function isEmail(value: string): boolean {
  return emailPattern.test(value.trim())
}

// firstInvalidReceiver 返回第一个非法邮箱；全部合法或为空时返回 null。
export function firstInvalidReceiver(raw: string): string | null {
  for (const item of raw.split(/[,;\n]/)) {
    const value = item.trim()
    if (value && !isEmail(value)) {
      return value
    }
  }
  return null
}
