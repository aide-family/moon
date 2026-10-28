import { del, get, post, put } from "@/api/http"

import type {
  ClusterItem,
  GetKubernetesConfigMapDetailParams,
  GetKubernetesConfigMapDetailResult,
  GetKubernetesCronJobDetailParams,
  GetKubernetesCronJobDetailResult,
  GetKubernetesIngressDetailParams,
  GetKubernetesIngressDetailResult,
  GetKubernetesJobDetailParams,
  GetKubernetesJobDetailResult,
  GetKubernetesPersistentVolumeClaimDetailParams,
  GetKubernetesPersistentVolumeClaimDetailResult,
  GetKubernetesResourceYamlParams,
  GetKubernetesResourceYamlResult,
  GetKubernetesSecretDetailParams,
  GetKubernetesSecretDetailResult,
  GetKubernetesServiceDetailParams,
  GetKubernetesServiceDetailResult,
  GetKubernetesWorkloadDetailParams,
  GetKubernetesWorkloadDetailResult,
  GetKubernetesWorkloadImagesParams,
  GetKubernetesWorkloadImagesResult,
  KubernetesCronJobItem,
  KubernetesJobItem,
  KubernetesPodLogsParams,
  KubernetesPodLogsResult,
  KubernetesWorkloadItem,
  ListClustersParams,
  ListClustersResult,
  ListKubernetesConfigMapsParams,
  ListKubernetesConfigMapsResult,
  ListKubernetesCronJobsParams,
  ListKubernetesCronJobsResult,
  ListKubernetesIngressesParams,
  ListKubernetesIngressesResult,
  ListKubernetesJobsParams,
  ListKubernetesJobsResult,
  ListKubernetesNamespacesResult,
  ListKubernetesNodesParams,
  ListKubernetesNodesResult,
  ListKubernetesPersistentVolumeClaimsParams,
  ListKubernetesPersistentVolumeClaimsResult,
  ListKubernetesPersistentVolumesParams,
  ListKubernetesPersistentVolumesResult,
  ListKubernetesPodsParams,
  ListKubernetesPodsResult,
  ListKubernetesSecretsParams,
  ListKubernetesSecretsResult,
  ListKubernetesServicesParams,
  ListKubernetesServicesResult,
  ListKubernetesWorkloadRevisionsParams,
  ListKubernetesWorkloadRevisionsResult,
  ListKubernetesWorkloadsParams,
  ListKubernetesWorkloadsResult,
  ProbeKubeconfigResult,
  RestartKubernetesWorkloadParams,
  RollbackKubernetesWorkloadParams,
  SaveClusterParams,
  ScaleKubernetesWorkloadParams,
  SetKubernetesCronJobSuspendParams,
  SetKubernetesWorkloadImageParams,
  TestClusterResult,
  TriggerKubernetesCronJobParams,
  UpdateClusterParams,
} from "./types"

// 路径参数可能是命名空间、节点名或工作负载名，必须编码，否则带 / 的名称会破坏路由。
function segment(value: string): string {
  return encodeURIComponent(value)
}

/** GET /v1/clusters */
export function listClusters(
  params: ListClustersParams = {}
): Promise<ListClustersResult> {
  return get<ListClustersResult>("/clusters", params)
}

/** GET /v1/cluster/{uid} */
export function getCluster(uid: string): Promise<{ cluster: ClusterItem }> {
  return get<{ cluster: ClusterItem }>(`/cluster/${segment(uid)}`)
}

/** POST /v1/cluster */
export function createCluster(
  body: SaveClusterParams
): Promise<{ cluster: ClusterItem }> {
  return post<{ cluster: ClusterItem }>("/cluster", body)
}

/** PUT /v1/cluster/{uid}；kubeconfig 留空表示保持原值。 */
export function updateCluster(
  body: UpdateClusterParams
): Promise<{ cluster: ClusterItem }> {
  return put<{ cluster: ClusterItem }>(`/cluster/${segment(body.uid)}`, body)
}

/** DELETE /v1/cluster/{uid} */
export function deleteCluster(uid: string): Promise<Record<string, never>> {
  return del<Record<string, never>>(`/cluster/${segment(uid)}`)
}

/** POST /v1/cluster/{uid}/test */
export function testCluster(uid: string): Promise<TestClusterResult> {
  return post<TestClusterResult>(`/cluster/${segment(uid)}/test`, {})
}

/** POST /v1/cluster/probe */
export function probeKubeconfig(
  kubeconfig: string
): Promise<ProbeKubeconfigResult> {
  return post<ProbeKubeconfigResult>("/cluster/probe", { kubeconfig })
}

