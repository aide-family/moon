package bo

import "time"

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
// Manifest 是 YAML；Scope 取 NAMESPACED / CLUSTER / UNKNOWN。
type ApplicationResource struct {
	Item          string
	Manifest      string
	APIVersion    string
	Kind          string
	Name          string
	Namespace     string
	Scope         string
	ClusterScoped bool
	Images        []string
	SecretKeys    []string
}

// ApplicationVersion 是发布包版本。
type ApplicationVersion struct {
	UID            string
	NamespaceUID   string
	ApplicationUID string
	Version        string
	Status         string
	Changelog      string
	Resources      []ApplicationResource
	// SecretValues 只在写入时携带明文；读取时只返回 SecretKeys。
	SecretValues map[string]string
	Images       []ApplicationVersionImage
	Validation   ApplicationValidation
	PublishedAt  *time.Time
	CreatedAt    time.Time
}

// ApplicationVersionImage 是版本里扫描到的镜像快照。
type ApplicationVersionImage struct {
	UID            string
	NamespaceUID   string
	ApplicationUID string
	VersionUID     string
	ResourceItem   string
	Container      string
	Image          string
	RegistryUID    string
	RegistryName   string
	Repository     string
	Tag            string
	Digest         string
	Platforms      []string
	External       bool
	ResolvedAt     *time.Time
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
	UID            string
	NamespaceUID   string
	DeploymentUID  string
	ReleaseUID     string
	BatchUID       string
	ClusterUID     string
	ClusterName    string
	Action         string
	Status         string
	FromVersionUID string
	ToVersionUID   string
	Values         ApplicationDeployValues
	Resources      []ResourceRef
	Diff           map[string]int32
	Message        string
	OperatorUID    string
	OperatorName   string
	LeaseOwner     string
	LeaseUntil     *time.Time
	StartedAt      *time.Time
	FinishedAt     *time.Time
	CreatedAt      time.Time
}

// ResourceRef 是 Kubernetes 资源标识。
type ResourceRef struct {
	APIVersion string
	Kind       string
	Namespace  string
	Name       string
}
