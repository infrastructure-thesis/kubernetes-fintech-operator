package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SettlementWorkerSpec defines the desired state
type SettlementWorkerSpec struct {
	// Replicas is the number of settlement worker pods
	Replicas *int32 `json:"replicas,omitempty"`

	// SchedulingPriority: "high", "medium", "low"
	SchedulingPriority string `json:"schedulingPriority,omitempty"`

	// SettlementType describes what type of settlement
	SettlementType string `json:"settlementType,omitempty"`

	// MaxInFlightTransactions: max concurrent transactions allowed
	MaxInFlightTransactions int64 `json:"maxInFlightTransactions,omitempty"`

	// GracefulShutdownTimeout in seconds
	GracefulShutdownTimeout int32 `json:"gracefulShutdownTimeout,omitempty"`

	// Image is the container image to run
	Image string `json:"image,omitempty"`
}

// SettlementWorkerStatus defines the observed state
type SettlementWorkerStatus struct {
	// Phase: Pending, Running, Completed, Failed
	Phase string `json:"phase,omitempty"`

	// Replicas is actual pod count
	Replicas int32 `json:"replicas,omitempty"`

	// ReadyReplicas is pods ready
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Conditions represent the latest available observations
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration for drift detection
	ObservedGeneration int64 `json:"observedGeneration"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.status.replicas`
// +kubebuilder:printcolumn:name="Ready",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// SettlementWorker is the Schema for settlement workers
type SettlementWorker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SettlementWorkerSpec   `json:"spec,omitempty"`
	Status SettlementWorkerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SettlementWorkerList contains a list of SettlementWorker
type SettlementWorkerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SettlementWorker `json:"items"`
}
