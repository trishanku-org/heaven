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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TrishankuHeavenSpec defines the desired state of TrishankuHeaven
type TrishankuHeavenSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// EventsEtcd specifies the configuration for events etcd.
	// +Optional
	EventsEtcd *EventsEtcdSpec `json:"eventsEtcd,omitempty"`

	// Gitcd specifies the Gitcd configuration.
	// +Required
	Gitcd GitcdSpec `json:"git"`

	// Apiserver specifies the Apiserver configuration.
	// +Optional
	Apiserver *ApiserverSpec `json:"apiserver,omitempty"`

	// ControllerUser specifies the user details the controller is to be configured to use.
	// +Required
	ControllerUser ControllerUserSpec `json:"controllerUser"`

	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	// template specifies the pod template for the controller deployment.
	// +Required
	Template PodTemplateSpec `json:"template"`
}

// ImageSpec specifies the image details for a container.
type ImageSpec struct {
	// Image specifies the container image.
	// +Optional
	Image *string `json:"image,omitempty"`

	//ImagePullPolicy specifies the image pull policy for the container image.
	// +Optional
	ImagePullPolicy *corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
}

type EventsEtcdSpec struct {
	// ImagePullPolicy specifies the image pull policy for the events etcd container image.
	// +Optional
	Local *EventsEtcdLocalSpec `json:"local,omitempty"`
}

type EventsEtcdLocalSpec ImageSpec

// GitcdSpec specifies the configuration for Gitcd
type GitcdSpec struct {
	// GitImage specifie the git container image.
	// +Optional
	GitImage *ImageSpec `json:"gitImage,omitempty"`

	// GitcdImage specifie the gitcd container image.
	// +Optional
	GitcdImage *ImageSpec `json:"gitcdImage,omitempty"`

	// CredentialsSecretName specifies the name of the secret containing the .git-credentials file.
	// +Required
	CredentialsSecretName string `json:"credentialsSecretName"`

	// ImagePullSecretName specifies the name of the secret containing the credentials to pull gitcd image.
	// +Optional
	ImagePullSecretName string `json:"imagePullSecretName"`

	// RemoteRepo specifies the URL or path to the remote repo for the git repo that will back gitcd.
	// +Optional
	RemoteRepo string `json:"remoteRepo,omitempty"`

	// Branches specifies the local and remote data and metadata branches for gitcd.
	// +Required
	Branches BranchesSpec `json:"branches"`

	// Pull specifies the pull configuration for gitcd.
	// +Required
	Pull PullSpec `json:"pull"`

	// PullRequest specifies the pull request configuration for gitcd.
	// *Optional
	PullRequest *PullRequestSpec `json:"pullRequest,omitempty"`
}

// BranchesSpec specifies the local and remote branches for Gitcd
type BranchesSpec struct {
	// Local specifies the local data and metadata branches.
	// +Required
	Local BranchSpec `json:"local"`

	// Remote specifies the remote data and metadata branches.
	// +Optional
	Remote *BranchSpec `json:"remote,omitempty"`
}

// BranchSpec specifies the data and metadata branches for Gitcd
type BranchSpec struct {
	// Data specifies the data branch.
	// +Required
	Data string `json:"data,omitempty"`

	// Metadata specifies the data branch.
	// +Required
	Metadata string `json:"metadata,omitempty"`
}

// PullSpec specifies the pull configuration for Gitcd.
type PullSpec struct {
	// RetentionPolicies specifies the retention policies while merging changes from remote.
	// +Optional
	RetentionPolicies RetentionPoliciesSpec `json:"retentionPolicies,omitempty"`

	// ConflictResolutions specifies the conflict resolution policy while merging changes from remote.
	// +Optional
	// +default:=default=1
	ConflictResolutions string `json:"conflictResolutions,omitempty"`

	// PushAfterMerge specifies if the local branch must be pushed to remote after merging from remote.
	// +Optional
	// +default=true
	PushAfterMerge bool `json:"pushAfterMerge,omitempty"`

	// TickerDuration species the ticker duration between automated pulls from remote.
	// +Required
	TickerDuration metav1.Duration `json:"tickerDuration"`
}

// RetentionPoliciesSpec specifies the merge retention policies for Gitcd.
type RetentionPoliciesSpec struct {
	// Include specifies the include retention policies.
	// +Optional
	// +default:=default=.*
	Include string `json:"include,omitempty"`

	// Exclude specifies the exclude retention policies.
	// +Optional
	// +default:=default=^$
	Exclude string `json:"exclude,omitempty"`
}

// PullRequestSpec specifies the pull request configuration (automated pull of local changes into remote) for Gitcd.
type PullRequestSpec struct {
	// ConflictResolutions specifies the conflict resolution policy while pulling local changes into remote.
	// +Optional
	// +default:=default=2
	ConflictResolutions string `json:"conflictResolutions,omitempty"`

	// PushAfterMerge specifies if the local branch must be pushed to remote after merging from remote.
	// +Optional
	// +default=true
	PushAfterMerge bool `json:"pushAfterMerge,omitempty"`

	// TickerDuration species the ticker duration between automated pull of local changes into remote.
	// +Required
	TickerDuration metav1.Duration `json:"tickerDuration"`
}

type ApiserverSpec ImageSpec

// ControllerUserSpec specifies the user and group details which the controller is to be configured to use.
type ControllerUserSpec struct {
	// UserName specifies the user name the controller is to be configured to use.
	// +Required
	UserName string `json:"userName"`

	// GroupName specifies the group name the controller is to be configured to use.
	// +Optional
	GroupName string `json:"groupName,omitempty"`

	// KubeconfigMountPath specifies the path where the generated controller kubeconfig should be mounted.
	// +Required
	KubeconfigMountPath string `json:"kubeconfigMountPath,omitempty"`
}

// PodTemplateSpec has been copied from corev1 package to annotate metadata to preserve unknown fields.
type PodTemplateSpec struct {
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +kubebuilder:pruning:PreserveUnknownFields
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Specification of the desired behavior of the pod.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status
	// +optional
	Spec corev1.PodSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// TrishankuHeavenStatus defines the observed state of TrishankuHeaven
type TrishankuHeavenStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TrishankuHeaven is the Schema for the trishankuheavens API
type TrishankuHeaven struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TrishankuHeavenSpec   `json:"spec,omitempty"`
	Status TrishankuHeavenStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TrishankuHeavenList contains a list of TrishankuHeaven
type TrishankuHeavenList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TrishankuHeaven `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TrishankuHeaven{}, &TrishankuHeavenList{})
}
