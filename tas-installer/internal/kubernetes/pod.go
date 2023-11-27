package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	MaxAttempts    = 30
	SleepInterval  = 10
	ErrPodNotFound = errors.New("pod not found")
)

func (kc *KubernetesClient) CheckPodStatus(namespace, podNamePrefix string) (string, error) {
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

func (kc *KubernetesClient) WaitForPodStatusRunning(namespace, podNamePrefix string) error {
	CurrentAttempt := 0
	fmt.Printf("Waiting for %s to reach a running state \n", podNamePrefix)
	for CurrentAttempt < MaxAttempts {
		phase, err := kc.CheckPodStatus(namespace, podNamePrefix)
		if err != nil {
			if err == ErrPodNotFound {
				fmt.Printf("No pods with the prefix '%s' found in namespace %s. Retrying in %d seconds... \n", podNamePrefix, namespace, SleepInterval)
			} else {
				return err
			}
		} else if phase == "Running" {
			fmt.Printf("%s is up and running in namespace %s. \n", podNamePrefix, namespace)
			return nil
		}

		CurrentAttempt++
		time.Sleep(time.Duration(SleepInterval) * time.Second)
	}
	return fmt.Errorf("Timed out. No pods with the prefix '%s' reached the 'Running' state within the specified time.", podNamePrefix)
}