/** GET /v1/cluster/{clusterUid}/namespaces */
export function listKubernetesNamespaces(
  clusterUid: string
): Promise<ListKubernetesNamespacesResult> {
  return get<ListKubernetesNamespacesResult>(
    `/cluster/${segment(clusterUid)}/namespaces`
  )
}

/** GET /v1/cluster/{clusterUid}/nodes */
export function listKubernetesNodes(
  clusterUid: string,
  params: ListKubernetesNodesParams = {}
): Promise<ListKubernetesNodesResult> {
  return get<ListKubernetesNodesResult>(
    `/cluster/${segment(clusterUid)}/nodes`,
    params
  )
}

/** PUT /v1/cluster/{clusterUid}/node/{name}/unschedulable */
export function setKubernetesNodeUnschedulable(
  clusterUid: string,
  name: string,
  unschedulable: boolean
): Promise<Record<string, never>> {
  return put<Record<string, never>>(
    `/cluster/${segment(clusterUid)}/node/${segment(name)}/unschedulable`,
    { unschedulable }
  )
}

/** GET /v1/cluster/{clusterUid}/workloads */
export function listKubernetesWorkloads(
  clusterUid: string,
  params: ListKubernetesWorkloadsParams = {}
): Promise<ListKubernetesWorkloadsResult> {
  return get<ListKubernetesWorkloadsResult>(
    `/cluster/${segment(clusterUid)}/workloads`,
    params
  )
}

/** PUT /v1/cluster/{clusterUid}/workload/scale */
export function scaleKubernetesWorkload(
  clusterUid: string,
  body: ScaleKubernetesWorkloadParams
): Promise<{ workload: KubernetesWorkloadItem }> {
  return put<{ workload: KubernetesWorkloadItem }>(
    `/cluster/${segment(clusterUid)}/workload/scale`,
    body
  )
}

/** POST /v1/cluster/{clusterUid}/workload/restart */
export function restartKubernetesWorkload(
  clusterUid: string,
  body: RestartKubernetesWorkloadParams
): Promise<{ workload: KubernetesWorkloadItem }> {
  return post<{ workload: KubernetesWorkloadItem }>(
    `/cluster/${segment(clusterUid)}/workload/restart`,
    body
  )
}

/** GET /v1/cluster/{clusterUid}/workload/images */
export function getKubernetesWorkloadImages(
  clusterUid: string,
  params: GetKubernetesWorkloadImagesParams
): Promise<GetKubernetesWorkloadImagesResult> {
  return get<GetKubernetesWorkloadImagesResult>(
    `/cluster/${segment(clusterUid)}/workload/images`,
    { kind: params.kind, namespace: params.namespace, name: params.name },
    undefined,
    { silent: true }
  )
}

/** PUT /v1/cluster/{clusterUid}/workload/image */
export function setKubernetesWorkloadImage(
  clusterUid: string,
  body: SetKubernetesWorkloadImageParams
): Promise<{ workload: KubernetesWorkloadItem }> {
  return put<{ workload: KubernetesWorkloadItem }>(
    `/cluster/${segment(clusterUid)}/workload/image`,
    body
  )
}

/** GET /v1/cluster/{clusterUid}/workload/revisions */
export function listKubernetesWorkloadRevisions(
  clusterUid: string,
  params: ListKubernetesWorkloadRevisionsParams
): Promise<ListKubernetesWorkloadRevisionsResult> {
  return get<ListKubernetesWorkloadRevisionsResult>(
    `/cluster/${segment(clusterUid)}/workload/revisions`,
    { kind: params.kind, namespace: params.namespace, name: params.name },
    undefined,
    { silent: true }
  )
}

/** POST /v1/cluster/{clusterUid}/workload/rollback */
export function rollbackKubernetesWorkload(
  clusterUid: string,
  body: RollbackKubernetesWorkloadParams
): Promise<{ workload: KubernetesWorkloadItem }> {
  return post<{ workload: KubernetesWorkloadItem }>(
    `/cluster/${segment(clusterUid)}/workload/rollback`,
    body
  )
}

/** GET /v1/cluster/{clusterUid}/pods */
export function listKubernetesPods(
  clusterUid: string,
  params: ListKubernetesPodsParams = {}
): Promise<ListKubernetesPodsResult> {
  return get<ListKubernetesPodsResult>(
    `/cluster/${segment(clusterUid)}/pods`,
    params
  )
}

