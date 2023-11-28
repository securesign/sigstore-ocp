package kubernetes

import (
	"context"
	"time"

	v1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kc *KubernetesClient) GetJob(namespace, jobName string) (*v1.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	jobs, err := kc.Clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	for _, job := range jobs.Items {
		if job.Name == jobName {
			return &job, nil
		}
	}

	return nil, nil
}

func (kc *KubernetesClient) DeleteJob(namespace, jobName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	deletePolicy := metav1.DeletePropagationBackground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	return kc.Clientset.BatchV1().Jobs(namespace).Delete(ctx, jobName, deleteOptions)
}
