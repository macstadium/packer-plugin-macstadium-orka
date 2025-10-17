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
	"reflect"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: JSON tags are required. When you add new fields, they must have JSON tags. Otherwise, they won't be serialized.

// The image that should be cached on the node
type OrkaImage struct {
	// The name of the image to cache
	Name string `json:"name"`
}

// OrkaNodeSpec describes the desired state of OrkaNode
type OrkaNodeSpec struct {
	// One or more tags setting node affinity. Node affinity indicates that the tagged OrkaNode is preferred for the deployment of VirtualMachineInstances with the same tag
	Tags []string `json:"tags"`
	// The name of a specific namespace to which a node is assigned. Only users with appropriate access to that namespace will have the ability to deploy VirtualMachineInstances on that OrkaNode
	Namespace string `json:"namespace"`
}

// NodeType is an enum providing information about the node type. Possible values are: FOUNDATION, SERVICE, WORKER, SANDBOX
type NodeType string

const (
	Foundation NodeType = "FOUNDATION"
	Service    NodeType = "SERVICE"
	Worker     NodeType = "WORKER"
	Sandbox    NodeType = "SANDBOX"
)

// NodeArchitecture is an enum providing information about the node architecture. Possible values are: arm64 (for Apple silicon), amd64 (for Intel)
type Architecture string

const (
	Amd64   Architecture = "amd64"
	Arm     Architecture = "arm64"
	Unknown Architecture = "unknown"
)

// NodePhase is an enum providing information about the node status. Possible values are: Ready, Not Ready
type NodePhase string

const (
	NodeReady    NodePhase = "READY"
	NodeNotReady NodePhase = "NOT READY"
)

// Describes a cached image on an OrkaNode
type OrkaImageStatus struct {
	// Names by which this image is known.
	Names []string `json:"names,omitempty"`
	// The virtual size of the image file in bytes
	SizeBytes int64 `json:"size,omitempty"`
	// The amount of storage the image file actually consumes in bytes
	SpaceUsedBytes int64 `json:"spaceUsed,omitempty"`
	// The status of the image. One of Ready, Caching, Failed
	OrkaImageState OrkaImageState `json:"status,omitempty"`
	// The error message if the image status is FAILED
	ErrorMessage string `json:"errorMessage,omitempty"`
}

// OrkaImageState is an enum providing information about the image status. Possible values are: Ready, Caching, Failed
type OrkaImageState string

const (
	OrkaImageStateReady   OrkaImageState = "Ready"
	OrkaImageStateCaching OrkaImageState = "Caching"
	OrkaImageStateFailed  OrkaImageState = "Failed"
)

// OrkaNodeStatus describes the observed state of OrkaNode
type OrkaNodeStatus struct {
	// The IP of the OrkaNode
	NodeIP string `json:"nodeIP"`
	// The amount of available CPU on the node
	AvailableCPU int `json:"availableCpu"`
	// The amount of available Memory on the node
	AvailableMemory string `json:"availableMemory"`
	// The amount of available GPU on the node
	AvailableGPU int `json:"availableGpu"`
	// The complete amount of CPU cores on the node when no VirtualMachineInstances are deployed
	AllocatableCPU int64 `json:"allocatableCpu"`
	// The complete amount of memory (in GiB) on the node when no VirtualMachineInstances are deployed
	AllocatableMemory string `json:"allocatableMemory"`
	// The complete amount of GPU cores on the node when no VirtualMachineInstances are deployed
	AllocatableGpu int64 `json:"allocatableGpu"`
	// The type of the OrkaNode. One of FOUNDATION, SERVICE, WORKER, SANDBOX
	NodeType NodeType `json:"nodeType"`
	// The status of the OrkaNode. One of READY, NOT READY
	Phase NodePhase `json:"phase"`
	// The list of cached images on the OrkaNode
	Images []OrkaImageStatus `json:"images,omitempty"`
	// The version of the kubelet
	KubeletVersion string `json:"kubeletVersion"`
	// The version Container Runtime
	ContainerRuntimeVersion string `json:"containerRuntimeVersion"`
	// The MacOS Version
	OSVersion string `json:"osVersion"`
}

