// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// QuotaPolicy specifies token quota configuration for inference services.
// Providing a list of backends in the AIGatewayRouteRule allows failover to a different service
// if token quota for a service had been exceeded.
//
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[-1:].type`
// +kubebuilder:metadata:labels="gateway.networking.k8s.io/policy=direct"
type QuotaPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              QuotaPolicySpec `json:"spec,omitempty"`
	// Status defines the status details of the QuotaPolicy.
	Status QuotaPolicyStatus `json:"status,omitempty"`
}

// QuotaPolicySpec specifies rules for computing token based costs of requests.
type QuotaPolicySpec struct {
	// TargetRefs are the names of the AIServiceBackend resources this QuotaPolicy is being attached to.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	// +kubebuilder:validation:XValidation:rule="self.all(ref, ref.group == 'aigateway.envoyproxy.io' && ref.kind == 'AIServiceBackend')", message="targetRefs must reference AIServiceBackend resources"
	TargetRefs []gwapiv1a2.LocalPolicyTargetReference `json:"targetRefs,omitempty"`
	// Quota for all models served by AIServiceBackend(s). This value can be overridden for specific models using the "PerModelQuotas"
	// configuration.
	//
	// +optional
	ServiceQuota ServiceQuotaDefinition `json:"serviceQuota,omitempty"`
	// PerModelQuotas specifies quota for different models served by the AIServiceBackend(s) where this
	// policy is attached.
	//
	// +kubebuilder:validation:MaxItems=128
	// +optional
	PerModelQuotas []PerModelQuota `json:"perModelQuotas,omitempty"`
	// CostTransparency opts this policy into surfacing the computed per-request
	// cost back to the caller. Exposure is additionally gated by the caller
	// sending the "x-ai-eg-cost-visibility: true" request header, so that both
	// the administrator and the caller must agree before pricing is revealed.
	// Disabled by default.
	//
	// +optional
	CostTransparency *CostTransparencyConfig `json:"costTransparency,omitempty"`
}

// +kubebuilder:validation:XValidation:rule="!(has(self.costExpression) && has(self.pricingRef))",message="serviceQuota.costExpression and serviceQuota.pricingRef are mutually exclusive"
// +kubebuilder:validation:XValidation:rule="!(has(self.quota) && has(self.monetaryQuota))",message="serviceQuota.quota and serviceQuota.monetaryQuota are mutually exclusive"
// +kubebuilder:validation:XValidation:rule="has(self.pricingRef) ? has(self.monetaryQuota) : true",message="serviceQuota.pricingRef requires serviceQuota.monetaryQuota"
type ServiceQuotaDefinition struct {
	// CostExpression specifies a CEL expression for computing the quota burndown of the LLM-related request.
	// If no expression is specified the "total_tokens" value is used.
	// For example:
	//
	//  * "input_tokens + cached_input_tokens * 0.1 + output_tokens * 6"
	//
	// CostExpression and PricingRef are mutually exclusive.
	//
	// +optional
	CostExpression *string `json:"costExpression,omitempty"`
	// PricingRef selects a ModelPricing resource that drives money-denominated
	// cost computation for this quota. When set, the data plane uses the
	// referenced pricing table (along with optional region and tier) to compute
	// per-request cost in the unit declared on the pricing table, bypassing
	// CostExpression.
	//
	// CostExpression and PricingRef are mutually exclusive.
	//
	// +optional
	PricingRef *PricingRef `json:"pricingRef,omitempty"`
	// Quota value applicable to all requests.
	// A response with 429 HTTP status code is sent back to the client when
	// the selected requests have exceeded the quota.
	//
	// Quota and MonetaryQuota are mutually exclusive.
	//
	// +optional
	Quota *QuotaValue `json:"quota,omitempty"`
	// MonetaryQuota expresses the quota as an amount of currency in the unit
	// declared on the referenced ModelPricing table. Required when PricingRef
	// is set, and mutually exclusive with Quota.
	//
	// +optional
	MonetaryQuota *MonetaryQuotaValue `json:"monetaryQuota,omitempty"`
}

// PricingRef references a ModelPricing resource and optionally pins the
// region/tier lookup used for factor resolution. When Region or Tier is left
// unset, the data plane falls back to the x-ai-eg-region / x-ai-eg-tier
// request headers, then to the ModelPricing defaults, then to the empty
// wildcard.
type PricingRef struct {
	// Name is the name of the ModelPricing resource in the same namespace as
	// the QuotaPolicy.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Name string `json:"name"`
	// Region pins the region used for factor lookup. Empty means "use request
	// header or pricing-table default".
	//
	// +optional
	// +kubebuilder:validation:MaxLength=63
	Region *string `json:"region,omitempty"`
	// Tier pins the service tier used for factor lookup. Empty means "use
	// request header or pricing-table default".
	//
	// +optional
	// +kubebuilder:validation:MaxLength=63
	Tier *string `json:"tier,omitempty"`
}

// MonetaryQuotaValue defines the quota limit in a currency unit (for example
// MicroDollar or CentiDollar) over a fixed time window.
type MonetaryQuotaValue struct {
	// Amount is the limit allotted for the specified Duration, expressed in the
	// Unit declared below. For example, Amount=1_000_000 with Unit=MicroDollar
	// represents a $1.00 cap.
	//
	// +kubebuilder:validation:Minimum=0
	Amount uint64 `json:"amount"`
	// Currency is the ISO-4217 currency code. Must match the currency declared
	// on the referenced ModelPricing.
	//
	// +kubebuilder:validation:Enum=USD;EUR;GBP;JPY
	Currency string `json:"currency"`
	// Unit is the minor currency unit the Amount is denominated in. Must match
	// the unit declared on the referenced ModelPricing.
	//
	// +kubebuilder:validation:Enum=MicroDollar;MicroEuro;CentiDollar
	Unit string `json:"unit"`
	// Duration is the time window over which Amount applies. Suffix units are:
	// * s - seconds
	// * m - minutes
	// * h - hours
	// * d - days
	Duration string `json:"duration"`
}

// CostTransparencyConfig gates whether a QuotaPolicy exposes per-request cost
// back to callers. Both this administrator gate AND the caller header
// "x-ai-eg-cost-visibility: true" must be present for the data plane to emit
// cost information on a response.
type CostTransparencyConfig struct {
	// Enabled turns on the administrator-side gate. Disabled by default.
	//
	// +kubebuilder:default=false
	Enabled bool `json:"enabled"`
}

type PerModelQuota struct {
	// Model name for which the quota is specified.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	ModelName *string `json:"modelName"`

	// Expression for computing request cost and rules for matching requests to quota buckets.
	//
	// +kubebuilder:validation:Required
	Quota QuotaDefinition `json:"quota"`
}

// QuotaDefinition specified expression for computing request cost and rules for matching requests to quota buckets.
type QuotaDefinition struct {
	// CostExpression specifies a CEL expression for computing the quota burndown of the LLM-related request.
	// If no expression is specified the "total_tokens" value is used.
	// For example:
	//
	//  * "input_tokens + cached_input_tokens * 0.1 + output_tokens * 6"
	//
	// +optional
	CostExpression *string `json:"costExpression,omitempty"`
	// The "Mode" determines how quota is charged to the "DefaultBucket" and matching "BucketRules".
	// In the "exclusive" mode the quota is charged to matching BucketRules or the DefaultBucket
	// if no BucketRules match the request. The request is denied if all matching buckets are out of quota.
	// In the "shared" mode the quota is charged to all matching "BucketRules" AND the "DefaultBucket"
	// and request is allowed only if the quota is available in all matching buckets.
	Mode QuotaBucketMode `json:"mode"`
	// Quota applicable to all traffic. This value can be overridden for specific classes of requests
	// using the "BucketRules" configuration.
	//
	// +optional
	DefaultBucket QuotaValue `json:"defaultBucket"`
	// BucketRules are a list of client selectors and quotas. If a request
	// matches multiple rules, each of their associated quotas get applied, so a
	// single request might burn down the quota for multiple rules.
	//
	// +optional
	BucketRules []QuotaRule `json:"bucketRules"`
}

// QuotaBucketMode specifies how quota is charged across the default bucket and
// any matching bucket rules.
//
// +kubebuilder:validation:Enum=Exclusive;Shared;MostGenerous
type QuotaBucketMode string

const (
	// QuotaBucketModeShared charges the request against every matching bucket
	// (including the default). The request is allowed only when every matching
	// bucket has capacity.
	QuotaBucketModeShared QuotaBucketMode = "Shared"
	// QuotaBucketModeExclusive charges the request against matching bucket rules
	// (or the default bucket when no rules match) and denies the request only
	// when all matching buckets are out of quota.
	QuotaBucketModeExclusive QuotaBucketMode = "Exclusive"
	// QuotaBucketModeMostGenerous charges the request against the single
	// matching bucket with the largest configured cap. Remaining matching
	// buckets are not touched. Ties are broken by declaration order.
	QuotaBucketModeMostGenerous QuotaBucketMode = "MostGenerous"

	// Deprecated: Use QuotaBucketModeShared.
	QuoteBucketModeShared = QuotaBucketModeShared
	// Deprecated: Use QuotaBucketModeExclusive.
	QuoteBucketModeExclusive = QuotaBucketModeExclusive
)

type QuotaRule struct {
	// ClientSelectors holds the list of conditions to select
	// specific clients using attributes from the traffic flow.
	// All individual select conditions must hold True for this rule
	// and its limit to be applied.
	//
	// If no client selectors are specified, the rule applies to all traffic of
	// the targeted AIServiceBackend.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=8
	ClientSelectors []egv1a1.RateLimitSelectCondition `json:"clientSelectors,omitempty"`
	// Quota value for given client selectors.
	// This quota is applied for traffic flows when the selectors
	// compute to True, causing the request to be counted towards the limit.
	// A response with 429 HTTP status code is sent back to the client when
	// the selected requests have exceeded the quota.
	Quota QuotaValue `json:"quota"`
	// ShadowMode indicates whether this quota rule runs in shadow mode.
	// When enabled, all quota checks are performed (cache lookups,
	// counter updates, telemetry generation), but the outcome is never enforced.
	// The request always succeeds, even if the configured quota is exceeded.
	//
	// +optional
	ShadowMode *bool `json:"shadowMode,omitempty"`
}

// QuotaValue defines the quota limits using sliding window.
type QuotaValue struct {
	// The limit alloted for a specified time window.
	Limit uint `json:"limit"`
	// Time window. The suffix is used to specify units. The following
	// suffixes are supported:
	// * s - seconds (the default unit)
	// * m - minutes
	// * h - hours
	Duration string `json:"duration"`
}

// QuotaPolicyList contains a list of QuotaPolicy
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
type QuotaPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QuotaPolicy `json:"items"`
}
