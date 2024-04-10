/*
Copyright 2024 The Moon Authors.
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// Provider of the cluster, this field is just for description
	// +optional
	Provider string `json:"provider,omitempty"`
	// Desired state of the cluster
	// +optional
	Disabled bool `json:"disabled,omitempty"`
	// Connect used to connect to cluster api server.
	// You can choose one of the following three ways to connect:
	// + ConnectConfig.Secret
	// + ConnectConfig.Config
	// + ConnectConfig.Token
	Connect ConnectConfig `json:"connect"`
	// Region represents the region of the member cluster locate in.
	// +optional
	Region Region `json:"region,omitempty"`
}

type ConnectConfig struct {
	// It is relatively safe to use Secret to save token and CABundle in the cluster.
	// It is recommended and has the highest priority.
	// If you want to do this, the data definition of Secret must meet the following conditions:
	// - secret.data.token
	// - secret.data.caBundle
	// +optional
	Secret *SecretRef `json:"secret,omitempty"`
	// Config needs to use a configuration file to connect. If you have defined a Secret,
	//it will use the Secret for encoding and decoding to ensure data security. Moderate recommendation.
	// config usually can be /etc/kubernetes/admin.conf or ~/.kube/config
	// +optional
	Config *ConfigRef `json:"config,omitempty"`
	// The Token display declares the token and CABundle connected to the cluster,
	// which is not safe, not recommended, and has the lowest priority.
	// +optional
	Token *TokenRef `json:"token,omitempty"`
	// InsecureSkipTLSVerification indicates that the cluster pool should not confirm the validity of the serving
	// certificate of the cluster it is connecting to. This will make the HTTPS connection between the cluster pool
	// and the member cluster insecure.
	// Defaults to false.
	// +optional
	InsecureSkipTLSVerification bool `json:"insecureSkipTLSVerification,omitempty"`
	// Kubernetes API Server endpoint.
	// hostname:port, IP or IP:port.
	// Example: https://10.10.0.1:6443
	// +optional
	Endpoint string `json:"endpoint,omitempty"`
	// ProxyURL is the proxy URL for the cluster.
	// If not empty, the multi-cluster control plane will use this proxy to talk to the cluster.
	// More details please refer to: https://github.com/kubernetes/client-go/issues/351
	// +optional
	ProxyURL string `json:"proxyURL,omitempty"`
	// ProxyHeader is the HTTP header required by proxy server.
	// The key in the key-value pair is HTTP header key and value is the associated header payloads.
	// For the header with multiple values, the values should be separated by comma(e.g. 'k1': 'v1,v2,v3').
	// +optional
	ProxyHeader map[string]string `json:"proxyHeader,omitempty"`
}

type ConfigRef struct {
	//Secret used to encode and decode Config to protect Config from being leaked.
	// +optional
	Secret *SecretRef `json:"secret,omitempty"`
	// The Config used to connect to the cluster.
	// There is no need to encrypt when joining.
	// When saving data, it will automatically use Secret for encryption. If Secret exists.
	Config []byte `json:"config,omitempty"`
}

type SecretRef struct {
	// Namespace is the namespace for the resource being referenced.
	Namespace string `json:"namespace"`

	// Name is the name of resource being referenced.
	Name string `json:"name"`
}

const (
	// SecretTokenKey is the name of secret token key.
	SecretTokenKey = "token"
	// SecretCADataKey is the name of secret caBundle key.
	SecretCADataKey = "caBundle"
)

type TokenRef struct {
	// CABundle contains the certificate authority information.
	// +optional
	CABundle []byte `json:"caBundle,omitempty"`

	// Token contain the token authority information.
	// +optional
	Token string `json:"token,omitempty"`
}

type Region struct {
	// Zone represents the zone of the member cluster locate in.
	// +optional
	Zone string `json:"zone,omitempty"`
	// Country represents the country of the member cluster locate in.
	// +optional
	Country string `json:"country,omitempty"`
	// Province represents the province of the member cluster locate in.
	// +optional
	Province string `json:"province,omitempty"`
	// City represents the city of the member cluster locate in.
	// +optional
	City string `json:"city,omitempty"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	Phase ClusterPhase `json:"phase"`
	// Version represents version of the member cluster.
	// +optional
	Version string `json:"version,omitempty"`

	// APIEnablements represents the list of APIs installed in the member cluster.
	// +optional
	APIEnablements []APIEnablement `json:"apiEnablements,omitempty"`

	// Conditions is an array of current cluster conditions.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// NodeSummary represents the summary of nodes status in the member cluster.
	// +optional
	NodeSummary *NodeSummary `json:"nodeSummary,omitempty"`
}

type ClusterPhase string

const (
	ClusterPhaseInitial     ClusterPhase = "Initial"
	ClusterPhaseHealthy     ClusterPhase = "Healthy"
	ClusterPhaseTerminating ClusterPhase = "Terminating"
)

// APIEnablement is a list of API resource, it is used to expose the name of the
// resources supported in a specific group and version.
type APIEnablement struct {
	// GroupVersion is the group and version this APIEnablement is for.
	GroupVersion string `json:"groupVersion,omitempty"`

	// Resources is a list of APIResource.
	// +optional
	Resources []APIResource `json:"resources,omitempty"`
}

// APIResource specifies the name and kind names for the resource.
type APIResource struct {
	// Name is the plural name of the resource.
	// +required
	Name string `json:"name,omitempty"`

	// Kind is the kind for the resource (e.g. 'Deployment' is the kind for resource 'deployments')
	// +required
	Kind string `json:"kind,omitempty"`
}

// NodeSummary represents the summary of nodes status in a specific cluster.
type NodeSummary struct {
	// TotalNum is the total number of nodes in the cluster.
	// +optional
	TotalNum int32 `json:"total,omitempty"`

	// ReadyNum is the number of ready nodes in the cluster.
	// +optional
	ReadyNum int32 `json:"ready,omitempty"`
}

// +genclient:nonNamespaced
// +k8s:openapi-gen=true
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="ENDPOINT",type="string",priority=1,JSONPath=".spec.endpoint",description="The cluster endpoint"
// +kubebuilder:printcolumn:name="ENABLE",type="boolean",priority=1,JSONPath=".spec.enabled",description="The cluster enable status"
// +kubebuilder:printcolumn:name="PROVIDER",type="string",priority=1,JSONPath=".spec.provider",description="The cluster provider"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".status.version",description="The cluster version"
// +kubebuilder:printcolumn:name="TOTAL",type="integer",JSONPath=".status.nodeSummary.total",description="The total number of node"
// +kubebuilder:printcolumn:name="READY",type="integer",JSONPath=".status.nodeSummary.ready",description="The ready number of node"
// +kubebuilder:printcolumn:name="AGE",type=date,JSONPath=".metadata.creationTimestamp"

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterSpec `json:"spec,omitempty"`
	// +optional
	Status ClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
