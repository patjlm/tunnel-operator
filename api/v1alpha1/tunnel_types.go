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
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	TunnelConditionCreatedType          string = "Created"
	TunnelConditionCreatedFailedReason  string = "CreationFailed"
	TunnelConditionCreatedExistsReason  string = "AlreadyExists"
	TunnelConditionCreatedSuccessReason string = "CreationSucceeded"
)

const (
	TunnelDefaultRun bool = false
)

// copied from https://github.com/cloudflare/cloudflared/blob/master/config/configuration.go
// OriginRequestConfig is a set of optional fields that users may set to
// customize how cloudflared sends requests to origin services. It is used to set
// up general config that apply to all rules, and also, specific per-rule
// config.
// Note: To specify a time.Duration in go-yaml, use e.g. "3s" or "24h".
type OriginRequestConfig struct {
	// HTTP proxy timeout for establishing a new connection
	ConnectTimeout *time.Duration `json:"connectTimeout,omitempty" yaml:"connectTimeout,omitempty"`
	// HTTP proxy timeout for completing a TLS handshake
	TLSTimeout *time.Duration `json:"tlsTimeout,omitempty" yaml:"tlsTimeout,omitempty"`
	// HTTP proxy TCP keepalive duration
	TCPKeepAlive *time.Duration `json:"tcpKeepAlive,omitempty" yaml:"tcpKeepAlive,omitempty"`
	// HTTP proxy should disable "happy eyeballs" for IPv4/v6 fallback
	NoHappyEyeballs *bool `json:"noHappyEyeballs,omitempty" yaml:"noHappyEyeballs,omitempty"`
	// HTTP proxy maximum keepalive connection pool size
	KeepAliveConnections *int `json:"keepAliveConnections,omitempty" yaml:"keepAliveConnections,omitempty"`
	// HTTP proxy timeout for closing an idle connection
	KeepAliveTimeout *time.Duration `json:"keepAliveTimeout,omitempty" yaml:"keepAliveTimeout,omitempty"`
	// Sets the HTTP Host header for the local webserver.
	HTTPHostHeader *string `json:"httpHostHeader,omitempty" yaml:"httpHostHeader,omitempty"`
	// Hostname on the origin server certificate.
	OriginServerName *string `json:"originServerName,omitempty" yaml:"originServerName,omitempty"`
	// Path to the CA for the certificate of your origin.
	// This option should be used only if your certificate is not signed by Cloudflare.
	CAPool *string `json:"caPool,omitempty" yaml:"caPool,omitempty"`
	// Disables TLS verification of the certificate presented by your origin.
	// Will allow any certificate from the origin to be accepted.
	// Note: The connection from your machine to Cloudflare's Edge is still encrypted.
	NoTLSVerify *bool `json:"noTLSVerify,omitempty" yaml:"noTLSVerify,omitempty"`
	// Disables chunked transfer encoding.
	// Useful if you are running a WSGI server.
	DisableChunkedEncoding *bool `json:"disableChunkedEncoding,omitempty" yaml:"disableChunkedEncoding,omitempty"`
	// Runs as jump host
	BastionMode *bool `json:"bastionMode,omitempty" yaml:"bastionMode,omitempty"`
	// Listen address for the proxy.
	ProxyAddress *string `json:"proxyAddress,omitempty" yaml:"proxyAddress,omitempty"`
	// Listen port for the proxy.
	ProxyPort *uint `json:"proxyPort,omitempty" yaml:"proxyPort,omitempty"`
	// Valid options are 'socks' or empty.
	ProxyType *string `json:"proxyType,omitempty" yaml:"proxyType,omitempty"`
	// IP rules for the proxy service
	IPRules []IngressIPRule `json:"ipRules,omitempty" yaml:"ipRules,omitempty"`
}

type IngressIPRule struct {
	Prefix *string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Ports  []int   `json:"ports,omitempty" yaml:"ports,omitempty"`
	Allow  bool    `json:"allow,omitempty" yaml:"allow,omitempty"`
}

type TunnelIngress struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// HostName is the hostname that can be used to reach this tunnel ingress
	HostName      string               `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	Path          *string              `json:"path,omitempty" yaml:"path,omitempty"`
	Service       *string              `json:"service,omitempty" yaml:"service,omitempty"`
	OriginRequest *OriginRequestConfig `json:"originRequest,omitempty" yaml:"originRequest,omitempty"`
}

// TunnelSpec defines the desired state of Tunnel
type TunnelSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// Name is the name of the tunnel to create
	Name string `json:"name"`

	// AccountSecret is a reference to a secret containing the cloudflare account API token
	AccountSecret *corev1.SecretReference `json:"accountSecret,omitempty"`

	// TunnelSecret is a reference to the secret to create with the tunnel information
	// TunnelSecret *corev1.SecretReference `json:"secret,omitempty"`
	TunnelSecretName *string `json:"secretName,omitempty"`

	Ingress *[]TunnelIngress `json:"ingress,omitempty"`

	Run            bool                   `json:"run,omitempty"`
	DeploymentSpec *appsv1.DeploymentSpec `json:"deploymentSpec,omitempty"`
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
	IngressHostnames []string `json:"hostnames,omitempty"`
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

func (t *Tunnel) BaseTunnelSecret() *corev1.Secret {
	name := t.Name
	if t.Spec.TunnelSecretName != nil {
		name = *t.Spec.TunnelSecretName
	}
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: t.Namespace,
		},
	}
}

func (t *Tunnel) DefaultDeploymentSpec() appsv1.DeploymentSpec {
	labelSelector := t.DefaultDeploymentLabelSelector()
	var replicas int32 = 1
	var optionalOpenshitCA = true
	var secretName = t.Name
	if t.Spec.TunnelSecretName != nil {
		secretName = *t.Spec.TunnelSecretName
	}

	return appsv1.DeploymentSpec{
		Replicas: &replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: labelSelector,
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: labelSelector,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{{
					Image: "cloudflare/cloudflared:2022.1.3",
					Name:  "cloudflared",
					Args: []string{
						"tunnel",
						"--config", "/config/config.yaml",
						"--metrics", "0.0.0.0:10000",
						"run",
						"--credentials-file", "/config/credentials.json",
					},
					Ports: []corev1.ContainerPort{{
						Name:          "metrics",
						ContainerPort: 10000,
					}},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "cloudflared-config",
							MountPath: "/config",
							ReadOnly:  true,
						},
						{
							Name:      "openshift-ca",
							MountPath: "/openshift-ca",
							ReadOnly:  true,
						},
					},
				}},
				Volumes: []corev1.Volume{
					{
						Name: "cloudflared-config",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: secretName,
							},
						},
					},
					{
						Name: "openshift-ca",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "openshift-ca",
								},
								Optional: &optionalOpenshitCA,
							},
						},
					},
				},
			},
		},
	}
}

func (t *Tunnel) DefaultDeploymentLabelSelector() map[string]string {
	return map[string]string{"app": "cloudflared-run", "tunnel-id": t.Status.TunnelID}
}

func (t *Tunnel) DeploymentForTunnelRun() *appsv1.Deployment {
	labels := t.DefaultDeploymentLabelSelector()

	deploymentSpec := t.DefaultDeploymentSpec()
	if t.Spec.DeploymentSpec != nil {
		deploymentSpec = *t.Spec.DeploymentSpec
		deploymentSpec.Selector.MatchLabels = labels
		deploymentSpec.Template.ObjectMeta.Labels = labels
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      t.Name,
			Namespace: t.Namespace,
			Labels:    labels,
		},
		Spec: deploymentSpec,
	}
	return dep
}
