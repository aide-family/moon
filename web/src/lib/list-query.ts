import { useSearchParams } from "react-router-dom"

export function useListQuery() {
  const [params, setParams] = useSearchParams()
  const page = Number(params.get("page") || "1")
  const pageSize = Number(params.get("pageSize") || "20")
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
