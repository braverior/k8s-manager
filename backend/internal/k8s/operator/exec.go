package operator

import (
	"context"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// ExecOperator 用于执行 Pod exec 操作
type ExecOperator struct {
	client *kubernetes.Clientset
	config *rest.Config
}

// NewExecOperator 创建 ExecOperator
func NewExecOperator(client *kubernetes.Clientset, config *rest.Config) *ExecOperator {
	return &ExecOperator{
		client: client,
		config: config,
	}
}

// TerminalSizeQueue 终端大小队列接口
type TerminalSizeQueue interface {
	remotecommand.TerminalSizeQueue
}

// ExecOptions exec 执行选项
type ExecOptions struct {
	Namespace     string
	PodName       string
	ContainerName string
	Command       []string
	Stdin         io.Reader
	Stdout        io.Writer
	Stderr        io.Writer
	TTY           bool
	SizeQueue     TerminalSizeQueue
}

// Exec 执行 Pod exec 命令
func (o *ExecOperator) Exec(ctx context.Context, opts *ExecOptions) error {
	req := o.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(opts.PodName).
		Namespace(opts.Namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: opts.ContainerName,
		Command:   opts.Command,
		Stdin:     opts.Stdin != nil,
		Stdout:    opts.Stdout != nil,
		Stderr:    opts.Stderr != nil,
		TTY:       opts.TTY,
	}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(o.config, "POST", req.URL())
	if err != nil {
		return err
	}

	streamOpts := remotecommand.StreamOptions{
		Stdin:  opts.Stdin,
		Stdout: opts.Stdout,
		Stderr: opts.Stderr,
		Tty:    opts.TTY,
	}

	if opts.TTY && opts.SizeQueue != nil {
		streamOpts.TerminalSizeQueue = opts.SizeQueue
	}

	return exec.StreamWithContext(ctx, streamOpts)
}

// GetContainers 获取 Pod 的容器列表
func (o *ExecOperator) GetContainers(ctx context.Context, namespace, podName string) ([]corev1.Container, []corev1.ContainerStatus, error) {
	pod, err := o.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, nil, err
	}
	return pod.Spec.Containers, pod.Status.ContainerStatuses, nil
}
