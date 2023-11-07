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

func init() {
	SchemeBuilder.Register(&Image{}, &ImageList{})
}

// NOTE: JSON tags are required. When you add new fields, they must have JSON tags. Otherwise, they won't be serialized.

// ImageSpec describes the desired state of Image
type ImageSpec struct {
	// The size of the image file in formatted bytes
	Size resource.Quantity `json:"size,omitempty"`
	// The user who initially created the image file represented by the Image
	Owner string `json:"owner,omitempty"`
	// The name of an Image, RemoteImage or VirtualMachineInstance to use as a source for a specific Image operation. Must match the SourceType
	Source string `json:"source,omitempty"`
	// The namespace of the source VM (for commit and save).  Uses 'orka-default' if not specified
	SourceNamespace string `json:"sourceNamespace,omitempty"`
	// Modifier used for Image operations. One of generated (for generate), local (for copy), remote (for pull), vm (for commit and save)
	SourceType SourceType `json:"sourceType,omitempty"`
	// A new name for the Image resulting from the Image save operation. Leave empty for commit operations
	Destination string `json:"destination,omitempty"`
	// (amd64-only) The automatically calculated MD5 checksum of the image file represented by the Image. Orka populates Checksum only after you explicitly request the checksum
	Checksum string `json:"checksum,omitempty"`
}

// ImageStatus describes the observed state of Image
type ImageStatus struct {
	// The current state of the image. One of Ready, Updating, Failed
	State State `json:"state,omitempty"`
	// The timestamp for the last Image update (in the ISO 8601 format)
	LastUpdatedTimestamp string `json:"lastUpdatedTimestamp,omitempty"`
	// The error message from the last failed operation with this Image
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// SourceType is an enum providing additional information about the type of Image operation being run. Possible values are: local (for copy), remote (for pull), vm (for commit and save), generated (for generate)
type SourceType string

const (
	Local      SourceType = "local"
	Remote     SourceType = "remote"
	Vm         SourceType = "vm"
	Generated  SourceType = "generated"
	UserUpload SourceType = "userUpload"
)

// State is an enum providing information about the Image state. Possible values are: Ready, Updating, Failed
type State string

const (
	Ready    State = "Ready"
	Updating State = "Updating"
	Failed   State = "Failed"
)

//+kubebuilder:printcolumn:name="Description",type=string,JSONPath=`.metadata.annotations['orka\.macstadium\.com/description']`
//+kubebuilder:printcolumn:name="Size",type=string,JSONPath=`.spec.size`
//+kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.metadata.labels['kubernetes\.io/arch']`
//+kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.state`
//+kubebuilder:printcolumn:name="Error",type=string,JSONPath=`.status.errorMessage`,priority=1
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:path=images

// Image is the Schema for the images API
type Image struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ImageSpec   `json:"spec,omitempty"`
	Status ImageStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ImageList contains a list of Image
type ImageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Image `json:"items"`
}
