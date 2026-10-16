/** Return an error message when value is blank after trim; otherwise null. */
export function requireTrimmed(value: string, message: string): string | null {
  if (value.trim() === "") {
    return message
  }
  return null
}

/** Return an error message when value is empty or the sentinel NONE; otherwise null. */
export function requireSelected(value: string, message: string): string | null {
  if (value.trim() === "" || value === "NONE") {
    return message
  }
  return null
}

/** First non-null message from checks, or null if all pass. */
export function firstError(...checks: Array<string | null>): string | null {
  for (const check of checks) {
    if (check) {
      return check
    }
  }
  return null
}
