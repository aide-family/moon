import { Database } from "lucide-react"

import elasticsearch from "@/assets/brands/elasticsearch.svg"
import jaeger from "@/assets/brands/jaeger.svg"
import loki from "@/assets/brands/loki.svg"
import prometheus from "@/assets/brands/prometheus.svg"
import victoriametrics from "@/assets/brands/victoriametrics.svg"
import vlogs from "@/assets/brands/vlogs.svg"
import { cn } from "@/lib/utils"

const brandLogos: Record<string, string> = {
  METRICS_PROMETHEUS: prometheus,
  METRICS_VICTORIA_METRICS: victoriametrics,
  LOGS_ELASTICSEARCH: elasticsearch,
  LOGS_LOKI: loki,
  LOGS_VLOGS: vlogs,
  TRACE_JAEGER: jaeger,
}

export function BrandIcon({
  driver,
  className,
}: {
  driver: string
  className?: string
}) {
  const logo = brandLogos[driver]
  return (
    <span
      className={cn(
        "inline-flex shrink-0 items-center justify-center",
        className
      )}
      aria-hidden
    >
      {logo ? (
        <img src={logo} alt="" className="size-full object-contain" />
      ) : (
        <Database className="size-full text-muted-foreground" />
      )}
    </span>
  )
}
