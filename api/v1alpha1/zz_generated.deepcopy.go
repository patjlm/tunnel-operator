//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	timex "time"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IngressIPRule) DeepCopyInto(out *IngressIPRule) {
	*out = *in
	if in.Prefix != nil {
		in, out := &in.Prefix, &out.Prefix
		*out = new(string)
		**out = **in
	}
	if in.Ports != nil {
		in, out := &in.Ports, &out.Ports
		*out = make([]int, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IngressIPRule.
func (in *IngressIPRule) DeepCopy() *IngressIPRule {
	if in == nil {
		return nil
	}
	out := new(IngressIPRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OriginRequestConfig) DeepCopyInto(out *OriginRequestConfig) {
	*out = *in
	if in.ConnectTimeout != nil {
		in, out := &in.ConnectTimeout, &out.ConnectTimeout
		*out = new(timex.Duration)
		**out = **in
	}
	if in.TLSTimeout != nil {
		in, out := &in.TLSTimeout, &out.TLSTimeout
		*out = new(timex.Duration)
		**out = **in
	}
	if in.TCPKeepAlive != nil {
		in, out := &in.TCPKeepAlive, &out.TCPKeepAlive
		*out = new(timex.Duration)
		**out = **in
	}
	if in.NoHappyEyeballs != nil {
		in, out := &in.NoHappyEyeballs, &out.NoHappyEyeballs
		*out = new(bool)
		**out = **in
	}
	if in.KeepAliveConnections != nil {
		in, out := &in.KeepAliveConnections, &out.KeepAliveConnections
		*out = new(int)
		**out = **in
	}
	if in.KeepAliveTimeout != nil {
		in, out := &in.KeepAliveTimeout, &out.KeepAliveTimeout
		*out = new(timex.Duration)
		**out = **in
	}
	if in.HTTPHostHeader != nil {
		in, out := &in.HTTPHostHeader, &out.HTTPHostHeader
		*out = new(string)
		**out = **in
	}
	if in.OriginServerName != nil {
		in, out := &in.OriginServerName, &out.OriginServerName
		*out = new(string)
		**out = **in
	}
	if in.CAPool != nil {
		in, out := &in.CAPool, &out.CAPool
		*out = new(string)
		**out = **in
	}
	if in.NoTLSVerify != nil {
		in, out := &in.NoTLSVerify, &out.NoTLSVerify
		*out = new(bool)
		**out = **in
	}
	if in.DisableChunkedEncoding != nil {
		in, out := &in.DisableChunkedEncoding, &out.DisableChunkedEncoding
		*out = new(bool)
		**out = **in
	}
	if in.BastionMode != nil {
		in, out := &in.BastionMode, &out.BastionMode
		*out = new(bool)
		**out = **in
	}
	if in.ProxyAddress != nil {
		in, out := &in.ProxyAddress, &out.ProxyAddress
		*out = new(string)
		**out = **in
	}
	if in.ProxyPort != nil {
		in, out := &in.ProxyPort, &out.ProxyPort
		*out = new(uint)
		**out = **in
	}
	if in.ProxyType != nil {
		in, out := &in.ProxyType, &out.ProxyType
		*out = new(string)
		**out = **in
	}
	if in.IPRules != nil {
		in, out := &in.IPRules, &out.IPRules
		*out = make([]IngressIPRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OriginRequestConfig.
func (in *OriginRequestConfig) DeepCopy() *OriginRequestConfig {
	if in == nil {
		return nil
	}
	out := new(OriginRequestConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Tunnel) DeepCopyInto(out *Tunnel) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Tunnel.
func (in *Tunnel) DeepCopy() *Tunnel {
	if in == nil {
		return nil
	}
	out := new(Tunnel)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Tunnel) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TunnelIngress) DeepCopyInto(out *TunnelIngress) {
	*out = *in
	if in.Path != nil {
		in, out := &in.Path, &out.Path
		*out = new(string)
		**out = **in
	}
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(string)
		**out = **in
	}
	if in.OriginRequest != nil {
		in, out := &in.OriginRequest, &out.OriginRequest
		*out = new(OriginRequestConfig)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TunnelIngress.
func (in *TunnelIngress) DeepCopy() *TunnelIngress {
	if in == nil {
		return nil
	}
	out := new(TunnelIngress)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TunnelList) DeepCopyInto(out *TunnelList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Tunnel, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TunnelList.
func (in *TunnelList) DeepCopy() *TunnelList {
	if in == nil {
		return nil
	}
	out := new(TunnelList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TunnelList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TunnelSpec) DeepCopyInto(out *TunnelSpec) {
	*out = *in
	if in.AccountSecret != nil {
		in, out := &in.AccountSecret, &out.AccountSecret
		*out = new(v1.SecretReference)
		**out = **in
	}
	if in.TunnelSecret != nil {
		in, out := &in.TunnelSecret, &out.TunnelSecret
		*out = new(v1.SecretReference)
		**out = **in
	}
	if in.Ingress != nil {
		in, out := &in.Ingress, &out.Ingress
		*out = new([]TunnelIngress)
		if **in != nil {
			in, out := *in, *out
			*out = make([]TunnelIngress, len(*in))
			for i := range *in {
				(*in)[i].DeepCopyInto(&(*out)[i])
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TunnelSpec.
func (in *TunnelSpec) DeepCopy() *TunnelSpec {
	if in == nil {
		return nil
	}
	out := new(TunnelSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TunnelStatus) DeepCopyInto(out *TunnelStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.IngressHostnames != nil {
		in, out := &in.IngressHostnames, &out.IngressHostnames
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TunnelStatus.
func (in *TunnelStatus) DeepCopy() *TunnelStatus {
	if in == nil {
		return nil
	}
	out := new(TunnelStatus)
	in.DeepCopyInto(out)
	return out
}