/** DELETE /v1/cluster/{clusterUid}/namespace/{namespace}/pod/{name} */
export function deleteKubernetesPod(
  clusterUid: string,
  namespace: string,
  name: string
): Promise<Record<string, never>> {
  return del<Record<string, never>>(
    `/cluster/${segment(clusterUid)}/namespace/${segment(namespace)}/pod/${segment(name)}`
  )
}

/** GET /v1/cluster/{clusterUid}/namespace/{namespace}/pod/{name}/logs */
export function getKubernetesPodLogs(
  clusterUid: string,
  params: KubernetesPodLogsParams
): Promise<KubernetesPodLogsResult> {
  return get<KubernetesPodLogsResult>(
    `/cluster/${segment(clusterUid)}/namespace/${segment(params.namespace)}/pod/${segment(params.name)}/logs`,
    {
      container: params.container,
      tailLines: params.tailLines,
      previous: params.previous,
    }
  )
}

/** GET /v1/cluster/{clusterUid}/jobs */
export function listKubernetesJobs(
  clusterUid: string,
  params: ListKubernetesJobsParams = {}
): Promise<ListKubernetesJobsResult> {
  return get<ListKubernetesJobsResult>(
    `/cluster/${segment(clusterUid)}/jobs`,
    params
  )
}

/** GET /v1/cluster/{clusterUid}/job-detail */
export function getKubernetesJobDetail(
  clusterUid: string,
  params: GetKubernetesJobDetailParams
): Promise<GetKubernetesJobDetailResult> {
  return get<GetKubernetesJobDetailResult>(
    `/cluster/${segment(clusterUid)}/job-detail`,
    { namespace: params.namespace, name: params.name },
    undefined,
    { silent: true }
  )
}

/** DELETE /v1/cluster/{clusterUid}/namespace/{namespace}/job/{name} */
export function deleteKubernetesJob(
  clusterUid: string,
  namespace: string,
  name: string
): Promise<Record<string, never>> {
  return del<Record<string, never>>(
    `/cluster/${segment(clusterUid)}/namespace/${segment(namespace)}/job/${segment(name)}`
  )
}

/** GET /v1/cluster/{clusterUid}/cronjobs */
export function listKubernetesCronJobs(
  clusterUid: string,
  params: ListKubernetesCronJobsParams = {}
): Promise<ListKubernetesCronJobsResult> {
  return get<ListKubernetesCronJobsResult>(
    `/cluster/${segment(clusterUid)}/cronjobs`,
    params
  )
}

/** GET /v1/cluster/{clusterUid}/cronjob-detail */
export function getKubernetesCronJobDetail(
  clusterUid: string,
  params: GetKubernetesCronJobDetailParams
): Promise<GetKubernetesCronJobDetailResult> {
  return get<GetKubernetesCronJobDetailResult>(
    `/cluster/${segment(clusterUid)}/cronjob-detail`,
    { namespace: params.namespace, name: params.name },
    undefined,
    { silent: true }
  )
}

/** PUT /v1/cluster/{clusterUid}/cronjob/suspend */
export function setKubernetesCronJobSuspend(
  clusterUid: string,
  body: SetKubernetesCronJobSuspendParams
): Promise<{ cronjob: KubernetesCronJobItem }> {
  return put<{ cronjob: KubernetesCronJobItem }>(
    `/cluster/${segment(clusterUid)}/cronjob/suspend`,
    body
  )
}

/** POST /v1/cluster/{clusterUid}/cronjob/trigger */
export function triggerKubernetesCronJob(
  clusterUid: string,
  body: TriggerKubernetesCronJobParams
): Promise<{ job: KubernetesJobItem }> {
  return post<{ job: KubernetesJobItem }>(
    `/cluster/${segment(clusterUid)}/cronjob/trigger`,
    body
  )
}

/** DELETE /v1/cluster/{clusterUid}/namespace/{namespace}/cronjob/{name} */
export function deleteKubernetesCronJob(
  clusterUid: string,
  namespace: string,
  name: string
): Promise<Record<string, never>> {
  return del<Record<string, never>>(
    `/cluster/${segment(clusterUid)}/namespace/${segment(namespace)}/cronjob/${segment(name)}`
  )
}

/** GET /v1/cluster/{clusterUid}/ingresses */
export function listKubernetesIngresses(
  clusterUid: string,
  params: ListKubernetesIngressesParams = {}
): Promise<ListKubernetesIngressesResult> {
  return get<ListKubernetesIngressesResult>(
    `/cluster/${segment(clusterUid)}/ingresses`,
    params
  )
}

