package operator

import (
	"context"
	"encoding/json"
	"fmt"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// HPA GroupVersionResource 定义
var (
	hpaGVRv2 = schema.GroupVersionResource{
		Group:    "autoscaling",
		Version:  "v2",
		Resource: "horizontalpodautoscalers",
	}
	hpaGVRv2beta2 = schema.GroupVersionResource{
		Group:    "autoscaling",
		Version:  "v2beta2",
		Resource: "horizontalpodautoscalers",
	}
)

type HPAOperator struct {
	client        *kubernetes.Clientset
	dynamicClient dynamic.Interface
	gvr           schema.GroupVersionResource
	apiVersion    string
}

// NewHPAOperator 创建 HPA 操作器，自动检测 API 版本
func NewHPAOperator(client *kubernetes.Clientset) *HPAOperator {
	return &HPAOperator{
		client: client,
	}
}

// NewHPAOperatorWithDynamic 创建带动态客户端的 HPA 操作器
func NewHPAOperatorWithDynamic(client *kubernetes.Clientset, dynamicClient dynamic.Interface) *HPAOperator {
	op := &HPAOperator{
		client:        client,
		dynamicClient: dynamicClient,
	}
	op.detectAPIVersion()
	return op
}

// detectAPIVersion 检测集群支持的 HPA API 版本
func (o *HPAOperator) detectAPIVersion() {
	// 优先尝试 v2
	_, resources, err := o.client.Discovery().ServerGroupsAndResources()
	if err != nil {
		// 默认使用 v2
		o.gvr = hpaGVRv2
		o.apiVersion = "autoscaling/v2"
		return
	}

	hasV2 := false
	hasV2beta2 := false

	for _, resourceList := range resources {
		if resourceList.GroupVersion == "autoscaling/v2" {
			for _, r := range resourceList.APIResources {
				if r.Name == "horizontalpodautoscalers" {
					hasV2 = true
					break
				}
			}
		}
		if resourceList.GroupVersion == "autoscaling/v2beta2" {
			for _, r := range resourceList.APIResources {
				if r.Name == "horizontalpodautoscalers" {
					hasV2beta2 = true
					break
				}
			}
		}
	}

	if hasV2 {
		o.gvr = hpaGVRv2
		o.apiVersion = "autoscaling/v2"
	} else if hasV2beta2 {
		o.gvr = hpaGVRv2beta2
		o.apiVersion = "autoscaling/v2beta2"
	} else {
		// 默认 v2
		o.gvr = hpaGVRv2
		o.apiVersion = "autoscaling/v2"
	}
}

// GetAPIVersion 返回当前使用的 API 版本
func (o *HPAOperator) GetAPIVersion() string {
	if o.apiVersion == "" {
		o.detectAPIVersion()
	}
	return o.apiVersion
}

// List 列出 HPA（返回 v2 类型，内部自动转换）
func (o *HPAOperator) List(ctx context.Context, namespace string) ([]autoscalingv2.HorizontalPodAutoscaler, error) {
	if o.dynamicClient != nil && o.apiVersion != "" {
		return o.listDynamic(ctx, namespace)
	}
	// 尝试 v2，失败则尝试 v2beta2
	return o.listWithFallback(ctx, namespace)
}

func (o *HPAOperator) listWithFallback(ctx context.Context, namespace string) ([]autoscalingv2.HorizontalPodAutoscaler, error) {
	// 先尝试 v2
	list, err := o.client.AutoscalingV2().HorizontalPodAutoscalers(namespace).List(ctx, metav1.ListOptions{})
	if err == nil {
		return list.Items, nil
	}

	// v2 失败，尝试 v2beta2
	listBeta, err := o.client.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 转换 v2beta2 到 v2
	result := make([]autoscalingv2.HorizontalPodAutoscaler, 0, len(listBeta.Items))
	for _, item := range listBeta.Items {
		converted, err := convertV2beta2ToV2(&item)
		if err != nil {
			return nil, fmt.Errorf("convert hpa failed: %w", err)
		}
		result = append(result, *converted)
	}
	return result, nil
}

func (o *HPAOperator) listDynamic(ctx context.Context, namespace string) ([]autoscalingv2.HorizontalPodAutoscaler, error) {
	var list *unstructured.UnstructuredList
	var err error

	if namespace == "" {
		list, err = o.dynamicClient.Resource(o.gvr).List(ctx, metav1.ListOptions{})
	} else {
		list, err = o.dynamicClient.Resource(o.gvr).Namespace(namespace).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, err
	}

	result := make([]autoscalingv2.HorizontalPodAutoscaler, 0, len(list.Items))
	for _, item := range list.Items {
		hpa, err := unstructuredToHPAv2(&item)
		if err != nil {
			return nil, err
		}
		result = append(result, *hpa)
	}
	return result, nil
}

// Get 获取单个 HPA
func (o *HPAOperator) Get(ctx context.Context, namespace, name string) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	if o.dynamicClient != nil && o.apiVersion != "" {
		return o.getDynamic(ctx, namespace, name)
	}
	return o.getWithFallback(ctx, namespace, name)
}

func (o *HPAOperator) getWithFallback(ctx context.Context, namespace, name string) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	// 先尝试 v2
	hpa, err := o.client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		return hpa, nil
	}

	// v2 失败，尝试 v2beta2
	hpaBeta, err := o.client.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return convertV2beta2ToV2(hpaBeta)
}

func (o *HPAOperator) getDynamic(ctx context.Context, namespace, name string) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	obj, err := o.dynamicClient.Resource(o.gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return unstructuredToHPAv2(obj)
}

