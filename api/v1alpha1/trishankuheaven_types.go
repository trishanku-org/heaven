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

	// Skip specifies if some resource generation should be skipped.
	// +optional
	Skip SkipSpec `json:"skip"`

	// Certificates specifies the certificates for the TrishankuHeaven.
	// +optional
	Certificates CertificatesSpec `json:"certificates"`

	// EventsEtcd specifies the configuration for events etcd.
	// +optional
	EventsEtcd *EventsEtcdSpec `json:"eventsEtcd,omitempty"`

	// Gitcd specifies the Gitcd configuration.
	// +optional
	Gitcd *GitcdSpec `json:"gitcd,omitempty"`

	// Apiserver specifies the Apiserver configuration.
	// +optional
	Apiserver *ApiserverSpec `json:"apiserver,omitempty"`

	// App specifies the app for which the trishanku heaven is to be set up.
	// +optional
	App *AppSpec `json:"app"`
}

// SkipSpec specifies if some resource generation should be skipped.
type SkipSpec struct {
	// Certificates specifies if the certificate generation should be skipped.
	// +optional
	Cerfificates bool `json:"certificates"`

	// Entrypoints specifies if the entrypoint configmap generation should be skipped.
	// +optional
	Entrypoints bool `json:"entrypoints"`

	// App specifies if the app generation should be skipped.
	// +optional
	App bool `json:"app"`
}

// CertificatesSpec specifies the certificates required for the TrishankuHeaven.
type CertificatesSpec struct {
	// CertificateAuthoritySecretName specifies the name for the secret containing the CA certificates.
	// The secret will be generated if not existing. The secret name will be generated if not specified.
	// +optional
	CertificateAuthoritySecretName *string `json:"certificateAuthoritySecretName,omitempty"`

	// ServiceAccountsSecretName specifies the name for the secret containing the service-accounts certificates.
	// The secret will be generated if not existing. The secret name will be generated if not specified.
	// +optional
	ServiceAccountsSecretName *string `json:"serviceAccountsSecretName,omitempty"`

	// KubernetesSecretName specifies the name for the secret containing the Kubernetes certificates.
	// The secret will be generated if not existing. The secret name will be generated if not specified.
	// +optional
	KubernetesSecretName *string `json:"kubernetesSecretName,omitempty"`

	// AdminSecretName specifies the name for the secret containing the admin certificates and kubeconfig.
	// The secret will be generated if not existing. The secret name will be generated if not specified.
	// +optional
	AdminSecretName *string `json:"adminSecretName,omitempty"`

	// Controller specifies the configuration for the controller certificates.
	// No controller certificates are generated if not specified.
	// This field is mandatory if app is to be deployed (that is, if `SkipSpec.Deploy == false`).
	// +optional
	Controller *ControllerCertificateSpec `json:"controller,omitempty"`
}

type ControllerCertificateSpec struct {
	// SecretName specifies the name of the scret containing the controller certificates and kubeconfig.
	// The secret will be generated if not existing. The secret name will be generated if not specified.
	// +optional
	SecretName *string `json:"secretName,omitempty"`

	// UserName specifies the user name the controller is to be configured to use.
	UserName string `json:"userName"`

	// GroupName specifies the group name the controller is to be configured to use.
	// +optional
	GroupName string `json:"groupName,omitempty"`
}

// ImageSpec specifies the image details for a container.
type ImageSpec struct {
	// Image specifies the container image.
	// +optional
	Image *string `json:"image,omitempty"`

	//ImagePullPolicy specifies the image pull policy for the container image.
	// +optional
	ImagePullPolicy *corev1.PullPolicy `json:"imagePullPolicy,omitempty"`
}

type EventsEtcdSpec struct {
	// ImagePullPolicy specifies the image pull policy for the events etcd container image.
	// +optional
	Local *EventsEtcdLocalSpec `json:"local,omitempty"`
}

type EventsEtcdLocalSpec ImageSpec

