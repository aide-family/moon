import type { JSONSchema } from "monaco-yaml"

// 内置的 Kubernetes 清单 JSON Schema（裁剪版）。
//
// 只声明编辑时最容易写错、也最有提示价值的字段，未列出的字段全部放行
// （additionalProperties: true），因此不会把合法清单标红。数据全部静态内置，
// 不请求后端，也不依赖 schema 接口。

const stringMap: JSONSchema = {
  type: "object",
  additionalProperties: { type: "string" },
}

const labels: JSONSchema = { type: "object", additionalProperties: { type: "string" } }

const metadataProperties: Record<string, JSONSchema> = {
  name: { type: "string", description: "资源名（DNS 子域名）" },
  namespace: { type: "string", description: "命名空间；集群级资源不填" },
  labels,
  annotations: labels,
  generateName: { type: "string" },
  finalizers: { type: "array", items: { type: "string" } },
}

/** 资源根级 metadata：name 必填。 */
const metadata: JSONSchema = {
  type: "object",
  properties: metadataProperties,
  required: ["name"],
  additionalProperties: true,
}

/**
 * Pod template / jobTemplate 内嵌的 metadata：只允许 labels/annotations 等，
 * 不要求 name（由上层控制器生成），因此不能复用根级 metadata 的 required。
 */
const podMetadata: JSONSchema = {
  type: "object",
  properties: metadataProperties,
  additionalProperties: true,
}

const containerPort: JSONSchema = {
  type: "object",
  properties: {
    name: { type: "string" },
    containerPort: { type: "integer", minimum: 1, maximum: 65535 },
    protocol: { type: "string", enum: ["TCP", "UDP", "SCTP"] },
    hostPort: { type: "integer" },
  },
  required: ["containerPort"],
  additionalProperties: true,
}

const servicePort: JSONSchema = {
  type: "object",
  properties: {
    name: { type: "string" },
    port: { type: "integer", minimum: 1, maximum: 65535 },
    targetPort: { type: ["integer", "string"] },
    nodePort: { type: "integer" },
    protocol: { type: "string", enum: ["TCP", "UDP", "SCTP"] },
    appProtocol: { type: "string" },
  },
  required: ["port"],
  additionalProperties: true,
}

const keySelector: JSONSchema = {
  type: "object",
  properties: {
    name: { type: "string" },
    key: { type: "string" },
    optional: { type: "boolean" },
  },
  required: ["key"],
  additionalProperties: true,
}

const envVar: JSONSchema = {
  type: "object",
  properties: {
    name: { type: "string" },
    value: { type: "string" },
    valueFrom: {
      type: "object",
      properties: {
        configMapKeyRef: keySelector,
        secretKeyRef: keySelector,
        fieldRef: { type: "object", properties: { fieldPath: { type: "string" } }, required: ["fieldPath"], additionalProperties: true },
        resourceFieldRef: { type: "object", additionalProperties: true },
      },
      additionalProperties: true,
    },
  },
  required: ["name"],
  additionalProperties: true,
}

const volumeMount: JSONSchema = {
  type: "object",
  properties: {
    name: { type: "string" },
    mountPath: { type: "string" },
    subPath: { type: "string" },
    readOnly: { type: "boolean" },
  },
  required: ["name", "mountPath"],
  additionalProperties: true,
}

const volume: JSONSchema = {
  type: "object",
  properties: {
    name: { type: "string" },
    emptyDir: { type: "object", additionalProperties: true },
    configMap: { type: "object", properties: { name: { type: "string" }, defaultMode: { type: "integer" }, optional: { type: "boolean" } }, additionalProperties: true },
    secret: { type: "object", properties: { secretName: { type: "string" }, defaultMode: { type: "integer" }, optional: { type: "boolean" } }, additionalProperties: true },
    persistentVolumeClaim: { type: "object", properties: { claimName: { type: "string" }, readOnly: { type: "boolean" } }, required: ["claimName"], additionalProperties: true },
    hostPath: { type: "object", properties: { path: { type: "string" }, type: { type: "string" } }, required: ["path"], additionalProperties: true },
    projected: { type: "object", additionalProperties: true },
    nfs: { type: "object", additionalProperties: true },
  },
  required: ["name"],
  additionalProperties: true,
}