// Create 创建 HPA
func (o *HPAOperator) Create(ctx context.Context, namespace string, hpa *autoscalingv2.HorizontalPodAutoscaler) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	hpa.Namespace = namespace
	if o.dynamicClient != nil && o.apiVersion != "" {
		return o.createDynamic(ctx, namespace, hpa)
	}
	return o.createWithFallback(ctx, namespace, hpa)
}

func (o *HPAOperator) createWithFallback(ctx context.Context, namespace string, hpa *autoscalingv2.HorizontalPodAutoscaler) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	// 先尝试 v2
	result, err := o.client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Create(ctx, hpa, metav1.CreateOptions{})
	if err == nil {
		return result, nil
	}

	// v2 失败，尝试 v2beta2
	hpaBeta, err := convertV2ToV2beta2(hpa)
	if err != nil {
		return nil, err
	}
	resultBeta, err := o.client.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Create(ctx, hpaBeta, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return convertV2beta2ToV2(resultBeta)
}

func (o *HPAOperator) createDynamic(ctx context.Context, namespace string, hpa *autoscalingv2.HorizontalPodAutoscaler) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	obj, err := hpaV2ToUnstructured(hpa, o.apiVersion)
	if err != nil {
		return nil, err
	}
	result, err := o.dynamicClient.Resource(o.gvr).Namespace(namespace).Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	return unstructuredToHPAv2(result)
}

// Update 更新 HPA
func (o *HPAOperator) Update(ctx context.Context, namespace string, hpa *autoscalingv2.HorizontalPodAutoscaler) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	if o.dynamicClient != nil && o.apiVersion != "" {
		return o.updateDynamic(ctx, namespace, hpa)
	}
	return o.updateWithFallback(ctx, namespace, hpa)
}

func (o *HPAOperator) updateWithFallback(ctx context.Context, namespace string, hpa *autoscalingv2.HorizontalPodAutoscaler) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	// 先尝试 v2
	result, err := o.client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Update(ctx, hpa, metav1.UpdateOptions{})
	if err == nil {
		return result, nil
	}

	// v2 失败，尝试 v2beta2
	hpaBeta, err := convertV2ToV2beta2(hpa)
	if err != nil {
		return nil, err
	}
	resultBeta, err := o.client.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Update(ctx, hpaBeta, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	return convertV2beta2ToV2(resultBeta)
}

func (o *HPAOperator) updateDynamic(ctx context.Context, namespace string, hpa *autoscalingv2.HorizontalPodAutoscaler) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	obj, err := hpaV2ToUnstructured(hpa, o.apiVersion)
	if err != nil {
		return nil, err
	}
	result, err := o.dynamicClient.Resource(o.gvr).Namespace(namespace).Update(ctx, obj, metav1.UpdateOptions{})
	if err != nil {
		return nil, err
	}
	return unstructuredToHPAv2(result)
}

// Delete 删除 HPA
func (o *HPAOperator) Delete(ctx context.Context, namespace, name string) error {
	if o.dynamicClient != nil && o.apiVersion != "" {
		err := o.dynamicClient.Resource(o.gvr).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
		if k8serrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	// 先尝试 v2
	err := o.client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err == nil || k8serrors.IsNotFound(err) {
		return nil
	}

	// v2 失败，尝试 v2beta2
	err = o.client.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if k8serrors.IsNotFound(err) {
		return nil
	}
	return err
}

// Exists 检查 HPA 是否存在
func (o *HPAOperator) Exists(ctx context.Context, namespace, name string) (bool, error) {
	_, err := o.Get(ctx, namespace, name)
	if k8serrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// convertV2beta2ToV2 将 v2beta2 HPA 转换为 v2
func convertV2beta2ToV2(hpa *autoscalingv2beta2.HorizontalPodAutoscaler) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	data, err := json.Marshal(hpa)
	if err != nil {
		return nil, err
	}
	var result autoscalingv2.HorizontalPodAutoscaler
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	result.APIVersion = "autoscaling/v2"
	return &result, nil
}

// convertV2ToV2beta2 将 v2 HPA 转换为 v2beta2
func convertV2ToV2beta2(hpa *autoscalingv2.HorizontalPodAutoscaler) (*autoscalingv2beta2.HorizontalPodAutoscaler, error) {
	data, err := json.Marshal(hpa)
	if err != nil {
		return nil, err
	}
	var result autoscalingv2beta2.HorizontalPodAutoscaler
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	result.APIVersion = "autoscaling/v2beta2"
	return &result, nil
}

// unstructuredToHPAv2 将 unstructured 转换为 v2 HPA
func unstructuredToHPAv2(obj *unstructured.Unstructured) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	var hpa autoscalingv2.HorizontalPodAutoscaler
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &hpa)
	if err != nil {
		return nil, err
	}
	hpa.APIVersion = "autoscaling/v2"
	return &hpa, nil
}

// hpaV2ToUnstructured 将 v2 HPA 转换为 unstructured
func hpaV2ToUnstructured(hpa *autoscalingv2.HorizontalPodAutoscaler, apiVersion string) (*unstructured.Unstructured, error) {
	hpa.APIVersion = apiVersion
	hpa.Kind = "HorizontalPodAutoscaler"
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(hpa)
	if err != nil {
		return nil, err
	}
	return &unstructured.Unstructured{Object: obj}, nil
}

// HPAToJSON 序列化 HPA 为 JSON
func HPAToJSON(hpa *autoscalingv2.HorizontalPodAutoscaler) (string, error) {
	data, err := json.MarshalIndent(hpa, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal hpa: %w", err)
	}
	return string(data), nil
}

// JSONToHPA 从 JSON 反序列化 HPA
func JSONToHPA(data string) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	var hpa autoscalingv2.HorizontalPodAutoscaler
	if err := json.Unmarshal([]byte(data), &hpa); err != nil {
		return nil, fmt.Errorf("failed to unmarshal hpa: %w", err)
	}
	return &hpa, nil
}
