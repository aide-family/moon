import { useSearchParams } from "react-router-dom"

function boundedInt(value: string | null, fallback: number, min: number, max: number): number {
  const parsed = Number(value)
  if (!Number.isInteger(parsed) || parsed < min || parsed > max) {
    return fallback
  }
  return parsed
}

export function useListQuery() {
  const [params, setParams] = useSearchParams()
  const page = boundedInt(params.get("page"), 1, 1, 1_000_000)
  const pageSize = boundedInt(params.get("pageSize"), 20, 1, 100)
  const name = params.get("name") || ""

  function update(next: { page?: number; name?: string }) {
    const query = new URLSearchParams(params)
    if (next.page !== undefined) {
      query.set("page", String(next.page))
    }
    if (next.name !== undefined) {
      if (next.name) {
        query.set("name", next.name)
      } else {
        query.delete("name")
      }
      query.set("page", "1")
    }
    query.set("pageSize", String(pageSize))
    setParams(query)
  }

  return { page, pageSize, name, update }
}
