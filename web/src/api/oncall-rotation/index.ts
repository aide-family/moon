import { del, get, post, put } from "@/api/http"

import type {
  ListOnCallRotationsParams,
  ListOnCallRotationsResult,
  OnCallRotationItem,
  OnCallRotationScheduleParams,
  OnCallRotationScheduleResult,
  SaveOnCallRotationParams,
} from "./types"

/** GET /v1/oncall-rotations */
export function listOnCallRotations(
  params: ListOnCallRotationsParams = {}
): Promise<ListOnCallRotationsResult> {
  return get<ListOnCallRotationsResult>("/oncall-rotations", params)
}

/** GET /v1/oncall-rotations/{uid} */
export function getOnCallRotation(
  uid: string
): Promise<{ rotation: OnCallRotationItem }> {
  return get<{ rotation: OnCallRotationItem }>(`/oncall-rotations/${uid}`)
}

/** POST /v1/oncall-rotations */
export function createOnCallRotation(
  body: SaveOnCallRotationParams
): Promise<{ rotation: OnCallRotationItem }> {
  return post<{ rotation: OnCallRotationItem }>("/oncall-rotations", body)
}

/** PUT /v1/oncall-rotations/{uid}：成员整体替换，按数组顺序重写 position。 */
export function updateOnCallRotation(
  body: SaveOnCallRotationParams & { uid: string }
): Promise<{ rotation: OnCallRotationItem }> {
  return put<{ rotation: OnCallRotationItem }>(
    `/oncall-rotations/${body.uid}`,
    body
  )
}

/** PUT /v1/oncall-rotations/{uid}/status */
export function updateOnCallRotationStatus(
  uid: string,
  status: string
): Promise<{ rotation: OnCallRotationItem }> {
  return put<{ rotation: OnCallRotationItem }>(
    `/oncall-rotations/${uid}/status`,
    { uid, status }
  )
}

/** DELETE /v1/oncall-rotations/{uid} */
export function deleteOnCallRotation(
  uid: string
): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/oncall-rotations/${uid}`)
}

/** GET /v1/oncall-rotations/{uid}/schedule?days=14 */
export function getOnCallRotationSchedule(
  params: OnCallRotationScheduleParams
): Promise<OnCallRotationScheduleResult> {
  const query = new URLSearchParams()
  if (params.days !== undefined) {
    query.set("days", String(params.days))
  }
  const suffix = query.size > 0 ? `?${query.toString()}` : ""
  return get<OnCallRotationScheduleResult>(
    `/oncall-rotations/${params.uid}/schedule${suffix}`
  )
}

export type {
  ListOnCallRotationsParams,
  ListOnCallRotationsResult,
  OnCallRotationItem,
  OnCallRotationMemberInput,
  OnCallRotationMemberItem,
  OnCallRotationPeriod,
  OnCallRotationScheduleParams,
  OnCallRotationScheduleResult,
  OnCallRotationShiftItem,
  OnCallRotationStatus,
  SaveOnCallRotationParams,
} from "./types"
