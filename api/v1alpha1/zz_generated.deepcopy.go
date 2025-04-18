//go:build !ignore_autogenerated

/*
                    GNU GENERAL PUBLIC LICENSE
                       Version 2, June 1991

 Copyright (C) 1989, 1991 Free Software Foundation, Inc.,
 51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA
 Everyone is permitted to copy and distribute verbatim copies
 of this license document, but changing it is not allowed.
*/

// SPDX-License-Identifier: GPL-2.0-only

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Hive) DeepCopyInto(out *Hive) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Hive.
func (in *Hive) DeepCopy() *Hive {
	if in == nil {
		return nil
	}
	out := new(Hive)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Hive) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HiveData) DeepCopyInto(out *HiveData) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HiveData.
func (in *HiveData) DeepCopy() *HiveData {
	if in == nil {
		return nil
	}
	out := new(HiveData)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HiveData) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HiveDataList) DeepCopyInto(out *HiveDataList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]HiveData, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HiveDataList.
func (in *HiveDataList) DeepCopy() *HiveDataList {
	if in == nil {
		return nil
	}
	out := new(HiveDataList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HiveDataList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HiveDataSpec) DeepCopyInto(out *HiveDataSpec) {
	*out = *in
	if in.HiveData != nil {
		in, out := &in.HiveData, &out.HiveData
		*out = make([]HiveDataType, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HiveDataSpec.
func (in *HiveDataSpec) DeepCopy() *HiveDataSpec {
	if in == nil {
		return nil
	}
	out := new(HiveDataSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HiveDataStatus) DeepCopyInto(out *HiveDataStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HiveDataStatus.
func (in *HiveDataStatus) DeepCopy() *HiveDataStatus {
	if in == nil {
		return nil
	}
	out := new(HiveDataStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HiveDataType) DeepCopyInto(out *HiveDataType) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HiveDataType.
func (in *HiveDataType) DeepCopy() *HiveDataType {
	if in == nil {
		return nil
	}
	out := new(HiveDataType)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HiveList) DeepCopyInto(out *HiveList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Hive, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HiveList.
func (in *HiveList) DeepCopy() *HiveList {
	if in == nil {
		return nil
	}
	out := new(HiveList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HiveList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HivePolicy) DeepCopyInto(out *HivePolicy) {
	*out = *in
	in.Match.DeepCopyInto(&out.Match)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HivePolicy.
func (in *HivePolicy) DeepCopy() *HivePolicy {
	if in == nil {
		return nil
	}
	out := new(HivePolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HivePolicyMatch) DeepCopyInto(out *HivePolicyMatch) {
	*out = *in
	if in.Pod != nil {
		in, out := &in.Pod, &out.Pod
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Namespace != nil {
		in, out := &in.Namespace, &out.Namespace
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Label != nil {
		in, out := &in.Label, &out.Label
		*out = make([]LabelType, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HivePolicyMatch.
func (in *HivePolicyMatch) DeepCopy() *HivePolicyMatch {
	if in == nil {
		return nil
	}
	out := new(HivePolicyMatch)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HiveStatus) DeepCopyInto(out *HiveStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HiveStatus.
func (in *HiveStatus) DeepCopy() *HiveStatus {
	if in == nil {
		return nil
	}
	out := new(HiveStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LabelType) DeepCopyInto(out *LabelType) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LabelType.
func (in *LabelType) DeepCopy() *LabelType {
	if in == nil {
		return nil
	}
	out := new(LabelType)
	in.DeepCopyInto(out)
	return out
}
