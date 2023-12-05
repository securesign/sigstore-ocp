package kubernetes

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	ErrNamespaceAlreadyExists error
)

func (kc *KubernetesClient) DeleteNamespaceIfExists(ns string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exists, err := kc.NamespaceExists(ctx, ns)
	if err != nil {
		return false, err
	}

	if exists {
		if err := kc.Clientset.CoreV1().Namespaces().Delete(ctx, ns, metav1.DeleteOptions{}); err != nil {
			return exists, err
		}
	}
	return exists, nil
}

func (kc *KubernetesClient) CreateNamespaceIfNotExists(ns string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exists, err := kc.NamespaceExists(ctx, ns)
	if err != nil {
		return err
	}

	if exists {
		return ErrNamespaceAlreadyExists
	}
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: ns,
		},
	}

	if _, err := kc.Clientset.CoreV1().Namespaces().Create(ctx, namespace, metav1.CreateOptions{}); err != nil {
		return err
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
