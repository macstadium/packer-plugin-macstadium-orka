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

// RemoteImageSpec defines the desired state of RemoteImage
type RemoteImageSpec struct {
	// The name of the image file represented by the Image. Must include a valid file extension (.img for amd64 or .orkasi for arm)
	ImageName string `json:"imageName"`
	// The size of the image file in formatted bytes
	Size resource.Quantity `json:"size,omitempty"`
}

//+kubebuilder:printcolumn:name="Size",type=string,JSONPath=`.spec.size`
//+kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.metadata.labels['kubernetes\.io/arch']`
//+kubebuilder:object:root=true

// RemoteImage is the Schema for the remoteimages API
type RemoteImage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RemoteImageSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// RemoteImageList contains a list of RemoteImage
type RemoteImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RemoteImage `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RemoteImage{}, &RemoteImageList{})
}
