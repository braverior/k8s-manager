package operator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

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
	list, err := o.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{ResourceVersion: "0"})
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
	Name        string `json:"name"`
	CPU         string `json:"cpu"`          // 如 "100m"
	Memory      string `json:"memory"`       // 如 "128Mi"
	CPUMillis   int64  `json:"cpu_millis"`   // CPU 用量（毫核），供聚合计算
	MemoryBytes int64  `json:"memory_bytes"` // 内存用量（字节），供聚合计算
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
		cpu := c.Usage.Cpu()
		mem := c.Usage.Memory()
		podMetrics.Containers = append(podMetrics.Containers, ContainerMetrics{
			Name:        c.Name,
			CPU:         formatPodCPU(cpu),
			Memory:      formatPodMemory(mem),
			CPUMillis:   cpu.MilliValue(),
			MemoryBytes: mem.Value(),
		})
	}

	return podMetrics, nil
}

// ListMetrics 获取命名空间下所有 Pod 的指标
func (o *PodOperator) ListMetrics(ctx context.Context, namespace string) ([]PodMetrics, error) {
	if o.metricsClient == nil {
		return nil, fmt.Errorf("metrics client not available")
	}

	list, err := o.metricsClient.MetricsV1beta1().PodMetricses(namespace).List(ctx, metav1.ListOptions{ResourceVersion: "0"})
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
			cpu := c.Usage.Cpu()
			mem := c.Usage.Memory()
			pm.Containers = append(pm.Containers, ContainerMetrics{
				Name:        c.Name,
				CPU:         formatPodCPU(cpu),
				Memory:      formatPodMemory(mem),
				CPUMillis:   cpu.MilliValue(),
				MemoryBytes: mem.Value(),
			})
		}
		result = append(result, pm)
	}

	return result, nil
}

// GetLogs 获取 Pod 容器日志
func (o *PodOperator) GetLogs(ctx context.Context, namespace, name, container string, tailLines int64, previous bool, timestamps bool) (string, error) {
	opts := &corev1.PodLogOptions{
		Container:  container,
		Previous:   previous,
		Timestamps: timestamps,
	}
	if tailLines > 0 {
		opts.TailLines = &tailLines
	}

	req := o.client.CoreV1().Pods(namespace).GetLogs(name, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer stream.Close()

	data, err := io.ReadAll(stream)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetEvents 获取 Pod 相关事件
func (o *PodOperator) GetEvents(ctx context.Context, namespace, name string) ([]corev1.Event, error) {
	events, err := o.client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Pod", name),
	})
	if err != nil {
		return nil, err
	}
	return events.Items, nil
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

// formatPodMemory 格式化内存为 Mi 单位
func formatPodMemory(q *resource.Quantity) string {
	if q == nil {
		return "0"
	}
	bytes := q.Value()
	mi := float64(bytes) / (1024 * 1024)
	if mi >= 1024 {
		gi := mi / 1024
		return fmt.Sprintf("%.1fGi", gi)
	}
	return fmt.Sprintf("%.0fMi", mi)
}
