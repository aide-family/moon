package service

import (
	"context"

	v1 "github.com/aide-family/moon/api/alioth/v1"
	"github.com/aide-family/moon/app/alioth/internal/biz/cluster"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// KubernetesService 把集群资源操作转成用例入参。只读操作走 Explorer，
// 改调度、改副本、重启和删 Pod 走 Operator。
type KubernetesService struct {
	v1.UnimplementedKubernetesServiceServer
	explorer *cluster.Explorer
	operator *cluster.Operator
}

func NewKubernetesService(explorer *cluster.Explorer, operator *cluster.Operator) *KubernetesService {
	return &KubernetesService{explorer: explorer, operator: operator}
}

func (s *KubernetesService) ListKubernetesNamespaces(ctx context.Context, req *v1.ListKubernetesNamespacesRequest) (*v1.ListKubernetesNamespacesResponse, error) {
	names, err := s.explorer.ListNamespaces(ctx, req.GetClusterUid())
	if err != nil {
		return nil, err
	}
	return &v1.ListKubernetesNamespacesResponse{Items: names}, nil
}

func (s *KubernetesService) ListKubernetesNodes(ctx context.Context, req *v1.ListKubernetesNodesRequest) (*v1.ListKubernetesNodesResponse, error) {
	items, total, err := s.explorer.ListNodes(ctx, req.GetClusterUid(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListKubernetesNodesResponse{Items: kubeNodeItems(items), Total: int32(total)}, nil
}

func (s *KubernetesService) SetKubernetesNodeUnschedulable(ctx context.Context, req *v1.SetKubernetesNodeUnschedulableRequest) (*v1.SetKubernetesNodeUnschedulableResponse, error) {
	if err := s.operator.SetNodeUnschedulable(ctx, req.GetClusterUid(), req.GetName(), req.GetUnschedulable()); err != nil {
		return nil, err
	}
	return &v1.SetKubernetesNodeUnschedulableResponse{}, nil
}

func (s *KubernetesService) ListKubernetesWorkloads(ctx context.Context, req *v1.ListKubernetesWorkloadsRequest) (*v1.ListKubernetesWorkloadsResponse, error) {
	items, total, err := s.explorer.ListWorkloads(ctx, req.GetClusterUid(), req.GetNamespace(), req.GetKind(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListKubernetesWorkloadsResponse{Items: kubeWorkloadItems(items), Total: int32(total)}, nil
}

func (s *KubernetesService) ScaleKubernetesWorkload(ctx context.Context, req *v1.ScaleKubernetesWorkloadRequest) (*v1.ScaleKubernetesWorkloadResponse, error) {
	item, err := s.operator.ScaleWorkload(ctx, req.GetClusterUid(), req.GetNamespace(), req.GetKind(), req.GetName(), req.GetReplicas())
	if err != nil {
		return nil, err
	}
	return &v1.ScaleKubernetesWorkloadResponse{Workload: kubeWorkloadItem(*item)}, nil
}

func (s *KubernetesService) RestartKubernetesWorkload(ctx context.Context, req *v1.RestartKubernetesWorkloadRequest) (*v1.RestartKubernetesWorkloadResponse, error) {
	item, err := s.operator.RestartWorkload(ctx, req.GetClusterUid(), req.GetNamespace(), req.GetKind(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.RestartKubernetesWorkloadResponse{Workload: kubeWorkloadItem(*item)}, nil
}

func (s *KubernetesService) ListKubernetesPods(ctx context.Context, req *v1.ListKubernetesPodsRequest) (*v1.ListKubernetesPodsResponse, error) {
	items, total, page, pageSize, err := s.explorer.ListPods(
		ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName(),
		req.GetOwnerKind(), req.GetOwnerName(), req.GetPage(), req.GetPageSize(),
	)
	if err != nil {
		return nil, err
	}
	return &v1.ListKubernetesPodsResponse{
		Items: kubePodItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize),
	}, nil
}

func (s *KubernetesService) DeleteKubernetesPod(ctx context.Context, req *v1.DeleteKubernetesPodRequest) (*v1.DeleteKubernetesPodResponse, error) {
	if err := s.operator.DeletePod(ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName()); err != nil {
		return nil, err
	}
	return &v1.DeleteKubernetesPodResponse{}, nil
}

func (s *KubernetesService) GetKubernetesPodLogs(ctx context.Context, req *v1.GetKubernetesPodLogsRequest) (*v1.GetKubernetesPodLogsResponse, error) {
	logs, err := s.explorer.PodLogs(
		ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName(),
		req.GetContainer(), req.GetTailLines(), req.GetPrevious(),
	)
	if err != nil {
		return nil, err
	}
	return &v1.GetKubernetesPodLogsResponse{Content: logs.Content, Container: logs.Container}, nil
}

func (s *KubernetesService) ListKubernetesIngresses(ctx context.Context, req *v1.ListKubernetesIngressesRequest) (*v1.ListKubernetesIngressesResponse, error) {
	items, total, page, pageSize, err := s.explorer.ListIngresses(
		ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName(), req.GetPage(), req.GetPageSize(),
	)
	if err != nil {
		return nil, err
	}
	return &v1.ListKubernetesIngressesResponse{
		Items: kubeIngressItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize),
	}, nil
}

func (s *KubernetesService) ListKubernetesServices(ctx context.Context, req *v1.ListKubernetesServicesRequest) (*v1.ListKubernetesServicesResponse, error) {
	items, total, page, pageSize, err := s.explorer.ListServices(
		ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName(), req.GetPage(), req.GetPageSize(),
	)
	if err != nil {
		return nil, err
	}
	return &v1.ListKubernetesServicesResponse{
		Items: kubeServiceItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize),
	}, nil
}

func (s *KubernetesService) ListKubernetesConfigMaps(ctx context.Context, req *v1.ListKubernetesConfigMapsRequest) (*v1.ListKubernetesConfigMapsResponse, error) {
	items, total, page, pageSize, err := s.explorer.ListConfigMaps(
		ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName(), req.GetPage(), req.GetPageSize(),
	)
	if err != nil {
		return nil, err
	}
	return &v1.ListKubernetesConfigMapsResponse{
		Items: kubeConfigMapItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize),
	}, nil
}

func (s *KubernetesService) ListKubernetesSecrets(ctx context.Context, req *v1.ListKubernetesSecretsRequest) (*v1.ListKubernetesSecretsResponse, error) {
	items, total, page, pageSize, err := s.explorer.ListSecrets(
		ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName(), req.GetPage(), req.GetPageSize(),
	)
	if err != nil {
		return nil, err
	}
	return &v1.ListKubernetesSecretsResponse{
		Items: kubeSecretItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize),
	}, nil
}

func (s *KubernetesService) ListKubernetesPersistentVolumes(ctx context.Context, req *v1.ListKubernetesPersistentVolumesRequest) (*v1.ListKubernetesPersistentVolumesResponse, error) {
	items, total, page, pageSize, err := s.explorer.ListPersistentVolumes(
		ctx, req.GetClusterUid(), req.GetName(), req.GetPage(), req.GetPageSize(),
	)
	if err != nil {
		return nil, err
	}
	return &v1.ListKubernetesPersistentVolumesResponse{
		Items: kubePersistentVolumeItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize),
	}, nil
}

func (s *KubernetesService) ListKubernetesPersistentVolumeClaims(ctx context.Context, req *v1.ListKubernetesPersistentVolumeClaimsRequest) (*v1.ListKubernetesPersistentVolumeClaimsResponse, error) {
	items, total, page, pageSize, err := s.explorer.ListPersistentVolumeClaims(
		ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName(), req.GetPage(), req.GetPageSize(),
	)
	if err != nil {
		return nil, err
	}
	return &v1.ListKubernetesPersistentVolumeClaimsResponse{
		Items: kubePersistentVolumeClaimItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize),
	}, nil
}

