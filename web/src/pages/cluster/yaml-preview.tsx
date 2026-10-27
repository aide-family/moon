import { Loader2 } from "lucide-react"
import { useEffect, useState } from "react"
import { useTranslation } from "react-i18next"

import { getKubernetesResourceYaml } from "@/api/cluster"
import { CodeBlock } from "@/components/code-block"
import { Button } from "@/components/ui/button"
import { failureMessage } from "@/lib/form-error"

interface ClusterYamlPreviewProps {
  clusterUid: string
  /** Node / Pod / Deployment / … / PersistentVolumeClaim。 */
  kind: string
  /** 集群级资源（Node、PersistentVolume）不传。 */
  namespace?: string
  name: string
}

/**
 * 资源 YAML 预览：懒加载 + 只在代码内容区滚动。
 * 区块本身铺满详情抽屉 YAML Tab 的剩余高度，不再用固定 max-h 把面板撑成小块，
 * 也不让外层抽屉替它承担滚动。
 */
export function ClusterYamlPreview({
  clusterUid,
  kind,
  namespace,
  name,
}: ClusterYamlPreviewProps) {
  const { t } = useTranslation()
  const [yaml, setYaml] = useState("")
  const [redacted, setRedacted] = useState(false)
  const [loading, setLoading] = useState(true)
  const [failed, setFailed] = useState(false)
  const [error, setError] = useState("")
  const [attempt, setAttempt] = useState(0)

  useEffect(() => {
    if (!clusterUid || !kind || !name) {
      return
    }
    let active = true
    async function load() {
      setLoading(true)
      setFailed(false)
      setError("")
      try {
        const result = await getKubernetesResourceYaml(clusterUid, {
          kind,
          namespace: namespace || undefined,
          name,
        })
        if (!active) {
          return
        }
        setYaml(result.yaml ?? "")
        setRedacted(Boolean(result.redacted))
      } catch (err) {
        if (!active) {
          return
        }
        setYaml("")
        setRedacted(false)
        setFailed(true)
        setError(failureMessage(err, ""))
      } finally {
        if (active) {
          setLoading(false)
        }
      }
    }
    void load()
    return () => {
      active = false
    }
  }, [clusterUid, kind, namespace, name, attempt])

  if (loading) {
    return (
      <div
        role="status"
        className="flex min-h-0 flex-1 items-center justify-center gap-2 text-sm text-muted-foreground"
      >
        <Loader2 aria-hidden="true" className="size-4 animate-spin" />
        {t("kube.yamlLoading")}
      </div>
    )
  }

  if (failed) {
    return (
      <div
        role="alert"
        className="flex min-h-0 flex-1 flex-col items-center justify-center gap-3 px-6 text-center"
      >
        <p className="max-w-md min-w-0 text-sm break-words text-destructive">
          {error || t("kube.yamlFailed")}
        </p>
        <Button
          type="button"
          variant="outline"
          onClick={() => setAttempt((value) => value + 1)}
        >
          {t("action.retry")}
        </Button>
      </div>
    )
  }

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-2">
      {redacted ? (
        <p className="shrink-0 text-xs break-words text-muted-foreground">
          {t("kube.yamlRedacted")}
        </p>
      ) : null}
      <CodeBlock
        fill
        language="yaml"
        label={t("kube.yaml")}
        code={yaml}
        copyable
        copyLabel={t("kube.yamlCopy")}
        copiedLabel={t("kube.yamlCopied")}
        copyErrorLabel={t("kube.yamlCopyFailed")}
        ariaLabel={t("kube.yaml")}
      />
    </div>
  )
}
