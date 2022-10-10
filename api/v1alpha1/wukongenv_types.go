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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HostAlias holds the mapping between IP and hostnames that will be injected as an entry in the
// pod's hosts file.
type WukongApp struct {
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// +kubebuilder:validation:Required
	Workload string `json:"workload"`
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`
	// +kubebuilder:validation:Required
	Image string `json:"image"`
	// +kubebuilder:validation:Required
	Size int32 `json:"size"`
	// +kubebuilder:validation:Required
	SkywalkingBackendAddr string `json:"skywalkingBackendAddr"`
	// +kubebuilder:validation:Required
	SkywalkingIson string `json:"skywalkingIson"`
}

// WukongEnvSpec defines the desired state of WukongEnv
type WukongEnvSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of WukongEnv. Edit wukongenv_types.go to remove/update

	// env prefix name
	Namespaces []string    `json:"namespaces"`
	Apps       []WukongApp `json:"apps"`
}

// WukongEnvStatus defines the observed state of WukongEnv
type WukongEnvStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Pods []string `json:"pods,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// WukongEnv is the Schema for the wukongenvs API
type WukongEnv struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WukongEnvSpec   `json:"spec,omitempty"`
	Status WukongEnvStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WukongEnvList contains a list of WukongEnv
type WukongEnvList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WukongEnv `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WukongEnv{}, &WukongEnvList{})
}