const probeAction: JSONSchema = {
  type: "object",
  properties: {
    httpGet: {
      type: "object",
      properties: {
        path: { type: "string" },
        port: { type: ["integer", "string"] },
        host: { type: "string" },
        scheme: { type: "string", enum: ["HTTP", "HTTPS"] },
      },
      required: ["port"],
      additionalProperties: true,
    },
    tcpSocket: {
      type: "object",
      properties: {
        host: { type: "string" },
        port: { type: ["integer", "string"] },
      },
      required: ["port"],
      additionalProperties: true,
    },
    exec: {
      type: "object",
      properties: { command: { type: "array", items: { type: "string" }, minItems: 1 } },
      required: ["command"],
      additionalProperties: true,
    },
    grpc: {
      type: "object",
      properties: { port: { type: "integer" }, service: { type: "string" } },
      required: ["port"],
      additionalProperties: true,
    },
    initialDelaySeconds: { type: "integer", minimum: 0 },
    periodSeconds: { type: "integer", minimum: 1 },
    timeoutSeconds: { type: "integer", minimum: 1 },
    successThreshold: { type: "integer", minimum: 1 },
    failureThreshold: { type: "integer", minimum: 1 },
    terminationGracePeriodSeconds: { type: "integer", minimum: 0 },
  },
  additionalProperties: true,
}

const container: JSONSchema = {
  type: "object",
  properties: {
    name: { type: "string" },
    image: { type: "string", description: "镜像引用，建议固定 digest 或 tag" },
    imagePullPolicy: { type: "string", enum: ["Always", "IfNotPresent", "Never"] },
    command: { type: "array", items: { type: "string" } },
    args: { type: "array", items: { type: "string" } },
    env: { type: "array", items: envVar },
    envFrom: { type: "array", items: { type: "object", additionalProperties: true } },
    ports: { type: "array", items: containerPort },
    resources: {
      type: "object",
      properties: {
        requests: { type: "object", additionalProperties: { type: ["string", "number"] } },
        limits: { type: "object", additionalProperties: { type: ["string", "number"] } },
      },
      additionalProperties: true,
    },
    volumeMounts: { type: "array", items: volumeMount },
    readinessProbe: probeAction,
    livenessProbe: probeAction,
    startupProbe: probeAction,
    securityContext: { type: "object", additionalProperties: true },
  },
  required: ["name", "image"],
  additionalProperties: true,
}

const podSpec: JSONSchema = {
  type: "object",
  properties: {
    containers: { type: "array", items: container, minItems: 1 },
    initContainers: { type: "array", items: container },
    volumes: { type: "array", items: volume },
    serviceAccountName: { type: "string" },
    restartPolicy: { type: "string", enum: ["Always", "OnFailure", "Never"] },
    nodeSelector: stringMap,
    tolerations: { type: "array", items: { type: "object", additionalProperties: true } },
    affinity: { type: "object", additionalProperties: true },
    imagePullSecrets: {
      type: "array",
      items: { type: "object", properties: { name: { type: "string" } }, additionalProperties: true },
    },
    securityContext: { type: "object", additionalProperties: true },
  },
  required: ["containers"],
  additionalProperties: true,
}

const workloadSpec = (extra: Record<string, JSONSchema>): JSONSchema => ({
  type: "object",
  properties: {
    replicas: { type: "integer", minimum: 0, maximum: 200, description: "副本数" },
    selector: {
      type: "object",
      properties: { matchLabels: labels, matchExpressions: { type: "array", items: { type: "object", additionalProperties: true } } },
      additionalProperties: true,
    },
    template: { type: "object", properties: { metadata: podMetadata, spec: podSpec }, additionalProperties: true },
    strategy: { type: "object", additionalProperties: true },
    minReadySeconds: { type: "integer", minimum: 0 },
    revisionHistoryLimit: { type: "integer", minimum: 0 },
    ...extra,
  },
  required: ["selector", "template"],
  additionalProperties: true,
})

function schema(kind: string, apiVersion: string, properties: Record<string, JSONSchema>, required: string[] = []): JSONSchema {
  return {
    $schema: "http://json-schema.org/draft-07/schema",
    title: `Kubernetes ${kind}`,
    type: "object",
    properties: {
      apiVersion: { type: "string", enum: [apiVersion] },
      kind: { type: "string", enum: [kind] },
      metadata,
      ...properties,
    },
    required: ["apiVersion", "kind", "metadata", ...required],
    additionalProperties: true,
  }
}

