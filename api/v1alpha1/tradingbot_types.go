package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TradingBotSpec defines the desired state
type TradingBotSpec struct {
	// Replicas is the number of trading bot pods
	Replicas *int32 `json:"replicas,omitempty"`

	// NodeAffinity: which node pool to use ("low-latency-pool")
	NodeAffinity string `json:"nodeAffinity,omitempty"`

	// CPUPinning enables CPU pinning for this bot
	CPUPinning bool `json:"cpuPinning,omitempty"`

	// TradingPair like "USD/EUR"
	TradingPair string `json:"tradingPair,omitempty"`

	// Image is the container image to run
	Image string `json:"image,omitempty"`
}

// TradingBotStatus defines the observed state
type TradingBotStatus struct {
	// Phase: Pending, Running, Failed
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

// TradingBot is the Schema for trading bots
type TradingBot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TradingBotSpec   `json:"spec,omitempty"`
	Status TradingBotStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TradingBotList contains a list of TradingBot
type TradingBotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TradingBot `json:"items"`
}
