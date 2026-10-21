import { get, post } from "@/api/http"

import type {
  AlertEventChart,
  AlertEventItem,
  AlertEventRealtimeStats,
  AlertEventStats,
  AlertEventSuppressResult,
  AlertEventUids,
  ListAlertEventsParams,
  ListAlertEventsResult,
} from "./types"

/** GET /v1/alert-events */
export function listAlertEvents(
  params: ListAlertEventsParams = {}
): Promise<ListAlertEventsResult> {
  return get<ListAlertEventsResult>("/alert-events", params)
}

/** GET /v1/alert-events/{uid} */
export function getAlertEvent(uid: string): Promise<{ event: AlertEventItem }> {
  return get<{ event: AlertEventItem }>(`/alert-events/${uid}`)
}

/** GET /v1/alert-events/{uid}/chart */
export function getAlertEventChart(uid: string): Promise<AlertEventChart> {
  return get<AlertEventChart>(`/alert-events/${uid}/chart`)
}

/** GET /v1/alert-event-stats */
export function alertEventStats(): Promise<AlertEventStats> {
  return get<AlertEventStats>("/alert-event-stats")
}

/** GET /v1/alert-event-realtime-stats */
export function alertEventRealtimeStats(): Promise<AlertEventRealtimeStats> {
  return get<AlertEventRealtimeStats>("/alert-event-realtime-stats")
}

/** POST /v1/alert-events/{uid}/intervene */
export function interveneAlertEvent(
  uid: string
): Promise<Record<string, never>> {
  return post<Record<string, never>>(`/alert-events/${uid}/intervene`, { uid })
}

/** POST /v1/alert-events/intervene */
export function interveneAlertEvents(
  body: AlertEventUids
): Promise<Record<string, never>> {
  return post<Record<string, never>>("/alert-events/intervene", body)
}

/** POST /v1/alert-events/{uid}/suppress */
export function suppressAlertEvent(
  uid: string,
  suppressedSeconds: number,
  permanent = false
): Promise<AlertEventSuppressResult> {
  return post<AlertEventSuppressResult>(`/alert-events/${uid}/suppress`, {
    uid,
    suppressedSeconds,
    permanent,
  })
}

/** POST /v1/alert-events/{uid}/recover */
export function recoverAlertEvent(uid: string): Promise<Record<string, never>> {
  return post<Record<string, never>>(`/alert-events/${uid}/recover`, { uid })
}

/** POST /v1/alert-events/recover */
export function recoverAlertEvents(
  body: AlertEventUids
): Promise<Record<string, never>> {
  return post<Record<string, never>>("/alert-events/recover", body)
}

export type {
  AlertEventChart,
  AlertEventItem,
  AlertEventRealtimeStats,
  AlertEventStats,
  AlertEventSuppressResult,
  AlertEventUids,
  AlertEventView,
  ListAlertEventsParams,
  ListAlertEventsResult,
  RealtimeAlertTabStat,
} from "./types"
