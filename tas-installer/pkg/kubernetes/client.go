package kubernetes

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesClient struct {
	Clientset         *kubernetes.Clientset
	ClusterBaseDomain string
	ClusterCommonName string
}

func InitKubeClient() (*KubernetesClient, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting user home dir: %w", err)
	}
	kubeConfigPath := filepath.Join(homeDir, ".kube", "config")
	fmt.Printf("Using kube config found at %s\n", kubeConfigPath)

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("error getting Kubernetes config: %w", err)
	}

	dns := kubeConfig.Host
	baseDomain, err := parseClusterDNS(dns)
	commonName := "apps." + baseDomain

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("error getting Kubernetes clientset: %w", err)
	}

	return &KubernetesClient{Clientset: clientset, ClusterBaseDomain: baseDomain, ClusterCommonName: commonName}, nil
}

func parseClusterDNS(dns string) (string, error) {
	parsedURL, err := url.Parse(dns)
	if err != nil {
		panic(err)
	}
	domain := parsedURL.Hostname()
	return domain, nil
}
