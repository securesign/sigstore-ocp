package kubernetes

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kc *KubernetesClient) CreateNamespace(ns string) error {
	exists, err := kc.namespaceExists(ns)
	if err != nil {
		return err
	}

	if !exists {
		namespace := &v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: ns,
			},
		}

		_, err := kc.Clientset.CoreV1().Namespaces().Create(context.TODO(), namespace, metav1.CreateOptions{})
		if err != nil {
			return err
		}
		fmt.Printf("%s namespace created successfully \n", ns)
	} else {
		fmt.Printf("%s namespace already exists, skipping create \n", ns)
	}
	return nil
}

func (kc *KubernetesClient) namespaceExists(namespace string) (bool, error) {
	namespaces, err := kc.Clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
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