const rule: JSONSchema = {
  type: "object",
  properties: {
    apiGroups: { type: "array", items: { type: "string" } },
    resources: { type: "array", items: { type: "string" } },
    resourceNames: { type: "array", items: { type: "string" } },
    verbs: { type: "array", items: { type: "string" } },
    nonResourceURLs: { type: "array", items: { type: "string" } },
  },
  additionalProperties: true,
}

const subject: JSONSchema = {
  type: "object",
  properties: {
    kind: { type: "string", enum: ["User", "Group", "ServiceAccount"] },
    name: { type: "string" },
    namespace: { type: "string" },
    apiGroup: { type: "string" },
  },
  required: ["kind", "name"],
  additionalProperties: true,
}

const roleRef: JSONSchema = {
  type: "object",
  properties: {
    apiGroup: { type: "string", enum: ["rbac.authorization.k8s.io"] },
    kind: { type: "string", enum: ["Role", "ClusterRole"] },
    name: { type: "string" },
  },
  required: ["apiGroup", "kind", "name"],
  additionalProperties: true,
}

const cronJobSpec: JSONSchema = {
  type: "object",
  properties: {
    schedule: { type: "string", description: "Cron 表达式，例如 0 * * * *" },
    concurrencyPolicy: { type: "string", enum: ["Allow", "Forbid", "Replace"] },
    suspend: { type: "boolean" },
    successfulJobsHistoryLimit: { type: "integer", minimum: 0 },
    failedJobsHistoryLimit: { type: "integer", minimum: 0 },
    startingDeadlineSeconds: { type: "integer", minimum: 0 },
    jobTemplate: {
      type: "object",
      properties: {
        metadata: podMetadata,
        spec: { type: "object", properties: { template: { type: "object", properties: { metadata: podMetadata, spec: podSpec }, additionalProperties: true } }, additionalProperties: true },
      },
      additionalProperties: true,
    },
  },
  required: ["schedule", "jobTemplate"],
  additionalProperties: true,
}

export interface KubernetesYamlSchema {
  kind: string
  apiVersion: string
  schema: JSONSchema
}

