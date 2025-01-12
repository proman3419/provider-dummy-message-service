/*
Copyright 2022 The Crossplane Authors.

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
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// MessageParameters are the configurable fields of a Message.
type MessageParameters struct {
	Content string `json:"content"`
}

// MessageObservation are the observable fields of a Message.
type MessageObservation struct {
	Content string `json:"content"`
}

// A MessageSpec defines the desired state of a Message.
type MessageSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       MessageParameters `json:"forProvider"`
}

// A MessageStatus represents the observed state of a Message.
type MessageStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          MessageObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Message is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,dummymessageservice}
type Message struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MessageSpec   `json:"spec"`
	Status MessageStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MessageList contains a list of Message
type MessageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Message `json:"items"`
}

// Message type metadata.
var (
	MessageKind             = reflect.TypeOf(Message{}).Name()
	MessageGroupKind        = schema.GroupKind{Group: Group, Kind: MessageKind}.String()
	MessageKindAPIVersion   = MessageKind + "." + SchemeGroupVersion.String()
	MessageGroupVersionKind = SchemeGroupVersion.WithKind(MessageKind)
)

func init() {
	SchemeBuilder.Register(&Message{}, &MessageList{})
}