/** GET /v1/cluster/{clusterUid}/services */
export function listKubernetesServices(
  clusterUid: string,
  params: ListKubernetesServicesParams = {}
): Promise<ListKubernetesServicesResult> {
  return get<ListKubernetesServicesResult>(
    `/cluster/${segment(clusterUid)}/services`,
    params
  )
}

/** GET /v1/cluster/{clusterUid}/configmaps */
export function listKubernetesConfigMaps(
  clusterUid: string,
  params: ListKubernetesConfigMapsParams = {}
): Promise<ListKubernetesConfigMapsResult> {
  return get<ListKubernetesConfigMapsResult>(
    `/cluster/${segment(clusterUid)}/configmaps`,
    params
  )
}

/** GET /v1/cluster/{clusterUid}/secrets */
export function listKubernetesSecrets(
  clusterUid: string,
  params: ListKubernetesSecretsParams = {}
): Promise<ListKubernetesSecretsResult> {
  return get<ListKubernetesSecretsResult>(
    `/cluster/${segment(clusterUid)}/secrets`,
    params
  )
}

/** GET /v1/cluster/{clusterUid}/persistentvolumes */
export function listKubernetesPersistentVolumes(
  clusterUid: string,
  params: ListKubernetesPersistentVolumesParams = {}
): Promise<ListKubernetesPersistentVolumesResult> {
  return get<ListKubernetesPersistentVolumesResult>(
    `/cluster/${segment(clusterUid)}/persistentvolumes`,
    params
  )
}

/** GET /v1/cluster/{clusterUid}/persistentvolumeclaims */
export function listKubernetesPersistentVolumeClaims(
  clusterUid: string,
  params: ListKubernetesPersistentVolumeClaimsParams = {}
): Promise<ListKubernetesPersistentVolumeClaimsResult> {
  return get<ListKubernetesPersistentVolumeClaimsResult>(
    `/cluster/${segment(clusterUid)}/persistentvolumeclaims`,
    params
  )
}

/** GET /v1/cluster/{clusterUid}/yaml */
export function getKubernetesResourceYaml(
  clusterUid: string,
  params: GetKubernetesResourceYamlParams
): Promise<GetKubernetesResourceYamlResult> {
  return get<GetKubernetesResourceYamlResult>(
    `/cluster/${segment(clusterUid)}/yaml`,
    {
      kind: params.kind,
      namespace: params.namespace,
      name: params.name,
    },
    undefined,
    { silent: true }
  )
}

/** GET /v1/cluster/{clusterUid}/workload-detail */
export function getKubernetesWorkloadDetail(
  clusterUid: string,
  params: GetKubernetesWorkloadDetailParams
): Promise<GetKubernetesWorkloadDetailResult> {
  return get<GetKubernetesWorkloadDetailResult>(
    `/cluster/${segment(clusterUid)}/workload-detail`,
    { kind: params.kind, namespace: params.namespace, name: params.name },
    undefined,
    { silent: true }
  )
}

/** GET /v1/cluster/{clusterUid}/ingress-detail */
export function getKubernetesIngressDetail(
  clusterUid: string,
  params: GetKubernetesIngressDetailParams
): Promise<GetKubernetesIngressDetailResult> {
  return get<GetKubernetesIngressDetailResult>(
    `/cluster/${segment(clusterUid)}/ingress-detail`,
    { namespace: params.namespace, name: params.name },
    undefined,
    { silent: true }
  )
}

/** GET /v1/cluster/{clusterUid}/service-detail */
export function getKubernetesServiceDetail(
  clusterUid: string,
  params: GetKubernetesServiceDetailParams
): Promise<GetKubernetesServiceDetailResult> {
  return get<GetKubernetesServiceDetailResult>(
    `/cluster/${segment(clusterUid)}/service-detail`,
    { namespace: params.namespace, name: params.name },
    undefined,
    { silent: true }
  )
}

/** GET /v1/cluster/{clusterUid}/configmap-detail */
export function getKubernetesConfigMapDetail(
  clusterUid: string,
  params: GetKubernetesConfigMapDetailParams
): Promise<GetKubernetesConfigMapDetailResult> {
  return get<GetKubernetesConfigMapDetailResult>(
    `/cluster/${segment(clusterUid)}/configmap-detail`,
    { namespace: params.namespace, name: params.name },
    undefined,
    { silent: true }
  )
}