//+kubebuilder:printcolumn:name="Available-CPU",type=integer,JSONPath=`.status.availableCpu`
//+kubebuilder:printcolumn:name="Available-Memory",type=string,JSONPath=`.status.availableMemory`
//+kubebuilder:printcolumn:name="Available-GPU",type=integer,JSONPath=`.status.availableGpu`,priority=2
//+kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
//+kubebuilder:printcolumn:name="IP",type=string,JSONPath=`.status.nodeIP`,priority=1
//+kubebuilder:printcolumn:name="Allocatable-CPU",type=integer,JSONPath=`.status.allocatableCpu`,priority=1
//+kubebuilder:printcolumn:name="Allocatable-Memory",type=string,JSONPath=`.status.allocatableMemory`,priority=1
//+kubebuilder:printcolumn:name="Allocatable-GPU",type=integer,JSONPath=`.status.allocatableGpu`,priority=2
//+kubebuilder:printcolumn:name="Version",type=string,JSONPath=`.status.kubeletVersion`,priority=1
//+kubebuilder:printcolumn:name="Runtime",type=string,JSONPath=`.status.containerRuntimeVersion`,priority=1
//+kubebuilder:printcolumn:name="OS",type=string,JSONPath=`.status.osVersion`,priority=1
//+kubebuilder:printcolumn:name="Architecture",type=string,JSONPath=`.metadata.labels['kubernetes\.io/arch']`,priority=1
//+kubebuilder:printcolumn:name="Tags",type=string,JSONPath=`.spec.tags`,priority=1
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OrkaNode is the Schema for the orkanodes API
type OrkaNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OrkaNodeSpec   `json:"spec,omitempty"`
	Status OrkaNodeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OrkaNodeList contains a list of OrkaNodes
type OrkaNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OrkaNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OrkaNode{}, &OrkaNodeList{})
}

func (left *OrkaNodeSpec) DeepEqual(right OrkaNodeSpec) bool {
	leftTags := left.DeepCopy().Tags
	rightTags := right.DeepCopy().Tags
	sort.Strings(leftTags)
	sort.Strings(rightTags)

	return reflect.DeepEqual(leftTags, rightTags) && left.Namespace == right.Namespace
}

// DeepEqual checks if two OrkaNodeStatus objects are equal
func (left *OrkaNodeStatus) DeepEqual(right OrkaNodeStatus) bool {
	leftImages := left.DeepCopy().Images
	if leftImages == nil {
		leftImages = []OrkaImageStatus{}
	}

	rightImages := right.DeepCopy().Images
	if rightImages == nil {
		rightImages = []OrkaImageStatus{}
	}

	return left.NodeIP == right.NodeIP &&
		left.AllocatableCPU == right.AllocatableCPU &&
		left.AllocatableMemory == right.AllocatableMemory &&
		left.AllocatableGpu == right.AllocatableGpu &&
		left.NodeType == right.NodeType &&
		left.Phase == right.Phase &&
		left.AvailableCPU == right.AvailableCPU &&
		left.AvailableMemory == right.AvailableMemory &&
		left.AvailableGPU == right.AvailableGPU &&
		left.KubeletVersion == right.KubeletVersion &&
		left.ContainerRuntimeVersion == right.ContainerRuntimeVersion &&
		left.OSVersion == right.OSVersion &&
		reflect.DeepEqual(leftImages, rightImages)
}

func (left *OrkaNode) DeepEqual(right *OrkaNode) bool {
	return left.Spec.DeepEqual(right.Spec) &&
		reflect.DeepEqual(left.ObjectMeta.Labels, right.ObjectMeta.Labels) &&
		left.ObjectMeta.Name == right.ObjectMeta.Name &&
		left.ObjectMeta.Namespace == right.ObjectMeta.Namespace &&
		left.Spec.DeepEqual(right.Spec) &&
		left.Status.DeepEqual(right.Status)
}
