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

func CheckPodStatus(namespace, podNamePrefix string) (string, error) {
	pods, err := Clientset.CoreV1().Pods(namespace).List(context.Background(), v1.ListOptions{})
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

func WaitForPodStatusRunning(namespace, podNamePrefix string) error {
	CurrentAttempt := 0
	fmt.Printf("Waiting for %s to reach a running state\n", podNamePrefix)
	for CurrentAttempt < MaxAttempts {
		phase, err := CheckPodStatus(namespace, podNamePrefix)
		if err != nil {
			if err == ErrPodNotFound {
				fmt.Printf("Pod %s not found, retrying...\n", podNamePrefix)
			} else {
				return err
			}
		} else if phase == "Running" {
			fmt.Printf("Pod %s has reached a running state\n", podNamePrefix)
			return nil
		}

		CurrentAttempt++
		time.Sleep(time.Duration(SleepInterval) * time.Second)
	}
	return fmt.Errorf("%s pod in namespace %s did not reach the 'Running' phase after %d attempts\n", podNamePrefix, namespace, MaxAttempts)
}
