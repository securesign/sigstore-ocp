package kubernetes

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesClient struct {
	Clientset         *kubernetes.Clientset
	DynamicClientSet  dynamic.Interface
	ClusterBaseDomain string
	ClusterCommonName string
}

func InitKubeClient(kubeConfigPath string) (*KubernetesClient, error) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, fmt.Errorf("error getting Kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("error getting Kubernetes clientset: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating dynamic client: %w", err)
	}

	kubeClient := &KubernetesClient{Clientset: clientset, DynamicClientSet: dynamicClient}

	baseDomain, err := kubeClient.getClusterBaseDomain()
	if err != nil {
		return nil, err
	}

	kubeClient.ClusterBaseDomain = baseDomain
	kubeClient.ClusterCommonName = "apps." + baseDomain

	return kubeClient, nil
}

func (kc *KubernetesClient) getClusterBaseDomain() (string, error) {
	groupVersionResource := schema.GroupVersionResource{Group: "config.openshift.io", Version: "v1", Resource: "dnses"}
	dnsCRD, err := kc.DynamicClientSet.Resource(groupVersionResource).Namespace("").Get(context.TODO(), "cluster", metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	baseDomain, found, err := unstructured.NestedString(dnsCRD.Object, "spec", "baseDomain")
	if !found || err != nil {
		return "", err
	}

	return baseDomain, nil
}
