/*
Copyright 2025.

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

// api/v1alpha1/sealedage_types.go
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SealedAgeTemplate defines the structure for Secret templates (you currently use only .type).
type SealedAgeTemplate struct {
	// +kubebuilder:validation:Optional
	Type string `json:"type,omitempty"`
}

// SealedAgeSpec defines the desired state of the SealedAge resource.
type SealedAgeSpec struct {
	// Encrypted data (AGE armored or binary); key = Secret field name.
	// +kubebuilder:validation:Optional
	EncryptedData map[string]string `json:"encryptedData,omitempty"`

	// Secret template (e.g., Type: Opaque).
	// +kubebuilder:validation:Optional
	Template SealedAgeTemplate `json:"template,omitempty"`

	// Optional: list of recipients (used for decryption logic if needed).
	// +kubebuilder:validation:Optional
	Recipients []string `json:"recipients,omitempty"`

	// Optional: behavior flag for delete/restore.
	// +kubebuilder:validation:Optional
	RestoreOnDelete *bool `json:"restoreOnDelete,omitempty"`
}

// SealedAgeStatus defines observed state and metadata for the SealedAge resource.
type SealedAgeStatus struct {
	// +kubebuilder:validation:Optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
	// +kubebuilder:validation:Optional
	SecretName string `json:"secretName,omitempty"`
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=sealedages,scope=Namespaced,shortName=sa
// +kubebuilder:printcolumn:name="Secret",type=string,JSONPath=`.status.secretName`
// +kubebuilder:printcolumn:name="Age",type=integer,JSONPath=`.metadata.generation`
type SealedAge struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SealedAgeSpec   `json:"spec,omitempty"`
	Status SealedAgeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type SealedAgeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SealedAge `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SealedAge{}, &SealedAgeList{})
}
