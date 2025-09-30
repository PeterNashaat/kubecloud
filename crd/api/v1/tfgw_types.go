/*
Copyright 2025.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Condition types for TFGW
const (
	// ConditionTypeReady indicates whether the TFGW is ready
	ConditionTypeReady = "Ready"

	// ConditionTypeError indicates whether there is an error with the TFGW
	ConditionTypeError = "Error"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TFGWSpec defines the desired state of TFGW.
type TFGWSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of TFGW. Edit tfgw_types.go to remove/update
	Foo string `json:"foo,omitempty"`

	Hostname string   `json:"hostname"`
	Backends []string `json:"backends"`
}

// TFGWStatus defines the observed state of TFGW.
type TFGWStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	FQDN    string `json:"fqdn"`
	Message string `json:"message"`

	// Conditions represent the latest available observations of an object's state
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Host",type=string,JSONPath=`.spec.hostname`
// +kubebuilder:printcolumn:name="Backends",type=string,JSONPath=`.spec.backends`
// +kubebuilder:printcolumn:name="FQDN",type=string,JSONPath=`.status.fqdn`

// TFGW is the Schema for the tfgws API.
type TFGW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TFGWSpec   `json:"spec,omitempty"`
	Status TFGWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TFGWList contains a list of TFGW.
type TFGWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TFGW `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TFGW{}, &TFGWList{})
}
