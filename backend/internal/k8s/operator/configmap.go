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

type ConfigMapOperator struct {
	client *kubernetes.Clientset
}

func NewConfigMapOperator(client *kubernetes.Clientset) *ConfigMapOperator {
	return &ConfigMapOperator{client: client}
}

func (o *ConfigMapOperator) List(ctx context.Context, namespace string) ([]corev1.ConfigMap, error) {
	list, err := o.client.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (o *ConfigMapOperator) Get(ctx context.Context, namespace, name string) (*corev1.ConfigMap, error) {
	return o.client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (o *ConfigMapOperator) Create(ctx context.Context, namespace string, cm *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	cm.Namespace = namespace
	return o.client.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
}

func (o *ConfigMapOperator) Update(ctx context.Context, namespace string, cm *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	return o.client.CoreV1().ConfigMaps(namespace).Update(ctx, cm, metav1.UpdateOptions{})
}

func (o *ConfigMapOperator) Delete(ctx context.Context, namespace, name string) error {
	err := o.client.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if k8serrors.IsNotFound(err) {
		return nil
	}
	return err
}

func (o *ConfigMapOperator) Exists(ctx context.Context, namespace, name string) (bool, error) {
	_, err := o.Get(ctx, namespace, name)
	if k8serrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func ConfigMapToJSON(cm *corev1.ConfigMap) (string, error) {
	data, err := json.MarshalIndent(cm, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal configmap: %w", err)
	}
	return string(data), nil
}

func JSONToConfigMap(data string) (*corev1.ConfigMap, error) {
	var cm corev1.ConfigMap
	if err := json.Unmarshal([]byte(data), &cm); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configmap: %w", err)
	}
	return &cm, nil
}
