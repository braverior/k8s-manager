package operator

import (
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ServiceOperator struct {
	client *kubernetes.Clientset
}

func NewServiceOperator(client *kubernetes.Clientset) *ServiceOperator {
	return &ServiceOperator{client: client}
}

func (o *ServiceOperator) List(ctx context.Context, namespace string) ([]corev1.Service, error) {
	list, err := o.client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (o *ServiceOperator) Get(ctx context.Context, namespace, name string) (*corev1.Service, error) {
	return o.client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (o *ServiceOperator) Create(ctx context.Context, namespace string, svc *corev1.Service) (*corev1.Service, error) {
	svc.Namespace = namespace
	return o.client.CoreV1().Services(namespace).Create(ctx, svc, metav1.CreateOptions{})
}

func (o *ServiceOperator) Update(ctx context.Context, namespace string, svc *corev1.Service) (*corev1.Service, error) {
	return o.client.CoreV1().Services(namespace).Update(ctx, svc, metav1.UpdateOptions{})
}

func (o *ServiceOperator) Delete(ctx context.Context, namespace, name string) error {
	err := o.client.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if k8serrors.IsNotFound(err) {
		return nil
	}
	return err
}

func (o *ServiceOperator) Exists(ctx context.Context, namespace, name string) (bool, error) {
	_, err := o.Get(ctx, namespace, name)
	if k8serrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func ServiceToJSON(svc *corev1.Service) (string, error) {
	data, err := json.MarshalIndent(svc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal service: %w", err)
	}
	return string(data), nil
}

func JSONToService(data string) (*corev1.Service, error) {
	var svc corev1.Service
	if err := json.Unmarshal([]byte(data), &svc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal service: %w", err)
	}
	return &svc, nil
}
