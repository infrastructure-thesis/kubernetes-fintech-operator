package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime" // <-- add this line
)

// +kubebuilder:validation:Enum=high;medium;low
type SchedulingPriority string

const (
	PriorityHigh   SchedulingPriority = "high"
	PriorityMedium SchedulingPriority = "medium"
	PriorityLow    SchedulingPriority = "low"
)

// PnlMonitoringConfig defines P&L monitoring behavior for auto-rollback.
type PnlMonitoringConfig struct {
	// Enabled enables P&L anomaly detection.
	// +kubebuilder:default=false
	Enabled bool `json:"enabled,omitempty"`

	// CheckInterval is how often to check P&L (seconds).
	// +kubebuilder:default=60
	CheckInterval int32 `json:"checkInterval,omitempty"`

	// RollbackThreshold is the loss threshold (e.g., "$100000") that triggers rollback.
	// +kubebuilder:validation:Pattern=`^\$\d+$`
	RollbackThreshold string `json:"rollbackThreshold,omitempty"`

	// MetricsEndpoint is the Prometheus endpoint for P&L metrics.
	MetricsEndpoint string `json:"metricsEndpoint,omitempty"`
}

// SettlementWorkerSpec defines the desired state of SettlementWorker.
type SettlementWorkerSpec struct {
	// Replicas is the number of settlement worker pods.
	// +kubebuilder:default=1
	Replicas *int32 `json:"replicas,omitempty"`

	// SchedulingPriority defines the QoS priority for scheduling.
	// +kubebuilder:default=medium
	SchedulingPriority SchedulingPriority `json:"schedulingPriority,omitempty"`

	// SettlementType describes the workload (e.g., "daily-reconciliation").
	// +kubebuilder:validation:MinLength=1
	SettlementType string `json:"settlementType,omitempty"`

	// MaxInFlightTransactions prevents accepting new transactions above this limit.
	// +kubebuilder:default=10000
	MaxInFlightTransactions int64 `json:"maxInFlightTransactions,omitempty"`

	// GracefulShutdownTimeout is the time (seconds) to wait for transaction drain.
	// +kubebuilder:default=30
	GracefulShutdownTimeout int32 `json:"gracefulShutdownTimeout,omitempty"`

	// ComplianceLabels enforces required tags (e.g., pci, fca).
	// +kubebuilder:validation:Optional
	ComplianceLabels map[string]string `json:"complianceLabels,omitempty"`

	// PnlMonitoring enables auto-rollback based on financial metrics.
	// +kubebuilder:validation:Optional
	PnlMonitoring *PnlMonitoringConfig `json:"pnlMonitoring,omitempty"`

	// Template is the pod template for the underlying Deployment.
	// +kubebuilder:validation:Required
	Template corev1.PodTemplateSpec `json:"template"`
}

// SettlementWorkerStatus defines the observed state of SettlementWorker.
type SettlementWorkerStatus struct {
	// Phase indicates the overall state: Pending, Running, Failed, etc.
	Phase string `json:"phase,omitempty"`

	// Replicas is the current number of replicas.
	Replicas int32 `json:"replicas,omitempty"`

	// ReadyReplicas is the number of ready replicas.
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Conditions represent the latest available observations.
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// LastRollbackReason explains why an auto-rollback was triggered.
	LastRollbackReason string `json:"lastRollbackReason,omitempty"`

	// LastRollbackTime is the timestamp of the last auto-rollback.
	LastRollbackTime *metav1.Time `json:"lastRollbackTime,omitempty"`

	// ProcessedTransactions is the total count of transactions processed.
	ProcessedTransactions int64 `json:"processedTransactions,omitempty"`

	// ObservedGeneration helps detect spec drift.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.status.replicas`
// +kubebuilder:printcolumn:name="Ready",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// SettlementWorker is the Schema for the settlementworkers API.
type SettlementWorker struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SettlementWorkerSpec   `json:"spec,omitempty"`
	Status SettlementWorkerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SettlementWorkerList contains a list of SettlementWorker.
type SettlementWorkerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SettlementWorker `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &SettlementWorker{}, &SettlementWorkerList{})
		return nil
	})
}
