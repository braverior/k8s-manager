package operator

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type EndpointsOperator struct {
	client *kubernetes.Clientset
}

func NewEndpointsOperator(client *kubernetes.Clientset) *EndpointsOperator {
	return &EndpointsOperator{client: client}
}

// Get 获取指定名称的 Endpoints
func (o *EndpointsOperator) Get(ctx context.Context, namespace, name string) (*corev1.Endpoints, error) {
	return o.client.CoreV1().Endpoints(namespace).Get(ctx, name, metav1.GetOptions{})
}

// GetByService 根据 Service 名称获取对应的 Endpoints
// 在 Kubernetes 中，Endpoints 的名称与对应的 Service 名称相同
func (o *EndpointsOperator) GetByService(ctx context.Context, namespace, serviceName string) (*corev1.Endpoints, error) {
	return o.Get(ctx, namespace, serviceName)
}

// List 列出指定命名空间的所有 Endpoints
func (o *EndpointsOperator) List(ctx context.Context, namespace string) ([]corev1.Endpoints, error) {
	list, err := o.client.CoreV1().Endpoints(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}
