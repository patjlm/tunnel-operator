/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	TunnelConditionCreatedType          string = "Created"
	TunnelConditionCreatedFailedReason  string = "CreationFailed"
	TunnelConditionCreatedExistsReason  string = "AlreadyExists"
	TunnelConditionCreatedSuccessReason string = "CreationSucceeded"
)

type TunnelIngress struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// HostName is the hostname that can be used to reach this tunnel ingress
	HostName string `json:"hostname" yaml:"hostname,omitempty"`

	// Service represents the backend service URL reached through this tunnel ingress
	Service *string `json:"service,omitempty" yaml:"service,omitempty"`
}

// TunnelSpec defines the desired state of Tunnel
type TunnelSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// Name is the name of the tunnel to create
	Name string `json:"name"`

	// AccountSecret is a reference to a secret containing the cloudflare account API token
	AccountSecret *corev1.SecretReference `json:"accountSecret,omitempty"`

	// TunnelSecret is a reference to the secret to create with the tunnel information
	TunnelSecret *corev1.SecretReference `json:"secret,omitempty"`

	Ingress *[]TunnelIngress `json:"ingress,omitempty"`
}

// TunnelStatus defines the observed state of Tunnel
type TunnelStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// AccountID is the ID of the cloudflare account in which this tunnel is created
	AccountID string `json:"accountid"`

	// TunnelID is the id of the created cloudflare tunnel
	TunnelID string `json:"tunnelid"`

	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions"`

	// IngressHostnames lists the hostnames recorded in DNS
	IngressHostnames []string `json:"hostnames"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Tunnel is the Schema for the tunnels API
type Tunnel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TunnelSpec   `json:"spec,omitempty"`
	Status TunnelStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TunnelList contains a list of Tunnel
type TunnelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tunnel `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tunnel{}, &TunnelList{})
}