export const KUBERNETES_YAML_SCHEMAS: KubernetesYamlSchema[] = [
  {
    kind: "Deployment",
    apiVersion: "apps/v1",
    schema: schema("Deployment", "apps/v1", { spec: workloadSpec({}) }, ["spec"]),
  },
  {
    kind: "StatefulSet",
    apiVersion: "apps/v1",
    schema: schema(
      "StatefulSet",
      "apps/v1",
      {
        spec: workloadSpec({
          serviceName: { type: "string" },
          volumeClaimTemplates: { type: "array", items: { type: "object", additionalProperties: true } },
          podManagementPolicy: { type: "string", enum: ["OrderedReady", "Parallel"] },
          updateStrategy: { type: "object", additionalProperties: true },
        }),
      },
      ["spec"],
    ),
  },
  {
    kind: "DaemonSet",
    apiVersion: "apps/v1",
    schema: schema(
      "DaemonSet",
      "apps/v1",
      { spec: { ...workloadSpec({ updateStrategy: { type: "object", additionalProperties: true } }), required: ["selector", "template"] } },
      ["spec"],
    ),
  },
  {
    kind: "Pod",
    apiVersion: "v1",
    schema: schema("Pod", "v1", { spec: podSpec }, ["spec"]),
  },
  {
    kind: "Service",
    apiVersion: "v1",
    schema: schema(
      "Service",
      "v1",
      {
        spec: {
          type: "object",
          properties: {
            type: { type: "string", enum: ["ClusterIP", "NodePort", "LoadBalancer", "ExternalName"] },
            selector: stringMap,
            ports: { type: "array", items: servicePort, minItems: 1 },
            clusterIP: { type: "string" },
            externalName: { type: "string" },
            sessionAffinity: { type: "string", enum: ["None", "ClientIP"] },
          },
          required: ["ports"],
          additionalProperties: true,
        },
      },
      ["spec"],
    ),
  },
  {
    kind: "Ingress",
    apiVersion: "networking.k8s.io/v1",
    schema: schema(
      "Ingress",
      "networking.k8s.io/v1",
      {
        spec: {
          type: "object",
          properties: {
            ingressClassName: { type: "string" },
            defaultBackend: { type: "object", additionalProperties: true },
            tls: {
              type: "array",
              items: {
                type: "object",
                properties: { hosts: { type: "array", items: { type: "string" } }, secretName: { type: "string" } },
                additionalProperties: true,
              },
            },
            rules: {
              type: "array",
              items: {
                type: "object",
                properties: {
                  host: { type: "string" },
                  http: {
                    type: "object",
                    properties: {
                      paths: {
                        type: "array",
                        items: {
                          type: "object",
                          properties: {
                            path: { type: "string" },
                            pathType: { type: "string", enum: ["Prefix", "Exact", "ImplementationSpecific"] },
                            backend: { type: "object", additionalProperties: true },
                          },
                          required: ["pathType", "backend"],
                          additionalProperties: true,
                        },
                      },
                    },
                    additionalProperties: true,
                  },
                },
                additionalProperties: true,
              },
            },
          },
          additionalProperties: true,
        },
      },
      ["spec"],
    ),
  },
  {
    kind: "ConfigMap",
    apiVersion: "v1",
    schema: schema("ConfigMap", "v1", {
      data: stringMap,
      binaryData: stringMap,
      immutable: { type: "boolean" },
    }),
  },
  {
    kind: "Secret",
    apiVersion: "v1",
    schema: schema(
      "Secret",
      "v1",
      {
        type: { type: "string", description: "例如 Opaque、kubernetes.io/tls" },
        data: stringMap,
        stringData: stringMap,
        immutable: { type: "boolean" },
      },
      ["type"],
    ),
  },
  {
    kind: "PersistentVolumeClaim",
    apiVersion: "v1",
    schema: schema(
      "PersistentVolumeClaim",
      "v1",
      {
        spec: {
          type: "object",
          properties: {
            accessModes: { type: "array", items: { type: "string" }, minItems: 1 },
            resources: {
              type: "object",
              properties: {
                requests: { type: "object", additionalProperties: { type: "string" } },
                limits: { type: "object", additionalProperties: { type: "string" } },
              },
              additionalProperties: true,
            },
            storageClassName: { type: "string" },
            volumeName: { type: "string" },
            volumeMode: { type: "string", enum: ["Filesystem", "Block"] },
          },
          required: ["accessModes", "resources"],
          additionalProperties: true,
        },
      },
      ["spec"],
    ),
  },
  {
    kind: "Job",
    apiVersion: "batch/v1",
    schema: schema(
      "Job",
      "batch/v1",
      {
        spec: {
          type: "object",
          properties: {
            backoffLimit: { type: "integer", minimum: 0 },
            completions: { type: "integer", minimum: 0 },
            parallelism: { type: "integer", minimum: 0 },
            ttlSecondsAfterFinished: { type: "integer", minimum: 0 },
            suspend: { type: "boolean" },
            template: { type: "object", properties: { metadata: podMetadata, spec: podSpec }, additionalProperties: true },
          },
          required: ["template"],
          additionalProperties: true,
        },
      },
      ["spec"],
    ),
  },
  {
    kind: "CronJob",
    apiVersion: "batch/v1",
    schema: schema("CronJob", "batch/v1", { spec: cronJobSpec }, ["spec"]),
  },
  {
    kind: "HorizontalPodAutoscaler",
    apiVersion: "autoscaling/v2",
    schema: schema(
      "HorizontalPodAutoscaler",
      "autoscaling/v2",
      {
        spec: {
          type: "object",
          properties: {
            scaleTargetRef: {
              type: "object",
              properties: {
                apiVersion: { type: "string" },
                kind: { type: "string" },
                name: { type: "string" },
              },
              required: ["kind", "name"],
              additionalProperties: true,
            },
            minReplicas: { type: "integer", minimum: 1 },
            maxReplicas: { type: "integer", minimum: 1 },
            metrics: { type: "array", items: { type: "object", additionalProperties: true } },
            behavior: { type: "object", additionalProperties: true },
          },
          required: ["scaleTargetRef", "maxReplicas"],
          additionalProperties: true,
        },
      },
      ["spec"],
    ),
  },
  {
    kind: "ServiceAccount",
    apiVersion: "v1",
    schema: schema("ServiceAccount", "v1", {
      automountServiceAccountToken: { type: "boolean" },
      imagePullSecrets: {
        type: "array",
        items: { type: "object", properties: { name: { type: "string" } }, additionalProperties: true },
      },
    }),
  },
  {
    kind: "Role",
    apiVersion: "rbac.authorization.k8s.io/v1",
    schema: schema("Role", "rbac.authorization.k8s.io/v1", { rules: { type: "array", items: rule } }, ["rules"]),
  },
  {
    kind: "ClusterRole",
    apiVersion: "rbac.authorization.k8s.io/v1",
    schema: schema("ClusterRole", "rbac.authorization.k8s.io/v1", {
      rules: { type: "array", items: rule },
      aggregationRule: { type: "object", additionalProperties: true },
    }),
  },
  {
    kind: "RoleBinding",
    apiVersion: "rbac.authorization.k8s.io/v1",
    schema: schema(
      "RoleBinding",
      "rbac.authorization.k8s.io/v1",
      { roleRef, subjects: { type: "array", items: subject, minItems: 1 } },
      ["roleRef", "subjects"],
    ),
  },
  {
    kind: "ClusterRoleBinding",
    apiVersion: "rbac.authorization.k8s.io/v1",
    schema: schema(
      "ClusterRoleBinding",
      "rbac.authorization.k8s.io/v1",
      { roleRef, subjects: { type: "array", items: subject, minItems: 1 } },
      ["roleRef", "subjects"],
    ),
  },
  {
    kind: "CustomResourceDefinition",
    apiVersion: "apiextensions.k8s.io/v1",
    schema: schema(
      "CustomResourceDefinition",
      "apiextensions.k8s.io/v1",
      {
        spec: {
          type: "object",
          properties: {
            group: { type: "string" },
            scope: { type: "string", enum: ["Namespaced", "Cluster"] },
            names: {
              type: "object",
              properties: {
                kind: { type: "string" },
                plural: { type: "string" },
                singular: { type: "string" },
                shortNames: { type: "array", items: { type: "string" } },
              },
              required: ["kind", "plural"],
              additionalProperties: true,
            },
            versions: {
              type: "array",
              items: {
                type: "object",
                properties: {
                  name: { type: "string" },
                  served: { type: "boolean" },
                  storage: { type: "boolean" },
                  schema: { type: "object", additionalProperties: true },
                  subresources: { type: "object", additionalProperties: true },
                  additionalPrinterColumns: { type: "array", items: { type: "object", additionalProperties: true } },
                },
                required: ["name", "served", "storage"],
                additionalProperties: true,
              },
              minItems: 1,
            },
          },
          required: ["group", "names", "scope", "versions"],
          additionalProperties: true,
        },
      },
      ["spec"],
    ),
  },
  {
    kind: "NetworkPolicy",
    apiVersion: "networking.k8s.io/v1",
    schema: schema(
      "NetworkPolicy",
      "networking.k8s.io/v1",
      {
        spec: {
          type: "object",
          properties: {
            podSelector: { type: "object", additionalProperties: true },
            policyTypes: { type: "array", items: { type: "string", enum: ["Ingress", "Egress"] } },
            ingress: { type: "array", items: { type: "object", additionalProperties: true } },
            egress: { type: "array", items: { type: "object", additionalProperties: true } },
          },
          additionalProperties: true,
        },
      },
      ["spec"],
    ),
  },
]

