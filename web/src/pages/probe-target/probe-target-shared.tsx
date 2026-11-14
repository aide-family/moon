import { useTranslation } from "react-i18next"

import type { ProbeResultItem } from "@/api/probe-target/types"

// 探测类型。页面按套餐授权展示可选类型，五种类型的目标写法与判定口径不同。
export const PROBE_HTTP = "PROBE_HTTP"
export const PROBE_TCP = "PROBE_TCP"
export const PROBE_ICMP = "PROBE_ICMP"
export const PROBE_DNS = "PROBE_DNS"
export const PROBE_TLS = "PROBE_TLS"

/**
 * 非 HTTP 探测结果的「探测详情」摘要：DNS 答案与记录数、ICMP 丢包/抖动/RTT 极值、
 * TLS 协议与套件/证书主体。HTTP 保留状态码与证书列，不使用这个单元格。
 */
export function ProbeDetailSummary({
  driver,
  result,
}: {
  driver: string
  result?: ProbeResultItem | null
}) {
  const { t } = useTranslation()
  if (!result) {
    return <>—</>
  }
  if (driver === PROBE_DNS) {
    if (!result.dnsAnswers && !result.answerCount) {
      return <>—</>
    }
    return (
      <span className="break-words">
        {t("probeTarget.detailDns", {
          count: result.answerCount,
          answers: result.dnsAnswers || "—",
        })}
      </span>
    )
  }
  if (driver === PROBE_ICMP) {
    return (
      <span className="break-words tabular-nums">
        {t("probeTarget.detailIcmp", {
          loss: result.packetLoss,
          jitter: result.jitterMs,
          min: result.minRttMs,
          max: result.maxRttMs,
        })}
      </span>
    )
  }
  if (driver === PROBE_TLS) {
    if (!result.tlsVersion && !result.tlsCipher && !result.certSubject) {
      return <>—</>
    }
    return (
      <span className="grid min-w-0 gap-0.5">
        <span className="break-words tabular-nums">
          {t("probeTarget.detailTlsProtocol", {
            version: result.tlsVersion || "—",
            cipher: result.tlsCipher || "—",
          })}
        </span>
        <span className="text-xs break-words text-muted-foreground">
          {result.certSubject
            ? t("probeTarget.detailTlsCert", {
                subject: result.certSubject,
              })
            : "—"}
        </span>
      </span>
    )
  }
  return <>—</>
}
