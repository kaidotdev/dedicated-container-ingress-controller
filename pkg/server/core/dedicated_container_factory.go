package core

import (
	"context"
	"sync"
	"time"

	"golang.org/x/xerrors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

type DedicatedContainerFactory struct {
	table     sync.Map
	clientset kubernetes.Interface
}

func NewDedicatedContainerFactory(clientset kubernetes.Interface) *DedicatedContainerFactory {
	return &DedicatedContainerFactory{clientset: clientset}
}

func (r *DedicatedContainerFactory) AddEntry(host string, podTemplateSpec v1.PodTemplateSpec) {
	r.table.Store(host, podTemplateSpec)
}

func (r *DedicatedContainerFactory) DeleteEntry(host string) {
	r.table.Delete(host)
}

func (r *DedicatedContainerFactory) HasEntry(host string) bool {
	_, ok := r.table.Load(host)
	return ok
}

func (r *DedicatedContainerFactory) Create(ctx context.Context, host string) (*v1.Pod, error) {
	v, ok := r.table.Load(host)
	if !ok {
		return nil, xerrors.Errorf("no backend found")
	}
	podTemplateSpec := v.(v1.PodTemplateSpec)
	objectMeta := podTemplateSpec.ObjectMeta
	objectMeta.GenerateName = host + "-"
	labels := map[string]string{
		"owner": "dedicated-container-ingress-controller",
	}
	for k, v := range objectMeta.Labels {
		labels[k] = v
	}
	objectMeta.Labels = labels
	createdPod, err := r.clientset.CoreV1().Pods(podTemplateSpec.Namespace).Create(ctx, &v1.Pod{
		ObjectMeta: objectMeta,
		Spec:       podTemplateSpec.Spec,
	}, metav1.CreateOptions{})
	if err != nil {
		return nil, xerrors.Errorf("failed to get pod: %w", err)
	}

	var pod *v1.Pod
	if err := wait.PollImmediate(100*time.Millisecond, 30*time.Second, func() (bool, error) {
		pod, err = r.clientset.CoreV1().Pods(podTemplateSpec.Namespace).Get(ctx, createdPod.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if pod.Status.Phase == v1.PodRunning {
			return true, nil
		}
		return false, nil
	}); err != nil {
		return nil, xerrors.Errorf("failed to wait pod creation: %w", err)
	}

	return pod, nil
}
