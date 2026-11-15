package bo

import "time"

// Registry 是一个租户接入的镜像仓库连接。密码只以密文存在于数据层，
// 这个对象只携带展示元数据。
type Registry struct {
	UID              string
	NamespaceUID     string
	Name             string
	Driver           string
	Endpoint         string
	Username         string
	AllowInsecureTLS bool
	Remark           string
	LastCheckedAt    int64
	LastError        string
	CreatedAt        time.Time
}

// RegistryProbe 是一次仓库连通性探测结果。
type RegistryProbe struct {
	Endpoint        string
	Driver          string
	RepositoryCount int32
}

// RegistryProject 是 Harbor 项目或 repository 路径前缀。
type RegistryProject struct {
	Name            string
	RepositoryCount int32
}

// RegistryRepository 是远端仓库中的一个 repository。
type RegistryRepository struct {
	Name      string
	Project   string
	TagCount  int32
	UpdatedAt int64
}

// RegistryPlatform 是多架构 index 中的一个平台。
type RegistryPlatform struct {
	OS           string
	Architecture string
	Variant      string
	Digest       string
	Size         int64
}

// RegistryTag 是镜像仓库中的一个 Tag。
type RegistryTag struct {
	Tag            string
	Digest         string
	Size           int64
	Platforms      []string
	PushedAt       int64
	ReferenceCount int32
}

// RegistryTagReference 是引用该镜像的发布包版本。
// M35 还没有应用商店，返回空；M36 接入 application_version_images 后填充。
type RegistryTagReference struct {
	ApplicationUID  string
	ApplicationName string
	VersionUID      string
	Version         string
	Status          string
}

// RegistryTagDetail 是 Tag 的 manifest 与平台详情。
type RegistryTagDetail struct {
	Tag          string
	Digest       string
	Size         int64
	MediaType    string
	Platforms    []RegistryPlatform
	ConfigDigest string
	Labels       []string
	CreatedAt    int64
	References   []RegistryTagReference
}

// RegistryListFilter 是仓库连接列表的过滤条件。
type RegistryListFilter struct {
	Name   string
	Driver string
}

// ImageSyncSource 描述同步任务的来源。
// Kind 为 CONNECTION 时使用 RegistryUID；为 EXTERNAL 时使用 Image 与临时凭据。
type ImageSyncSource struct {
	Kind             string
	RegistryUID      string
	Image            string
	Username         string
	Password         string
	AllowInsecureTLS bool
}

// ImageSync 是一条镜像同步任务。
type ImageSync struct {
	UID                    string
	NamespaceUID           string
	SourceKind             string
	SourceRegistryUID      string
	SourceRegistryName     string
	SourceEndpoint         string
	SourceImage            string
	SourceUsername         string
	SourceRepository       string
	SourceTag              string
	SourceDigest           string
	SourcePlatforms        []string
	SourceCredentialCipher string
	SourceAllowInsecureTLS bool
	TargetRegistryUID      string
	TargetRegistryName     string
	TargetEndpoint         string
	TargetRepository       string
	TargetTag              string
	TargetDigest           string
	Mode                   string
	Overwrite              bool
	Status                 string
	Stage                  string
	CopiedBlobs            int32
	TotalBlobs             int32
	Message                string
	Attempt                int32
	OperatorUID            string
	OperatorName           string
	LeaseOwner             string
	LeaseUntil             *time.Time
	StartedAt              *time.Time
	FinishedAt             *time.Time
	CreatedAt              time.Time
}

// ImageSyncFilter 是同步任务列表的过滤条件。
type ImageSyncFilter struct {
	Status      string
	SourceKind  string
	RegistryUID string
}

// ImageSyncInput 是创建同步任务的入参。
type ImageSyncInput struct {
	SourceKind             string
	SourceRegistryUID      string
	SourceImage            string
	SourceUsername         string
	SourcePassword         string
	SourceAllowInsecureTLS bool
	TargetRegistryUID      string
	TargetRepository       string
	TargetTag              string
	Platforms              []string
	Overwrite              bool
}

// ImageSyncFailure 是批量创建时某一条的失败原因。
type ImageSyncFailure struct {
	Index   int32
	Message string
}

// ImageSyncPreview 是创建前的解析结果。
type ImageSyncPreview struct {
	SourceImage     string
	SourceDigest    string
	SourcePlatforms []string
	TargetExists    bool
	TargetDigest    string
}
