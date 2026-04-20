// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	// ConditionTypeAccepted is a condition type for the reconciliation result
	// where resources are accepted.
	ConditionTypeAccepted = "Accepted"
	// ConditionTypeNotAccepted is a condition type for the reconciliation result
	// where resources are not accepted.
	ConditionTypeNotAccepted = "NotAccepted"
	// ConditionTypePricingResolved is set on a QuotaPolicy when every model
	// served by the targeted backends has a reachable pricing row in the
	// referenced ModelPricing (via region/tier fallback).
	ConditionTypePricingResolved = "PricingResolved"
	// ConditionTypeModelNotPriced is set on a QuotaPolicy when one or more
	// served models lack a matching pricing row. Runtime behavior is
	// fail-open-with-warning by default.
	ConditionTypeModelNotPriced = "ModelNotPriced"
	// ConditionTypeBackendTrafficPolicyGenerated is set on a QuotaPolicy when
	// the controller has generated a BackendTrafficPolicy for the targeted
	// backend(s) and Envoy Gateway has accepted it.
	ConditionTypeBackendTrafficPolicyGenerated = "BackendTrafficPolicyGenerated"
)

// AIGatewayRouteStatus contains the conditions by the reconciliation result.
type AIGatewayRouteStatus struct {
	// Conditions is the list of conditions by the reconciliation result.
	// Currently, at most one condition is set.
	//
	// Known .status.conditions.type are: "Accepted", "NotAccepted".
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// AIServiceBackendStatus contains the conditions by the reconciliation result.
type AIServiceBackendStatus struct {
	// Conditions is the list of conditions by the reconciliation result.
	// Currently, at most one condition is set.
	//
	// Known .status.conditions.type are: "Accepted", "NotAccepted".
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// BackendSecurityPolicyStatus contains the conditions by the reconciliation result.
type BackendSecurityPolicyStatus struct {
	// Conditions is the list of conditions by the reconciliation result.
	// Currently, at most one condition is set.
	//
	// Known .status.conditions.type are: "Accepted", "NotAccepted".
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// MCPRouteStatus contains the conditions by the reconciliation result.
type MCPRouteStatus struct {
	// Conditions is the list of conditions by the reconciliation result.
	// Currently, at most one condition is set.
	//
	// Known .status.conditions.type are: "Accepted", "NotAccepted".
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// QuotaPolicyStatus contains the conditions by the reconciliation result.
type QuotaPolicyStatus struct {
	// Conditions is the list of conditions by the reconciliation result.
	//
	// Known .status.conditions.type are: "Accepted", "NotAccepted",
	// "PricingResolved", "ModelNotPriced", "BackendTrafficPolicyGenerated".
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// ModelPricingStatus contains the conditions by the reconciliation result.
type ModelPricingStatus struct {
	// Conditions is the list of conditions by the reconciliation result.
	//
	// Known .status.conditions.type are: "Accepted", "NotAccepted".
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}
