package bo

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// Application 是一个应用（发布包身份）。
type Application struct {
	UID           string
	NamespaceUID  string
	Name          string
	DisplayName   string
	Category      string
	Description   string
	Readme        string
	VersionCount  int32
	LatestVersion string
	CreatedAt     time.Time
}

// ApplicationListFilter 是应用列表过滤条件。
type ApplicationListFilter struct {
	Name     string
	Category string
}

// ApplicationResource 是发布包里的一个 Kubernetes 资源。
// Manifest 是 YAML；Scope 取 NAMESPACED / CLUSTER / UNKNOWN。UID 是资源在版本内
// 的稳定标识；Component 是资源所属的服务（组件）名；ContentHash 是
// canonical manifest 的 sha256。
type ApplicationResource struct {
	UID           string
	Item          string
	Component     string
	Manifest      string
	APIVersion    string
	Kind          string
	Name          string
	Namespace     string
	Scope         string
	ClusterScoped bool
	Images        []string
	SecretKeys    []string
	ContentHash   string
	// ImageRefs 是清单里容器镜像的定位，落库时写入镜像快照的 container/path。
	ImageRefs []ApplicationImageRef
}

// ApplicationImageRef 是资源清单里一个容器镜像的定位。
type ApplicationImageRef struct {
	Image         string
	Container     string
	ContainerPath string
}

// ApplicationResourceChange 是资源级写入（新增/编辑）的入参。
// SecretValues 以 "item/key" 为键；ClearSecretKeys 是相对 item 的 key，
// 只从版本级密文里删除这些键，未提供的键保持原值。
type ApplicationResourceChange struct {
	NamespaceUID     string
	ApplicationUID   string
	VersionUID       string
	ResourceUID      string
	Resource         ApplicationResource
	SecretValues     map[string]string
	ClearSecretKeys  []string
	ExpectedRevision int32
}

// ApplicationVersionCopy 是复制版本的入参。
type ApplicationVersionCopy struct {
	NamespaceUID     string
	ApplicationUID   string
	SourceVersionUID string
	Version          string
	Changelog        string
}

// ApplicationVersionComponent 是版本内的一个服务（组件），由若干 Kubernetes
// 资源组成。资源变更只作用于 DRAFT 版本。
type ApplicationVersionComponent struct {
	UID         string
	Name        string
	DisplayName string
	// Description 是服务备注，支持 Markdown。
	Description   string
	SortOrder     int32
	Resources     []ApplicationResource
	ResourceCount int32
}

// ApplicationVersion 是发布包版本。
type ApplicationVersion struct {
	UID            string
	NamespaceUID   string
	ApplicationUID string
	Version        string
	Status         string
	Changelog      string
	// Resources 是扁平视图，按组件顺序排列，保留给旧客户端。
	Resources  []ApplicationResource
	Components []ApplicationVersionComponent
	// SecretValues 只在写入时携带明文；读取时只返回 SecretKeys。
	SecretValues map[string]string
	Images       []ApplicationVersionImage
	Validation   ApplicationValidation
	// Revision 是版本内容的乐观锁版本号，每次内容变更 +1。
	Revision int32
	// CompositionVersion 为 1 时组件与资源来自归一化表，否则回退 spec。
	CompositionVersion   int32
	AllowedActions       []string
	ComponentCount       int32
	ResourceCount        int32
	PublishedAt          *time.Time
	StatusChangedAt      *time.Time
	DeprecatedAt         *time.Time
	CopiedFromVersionUID string
	CreatedAt            time.Time
}

// ApplicationVersionImage 是版本里扫描到的镜像快照。
type ApplicationVersionImage struct {
	UID            string
	NamespaceUID   string
	ApplicationUID string
	VersionUID     string
	ResourceItem   string
	ResourceUID    string
	Component      string
	Container      string
	ContainerPath  string
	Image          string
	RegistryUID    string
	RegistryName   string
	Repository     string
	Tag            string
	Digest         string
	Project        string
	Source         string
	ResolvedImage  string
	Platforms      []string
	External       bool
	ResolvedAt     *time.Time
}

// ApplicationManifestHash 返回 canonical manifest 的 sha256，用于并发编辑提示。
func ApplicationManifestHash(manifest string) string {
	sum := sha256.Sum256([]byte(manifest))
	return hex.EncodeToString(sum[:])
}

// ApplicationValidation 是版本校验结果。
type ApplicationValidation struct {
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

// ApplicationVersionFilter 是版本列表过滤条件。
type ApplicationVersionFilter struct {
	Status string
}

// ApplicationDeployValues 是部署时允许的参数覆盖。
type ApplicationDeployValues struct {
	Replicas     map[string]int32
	ImageTags    map[string]string
	IngressHosts map[string]string
	SecretValues map[string]string
}

// DeploymentCluster 是部署组里的一个集群快照。
type DeploymentCluster struct {
	UID        string
	Name       string
	Status     string
	ReleaseUID string
	Version    string
	Message    string
	// CurrentVersionUID 是集群最后一次成功发布的版本；InSync 表示它与部署组
	// 期望版本一致。
	CurrentVersionUID string
	CurrentVersion    string
	InSync            bool
}

// ApplicationDeployment 是一次多集群部署的聚合。
type ApplicationDeployment struct {
	UID             string
	NamespaceUID    string
	Name            string
	ApplicationUID  string
	ApplicationName string
	VersionUID      string
	Version         string
	TargetNamespace string
	Clusters        []DeploymentCluster
	Values          ApplicationDeployValues
	Images          []ApplicationVersionImage
	Status          string
	Message         string
	CreatedBy       string
	CreatedAt       time.Time
	DeployedAt      *time.Time
	Releases        []ApplicationRelease
}

// ApplicationRelease 是部署组在一个集群中的实例。
type ApplicationRelease struct {
	UID              string
	NamespaceUID     string
	DeploymentUID    string
	ClusterUID       string
	ClusterName      string
	TargetNamespace  string
	Name             string
	ApplicationUID   string
	ApplicationName  string
	VersionUID       string
	Version          string
	Status           string
	Values           ApplicationDeployValues
	Images           []ApplicationVersionImage
	Message          string
	LastOperationUID string
	DeployedAt       *time.Time
	CreatedAt        time.Time
}

// ApplicationReleaseFilter 是部署组列表过滤条件。
type ApplicationReleaseFilter struct {
	Name           string
	Status         string
	ApplicationUID string
	ClusterUID     string
}

// ApplicationOperation 是某个集群实例的一次操作。
type ApplicationOperation struct {
	UID                 string
	NamespaceUID        string
	ApplicationUID      string
	DeploymentUID       string
	ReleaseUID          string
	BatchUID            string
	ClusterUID          string
	ClusterName         string
	Action              string
	Status              string
	FromVersionUID      string
	ToVersionUID        string
	Values              ApplicationDeployValues
	Resources           []ResourceRef
	Diff                map[string]int32
	Message             string
	OperatorUID         string
	OperatorName        string
	KeepPVC             bool
	DeleteClusterScoped bool
	LeaseOwner          string
	LeaseUntil          *time.Time
	StartedAt           *time.Time
	FinishedAt          *time.Time
	CreatedAt           time.Time
}

// ApplicationImageReference 是发布包版本对镜像的引用。
type ApplicationImageReference struct {
	ApplicationUID  string
	ApplicationName string
	VersionUID      string
	Version         string
	Status          string
}

// ResourceRef 是 Kubernetes 资源标识。
type ResourceRef struct {
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
}
