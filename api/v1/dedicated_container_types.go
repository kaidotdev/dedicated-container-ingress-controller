package v1

import (
	v1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DedicatedContainerIngressSpec struct {
	// Host is set for load-balancer ingress points that are DNS based
	Host string `json:"host"`
	// PodTemplate describes a template for creating copies of a predefined pod.
	Template v1.PodTemplateSpec `json:"template"`
}

type DedicatedContainerIngressStatus struct{}

// +kubebuilder:object:root=true

type DedicatedContainerIngress struct {
	metaV1.TypeMeta   `json:",inline"`
	metaV1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DedicatedContainerIngressSpec   `json:"spec,omitempty"`
	Status DedicatedContainerIngressStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type DedicatedContainerIngressList struct {
	metaV1.TypeMeta `json:",inline"`
	metaV1.ListMeta `json:"metadata,omitempty"`
	Items           []DedicatedContainerIngress `json:"items"`
}

func init() { // nolint:gochecknoinits
	SchemeBuilder.Register(&DedicatedContainerIngress{}, &DedicatedContainerIngressList{})
}
