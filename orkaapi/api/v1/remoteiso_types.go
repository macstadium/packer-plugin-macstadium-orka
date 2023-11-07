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

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RemoteIsoSpec defines the desired state of RemoteIso
type RemoteIsoSpec struct {
	// The name of the ISO file represented by the Iso
	IsoName string `json:"isoName"`
	// The size of the ISO file in formatted bytes
	Size resource.Quantity `json:"size,omitempty"`
}

//+kubebuilder:printcolumn:name="Size",type=string,JSONPath=`.spec.size`
//+kubebuilder:object:root=true
//+kubebuilder:resource:path=remoteisos

// RemoteIso is the Schema for the remoteisos API
type RemoteIso struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RemoteIsoSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// RemoteIsoList contains a list of RemoteIso
type RemoteIsoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RemoteIso `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RemoteIso{}, &RemoteIsoList{})
}