func (s *KubernetesService) GetKubernetesResourceYAML(ctx context.Context, req *v1.GetKubernetesResourceYAMLRequest) (*v1.GetKubernetesResourceYAMLResponse, error) {
	item, err := s.explorer.ResourceYAML(ctx, req.GetClusterUid(), req.GetKind(), req.GetNamespace(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.GetKubernetesResourceYAMLResponse{Yaml: item.YAML, Redacted: item.Redacted}, nil
}

func (s *KubernetesService) GetKubernetesWorkloadDetail(ctx context.Context, req *v1.GetKubernetesWorkloadDetailRequest) (*v1.GetKubernetesWorkloadDetailResponse, error) {
	detail, err := s.explorer.WorkloadDetail(ctx, req.GetClusterUid(), req.GetKind(), req.GetNamespace(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.GetKubernetesWorkloadDetailResponse{
		Workload:    kubeWorkloadItem(detail.Workload),
		Replicasets: kubeReplicaSetItems(detail.ReplicaSets),
		Pods:        kubePodItems(detail.Pods),
	}, nil
}

func (s *KubernetesService) GetKubernetesIngressDetail(ctx context.Context, req *v1.GetKubernetesIngressDetailRequest) (*v1.GetKubernetesIngressDetailResponse, error) {
	detail, err := s.explorer.IngressDetail(ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.GetKubernetesIngressDetailResponse{
		Ingress:  kubeIngressItem(detail.Ingress),
		Services: kubeServiceItems(detail.Services),
	}, nil
}

func (s *KubernetesService) GetKubernetesServiceDetail(ctx context.Context, req *v1.GetKubernetesServiceDetailRequest) (*v1.GetKubernetesServiceDetailResponse, error) {
	detail, err := s.explorer.ServiceDetail(ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.GetKubernetesServiceDetailResponse{
		Service:   kubeServiceItem(detail.Service),
		Endpoints: kubeEndpointItems(detail.Endpoints),
	}, nil
}

func (s *KubernetesService) GetKubernetesConfigMapDetail(ctx context.Context, req *v1.GetKubernetesConfigMapDetailRequest) (*v1.GetKubernetesConfigMapDetailResponse, error) {
	detail, err := s.explorer.ConfigMapDetail(ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.GetKubernetesConfigMapDetailResponse{
		Configmap:        kubeConfigMapItem(detail.ConfigMap),
		RelatedWorkloads: kubeWorkloadRefItems(detail.RelatedWorkloads),
	}, nil
}

func (s *KubernetesService) GetKubernetesSecretDetail(ctx context.Context, req *v1.GetKubernetesSecretDetailRequest) (*v1.GetKubernetesSecretDetailResponse, error) {
	detail, err := s.explorer.SecretDetail(ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.GetKubernetesSecretDetailResponse{
		Secret:           kubeSecretItem(detail.Secret),
		RelatedWorkloads: kubeWorkloadRefItems(detail.RelatedWorkloads),
	}, nil
}

func (s *KubernetesService) GetKubernetesPersistentVolumeClaimDetail(ctx context.Context, req *v1.GetKubernetesPersistentVolumeClaimDetailRequest) (*v1.GetKubernetesPersistentVolumeClaimDetailResponse, error) {
	detail, err := s.explorer.PersistentVolumeClaimDetail(ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.GetKubernetesPersistentVolumeClaimDetailResponse{
		Claim:            kubePersistentVolumeClaimItem(detail.Claim),
		RelatedWorkloads: kubeWorkloadRefItems(detail.RelatedWorkloads),
	}, nil
}

func (s *KubernetesService) GetKubernetesWorkloadImages(ctx context.Context, req *v1.GetKubernetesWorkloadImagesRequest) (*v1.GetKubernetesWorkloadImagesResponse, error) {
	items, err := s.explorer.WorkloadImages(ctx, req.GetClusterUid(), req.GetNamespace(), req.GetKind(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.GetKubernetesWorkloadImagesResponse{Images: kubeWorkloadImageItems(items)}, nil
}

func (s *KubernetesService) SetKubernetesWorkloadImage(ctx context.Context, req *v1.SetKubernetesWorkloadImageRequest) (*v1.SetKubernetesWorkloadImageResponse, error) {
	item, err := s.operator.SetWorkloadImage(
		ctx, req.GetClusterUid(), req.GetNamespace(), req.GetKind(), req.GetName(),
		req.GetContainer(), req.GetInit(), req.GetImage(),
	)
	if err != nil {
		return nil, err
	}
	return &v1.SetKubernetesWorkloadImageResponse{Workload: kubeWorkloadItem(*item)}, nil
}

func (s *KubernetesService) ListKubernetesWorkloadRevisions(ctx context.Context, req *v1.ListKubernetesWorkloadRevisionsRequest) (*v1.ListKubernetesWorkloadRevisionsResponse, error) {
	items, err := s.explorer.WorkloadRevisions(ctx, req.GetClusterUid(), req.GetNamespace(), req.GetKind(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.ListKubernetesWorkloadRevisionsResponse{Revisions: kubeWorkloadRevisionItems(items)}, nil
}

func (s *KubernetesService) RollbackKubernetesWorkload(ctx context.Context, req *v1.RollbackKubernetesWorkloadRequest) (*v1.RollbackKubernetesWorkloadResponse, error) {
	item, err := s.operator.RollbackWorkload(
		ctx, req.GetClusterUid(), req.GetNamespace(), req.GetKind(), req.GetName(), req.GetRevision(),
	)
	if err != nil {
		return nil, err
	}
	return &v1.RollbackKubernetesWorkloadResponse{Workload: kubeWorkloadItem(*item)}, nil
}

func (s *KubernetesService) ListKubernetesJobs(ctx context.Context, req *v1.ListKubernetesJobsRequest) (*v1.ListKubernetesJobsResponse, error) {
	items, total, page, pageSize, err := s.explorer.ListJobs(
		ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName(), req.GetPage(), req.GetPageSize(),
	)
	if err != nil {
		return nil, err
	}
	return &v1.ListKubernetesJobsResponse{
		Items: kubeJobItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize),
	}, nil
}

func (s *KubernetesService) GetKubernetesJobDetail(ctx context.Context, req *v1.GetKubernetesJobDetailRequest) (*v1.GetKubernetesJobDetailResponse, error) {
	detail, err := s.explorer.JobDetail(ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.GetKubernetesJobDetailResponse{
		Job:  kubeJobItem(detail.Job),
		Pods: kubePodItems(detail.Pods),
	}, nil
}

func (s *KubernetesService) DeleteKubernetesJob(ctx context.Context, req *v1.DeleteKubernetesJobRequest) (*v1.DeleteKubernetesJobResponse, error) {
	if err := s.operator.DeleteJob(ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName()); err != nil {
		return nil, err
	}
	return &v1.DeleteKubernetesJobResponse{}, nil
}

func (s *KubernetesService) ListKubernetesCronJobs(ctx context.Context, req *v1.ListKubernetesCronJobsRequest) (*v1.ListKubernetesCronJobsResponse, error) {
	items, total, page, pageSize, err := s.explorer.ListCronJobs(
		ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName(), req.GetPage(), req.GetPageSize(),
	)
	if err != nil {
		return nil, err
	}
	return &v1.ListKubernetesCronJobsResponse{
		Items: kubeCronJobItems(items), Total: int32(total), Page: int32(page), PageSize: int32(pageSize),
	}, nil
}

func (s *KubernetesService) GetKubernetesCronJobDetail(ctx context.Context, req *v1.GetKubernetesCronJobDetailRequest) (*v1.GetKubernetesCronJobDetailResponse, error) {
	detail, err := s.explorer.CronJobDetail(ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.GetKubernetesCronJobDetailResponse{
		Cronjob: kubeCronJobItem(detail.CronJob),
		Jobs:    kubeJobItems(detail.Jobs),
	}, nil
}

func (s *KubernetesService) SetKubernetesCronJobSuspend(ctx context.Context, req *v1.SetKubernetesCronJobSuspendRequest) (*v1.SetKubernetesCronJobSuspendResponse, error) {
	item, err := s.operator.SetCronJobSuspend(ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName(), req.GetSuspend())
	if err != nil {
		return nil, err
	}
	return &v1.SetKubernetesCronJobSuspendResponse{Cronjob: kubeCronJobItem(*item)}, nil
}

func (s *KubernetesService) TriggerKubernetesCronJob(ctx context.Context, req *v1.TriggerKubernetesCronJobRequest) (*v1.TriggerKubernetesCronJobResponse, error) {
	item, err := s.operator.TriggerCronJob(ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName())
	if err != nil {
		return nil, err
	}
	return &v1.TriggerKubernetesCronJobResponse{Job: kubeJobItem(*item)}, nil
}

func (s *KubernetesService) DeleteKubernetesCronJob(ctx context.Context, req *v1.DeleteKubernetesCronJobRequest) (*v1.DeleteKubernetesCronJobResponse, error) {
	if err := s.operator.DeleteCronJob(ctx, req.GetClusterUid(), req.GetNamespace(), req.GetName()); err != nil {
		return nil, err
	}
	return &v1.DeleteKubernetesCronJobResponse{}, nil
}

func kubeNodeItems(items []bo.KubeNode) []*v1.KubernetesNodeItem {
	out := make([]*v1.KubernetesNodeItem, 0, len(items))
	for _, item := range items {
		out = append(out, &v1.KubernetesNodeItem{
			Name:           item.Name,
			Status:         item.Status,
			Schedulable:    item.Schedulable,
			Roles:          item.Roles,
			InternalIp:     item.InternalIP,
			KubeletVersion: item.KubeletVersion,
			OsImage:        item.OSImage,
			CpuCapacity:    item.CPUCapacity,
			MemoryCapacity: item.MemoryCapacity,
			PodCount:       item.PodCount,
			CreatedAt:      item.CreatedAt,
		})
	}
	return out
}

func kubeWorkloadItem(in bo.KubeWorkload) *v1.KubernetesWorkloadItem {
	return &v1.KubernetesWorkloadItem{
		Namespace:         in.Namespace,
		Kind:              in.Kind,
		Name:              in.Name,
		DesiredReplicas:   in.DesiredReplicas,
		ReadyReplicas:     in.ReadyReplicas,
		UpdatedReplicas:   in.UpdatedReplicas,
		AvailableReplicas: in.AvailableReplicas,
		Images:            in.Images,
		CreatedAt:         in.CreatedAt,
	}
}

func kubeWorkloadItems(items []bo.KubeWorkload) []*v1.KubernetesWorkloadItem {
	out := make([]*v1.KubernetesWorkloadItem, 0, len(items))
	for _, item := range items {
		out = append(out, kubeWorkloadItem(item))
	}
	return out
}

func kubePodItem(in bo.KubePod) *v1.KubernetesPodItem {
	containers := make([]*v1.KubernetesContainerItem, 0, len(in.Containers))
	for _, container := range in.Containers {
		containers = append(containers, &v1.KubernetesContainerItem{
			Name:     container.Name,
			Ready:    container.Ready,
			Restarts: container.Restarts,
			Image:    container.Image,
		})
	}
	return &v1.KubernetesPodItem{
		Namespace:  in.Namespace,
		Name:       in.Name,
		Phase:      in.Phase,
		Ready:      in.Ready,
		Restarts:   in.Restarts,
		NodeName:   in.NodeName,
		PodIp:      in.PodIP,
		OwnerKind:  in.OwnerKind,
		OwnerName:  in.OwnerName,
		Containers: containers,
		CreatedAt:  in.CreatedAt,
	}
}

func kubePodItems(items []bo.KubePod) []*v1.KubernetesPodItem {
	out := make([]*v1.KubernetesPodItem, 0, len(items))
	for _, item := range items {
		out = append(out, kubePodItem(item))
	}
	return out
}

func kubeIngressItem(in bo.KubeIngress) *v1.KubernetesIngressItem {
	paths := make([]*v1.KubernetesIngressPathItem, 0, len(in.Paths))
	for _, path := range in.Paths {
		paths = append(paths, &v1.KubernetesIngressPathItem{
			Host:        path.Host,
			Path:        path.Path,
			PathType:    path.PathType,
			ServiceName: path.ServiceName,
			ServicePort: path.ServicePort,
		})
	}
	return &v1.KubernetesIngressItem{
		Namespace: in.Namespace,
		Name:      in.Name,
		ClassName: in.ClassName,
		Hosts:     in.Hosts,
		Paths:     paths,
		Tls:       in.TLS,
		CreatedAt: in.CreatedAt,
	}
}

func kubeIngressItems(items []bo.KubeIngress) []*v1.KubernetesIngressItem {
	out := make([]*v1.KubernetesIngressItem, 0, len(items))
	for _, item := range items {
		out = append(out, kubeIngressItem(item))
	}
	return out
}

func kubeServiceItem(in bo.KubeService) *v1.KubernetesServiceItem {
	ports := make([]*v1.KubernetesServicePortItem, 0, len(in.Ports))
	for _, port := range in.Ports {
		ports = append(ports, &v1.KubernetesServicePortItem{
			Name:       port.Name,
			Protocol:   port.Protocol,
			Port:       port.Port,
			TargetPort: port.TargetPort,
			NodePort:   port.NodePort,
		})
	}
	return &v1.KubernetesServiceItem{
		Namespace:      in.Namespace,
		Name:           in.Name,
		Type:           in.Type,
		ClusterIp:      in.ClusterIP,
		ExternalIps:    in.ExternalIPs,
		LoadBalancerIp: in.LoadBalancerIP,
		Ports:          ports,
		Selector:       in.Selector,
		CreatedAt:      in.CreatedAt,
	}
}

func kubeServiceItems(items []bo.KubeService) []*v1.KubernetesServiceItem {
	out := make([]*v1.KubernetesServiceItem, 0, len(items))
	for _, item := range items {
		out = append(out, kubeServiceItem(item))
	}
	return out
}

func kubeConfigMapItem(in bo.KubeConfigMap) *v1.KubernetesConfigMapItem {
	return &v1.KubernetesConfigMapItem{
		Namespace:      in.Namespace,
		Name:           in.Name,
		DataKeys:       in.DataKeys,
		BinaryDataKeys: in.BinaryDataKeys,
		Immutable:      in.Immutable,
		CreatedAt:      in.CreatedAt,
	}
}

func kubeConfigMapItems(items []bo.KubeConfigMap) []*v1.KubernetesConfigMapItem {
	out := make([]*v1.KubernetesConfigMapItem, 0, len(items))
	for _, item := range items {
		out = append(out, kubeConfigMapItem(item))
	}
	return out
}

func kubeSecretItem(in bo.KubeSecret) *v1.KubernetesSecretItem {
	return &v1.KubernetesSecretItem{
		Namespace: in.Namespace,
		Name:      in.Name,
		Type:      in.Type,
		DataKeys:  in.DataKeys,
		Immutable: in.Immutable,
		CreatedAt: in.CreatedAt,
	}
}

func kubeSecretItems(items []bo.KubeSecret) []*v1.KubernetesSecretItem {
	out := make([]*v1.KubernetesSecretItem, 0, len(items))
	for _, item := range items {
		out = append(out, kubeSecretItem(item))
	}
	return out
}

func kubePersistentVolumeItems(items []bo.KubePersistentVolume) []*v1.KubernetesPersistentVolumeItem {
	out := make([]*v1.KubernetesPersistentVolumeItem, 0, len(items))
	for _, item := range items {
		out = append(out, &v1.KubernetesPersistentVolumeItem{
			Name:          item.Name,
			Capacity:      item.Capacity,
			StorageClass:  item.StorageClass,
			AccessModes:   item.AccessModes,
			Claim:         item.Claim,
			Phase:         item.Phase,
			ReclaimPolicy: item.ReclaimPolicy,
			CreatedAt:     item.CreatedAt,
		})
	}
	return out
}

func kubePersistentVolumeClaimItem(in bo.KubePersistentVolumeClaim) *v1.KubernetesPersistentVolumeClaimItem {
	return &v1.KubernetesPersistentVolumeClaimItem{
		Namespace:    in.Namespace,
		Name:         in.Name,
		Capacity:     in.Capacity,
		StorageClass: in.StorageClass,
		AccessModes:  in.AccessModes,
		VolumeName:   in.VolumeName,
		Phase:        in.Phase,
		CreatedAt:    in.CreatedAt,
	}
}

func kubePersistentVolumeClaimItems(items []bo.KubePersistentVolumeClaim) []*v1.KubernetesPersistentVolumeClaimItem {
	out := make([]*v1.KubernetesPersistentVolumeClaimItem, 0, len(items))
	for _, item := range items {
		out = append(out, kubePersistentVolumeClaimItem(item))
	}
	return out
}

func kubeReplicaSetItems(items []bo.KubeReplicaSet) []*v1.KubernetesReplicaSetItem {
	out := make([]*v1.KubernetesReplicaSetItem, 0, len(items))
	for _, item := range items {
		out = append(out, &v1.KubernetesReplicaSetItem{
			Namespace:       item.Namespace,
			Name:            item.Name,
			DesiredReplicas: item.DesiredReplicas,
			ReadyReplicas:   item.ReadyReplicas,
			CurrentReplicas: item.CurrentReplicas,
			Images:          item.Images,
			CreatedAt:       item.CreatedAt,
		})
	}
	return out
}

func kubeEndpointItems(items []bo.KubeEndpoint) []*v1.KubernetesEndpointItem {
	out := make([]*v1.KubernetesEndpointItem, 0, len(items))
	for _, item := range items {
		ports := make([]*v1.KubernetesEndpointPortItem, 0, len(item.Ports))
		for _, port := range item.Ports {
			ports = append(ports, &v1.KubernetesEndpointPortItem{
				Name:     port.Name,
				Port:     port.Port,
				Protocol: port.Protocol,
			})
		}
		out = append(out, &v1.KubernetesEndpointItem{
			Address:    item.Address,
			Hostname:   item.Hostname,
			NodeName:   item.NodeName,
			TargetKind: item.TargetKind,
			TargetName: item.TargetName,
			Ready:      item.Ready,
			Ports:      ports,
		})
	}
	return out
}

func kubeWorkloadRefItems(items []bo.KubeWorkloadRef) []*v1.KubernetesWorkloadRefItem {
	out := make([]*v1.KubernetesWorkloadRefItem, 0, len(items))
	for _, item := range items {
		out = append(out, &v1.KubernetesWorkloadRefItem{
			Kind:      item.Kind,
			Namespace: item.Namespace,
			Name:      item.Name,
		})
	}
	return out
}

func kubeWorkloadImageItem(in bo.KubeWorkloadImage) *v1.KubernetesWorkloadImageItem {
	versions := make([]*v1.KubernetesImageVersionItem, 0, len(in.Versions))
	for _, version := range in.Versions {
		versions = append(versions, &v1.KubernetesImageVersionItem{
			Image:     version.Image,
			Revision:  version.Revision,
			CreatedAt: version.CreatedAt,
		})
	}
	return &v1.KubernetesWorkloadImageItem{
		Container:  in.Container,
		Init:       in.Init,
		Image:      in.Image,
		Repository: in.Repository,
		Tag:        in.Tag,
		Versions:   versions,
	}
}

func kubeWorkloadImageItems(items []bo.KubeWorkloadImage) []*v1.KubernetesWorkloadImageItem {
	out := make([]*v1.KubernetesWorkloadImageItem, 0, len(items))
	for _, item := range items {
		out = append(out, kubeWorkloadImageItem(item))
	}
	return out
}

func kubeWorkloadRevisionItems(items []bo.KubeWorkloadRevision) []*v1.KubernetesWorkloadRevisionItem {
	out := make([]*v1.KubernetesWorkloadRevisionItem, 0, len(items))
	for _, item := range items {
		out = append(out, &v1.KubernetesWorkloadRevisionItem{
			Revision:  item.Revision,
			Name:      item.Name,
			Current:   item.Current,
			Replicas:  item.Replicas,
			Images:    item.Images,
			CreatedAt: item.CreatedAt,
		})
	}
	return out
}

func kubeJobItem(in bo.KubeJob) *v1.KubernetesJobItem {
	return &v1.KubernetesJobItem{
		Namespace:      in.Namespace,
		Name:           in.Name,
		Status:         in.Status,
		Completions:    in.Completions,
		Parallelism:    in.Parallelism,
		Active:         in.Active,
		Succeeded:      in.Succeeded,
		Failed:         in.Failed,
		OwnerKind:      in.OwnerKind,
		OwnerName:      in.OwnerName,
		StartTime:      in.StartTime,
		CompletionTime: in.CompletionTime,
		CreatedAt:      in.CreatedAt,
		Images:         in.Images,
	}
}

func kubeJobItems(items []bo.KubeJob) []*v1.KubernetesJobItem {
	out := make([]*v1.KubernetesJobItem, 0, len(items))
	for _, item := range items {
		out = append(out, kubeJobItem(item))
	}
	return out
}

func kubeCronJobItem(in bo.KubeCronJob) *v1.KubernetesCronJobItem {
	return &v1.KubernetesCronJobItem{
		Namespace:         in.Namespace,
		Name:              in.Name,
		Schedule:          in.Schedule,
		TimeZone:          in.TimeZone,
		Suspend:           in.Suspend,
		ConcurrencyPolicy: in.ConcurrencyPolicy,
		LastScheduleTime:  in.LastScheduleTime,
		ActiveCount:       in.ActiveCount,
		Images:            in.Images,
		CreatedAt:         in.CreatedAt,
	}
}

func kubeCronJobItems(items []bo.KubeCronJob) []*v1.KubernetesCronJobItem {
	out := make([]*v1.KubernetesCronJobItem, 0, len(items))
	for _, item := range items {
		out = append(out, kubeCronJobItem(item))
	}
	return out
}
