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
  const type = params.get("type") || ""
  const app = params.get("app") || ""
  const reviewStatus = params.get("reviewStatus") || ""

  function update(next: {
    page?: number
    name?: string
    type?: string
    app?: string
    reviewStatus?: string
  }) {
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
    if (next.type !== undefined) {
      if (next.type) {
        query.set("type", next.type)
      } else {
        query.delete("type")
      }
      query.set("page", "1")
    }
    if (next.app !== undefined) {
      if (next.app) {
        query.set("app", next.app)
      } else {
        query.delete("app")
      }
      query.set("page", "1")
    }
    if (next.reviewStatus !== undefined) {
      if (next.reviewStatus) {
        query.set("reviewStatus", next.reviewStatus)
      } else {
        query.delete("reviewStatus")
      }
      query.set("page", "1")
    }
    query.set("pageSize", String(pageSize))
    setParams(query)
  }

  return { page, pageSize, name, type, app, reviewStatus, update }
}
