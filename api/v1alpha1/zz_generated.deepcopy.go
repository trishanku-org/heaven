//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2022 Amshuman K R <amshuman.kr@gmail.com>.

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
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ApiserverSpec) DeepCopyInto(out *ApiserverSpec) {
	*out = *in
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		*out = new(string)
		**out = **in
	}
	if in.ImagePullPolicy != nil {
		in, out := &in.ImagePullPolicy, &out.ImagePullPolicy
		*out = new(v1.PullPolicy)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ApiserverSpec.
func (in *ApiserverSpec) DeepCopy() *ApiserverSpec {
	if in == nil {
		return nil
	}
	out := new(ApiserverSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AppSpec) DeepCopyInto(out *AppSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.EntrypointsConfigMapName != nil {
		in, out := &in.EntrypointsConfigMapName, &out.EntrypointsConfigMapName
		*out = new(string)
		**out = **in
	}
	in.PodTemplate.DeepCopyInto(&out.PodTemplate)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AppSpec.
func (in *AppSpec) DeepCopy() *AppSpec {
	if in == nil {
		return nil
	}
	out := new(AppSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AutomatedMerge) DeepCopyInto(out *AutomatedMerge) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AutomatedMerge.
func (in *AutomatedMerge) DeepCopy() *AutomatedMerge {
	if in == nil {
		return nil
	}
	out := new(AutomatedMerge)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AutomatedMerge) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AutomatedMergeAppSpec) DeepCopyInto(out *AutomatedMergeAppSpec) {
	*out = *in
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
	if in.EntrypointsConfigMapName != nil {
		in, out := &in.EntrypointsConfigMapName, &out.EntrypointsConfigMapName
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AutomatedMergeAppSpec.
func (in *AutomatedMergeAppSpec) DeepCopy() *AutomatedMergeAppSpec {
	if in == nil {
		return nil
	}
	out := new(AutomatedMergeAppSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AutomatedMergeList) DeepCopyInto(out *AutomatedMergeList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]AutomatedMerge, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AutomatedMergeList.
func (in *AutomatedMergeList) DeepCopy() *AutomatedMergeList {
	if in == nil {
		return nil
	}
	out := new(AutomatedMergeList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *AutomatedMergeList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AutomatedMergeSpec) DeepCopyInto(out *AutomatedMergeSpec) {
	*out = *in
	in.Gitcd.DeepCopyInto(&out.Gitcd)
	in.App.DeepCopyInto(&out.App)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AutomatedMergeSpec.
func (in *AutomatedMergeSpec) DeepCopy() *AutomatedMergeSpec {
	if in == nil {
		return nil
	}
	out := new(AutomatedMergeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AutomatedMergeStatus) DeepCopyInto(out *AutomatedMergeStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AutomatedMergeStatus.
func (in *AutomatedMergeStatus) DeepCopy() *AutomatedMergeStatus {
	if in == nil {
		return nil
	}
	out := new(AutomatedMergeStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BranchSpec) DeepCopyInto(out *BranchSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BranchSpec.
func (in *BranchSpec) DeepCopy() *BranchSpec {
	if in == nil {
		return nil
	}
	out := new(BranchSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CertificatesSpec) DeepCopyInto(out *CertificatesSpec) {
	*out = *in
	if in.CertificateAuthoritySecretName != nil {
		in, out := &in.CertificateAuthoritySecretName, &out.CertificateAuthoritySecretName
		*out = new(string)
		**out = **in
	}
	if in.ServiceAccountsSecretName != nil {
		in, out := &in.ServiceAccountsSecretName, &out.ServiceAccountsSecretName
		*out = new(string)
		**out = **in
	}
	if in.KubernetesSecretName != nil {
		in, out := &in.KubernetesSecretName, &out.KubernetesSecretName
		*out = new(string)
		**out = **in
	}
	if in.AdminSecretName != nil {
		in, out := &in.AdminSecretName, &out.AdminSecretName
		*out = new(string)
		**out = **in
	}
	if in.Controller != nil {
		in, out := &in.Controller, &out.Controller
		*out = new(ControllerCertificateSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CertificatesSpec.
func (in *CertificatesSpec) DeepCopy() *CertificatesSpec {
	if in == nil {
		return nil
	}
	out := new(CertificatesSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ControllerCertificateSpec) DeepCopyInto(out *ControllerCertificateSpec) {
	*out = *in
	if in.SecretName != nil {
		in, out := &in.SecretName, &out.SecretName
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControllerCertificateSpec.
func (in *ControllerCertificateSpec) DeepCopy() *ControllerCertificateSpec {
	if in == nil {
		return nil
	}
	out := new(ControllerCertificateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EventsEtcdLocalSpec) DeepCopyInto(out *EventsEtcdLocalSpec) {
	*out = *in
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		*out = new(string)
		**out = **in
	}
	if in.ImagePullPolicy != nil {
		in, out := &in.ImagePullPolicy, &out.ImagePullPolicy
		*out = new(v1.PullPolicy)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EventsEtcdLocalSpec.
func (in *EventsEtcdLocalSpec) DeepCopy() *EventsEtcdLocalSpec {
	if in == nil {
		return nil
	}
	out := new(EventsEtcdLocalSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EventsEtcdSpec) DeepCopyInto(out *EventsEtcdSpec) {
	*out = *in
	if in.Local != nil {
		in, out := &in.Local, &out.Local
		*out = new(EventsEtcdLocalSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EventsEtcdSpec.
func (in *EventsEtcdSpec) DeepCopy() *EventsEtcdSpec {
	if in == nil {
		return nil
	}
	out := new(EventsEtcdSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitHubRepository) DeepCopyInto(out *GitHubRepository) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitHubRepository.
func (in *GitHubRepository) DeepCopy() *GitHubRepository {
	if in == nil {
		return nil
	}
	out := new(GitHubRepository)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GitHubRepository) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitHubRepositoryList) DeepCopyInto(out *GitHubRepositoryList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GitHubRepository, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitHubRepositoryList.
func (in *GitHubRepositoryList) DeepCopy() *GitHubRepositoryList {
	if in == nil {
		return nil
	}
	out := new(GitHubRepositoryList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GitHubRepositoryList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitHubRepositorySpec) DeepCopyInto(out *GitHubRepositorySpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitHubRepositorySpec.
func (in *GitHubRepositorySpec) DeepCopy() *GitHubRepositorySpec {
	if in == nil {
		return nil
	}
	out := new(GitHubRepositorySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitHubRepositoryStatus) DeepCopyInto(out *GitHubRepositoryStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitHubRepositoryStatus.
func (in *GitHubRepositoryStatus) DeepCopy() *GitHubRepositoryStatus {
	if in == nil {
		return nil
	}
	out := new(GitHubRepositoryStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitSpec) DeepCopyInto(out *GitSpec) {
	*out = *in
	out.Branches = in.Branches
	if in.Remotes != nil {
		in, out := &in.Remotes, &out.Remotes
		*out = make([]RemoteSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitSpec.
func (in *GitSpec) DeepCopy() *GitSpec {
	if in == nil {
		return nil
	}
	out := new(GitSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GitcdSpec) DeepCopyInto(out *GitcdSpec) {
	*out = *in
	if in.GitImage != nil {
		in, out := &in.GitImage, &out.GitImage
		*out = new(ImageSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.GitcdImage != nil {
		in, out := &in.GitcdImage, &out.GitcdImage
		*out = new(ImageSpec)
		(*in).DeepCopyInto(*out)
	}
	in.Git.DeepCopyInto(&out.Git)
	out.Pull = in.Pull
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GitcdSpec.
func (in *GitcdSpec) DeepCopy() *GitcdSpec {
	if in == nil {
		return nil
	}
	out := new(GitcdSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ImageSpec) DeepCopyInto(out *ImageSpec) {
	*out = *in
	if in.Image != nil {
		in, out := &in.Image, &out.Image
		*out = new(string)
		**out = **in
	}
	if in.ImagePullPolicy != nil {
		in, out := &in.ImagePullPolicy, &out.ImagePullPolicy
		*out = new(v1.PullPolicy)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ImageSpec.
func (in *ImageSpec) DeepCopy() *ImageSpec {
	if in == nil {
		return nil
	}
	out := new(ImageSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodTemplateSpec) DeepCopyInto(out *PodTemplateSpec) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodTemplateSpec.
func (in *PodTemplateSpec) DeepCopy() *PodTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(PodTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PullSpec) DeepCopyInto(out *PullSpec) {
	*out = *in
	out.TickerDuration = in.TickerDuration
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PullSpec.
func (in *PullSpec) DeepCopy() *PullSpec {
	if in == nil {
		return nil
	}
	out := new(PullSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RemoteSpec) DeepCopyInto(out *RemoteSpec) {
	*out = *in
	if in.Branches != nil {
		in, out := &in.Branches, &out.Branches
		*out = new(BranchSpec)
		**out = **in
	}
	out.RetentionPolicies = in.RetentionPolicies
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RemoteSpec.
func (in *RemoteSpec) DeepCopy() *RemoteSpec {
	if in == nil {
		return nil
	}
	out := new(RemoteSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RetentionPoliciesSpec) DeepCopyInto(out *RetentionPoliciesSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RetentionPoliciesSpec.
func (in *RetentionPoliciesSpec) DeepCopy() *RetentionPoliciesSpec {
	if in == nil {
		return nil
	}
	out := new(RetentionPoliciesSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SkipSpec) DeepCopyInto(out *SkipSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SkipSpec.
func (in *SkipSpec) DeepCopy() *SkipSpec {
	if in == nil {
		return nil
	}
	out := new(SkipSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrishankuHeaven) DeepCopyInto(out *TrishankuHeaven) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrishankuHeaven.
func (in *TrishankuHeaven) DeepCopy() *TrishankuHeaven {
	if in == nil {
		return nil
	}
	out := new(TrishankuHeaven)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TrishankuHeaven) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrishankuHeavenList) DeepCopyInto(out *TrishankuHeavenList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TrishankuHeaven, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrishankuHeavenList.
func (in *TrishankuHeavenList) DeepCopy() *TrishankuHeavenList {
	if in == nil {
		return nil
	}
	out := new(TrishankuHeavenList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TrishankuHeavenList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrishankuHeavenSpec) DeepCopyInto(out *TrishankuHeavenSpec) {
	*out = *in
	out.Skip = in.Skip
	in.Certificates.DeepCopyInto(&out.Certificates)
	if in.EventsEtcd != nil {
		in, out := &in.EventsEtcd, &out.EventsEtcd
		*out = new(EventsEtcdSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Gitcd != nil {
		in, out := &in.Gitcd, &out.Gitcd
		*out = new(GitcdSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.Apiserver != nil {
		in, out := &in.Apiserver, &out.Apiserver
		*out = new(ApiserverSpec)
		(*in).DeepCopyInto(*out)
	}
	if in.App != nil {
		in, out := &in.App, &out.App
		*out = new(AppSpec)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrishankuHeavenSpec.
func (in *TrishankuHeavenSpec) DeepCopy() *TrishankuHeavenSpec {
	if in == nil {
		return nil
	}
	out := new(TrishankuHeavenSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrishankuHeavenStatus) DeepCopyInto(out *TrishankuHeavenStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrishankuHeavenStatus.
func (in *TrishankuHeavenStatus) DeepCopy() *TrishankuHeavenStatus {
	if in == nil {
		return nil
	}
	out := new(TrishankuHeavenStatus)
	in.DeepCopyInto(out)
	return out
}
