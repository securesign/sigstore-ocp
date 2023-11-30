package kubernetes

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ErrNamespaceAlreadyExists error
)

func (kc *KubernetesClient) CreateNamespaceIfNotExists(ns string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exists, err := kc.NamespaceExists(ctx, ns)
	if err != nil {
		return err
	}

	if !exists {
		namespace := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: ns,
				Labels: map[string]string{
					"app.kubernetes.io/managed-by": "Helm",
				},
				Annotations: map[string]string{
					"meta.helm.sh/release-name":      "trusted-artifact-signer",
					"meta.helm.sh/release-namespace": "trusted-artifact-signer",
				},
			},
		}

		_, err := kc.Clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{})
		if err != nil {
			return err
		}
	} else {
		return ErrNamespaceAlreadyExists
	}
	return nil
}

func (kc *KubernetesClient) NamespaceExists(ctx context.Context, namespace string) (bool, error) {
	namespaces, err := kc.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, err
	}

	for _, ns := range namespaces.Items {
		if ns.Name == namespace {
			return true, nil
		}
	}
	return false, nil
}
