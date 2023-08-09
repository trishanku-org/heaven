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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AutomatedMergeSpec defines the desired state of AutomatedMerge
type AutomatedMergeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Gitcd specifies the Gitcd configuration.
	Gitcd GitcdSpec `json:"gitcd,omitempty"`

	// App specifies the app for which the automated merge is to be set up.
	// +optional
	App AutomatedMergeAppSpec `json:"app"`
}

// AutomatedMergeAppSpec defines the desired state of the app for which the automated merge is to be set up.
type AutomatedMergeAppSpec struct {
	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	// EntrypointsConfigMapName specifies the name of the configmap containing the trishanku entrypoint scripts for the
	// The confimapg will be generated if not existing. The configmap name will be generated if not specified.
	// +optional
	EntrypointsConfigMapName *string `json:"entrypointsConfigMapName,omitempty"`

	// CommitterName specifies the user name the gitcd is to be configured to use.
	CommitterName string `json:"committerName"`
}

// AutomatedMergeStatus defines the observed state of AutomatedMerge
type AutomatedMergeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AutomatedMerge is the Schema for the automatedmerges API
type AutomatedMerge struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AutomatedMergeSpec   `json:"spec,omitempty"`
	Status AutomatedMergeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AutomatedMergeList contains a list of AutomatedMerge
type AutomatedMergeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AutomatedMerge `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AutomatedMerge{}, &AutomatedMergeList{})
}
