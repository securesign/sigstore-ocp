package kubernetes

import (
	"context"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	MaxAttempts      = 30
	SleepInterval    = 10
	ErrPodNotRunning error
	ErrPodNotFound   error
)

func (kc *KubernetesClient) getPodStatus(namespace, podNamePrefix string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pods, err := kc.Clientset.CoreV1().Pods(namespace).List(ctx, v1.ListOptions{})
	if err != nil {
		return "", err
	}

	for _, pod := range pods.Items {
		if strings.HasPrefix(pod.Name, podNamePrefix) {
			return string(pod.Status.Phase), nil
		}
	}

	return "", ErrPodNotFound
}

func (kc *KubernetesClient) WaitForPodStatusRunning(namespace, podNamePrefix string, status func(error)) error {
	CurrentAttempt := 0
	for CurrentAttempt < MaxAttempts {
		phase, err := kc.getPodStatus(namespace, podNamePrefix)
		if err != nil {
			if err == ErrPodNotFound {
				status(ErrPodNotFound)
			} else {
				return err
			}
		}

		if phase == "Running" {
			return nil
		} else {
			status(ErrPodNotRunning)
		}

		CurrentAttempt++
		time.Sleep(time.Duration(SleepInterval) * time.Second)
	}
	return fmt.Errorf("Timed out. No pods with the prefix '%s' reached the 'Running' state within the specified time.", podNamePrefix)
}