/** 无法识别 kind 时的兜底 schema：只提示 apiVersion / kind / metadata。 */
export const GENERIC_KUBERNETES_SCHEMA: JSONSchema = {
  $schema: "http://json-schema.org/draft-07/schema",
  title: "Kubernetes resource",
  type: "object",
  properties: {
    apiVersion: { type: "string", description: "资源所属 API 组与版本，例如 apps/v1" },
    kind: { type: "string", description: "资源类型，例如 Deployment" },
    metadata,
  },
  required: ["apiVersion", "kind", "metadata"],
  additionalProperties: true,
}

const knownKinds = new Set(KUBERNETES_YAML_SCHEMAS.map((item) => item.kind))

/** 内置 schema 覆盖的 kind 列表（用于资源类型下拉）。 */
export const KUBERNETES_YAML_KINDS = KUBERNETES_YAML_SCHEMAS.map((item) => item.kind)

/** 判断 kind 是否有内置 schema；无法识别时编辑器退化为通用提示。 */
export function isKnownKubernetesKind(kind: string): boolean {
  return knownKinds.has(kind)
}

/** 从清单原文读取 kind，读取失败返回空串。 */
export function readManifestKind(manifest: string): string {
  return /^\s*kind:\s*["']?([A-Za-z0-9.-]+)["']?\s*$/m?.exec(manifest)?.[1] ?? ""
}
