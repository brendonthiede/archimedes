//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2021.

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

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArchimedesProperty) DeepCopyInto(out *ArchimedesProperty) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArchimedesProperty.
func (in *ArchimedesProperty) DeepCopy() *ArchimedesProperty {
	if in == nil {
		return nil
	}
	out := new(ArchimedesProperty)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ArchimedesProperty) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArchimedesPropertyList) DeepCopyInto(out *ArchimedesPropertyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ArchimedesProperty, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArchimedesPropertyList.
func (in *ArchimedesPropertyList) DeepCopy() *ArchimedesPropertyList {
	if in == nil {
		return nil
	}
	out := new(ArchimedesPropertyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ArchimedesPropertyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArchimedesPropertySpec) DeepCopyInto(out *ArchimedesPropertySpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArchimedesPropertySpec.
func (in *ArchimedesPropertySpec) DeepCopy() *ArchimedesPropertySpec {
	if in == nil {
		return nil
	}
	out := new(ArchimedesPropertySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ArchimedesPropertyStatus) DeepCopyInto(out *ArchimedesPropertyStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ArchimedesPropertyStatus.
func (in *ArchimedesPropertyStatus) DeepCopy() *ArchimedesPropertyStatus {
	if in == nil {
		return nil
	}
	out := new(ArchimedesPropertyStatus)
	in.DeepCopyInto(out)
	return out
}
