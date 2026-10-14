import { request } from "@/api/http"

export type Health = {
  status: string
}

export async function checkHealth(): Promise<Health> {
  const response = await request("/health")
  if (!response.ok) {
    throw new Error(`health check failed: ${response.status}`)
  }
  return (await response.json()) as Health
}
