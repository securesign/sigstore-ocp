package kubernetes

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kc *KubernetesClient) SecretExists(secretName, namespace string) (bool, error) {
	secrets, err := kc.Clientset.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return false, err
	}

	for _, secret := range secrets.Items {
		if secretName == secret.Name {
			return true, nil
		}
	}

	return false, nil
}

func (kc *KubernetesClient) CreateSecret(secretName, namespace string, secret *v1.Secret) error {
	exists, err := kc.SecretExists(secretName, namespace)
	if err != nil {
		return err
	}
	if !exists {
		_, err = kc.Clientset.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	}
	return err
}

func (kc *KubernetesClient) UpdateSecretData(secretName, namespace, key string, data []byte) error {
	secret, err := kc.Clientset.CoreV1().Secrets(namespace).Get(context.TODO(), secretName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	secret.Data[key] = data
	_, err = kc.Clientset.CoreV1().Secrets(namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
	return err
}
