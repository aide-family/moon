package rest

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/aide-family/moon/api/cluster/v1beta1"
	rsautil "github.com/aide-family/moon/pkg/util/rsa"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"net/http"
	"net/url"
)

const PrivateKey = "privateKey"

// BuildConfig return rest config for cluster.
func BuildConfig(clusterName string, connect v1beta1.ConnectConfig, secretGetter SecretGetter) (*rest.Config, error) {
	if len(connect.Endpoint) == 0 {
		return nil, fmt.Errorf("cluster %s api endpoint cannot be empty", clusterName)
	}

	config, err := clientcmd.BuildConfigFromFlags(connect.Endpoint, "")
	if err != nil {
		return nil, err
	}

	// Handle configuration.
	switch {
	case connect.Secret != nil:
		if err = buildConfigWithSecret(connect.Secret, secretGetter, connect.InsecureSkipTLSVerification, config); err != nil {
			return nil, fmt.Errorf("cluster %s build config with secret failed: %s", clusterName, err)
		}
	case connect.Config != nil:
		if config, err = buildConfigWithConfig(connect.Config, secretGetter); err != nil {
			return nil, fmt.Errorf("cluster %s build config with config failed: %s", clusterName, err)
		}
	case connect.Token != nil:
		if err = buildConfigWithToken(connect.Token, connect.InsecureSkipTLSVerification, config); err != nil {
			return nil, fmt.Errorf("cluster %s build config with token failed: %s", clusterName, err)
		}
	default:
		return nil, fmt.Errorf("cluster %s secret, config, token cannot be empty as the same time", clusterName)
	}

	// Handle proxy configuration.
	if connect.ProxyURL != "" {
		proxy, err := url.Parse(connect.ProxyURL)
		if err != nil {
			klog.Errorf("parse proxy error. %v", err)
			return nil, err
		}
		config.Proxy = http.ProxyURL(proxy)
	}
	return config, nil
}

func buildConfigWithConfig(ref *v1beta1.ConfigRef, secretGetter SecretGetter) (*rest.Config, error) {
	kubeConfig := ref.Config
	if ref.Secret != nil {
		if secretGetter == nil {
			return nil, fmt.Errorf("secret getter is required")
		}
		secret, err := secretGetter(types.NamespacedName{
			Namespace: ref.Secret.Namespace,
			Name:      ref.Secret.Name,
		})
		if err != nil {
			return nil, err
		}
		privateKey, err := SecretToRSACerts(secret)
		kubeConfig, err = rsautil.RSADecryptByPrivateKey(kubeConfig, privateKey)
		if err != nil {
			return nil, err
		}
	}
	clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeConfig)
	if err != nil {
		return nil, err
	}
	return clientConfig.ClientConfig()
}

func buildConfigWithSecret(ref *v1beta1.SecretRef, secretGetter SecretGetter,
	insecureSkipTLSVerification bool, config *rest.Config) error {
	if secretGetter == nil {
		return fmt.Errorf("secret getter is required")
	}
	secret, err := secretGetter(types.NamespacedName{
		Namespace: ref.Namespace,
		Name:      ref.Name,
	})
	if err != nil {
		return err
	}
	config.BearerToken = string(secret.Data[v1beta1.SecretTokenKey])
	if insecureSkipTLSVerification {
		config.TLSClientConfig.Insecure = true
	} else {
		if len(secret.Data[v1beta1.SecretCADataKey]) == 0 {
			return fmt.Errorf("cannot missing the CA data")
		}
		config.TLSClientConfig.CAData = secret.Data[v1beta1.SecretCADataKey]
	}
	return nil
}

func buildConfigWithToken(ref *v1beta1.TokenRef, insecureSkipTLSVerification bool, config *rest.Config) error {
	config.BearerToken = ref.Token
	if insecureSkipTLSVerification {
		config.TLSClientConfig.Insecure = true
	} else {
		if len(ref.CABundle) == 0 {
			return fmt.Errorf("cannot missing the CA data")
		}
		config.TLSClientConfig.CAData = ref.CABundle
	}
	return nil
}

type ClusterGetter func(string) (*v1beta1.Cluster, error)
type SecretGetter func(types.NamespacedName) (*v1.Secret, error)

func BuildSecret(key types.NamespacedName) (*v1.Secret, *rsa.PrivateKey, error) {
	privateKey, err := rsautil.NewPrivateKey()
	if err != nil {
		return nil, nil, err
	}
	secret := CertsToSecret(rsautil.EncodePrivateKeyPEM(privateKey), key)
	return secret, privateKey, nil
}

func SecretToRSACerts(secret *v1.Secret) (*rsa.PrivateKey, error) {
	if secret.Data == nil {
		return nil, fmt.Errorf("secret data is nil")
	}
	buf := secret.Data[PrivateKey]
	block, _ := pem.Decode(buf)
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

func CertsToSecret(key []byte, sec types.NamespacedName) *v1.Secret {
	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: sec.Namespace,
			Name:      sec.Name,
		},
		Data: map[string][]byte{
			PrivateKey: key,
		},
	}
}
