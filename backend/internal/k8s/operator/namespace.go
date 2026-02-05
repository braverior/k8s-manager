package operator

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type NamespaceOperator struct {
	client *kubernetes.Clientset
}

func NewNamespaceOperator(client *kubernetes.Clientset) *NamespaceOperator {
	return &NamespaceOperator{client: client}
}

func (o *NamespaceOperator) List(ctx context.Context) ([]corev1.Namespace, error) {
	list, err := o.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}
