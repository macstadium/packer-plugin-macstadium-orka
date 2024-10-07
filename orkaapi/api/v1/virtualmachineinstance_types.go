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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: JSON tags are required. When you add new fields, they must have JSON tags. Otherwise, they won't be serialized.

// VirtualMachineInstanceSpec describes the desired state of VirtualMachineInstance
type VirtualMachineInstanceSpec struct {
	// The ID of the Image to use for the VirtualMachineInstance
	Image string `json:"image"`
	// The number of OrkaNode CPU cores dedicated to the VirtualMachineInstance
	CPU int `json:"cpu"`
	// The name of the node where you want the VirtualMachineInstance to run. If not specified, the VirtualMachineInstance will run on the first available node that matches the criteria (e.g., available CPU and memory, tags, etc.)
	NodeName *string `json:"nodeName,omitempty"`
	// A custom port pairing to enable traffic forwarding. Must be provided in the <NODE_PORT>:<VM_PORT> format (e.g., 1337:3000)
	ReservedPorts string `json:"reservedPorts,omitempty"`
	// (Intel-only) Attaches the specified ISO (by name) to let you install macOS from scratch
	ISO *string `json:"iso,omitempty"`
	// Custom metadata to be passed to the VirtualMachineInstance
	CustomVMMetadata map[string]string `json:"customVMMetadata,omitempty"`
	// A custom serial number for the VirtualMachineInstance
	SystemSerial *string `json:"systemSerial,omitempty"`
	// A string setting node affinity. Node affinity indicates that the tagged OrkaNode is preferred for the deployment of VirtualMachineInstances with the same tag
	Tag *string `json:"tag,omitempty"`
	// Boolean setting if the Tag is required. When true, Orka never attempts to deploy to OrkaNodes without the specified Tag
	TagRequired *bool `json:"tagRequired,omitempty"`
	// Boolean setting if the VNC console is enabled on the VirtualMachineInstance. When enabled, GPUPassthrough must be disabled
	VNCConsole *bool `json:"vncConsole,omitempty"`
	// The scheduler handling the deployment. One of: default, most-allocated. When set to most-allocated, VirtualMachineInstances are scheduled to OrkaNodes having most of their resources allocated. The default setting keeps used vs free resources balanced between OrkaNodes
	Scheduler *string `json:"scheduler,omitempty"`
	// Boolean setting if legacy IO is enabled
	LegacyIO *bool `json:"legacyIO,omitempty"`
	// Boolean setting if Network boost is enabled
	NetBoost *bool `json:"netBoost,omitempty"`
	// Boolean setting if GPU passthrough is enabled. When enabled, VncConsole must be disabled
	GPUPassthrough *bool `json:"gpuPassthrough,omitempty"`
	// Memory in GiB. Rounded to the nearest 0.1 GiB. If not specified, the VirtualMachineInstance will use the default memory value set in the Orka configuration or will be automatically calculated based on the number of CPU cores
	Memory *float64 `json:"memory,omitempty"`
}

// VMPhase is an enum providing information about the state of the VirtualMachineInstance deployment
type VMPhase string

const (
	// VMRunning indicates that the VirtualMachineInstance is successfully deployed and running
	VMRunning VMPhase = "Running"
	// VMFailed indicates that the corresponding VirtualMachineInstance pod is NOT in a running phase and there are errors in its status field
	VMFailed VMPhase = "Failed"
	// VMPending indicates that the corresponding VirtualMachineInstance is currently deploying and still not running
	VMPending VMPhase = "Pending"
)

// VirtualMachineInstanceStatus describes the observed state of the VirtualMachineInstance
type VirtualMachineInstanceStatus struct {
	Phase VMPhase `json:"phase,omitempty"`
	// The amount of memory allocated to the VirtualMachineInstance
	Memory string `json:"memory"`
	// The name of the OrkaNode on which the VirtualMachineInstance is running
	NodeName string `json:"nodeName"`
	// The IP of the OrkaNode on which the VirtualMachineInstance is running
	HostIP string `json:"hostIP"`
	// The SSH port assigned to the VirtualMachineInstance
	SSHPort *int `json:"sshPort,omitempty"`
	// The VNC port assigned to the VirtualMachineInstance
	VNCPort *int `json:"vncPort,omitempty"`
	// The Screen Sharing port assigned to the VirtualMachineInstance
	ScreenSharePort *int `json:"screenSharePort,omitempty"`
	// Any port warnings that have occurred during the deployment
	PortWarnings string `json:"portWarnings"`
	StartTime    int64  `json:"startTime"`
	// The error message if the deployment failed
	ErrorMessage string `json:"errorMessage"`
}

//+kubebuilder:printcolumn:name="IP",type=string,JSONPath=`.status.hostIP`
//+kubebuilder:printcolumn:name="SSH",type=integer,JSONPath=`.status.sshPort`
//+kubebuilder:printcolumn:name="VNC",type=integer,JSONPath=`.status.vncPort`
//+kubebuilder:printcolumn:name="Screenshare",type=integer,JSONPath=`.status.screenSharePort`
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
//+kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`,priority=1
//+kubebuilder:printcolumn:name="CPU",type=integer,JSONPath=`.spec.cpu`,priority=1
//+kubebuilder:printcolumn:name="Memory",type=string,JSONPath=`.status.memory`,priority=1
//+kubebuilder:printcolumn:name="Node",type=string,JSONPath=`.status.nodeName`,priority=1
//+kubebuilder:printcolumn:name="Architecture",type=string,JSONPath=`.metadata.labels['kubernetes\.io/arch']`,priority=1
//+kubebuilder:printcolumn:name="Reserved-Ports",type=string,JSONPath=`.spec.reservedPorts`,priority=1
//+kubebuilder:printcolumn:name="GPU-Passthrough",type=boolean,JSONPath=`.spec.gpuPassthrough`,priority=1
//+kubebuilder:printcolumn:name="Owner",type=string,JSONPath=`.metadata.annotations.orka\.macstadium\.com/created-by`,priority=1
//+kubebuilder:printcolumn:name="Deploy-Date",type=string,JSONPath=`.metadata.creationTimestamp`,priority=1
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=vm;vms

// VirtualMachineInstance is the Schema for the virtualmachineinstances API
type VirtualMachineInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   VirtualMachineInstanceSpec   `json:"spec"`
	Status VirtualMachineInstanceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VirtualMachineInstanceList contains a list of VirtualMachineInstance
type VirtualMachineInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualMachineInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VirtualMachineInstance{}, &VirtualMachineInstanceList{})
}
