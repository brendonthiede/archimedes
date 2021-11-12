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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ArchimedesPropertySpec defines the desired state of ArchimedesProperty
type ArchimedesPropertySpec struct {
	//ConfigMapName is the name of the config map to be created
	ConfigMapName string `json:"name,omitempty"`
	//Repo is the application repo url
	RepoUrl string `json:"repoUrl,omitempty"`
	//Revision is the branch, commit hash or tag of the repo
	Revision string `json:"revision,omitempty"`
	//PropertiesPath is the path to the applications properties template
	//example: config/properties.tpl
	PropertiesPath string `json:"propertiesPath,omitempty"`
	//SourceConfig is yaml containing data to be merged with the properties template
	SourceConfig string `json:"sourceConfig,omitempty"`
	//PropertyType the format you wish to store the merged results as (keys or file)
	PropertyType string `json:"propertyType,omitempty"`
	//KeyName is the name of the key used if the PropertyType is file
	KeyName string `json:"KeyName,omitempty"`
}

// ArchimedesPropertyStatus defines the observed state of ArchimedesProperty
type ArchimedesPropertyStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ArchimedesProperty is the Schema for the archimedesproperties API
// +kubebuilder:printcolumn:name="Succeeded",type=string,JSONPath=`.status.conditions[?(@.type=="ConfigmapCreated")].status`,description="Indicates if the ConfigMap was created/updated successfully"
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.status.conditions[?(@.type=="ConfigmapCreated")].reason`,description="Reason for the current status"
// +kubebuilder:printcolumn:name="Message",type=string,JSONPath=`.status.conditions[?(@.type=="ConfigmapCreated")].message`,description="Message with more information, regarding the current status"
// +kubebuilder:printcolumn:name="Last Transition",type=date,JSONPath=`.status.conditions[?(@.type=="ConfigmapCreated")].lastTransitionTime`,description="Time when the condition was updated the last time"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="Time when this ConfigMap was created"
// +kubebuilder:subresource:statu
type ArchimedesProperty struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ArchimedesPropertySpec   `json:"spec,omitempty"`
	Status ArchimedesPropertyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ArchimedesPropertyList contains a list of ArchimedesProperty
type ArchimedesPropertyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ArchimedesProperty `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ArchimedesProperty{}, &ArchimedesPropertyList{})
}