/** GET /v1/cluster/{clusterUid}/secret-detail */
export function getKubernetesSecretDetail(
  clusterUid: string,
  params: GetKubernetesSecretDetailParams
): Promise<GetKubernetesSecretDetailResult> {
  return get<GetKubernetesSecretDetailResult>(
    `/cluster/${segment(clusterUid)}/secret-detail`,
    { namespace: params.namespace, name: params.name },
    undefined,
    { silent: true }
  )
}

/** GET /v1/cluster/{clusterUid}/persistentvolumeclaim-detail */
export function getKubernetesPersistentVolumeClaimDetail(
  clusterUid: string,
  params: GetKubernetesPersistentVolumeClaimDetailParams
): Promise<GetKubernetesPersistentVolumeClaimDetailResult> {
  return get<GetKubernetesPersistentVolumeClaimDetailResult>(
    `/cluster/${segment(clusterUid)}/persistentvolumeclaim-detail`,
    { namespace: params.namespace, name: params.name },
    undefined,
    { silent: true }
  )
}

export type {
  ClusterItem,
  ClusterProbe,
  GetKubernetesConfigMapDetailParams,
  GetKubernetesConfigMapDetailResult,
  GetKubernetesCronJobDetailParams,
  GetKubernetesCronJobDetailResult,
  GetKubernetesIngressDetailParams,
  GetKubernetesIngressDetailResult,
  GetKubernetesJobDetailParams,
  GetKubernetesJobDetailResult,
  GetKubernetesPersistentVolumeClaimDetailParams,
  GetKubernetesPersistentVolumeClaimDetailResult,
  GetKubernetesResourceYamlParams,
  GetKubernetesResourceYamlResult,
  GetKubernetesSecretDetailParams,
  GetKubernetesSecretDetailResult,
  GetKubernetesServiceDetailParams,
  GetKubernetesServiceDetailResult,
  GetKubernetesWorkloadDetailParams,
  GetKubernetesWorkloadDetailResult,
  GetKubernetesWorkloadImagesParams,
  GetKubernetesWorkloadImagesResult,
  KubernetesConfigMapItem,
  KubernetesCronJobItem,
  KubernetesEndpointItem,
  KubernetesEndpointPortItem,
  KubernetesImageVersionItem,
  KubernetesIngressItem,
  KubernetesIngressPathItem,
  KubernetesJobItem,
  KubernetesJobStatus,
  KubernetesNodeItem,
  KubernetesPersistentVolumeClaimItem,
  KubernetesPersistentVolumeItem,
  KubernetesPodContainerItem,
  KubernetesPodItem,
  KubernetesPodLogsParams,
  KubernetesPodLogsResult,
  KubernetesReplicaSetItem,
  KubernetesResourceKind,
  KubernetesSecretItem,
  KubernetesServiceItem,
  KubernetesServicePortItem,
  KubernetesWorkloadImageItem,
  KubernetesWorkloadItem,
  KubernetesWorkloadKind,
  KubernetesWorkloadRefItem,
  KubernetesWorkloadRevisionItem,
  ListClustersParams,
  ListClustersResult,
  ListKubernetesConfigMapsParams,
  ListKubernetesConfigMapsResult,
  ListKubernetesCronJobsParams,
  ListKubernetesCronJobsResult,
  ListKubernetesIngressesParams,
  ListKubernetesIngressesResult,
  ListKubernetesJobsParams,
  ListKubernetesJobsResult,
  ListKubernetesNamespacesResult,
  ListKubernetesNodesParams,
  ListKubernetesNodesResult,
  ListKubernetesPersistentVolumeClaimsParams,
  ListKubernetesPersistentVolumeClaimsResult,
  ListKubernetesPersistentVolumesParams,
  ListKubernetesPersistentVolumesResult,
  ListKubernetesPodsParams,
  ListKubernetesPodsResult,
  ListKubernetesSecretsParams,
  ListKubernetesSecretsResult,
  ListKubernetesServicesParams,
  ListKubernetesServicesResult,
  ListKubernetesWorkloadRevisionsParams,
  ListKubernetesWorkloadRevisionsResult,
  ListKubernetesWorkloadsParams,
  ListKubernetesWorkloadsResult,
  ProbeKubeconfigResult,
  RestartKubernetesWorkloadParams,
  RollbackKubernetesWorkloadParams,
  SaveClusterParams,
  ScaleKubernetesWorkloadParams,
  SetKubernetesCronJobSuspendParams,
  SetKubernetesWorkloadImageParams,
  TestClusterResult,
  TriggerKubernetesCronJobParams,
  UpdateClusterParams,
} from "./types"
