package operator

import (
	"context"
	"encoding/json"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type DeploymentOperator struct {
	client *kubernetes.Clientset
}

func NewDeploymentOperator(client *kubernetes.Clientset) *DeploymentOperator {
	return &DeploymentOperator{client: client}
}

func (o *DeploymentOperator) List(ctx context.Context, namespace string) ([]appsv1.Deployment, error) {
	list, err := o.client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

func (o *DeploymentOperator) Get(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	return o.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (o *DeploymentOperator) Create(ctx context.Context, namespace string, deploy *appsv1.Deployment) (*appsv1.Deployment, error) {
	deploy.Namespace = namespace
	return o.client.AppsV1().Deployments(namespace).Create(ctx, deploy, metav1.CreateOptions{})
}

func (o *DeploymentOperator) Update(ctx context.Context, namespace string, deploy *appsv1.Deployment) (*appsv1.Deployment, error) {
	return o.client.AppsV1().Deployments(namespace).Update(ctx, deploy, metav1.UpdateOptions{})
}

func (o *DeploymentOperator) Delete(ctx context.Context, namespace, name string) error {
	err := o.client.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if k8serrors.IsNotFound(err) {
		return nil
	}
	return err
}

func (o *DeploymentOperator) Exists(ctx context.Context, namespace, name string) (bool, error) {
	_, err := o.Get(ctx, namespace, name)
	if k8serrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func DeploymentToJSON(deploy *appsv1.Deployment) (string, error) {
	data, err := json.MarshalIndent(deploy, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal deployment: %w", err)
	}
	return string(data), nil
}

func JSONToDeployment(data string) (*appsv1.Deployment, error) {
	var deploy appsv1.Deployment
	if err := json.Unmarshal([]byte(data), &deploy); err != nil {
		return nil, fmt.Errorf("failed to unmarshal deployment: %w", err)
	}
	return &deploy, nil
}