// GitcdSpec specifies the configuration for Gitcd
type GitcdSpec struct {
	// GitImage specifie the git container image.
	// +optional
	GitImage *ImageSpec `json:"gitImage,omitempty"`

	// GitcdImage specifie the gitcd container image.
	// +optional
	GitcdImage *ImageSpec `json:"gitcdImage,omitempty"`

	// CredentialsSecretName specifies the name of the secret containing the .git-credentials file.
	CredentialsSecretName string `json:"credentialsSecretName"`

	// ImagePullSecretName specifies the name of the secret containing the credentials to pull gitcd image.
	// +optional
	ImagePullSecretName string `json:"imagePullSecretName"`

	//Git specifies the configuration for Git repo that will back gitcd.
	Git GitSpec `json:"git"`

	// Pull specifies the pull configuration for gitcd.
	Pull PullSpec `json:"pull"`
}

// GitSpec specifies the configuration for Git repo that will back gitcd.
type GitSpec struct {
	// Branches specifies the local data and metadata branches.
	Branches BranchSpec `json:"branches"`

	// Remotes specifies the details about one or more remote repos and data/metadata branches for the git repo that will back gitcd.
	Remotes []RemoteSpec `json:"remotes"`
}

// RemoteSpec specifies the details about one or more remote repos and data/metadata branches for the git repo that will back gitcd.
type RemoteSpec struct {
	// Name specifies the name of the remote repo. Defaults to origin.
	// +optional
	Name string `json:"name"`

	// RemoteRepo specifies the URL or path to the remote repo for the git repo that will back gitcd.
	// +optional
	Repo string `json:"repo"`

	// Branches specifies the data and metadata branches.
	// +optional
	Branches *BranchSpec `json:"branches"`

	// RetentionPolicies specifies the retention policies while merging changes from remote.
	// +optional
	RetentionPolicies RetentionPoliciesSpec `json:"retentionPolicies,omitempty"`

	// ConflictResolution specifies the conflict resolution policy while merging changes from remote.
	// +optional
	// +default:=1
	ConflictResolution int8 `json:"conflictResolution,omitempty"`
}

// BranchSpec specifies the data and metadata branches for Gitcd
type BranchSpec struct {
	// Data specifies the data branch.
	Data string `json:"data,omitempty"`

	// Metadata specifies the data branch.
	Metadata string `json:"metadata,omitempty"`

	// NoCreateBranch specifies if branches should not be created if they do not exist.
	// +optional
	NoCreateBranch bool `json:"noCreateBranch,omitempty"`
}

// PullSpec specifies the pull configuration for Gitcd.
type PullSpec struct {
	// PushAfterMerge specifies if the local branch must be pushed to remote after merging from remote.
	// +optional
	// +default=true
	PushAfterMerge bool `json:"pushAfterMerge,omitempty"`

	// PushOnPullFailure specifies if the local branch must be pushed to remote even if fetching from remote fails.
	// +optional
	// +default=true
	PushOnPullFailure bool `json:"pushOnPullFailure,omitempty"`

	// TickerDuration species the ticker duration between automated pulls from remote.
	TickerDuration metav1.Duration `json:"tickerDuration"`
}

// RetentionPoliciesSpec specifies the merge retention policies for Gitcd.
type RetentionPoliciesSpec struct {
	// Include specifies the include retention policies.
	// +optional
	// +default:=.*
	Include string `json:"include,omitempty"`

	// Exclude specifies the exclude retention policies.
	// +optional
	// +default:=^$
	Exclude string `json:"exclude,omitempty"`
}

type ApiserverSpec ImageSpec

type AppSpec struct {
	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	// EntrypointsConfigMapName specifies the name of the configmap containing the trishanku entrypoint scripts for the
	// The confimapg will be generated if not existing. The configmap name will be generated if not specified.
	// +optional
	EntrypointsConfigMapName *string `json:"entrypointsConfigMapName,omitempty"`

	// KubeconfigMountPath specifies the path where the generated controller kubeconfig should be mounted.
	KubeconfigMountPath string `json:"kubeconfigMountPath,omitempty"`

	// PodTemplate specifies the pod template for the controller deployment.
	PodTemplate PodTemplateSpec `json:"podTemplate"`
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
