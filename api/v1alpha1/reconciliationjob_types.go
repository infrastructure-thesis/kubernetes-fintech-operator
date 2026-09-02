package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReconciliationJobSpec defines the desired state
type ReconciliationJobSpec struct {
	// Replicas is the number of reconciliation pods
	Replicas *int32 `json:"replicas,omitempty"`

	// BatchSize: transactions per batch (e.g., 100000)
	BatchSize int64 `json:"batchSize,omitempty"`

	// CheckpointInterval: save state every N transactions
	CheckpointInterval int64 `json:"checkpointInterval,omitempty"`

	// Image is the container image to run
	Image string `json:"image,omitempty"`
}

// ReconciliationJobStatus defines the observed state
type ReconciliationJobStatus struct {
	// Phase: Pending, Running, Completed, Failed
	Phase string `json:"phase,omitempty"`

	// Replicas is actual pod count
	Replicas int32 `json:"replicas,omitempty"`

	// ReadyReplicas is pods ready
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Conditions represent the latest available observations
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ProcessedTransactions: total processed so far
	ProcessedTransactions int64 `json:"processedTransactions,omitempty"`

	// ObservedGeneration for drift detection
	ObservedGeneration int64 `json:"observedGeneration"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.status.replicas`
// +kubebuilder:printcolumn:name="Ready",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`

// ReconciliationJob is the Schema for reconciliation jobs
type ReconciliationJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReconciliationJobSpec   `json:"spec,omitempty"`
	Status ReconciliationJobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ReconciliationJobList contains a list of ReconciliationJob
type ReconciliationJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReconciliationJob `json:"items"`
}
