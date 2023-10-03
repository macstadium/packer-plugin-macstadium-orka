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

// TODO: Add validation for the VncConsole property, because it cannot be enabled if GPU Passthrough is also enabled

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: JSON tags are required. When you add new fields, they must have JSON tags. Otherwise, they won't be serialized.

// VirtualMachineConfigSpec describes the desired state of VirtualMachineConfig
type VirtualMachineConfigSpec struct {
	// The name of the Image to use for the VirtualMachineInstances created from the VirtualMachineConfig (.img for Intel, .orkasi for Apple silicon)
	Image string `json:"image"`
	// +kubebuilder:validation:Minimum=2
	// The number of CPU cores to allocate to VirtualMachineInstances created from the VirtualMachineConfig
	CPU int `json:"cpu"`
	// Memory in GiB. Rounded to the nearest 0.1 GiB. If not specified, will be automatically calculated based on the number of CPU cores
	Memory *float64 `json:"memory,omitempty"`
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=65
	// (Intel-only) Attaches the specified ISO (by name) to let you install macOS from scratch on a VirtualMachineInstance created from the VirtualMachineConfig. When specified, make sure that the Image field uses an empty disk generated with the respective operation
	ISO *string `json:"iso,omitempty"`
	// Boolean setting if the VNC console is enabled for VirtualMachineInstances created from the VirtualMachineConfig. When enabled, GPUPassthrough must be disabled
	VNCConsole *bool `json:"vncConsole,omitempty"`
	// Boolean setting if Network boost is enabled for VirtualMachineInstances created from the VirtualMachineConfig
	NetBoost *bool `json:"netBoost,omitempty"`
	// Boolean setting if GPU passthrough is enabled for VirtualMachineInstances created from the VirtualMachineConfig
	GPUPassthrough *bool `json:"gpuPassthrough,omitempty"`
	// +kubebuilder:validation:MinLength=8
	// +kubebuilder:validation:MaxLength=12
	// A custom serial number for the VirtualMachineInstances created from the VirtualMachineConfig. The provided serial number must be a valid Mac serial number
	SystemSerial *string `json:"systemSerial,omitempty"`
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=32
	// When specified, VirtualMachineInstances created from the VirtualMachineConfig will be scheduled for deployment on OrkaNodes labeled with the same Tag
	Tag *string `json:"tag,omitempty"`
	// Boolean setting if the Tag is required. When true, VirtualMachineInstances will be deployed only on Nodes matching the specified Tag
	TagRequired *bool `json:"tagRequired,omitempty"`
	// +kubebuilder:validation:MaxLength=32
	// The scheduler to use for the deployment of VirtualMachineInstances created from the VirtualMachineConfig. One of: default, most-allocated. When set to most-allocated, VirtualMachineInstances are scheduled to OrkaNodes having most of their resources allocated. The default setting keeps used vs free resources balanced between OrkaNodes
	Scheduler *string `json:"scheduler,omitempty"`
	// The name of the node where you want the VirtualMachineInstance to run. If not specified, the VirtualMachineInstance will run on the first available node that matches the criteria (e.g., available CPU and memory, tags, groups, etc.)
	NodeName *string `json:"nodeName,omitempty"`
}

// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
// +kubebuilder:printcolumn:name="CPU",type=integer,JSONPath=`.spec.cpu`
// +kubebuilder:printcolumn:name="Memory",type=string,JSONPath=`.spec.memory`
// +kubebuilder:printcolumn:name="ISO",type=string,JSONPath=`.spec.iso`,priority=1
// +kubebuilder:printcolumn:name="VNC",type=boolean,JSONPath=`.spec.vncConsole`,priority=1
// +kubebuilder:printcolumn:name="Scheduler",type=string,JSONPath=`.spec.scheduler`,priority=1
// +kubebuilder:printcolumn:name="Tag",type=string,JSONPath=`.spec.tag`,priority=1
// +kubebuilder:printcolumn:name="Tag_Required",type=boolean,JSONPath=`.spec.tagRequired`,priority=1
// +kubebuilder:printcolumn:name="GPU",type=boolean,JSONPath=`.spec.gpuPassthrough`,priority=1
// +kubebuilder:printcolumn:name="Net_Boost",type=boolean,JSONPath=`.spec.netBoost`,priority=1
// +kubebuilder:printcolumn:name="System_Serial",type=string,JSONPath=`.spec.systemSerial`,priority=1
// +kubebuilder:printcolumn:name="Owner",type=string,JSONPath=`.metadata.annotations.orka\.macstadium\.com/created-by`,priority=1
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=vmconfig;vmconfigs

// VirtualMachineConfig is the Schema for the virtualmachineconfigs API
type VirtualMachineConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VirtualMachineConfigSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// VirtualMachineConfigList contains a list of VirtualMachineConfig
type VirtualMachineConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualMachineConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtualMachineConfig{}, &VirtualMachineConfigList{})
}
