package operator

import (
	"context"
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

type PodOperator struct {
	client        *kubernetes.Clientset
	metricsClient metricsv.Interface
}

func NewPodOperator(client *kubernetes.Clientset, metricsClient metricsv.Interface) *PodOperator {
	return &PodOperator{
		client:        client,
		metricsClient: metricsClient,
	}
}

// List 列出命名空间下的所有 Pod
func (o *PodOperator) List(ctx context.Context, namespace string) ([]corev1.Pod, error) {
	list, err := o.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// ListByLabels 按标签筛选 Pod
func (o *PodOperator) ListByLabels(ctx context.Context, namespace string, labelSelector string) ([]corev1.Pod, error) {
	list, err := o.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// Get 获取单个 Pod
func (o *PodOperator) Get(ctx context.Context, namespace, name string) (*corev1.Pod, error) {
	return o.client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

// Delete 删除 Pod（用于重启 Pod）
func (o *PodOperator) Delete(ctx context.Context, namespace, name string) error {
	return o.client.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

// PodMetrics Pod 资源使用指标
type PodMetrics struct {
	Name       string             `json:"name"`
	Namespace  string             `json:"namespace"`
	Containers []ContainerMetrics `json:"containers"`
}

// ContainerMetrics 容器资源使用指标
type ContainerMetrics struct {
	Name   string `json:"name"`
	CPU    string `json:"cpu"`    // 如 "100m"
	Memory string `json:"memory"` // 如 "128Mi"
}

// GetMetrics 获取 Pod 的 CPU/内存指标
func (o *PodOperator) GetMetrics(ctx context.Context, namespace, name string) (*PodMetrics, error) {
	if o.metricsClient == nil {
		return nil, fmt.Errorf("metrics client not available")
	}

	metrics, err := o.metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	podMetrics := &PodMetrics{
		Name:       metrics.Name,
		Namespace:  metrics.Namespace,
		Containers: make([]ContainerMetrics, 0, len(metrics.Containers)),
	}

	for _, c := range metrics.Containers {
		podMetrics.Containers = append(podMetrics.Containers, ContainerMetrics{
			Name:   c.Name,
			CPU:    formatPodCPU(c.Usage.Cpu()),
			Memory: formatPodMemory(c.Usage.Memory()),
		})
	}

	return podMetrics, nil
}

// ListMetrics 获取命名空间下所有 Pod 的指标
func (o *PodOperator) ListMetrics(ctx context.Context, namespace string) ([]PodMetrics, error) {
	if o.metricsClient == nil {
		return nil, fmt.Errorf("metrics client not available")
	}

	list, err := o.metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	result := make([]PodMetrics, 0, len(list.Items))
	for _, m := range list.Items {
		pm := PodMetrics{
			Name:       m.Name,
			Namespace:  m.Namespace,
			Containers: make([]ContainerMetrics, 0, len(m.Containers)),
		}
		for _, c := range m.Containers {
			pm.Containers = append(pm.Containers, ContainerMetrics{
				Name:   c.Name,
				CPU:    formatPodCPU(c.Usage.Cpu()),
				Memory: formatPodMemory(c.Usage.Memory()),
			})
		}
		result = append(result, pm)
	}

	return result, nil
}

func PodToJSON(pod *corev1.Pod) (string, error) {
	data, err := json.MarshalIndent(pod, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal pod: %w", err)
	}
	return string(data), nil
}

// formatPodCPU 格式化 CPU 为毫核(m)格式
func formatPodCPU(q *resource.Quantity) string {
	if q == nil {
		return "0"
	}
	milliValue := q.MilliValue()
	if milliValue >= 1000 {
		cores := float64(milliValue) / 1000
		if cores == float64(int64(cores)) {
			return fmt.Sprintf("%d", int64(cores))
		}
		return fmt.Sprintf("%.2f", cores)
	}
	return fmt.Sprintf("%dm", milliValue)
}

// formatPodMemory 格式化内存为易读格式
func formatPodMemory(q *resource.Quantity) string {
	if q == nil {
		return "0"
	}
	bytes := q.Value()
	if bytes >= 1024*1024*1024 {
		return q.String()
	}
	mi := float64(bytes) / (1024 * 1024)
	if mi >= 1024 {
		gi := mi / 1024
		return resource.NewQuantity(int64(gi*1024*1024*1024), resource.BinarySI).String()
	}
	return resource.NewQuantity(int64(mi*1024*1024), resource.BinarySI).String()
}
