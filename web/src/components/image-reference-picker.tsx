import { useEffect, useMemo, useState } from "react"
import { useTranslation } from "react-i18next"

import { useSession } from "@/auth/session"

import {
  listRegistries,
  listRegistryProjects,
  listRegistryRepositories,
  listRegistryTags,
} from "@/api/registry"
import type {
  RegistryItem,
  RegistryProjectItem,
  RegistryRepositoryItem,
  RegistryTagItem,
} from "@/api/registry/types"
import { SearchableSelect } from "@/components/searchable-select"
import { Input } from "@/components/ui/input"
import { failureMessage } from "@/lib/form-error"

export interface ImageSelection {
  image: string
  registryUid: string
  registryName: string
  endpoint: string
  project: string
  repository: string
  tag: string
  digest: string
  platforms: string[]
}

interface ImageReferencePickerProps {
  value: string
  onChange: (value: string) => void
  /** 选中仓库镜像时回传完整引用（含 digest），手输外部镜像时回传 null。 */
  onSelect?: (selection: ImageSelection | null) => void
  disabled?: boolean
  invalid?: boolean
  id?: string
  className?: string
  "aria-describedby"?: string
}

/** 去掉协议与末尾斜杠，得到可用于拼镜像引用的仓库地址。 */
function registryHost(endpoint: string): string {
  return endpoint.replace(/^https?:\/\//i, "").replace(/\/+$/, "")
}

function joinImage(input: { registryName: string; project: string; repository: string; tag: string }): string {
  const path = [input.registryName, input.project, input.repository].filter(Boolean).join("/")
  return input.tag ? `${path}:${input.tag}` : path
}

/**
 * 从镜像仓库选择镜像：registry → project → repository → tag 级联。
 * 仍保留下方输入框手输外部镜像（例如 Docker Hub 或未接入的仓库），
 * 此时提示 digest 未固定。
 */
export function ImageReferencePicker({
  value,
  onChange,
  onSelect,
  disabled = false,
  invalid = false,
  id,
  className,
  "aria-describedby": ariaDescribedBy,
}: ImageReferencePickerProps) {
  const { t } = useTranslation()
  const session = useSession()
  const registryAllowed = session.can("registry:read")
  const [registries, setRegistries] = useState<RegistryItem[]>([])
  const [registryUid, setRegistryUid] = useState("")
  const [projects, setProjects] = useState<RegistryProjectItem[]>([])
  const [project, setProject] = useState("")
  const [repositories, setRepositories] = useState<RegistryRepositoryItem[]>([])
  const [repository, setRepository] = useState("")
  const [tags, setTags] = useState<RegistryTagItem[]>([])
  const [tag, setTag] = useState("")
  const [selection, setSelection] = useState<ImageSelection | null>(null)
  const [error, setError] = useState("")

  const registry = useMemo(
    () => registries.find((item) => item.uid === registryUid),
    [registries, registryUid],
  )
  const displayError = registryAllowed ? error : t("appStore.imagePicker.registryUnavailable")

  useEffect(() => {
    if (!registryAllowed) {
      return
    }
    let active = true
    void listRegistries({ page: 1, pageSize: 100 }, { allowForbidden: true, silent: true })
      .then((result) => {
        if (!active) return
        setRegistries(result.items)
        // 当前值命中某个仓库地址时，自动选中最匹配的仓库，便于继续选择。
        const matched = result.items.find((item) => value.startsWith(`${registryHost(item.endpoint)}/`))
        if (matched) {
          setRegistryUid((current) => current || matched.uid)
        }
      })
      .catch((err) => {
        if (active) setError(failureMessage(err, t("appStore.imagePicker.registryUnavailable")))
      })
    return () => {
      active = false
    }
    // 只在仓库能力与首次加载变化时刷新；value 的变化由用户选择驱动。
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [registryAllowed])

  useEffect(() => {
    if (!registryUid) {
      return
    }
    let active = true
    void listRegistryProjects(registryUid)
      .then((result) => {
        if (active) setProjects(result.items)
      })
      .catch(() => {
        if (active) setProjects([])
      })
    return () => {
      active = false
    }
  }, [registryUid])

  useEffect(() => {
    if (!registryUid) {
      return
    }
    let active = true
    void listRegistryRepositories(registryUid, { project, page: 1, pageSize: 100 })
      .then((result) => {
        if (active) setRepositories(result.items)
      })
      .catch(() => {
        if (active) setRepositories([])
      })
    return () => {
      active = false
    }
  }, [registryUid, project])

  useEffect(() => {
    if (!registryUid || !repository) {
      return
    }
    let active = true
    void listRegistryTags(registryUid, { project, repository, page: 1, pageSize: 100 })
      .then((result) => {
        if (active) setTags(result.items)
      })
      .catch(() => {
        if (active) setTags([])
      })
    return () => {
      active = false
    }
  }, [registryUid, project, repository])

  function pickTag(next: string) {
    setTag(next)
    if (!registry || !repository) {
      return
    }
    const item = tags.find((entry) => entry.tag === next)
    const image = joinImage({
      registryName: registryHost(registry.endpoint),
      project,
      repository,
      tag: next,
    })
    const nextSelection: ImageSelection = {
      image,
      registryUid: registry.uid,
      registryName: registry.name,
      endpoint: registryHost(registry.endpoint),
      project,
      repository,
      tag: next,
      digest: item?.digest ?? "",
      platforms: item?.platforms ?? [],
    }
    setSelection(nextSelection)
    onChange(image)
    onSelect?.(nextSelection)
  }

  function changeManual(next: string) {
    setSelection(null)
    onSelect?.(null)
    onChange(next)
  }

  const external = value.trim() !== "" && selection === null

  return (
    <div className={className}>
      <div className="grid gap-2 sm:grid-cols-2">
        <SearchableSelect
          value={registryUid}
          options={registries.map((item) => ({
            value: item.uid,
            label: item.name,
            description: item.endpoint,
          }))}
          onChange={(next) => {
            setRegistryUid(next)
            setProject("")
            setRepository("")
            setTag("")
            setSelection(null)
          }}
          placeholder={t("appStore.imagePicker.registryPlaceholder")}
          searchPlaceholder={t("appStore.imagePicker.search")}
          emptyText={t("appStore.imagePicker.emptyRegistries")}
          clearLabel={t("appStore.imagePicker.clear")}
          disabled={disabled || !registryAllowed}
          invalid={invalid}
          ariaLabel={t("appStore.imagePicker.registry")}
          aria-describedby={ariaDescribedBy}
        />
        <SearchableSelect
          value={project}
          options={projects.map((item) => ({
            value: item.name,
            label: item.name,
            description: t("appStore.imagePicker.repositoryCount", { count: item.repositoryCount }),
          }))}
          onChange={(next) => {
            setProject(next)
            setRepository("")
            setTag("")
          }}
          placeholder={t("appStore.imagePicker.projectPlaceholder")}
          searchPlaceholder={t("appStore.imagePicker.search")}
          emptyText={t("appStore.imagePicker.emptyProjects")}
          clearLabel={t("appStore.imagePicker.clear")}
          disabled={disabled || !registryUid}
          invalid={invalid}
          ariaLabel={t("appStore.imagePicker.project")}
        />
        <SearchableSelect
          value={repository}
          options={repositories.map((item) => ({
            value: item.name,
            label: item.name,
            description: t("appStore.imagePicker.tagCount", { count: item.tagCount }),
          }))}
          onChange={(next) => {
            setRepository(next)
            setTag("")
          }}
          placeholder={t("appStore.imagePicker.repositoryPlaceholder")}
          searchPlaceholder={t("appStore.imagePicker.search")}
          emptyText={t("appStore.imagePicker.emptyRepositories")}
          clearLabel={t("appStore.imagePicker.clear")}
          disabled={disabled || !registryUid}
          invalid={invalid}
          ariaLabel={t("appStore.imagePicker.repository")}
        />
        <SearchableSelect
          value={tag}
          options={tags.map((item) => ({
            value: item.tag,
            label: item.tag,
            description: [item.digest?.slice(0, 19) || "", item.platforms?.join(" · ") || ""]
              .filter(Boolean)
              .join(" "),
          }))}
          onChange={pickTag}
          placeholder={t("appStore.imagePicker.tagPlaceholder")}
          searchPlaceholder={t("appStore.imagePicker.search")}
          emptyText={t("appStore.imagePicker.emptyTags")}
          disabled={disabled || !registryUid || !repository}
          invalid={invalid}
          ariaLabel={t("appStore.imagePicker.tag")}
        />
      </div>
      <Input
        id={id}
        className="mt-2 font-mono"
        value={value}
        disabled={disabled}
        aria-invalid={invalid || undefined}
        aria-describedby={ariaDescribedBy}
        placeholder={t("appStore.imagePicker.imagePlaceholder")}
        onChange={(event) => changeManual(event.target.value)}
      />
      <div className="mt-1 grid gap-0.5 text-xs">
        {selection ? (
          <p className="text-muted-foreground">
            {t("appStore.imagePicker.pickedHint", {
              registry: selection.registryName,
              repository: selection.repository,
              tag: selection.tag,
            })}
            {selection.digest ? ` · ${t("appStore.imagePicker.digest", { digest: selection.digest })}` : null}
          </p>
        ) : null}
        {external ? (
          <p className="text-warning">{t("appStore.imagePicker.externalHint")}</p>
        ) : null}
        {displayError ? <p className="text-destructive">{displayError}</p> : null}
        {!selection && !external && !displayError ? (
          <p className="text-muted-foreground">{t("appStore.imagePicker.manualHint")}</p>
        ) : null}
      </div>
    </div>
  )
}
