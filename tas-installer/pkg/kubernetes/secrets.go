package kubernetes

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kc *KubernetesClient) SecretExists(ctx context.Context, secretName, namespace string) (bool, error) {
	secrets, err := kc.Clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	exists, err := kc.SecretExists(ctx, secretName, namespace)
	if err != nil {
		return err
	}
	if !exists {
		_, err = kc.Clientset.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
		if err == nil {
			fmt.Printf("Secret: %s created successfully\n", secretName)
		}
	}
	return err
}

func (kc *KubernetesClient) UpdateSecretData(secretName, namespace, key string, data []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	secret, err := kc.Clientset.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	secret.Data[key] = data
	_, err = kc.Clientset.CoreV1().Secrets(namespace).Update(ctx, secret, metav1.UpdateOptions{})
	return err
}
