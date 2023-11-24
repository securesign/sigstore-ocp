package kubernetes

import (
	"context"

	v1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (kc *KubernetesClient) GetJob(namespace, jobName string) (*v1.Job, error) {
	jobs, err := kc.Clientset.BatchV1().Jobs(namespace).List(context.TODO(), metav1.ListOptions{})
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

	deletePolicy := metav1.DeletePropagationBackground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	err := kc.Clientset.BatchV1().Jobs(namespace).Delete(context.TODO(), jobName, deleteOptions)
	if err != nil {
		return err
	}
	return nil
}
