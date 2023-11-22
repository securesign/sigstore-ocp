package kubernetes

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PullSecretExists(secretName, namespace string) (bool, error) {
	secrets, err := Clientset.CoreV1().Secrets(namespace).List(context.TODO(), metav1.ListOptions{})
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

func CreatePullSecret(secretName, namespace, filename string, secretData []byte) error {
	exists, err := PullSecretExists(secretName, namespace)
	if err != nil {
		return err
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			filename: secretData,
		},
	}

	if exists {
		_, err = Clientset.CoreV1().Secrets(namespace).Update(context.TODO(), secret, metav1.UpdateOptions{})
	} else {
		_, err = Clientset.CoreV1().Secrets(namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	}

	return err
}
