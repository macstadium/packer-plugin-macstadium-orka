/*
Copyright 2023.

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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: JSON tags are required. When you add new fields, they must have JSON tags. Otherwise, they won't be serialized.

// IsoSpec describes the desired state of Iso
type IsoSpec struct {
	// The size of the ISO file in formatted bytes
	Size resource.Quantity `json:"size,omitempty"`
	// The name of an Iso or RemoteIso to use as a source for a specific Iso operation.
	Source string `json:"source,omitempty"`
	// Modifier used for Iso operations. One of local (for copy) or remote (for pull)
	SourceType SourceType `json:"sourceType,omitempty"`
}

// IsoStatus describes the observed state of Iso
type IsoStatus struct {
	// The current state of the Iso. One of Ready, Updating, Failed
	State State `json:"state,omitempty"`
	// The timestamp for the last Iso update (in the ISO 8601 format)
	LastUpdatedTimestamp string `json:"lastUpdatedTimestamp,omitempty"`
	// The error message from the last failed operation with this Iso
	ErrorMessage string `json:"errorMessage,omitempty"`
}

//+kubebuilder:printcolumn:name="Description",type=string,JSONPath=`.metadata.annotations['orka\.macstadium\.com/description']`
//+kubebuilder:printcolumn:name="Size",type=string,JSONPath=`.spec.size`
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Error",type=string,JSONPath=`.status.errorMessage`,priority=1
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=isos

// Iso is the Schema for the isos API
type Iso struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IsoSpec   `json:"spec,omitempty"`
	Status IsoStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IsoList contains a list of Iso
type IsoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Iso `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Iso{}, &IsoList{})
}
