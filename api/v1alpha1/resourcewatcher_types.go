/*
Copyright 2021.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ResourceWatcherSpec defines the desired state of ResourceWatcher
type ResourceWatcherSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Handler    string   `json:"handler"`
	Namespaces []string `json:"namespaces"`
	// Resources is the list of resources to watch
	Resources []ApiServerResource `json:"resources"`
	// ServiceAccountName is the name of the ServiceAccount to use to run this
	// source.
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`
	// Target is a reference to an object that will resolve to a domain name to use as the sink.
	// +optional
	Target *corev1.ObjectReference `json:"sink,omitempty"`
}

// ResourceWatcherStatus defines the observed state of ResourceWatcher
type ResourceWatcherStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ResourceWatcher is the Schema for the resourcewatchers API
type ResourceWatcher struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceWatcherSpec   `json:"spec,omitempty"`
	Status ResourceWatcherStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ResourceWatcherList contains a list of ResourceWatcher
type ResourceWatcherList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceWatcher `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ResourceWatcher{}, &ResourceWatcherList{})
}

// ApiServerResource defines the resource to watch
type ApiServerResource struct {
	// API version of the resource to watch.
	APIVersion string `json:"apiVersion"`

	// Kind of the resource to watch.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds
	Kind string `json:"kind"`

	// LabelSelector restricts this source to objects with the selected labels
	// More info: http://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	LabelSelector metav1.LabelSelector `json:"labelSelector"`

	// ControllerSelector restricts this source to objects with a controlling owner reference of the specified kind.
	// Only apiVersion and kind are used. Both are optional.
	ControllerSelector metav1.OwnerReference `json:"controllerSelector"`

	// If true, send an event referencing the object controlling the resource
	Controller bool `json:"controller"`

	// NameSelector is the list of resource name watched
	NameSelector []string `json:"nameSelector"`
}
