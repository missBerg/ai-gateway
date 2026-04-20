// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ModelPricing describes the per-model pricing factors used by QuotaPolicy
// resources that reference money-denominated limits.
//
// A ModelPricing resource carries one currency + unit for the whole table, and
// one entry per model with a default set of pricing factors. Entries may
// optionally declare (region, tier) overrides to express regional or
// service-tier price variance without duplicating the table.
//
// Pricing factors are authored as decimal strings (e.g. "3.0" for $3 per 1M
// input tokens) so that CRD authors can express provider pricing verbatim.
// The controller pre-scales each factor to a uint64 micro-units-per-1M-tokens
// integer before serializing into the data plane filter configuration, so the
// ExtProc hot path can compute cost with plain uint64 arithmetic.
//
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.conditions[-1:].type`
type ModelPricing struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec defines the details of the ModelPricing.
	Spec ModelPricingSpec `json:"spec,omitempty"`
	// Status defines the status details of the ModelPricing.
	Status ModelPricingStatus `json:"status,omitempty"`
}

// ModelPricingList contains a list of ModelPricing.
//
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
type ModelPricingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ModelPricing `json:"items"`
}

// ModelPricingSpec specifies the pricing factors for a set of models.
//
// +kubebuilder:validation:XValidation:rule="self.models.all(m, !has(m.overrides) || m.overrides.all(o, has(o.region) || has(o.tier)))",message="each pricing override must specify at least one of region or tier"
type ModelPricingSpec struct {
	// Currency is the ISO-4217 currency code the pricing factors are denominated in.
	// All entries in this table share the same currency.
	//
	// +kubebuilder:validation:Enum=USD;EUR;GBP;JPY
	Currency string `json:"currency"`

	// Unit is the minor-unit the data plane will use for cost metadata values.
	// MicroDollar (10^-6 USD) is the default for USD. Use CentiDollar (10^-2 USD)
	// when a single request may accrue more than $1000 of cost, because Envoy
	// enforces a hits_addend ceiling of 1_000_000_000 per descriptor.
	//
	// +kubebuilder:validation:Enum=MicroDollar;MicroEuro;CentiDollar
	Unit string `json:"unit"`

	// PerMillionTokens controls whether the factor values are expressed per
	// 1M tokens (true, the default and recommended form) or per-token (false).
	// The controller performs any scaling required before emitting the pre-scaled
	// integer factors into the filter configuration.
	//
	// +kubebuilder:default=true
	// +optional
	PerMillionTokens bool `json:"perMillionTokens,omitempty"`

	// DefaultRegion is the region used for model lookup when a request does not
	// carry an explicit region (either via QuotaPolicy pricingRef.region or the
	// x-ai-eg-region header). Empty means "no regional default".
	//
	// +optional
	// +kubebuilder:validation:MaxLength=63
	DefaultRegion *string `json:"defaultRegion,omitempty"`

	// DefaultTier is the service tier used for model lookup when a request does
	// not carry an explicit tier. Empty means "no tier default".
	//
	// +optional
	// +kubebuilder:validation:MaxLength=63
	DefaultTier *string `json:"defaultTier,omitempty"`

	// Models lists the per-model pricing entries. Model names must be unique.
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=256
	// +kubebuilder:validation:XValidation:rule="self.all(m1, self.exists_one(m2, m1.name == m2.name))",message="models[].name must be unique within a ModelPricing"
	Models []ModelPricingEntry `json:"models"`
}

// ModelPricingEntry is the pricing row for a single model.
type ModelPricingEntry struct {
	// Name is the model identifier matched against the request model name
	// (after any modelNameOverride has been applied on the route).
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	Name string `json:"name"`

	// Default contains the baseline pricing factors applied when no override
	// matches the request's resolved (region, tier) tuple.
	//
	// +kubebuilder:validation:Required
	Default PricingFactors `json:"default"`

	// Overrides expresses per-(region, tier) variance on top of Default. Each
	// override must specify at least one of region or tier; the canonical
	// matching key is "region|tier" with empty strings acting as wildcards.
	// A pair (region, tier) may appear at most once.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=16
	// +kubebuilder:validation:XValidation:rule="self.all(o1, self.exists_one(o2, (has(o1.region) ? o1.region : '') == (has(o2.region) ? o2.region : '') && (has(o1.tier) ? o1.tier : '') == (has(o2.tier) ? o2.tier : '')))",message="each (region, tier) pair may appear at most once in overrides"
	Overrides []PricingOverride `json:"overrides,omitempty"`

	// ExtensionFactors captures forward-compatible extension token classes
	// (e.g. provider-specific audio or image tokens) as decimal strings keyed
	// by the extension name. The data plane treats unknown extension keys as
	// zero-cost; operators opt in per provider.
	//
	// +optional
	// +kubebuilder:validation:MaxProperties=16
	ExtensionFactors map[string]string `json:"extensionFactors,omitempty"`
}

// PricingOverride is an optional set of factor overrides for a particular
// (region, tier) tuple. Fields left unset on the inner PricingFactors inherit
// from the entry's Default.
type PricingOverride struct {
	// Region is the region this override applies to. Empty means "any region".
	//
	// +optional
	// +kubebuilder:validation:MaxLength=63
	Region *string `json:"region,omitempty"`

	// Tier is the service tier this override applies to. Empty means "any tier".
	//
	// +optional
	// +kubebuilder:validation:MaxLength=63
	Tier *string `json:"tier,omitempty"`

	// PricingFactors is the set of factors to merge on top of Default for this
	// (region, tier). Any unset factor falls through to the entry's Default.
	PricingFactors `json:",inline"`
}

// PricingFactors is the set of per-token-class cost factors for a single model.
// All values are non-negative decimal strings; the controller pre-scales each
// to uint64 micro-units-per-1M-tokens before the data plane consumes them.
//
// The factor names mirror LLMRequestCostType so operators can map pricing
// inputs 1:1 with the usage fields surfaced by existing LLM providers.
type PricingFactors struct {
	// InputTokenCost is the per-1M-input-token cost, as a decimal string.
	// Example: "3.0" meaning $3.00 per 1M input tokens.
	//
	// +optional
	// +kubebuilder:validation:Pattern=`^[0-9]+(\.[0-9]+)?$`
	// +kubebuilder:validation:MaxLength=32
	InputTokenCost *string `json:"inputTokenCost,omitempty"`

	// OutputTokenCost is the per-1M-output-token cost, as a decimal string.
	//
	// +optional
	// +kubebuilder:validation:Pattern=`^[0-9]+(\.[0-9]+)?$`
	// +kubebuilder:validation:MaxLength=32
	OutputTokenCost *string `json:"outputTokenCost,omitempty"`

	// CachedInputTokenCost is the per-1M-cached-input-token cost, as a decimal string.
	//
	// +optional
	// +kubebuilder:validation:Pattern=`^[0-9]+(\.[0-9]+)?$`
	// +kubebuilder:validation:MaxLength=32
	CachedInputTokenCost *string `json:"cachedInputTokenCost,omitempty"`

	// CacheCreationInputTokenCost is the per-1M-cache-creation-input-token cost, as a decimal string.
	//
	// +optional
	// +kubebuilder:validation:Pattern=`^[0-9]+(\.[0-9]+)?$`
	// +kubebuilder:validation:MaxLength=32
	CacheCreationInputTokenCost *string `json:"cacheCreationInputTokenCost,omitempty"`

	// ReasoningTokenCost is the per-1M-reasoning-token cost, as a decimal string.
	//
	// +optional
	// +kubebuilder:validation:Pattern=`^[0-9]+(\.[0-9]+)?$`
	// +kubebuilder:validation:MaxLength=32
	ReasoningTokenCost *string `json:"reasoningTokenCost,omitempty"`
}
