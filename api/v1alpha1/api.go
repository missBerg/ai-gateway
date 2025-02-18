// Copyright Envoy AI Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// +kubebuilder:object:root=true

// AIGatewayRoute combines multiple AIServiceBackends and attaches them to Gateway resources.
//
// It provides a unified AI API interface for your Gateway, allowing downstream clients
// to interact with multiple AI backends using a consistent schema. Key features:
//
// - Defines a unified API schema for client requests
// - Routes traffic to appropriate AI backends
// - Handles schema transformations between client and backend
// - Manages upstream authentication
// - Manages rate limiting
//
// **The AI Gateway controller generates these Kubernetes resources:**
//
// For API Processing:
//   - Deployment, Service, and ConfigMap (named: ai-eg-route-extproc-${name})
//
// For Routing:
//   - HTTPRoute (named same as AIGatewayRoute)
//   - EnvoyExtensionPolicy (named: ai-eg-route-extproc-${name})
//   - HTTPRouteFilter for hostname rewrites (named: ai-eg-host-rewrite)
//
// All resources are created in the same namespace as the AIGatewayRoute.
//
// While these implementation details may change, you can:
// - Use these resources as reference for custom configurations
// - Use EnvoyPatchPolicy to modify the generated resources (e.g., adding custom filters)
type AIGatewayRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec defines the details of the AIGatewayRoute.
	Spec AIGatewayRouteSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// AIGatewayRouteList contains a list of AIGatewayRoute.
type AIGatewayRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AIGatewayRoute `json:"items"`
}

// AIGatewayRouteSpec defines the desired state of an AIGatewayRoute.
type AIGatewayRouteSpec struct {
	// TargetRefs specifies which Gateway resources this AIGatewayRoute should attach to.
	// These references determine which Gateways will handle the AI traffic routing defined here.
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=128
	TargetRefs []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName `json:"targetRefs"`

	// APISchema defines the API schema that client requests must follow when sending requests to the Gateway. The AI Gateway will automatically transform these requests to match the schema required by the selected backend service.
	//
	// Currently supports OpenAI schema only.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule="self.name == 'OpenAI'"
	APISchema VersionedAPISchema `json:"schema"`

	// Rules defines how incoming requests are matched and routed to AI backends. Each rule extends the Gateway API's HTTPRoute concept with AI-specific features.\n\n
	// The AI Gateway controller generates an HTTPRoute based on these rules and adds, AI Gateway filter for request/response transformation, Model-based routing using the `x-ai-eg-model` header, and other AI-specific processing.\n\n
	// Rules are matched according to Gateway API HTTPRoute specifications.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxItems=128
	Rules []AIGatewayRouteRule `json:"rules"`

	// Configures how the AI Gateway processes and transforms requests/responses.\n\n
	// The filter handles:
	// - Request/response transformations between different AI API schemas
	// - Model information extraction for routing
	// - Request validation and preprocessing\n\n
	//
	// <sub>Currently implemented as an external process filter, with potential future extensions to other filter types (see: https://github.com/envoyproxy/ai-gateway/issues/90)</sub>
	FilterConfig *AIGatewayFilterConfig `json:"filterConfig,omitempty"`

	// LLMRequestCosts configures token usage tracking for LLM requests.
	// Each metric is stored in Envoy's dynamic metadata under "io.envoy.ai_gateway".
	//
	// Example Configuration:
	// ```yaml
	// llmRequestCosts:
	// - metadataKey: llm_input_token
	//   type: InputToken
	// - metadataKey: llm_output_token
	//   type: OutputToken
	// - metadataKey: llm_total_token
	//   type: TotalToken
	// ```
	//
	// Rate Limiting Integration:
	// The captured metrics can be used with Envoy Gateway's `BackendTrafficPolicy` for token-based rate limiting.<br/>
	// <b>For example, you can:</b>
	// - Set separate limits for input, output, and total tokens
	// - Apply limits per user using request headers
	// - Track token usage across multiple requests
	//
	// <hr><sub>See full rate limiting examples in our documentation: https://aigateway.envoyproxy.io/docs/capabilities/usage-based-ratelimiting</sub>
	//
	// +optional
	// +kubebuilder:validation:MaxItems=36
	LLMRequestCosts []LLMRequestCost `json:"llmRequestCosts,omitempty"`
}

// AIGatewayRouteRule defines how to match and route incoming requests to AI backends.
type AIGatewayRouteRule struct {
	// BackendRefs specifies the AI service backends to route traffic to.\n\n
	// Each backend can have a weight for traffic distribution (load balancing).\n\n
	// <b>Note:</b> Backends must be in the same namespace as the AIGatewayRoute.\n\n
	//
	// +optional
	// +kubebuilder:validation:MaxItems=128
	BackendRefs []AIGatewayRouteRuleBackendRef `json:"backendRefs,omitempty"`

	// Matches defines conditions for when this rule applies.\n\n
	// Based on Gateway API's HTTPRouteMatch with AI-specific extensions.\n\n
	// There is a special header for routing: `x-ai-eg-model`. It is used for model-based routing. The model name is automatically extracted from the request content.
	//
	// +optional
	// +kubebuilder:validation:MaxItems=128
	Matches []AIGatewayRouteRuleMatch `json:"matches,omitempty"`
}

// AIGatewayRouteRuleBackendRef references an AIServiceBackend with an optional weight.
type AIGatewayRouteRuleBackendRef struct {
	// Name of the AIServiceBackend to route to.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Weight determines the proportion of traffic to route to this backend.\n\n
	// Higher weights receive more traffic relative to other backends.\n\n
	// For details on weight-based routing, see the Gateway API documentation.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=1
	Weight int `json:"weight,omitempty"`
}

type AIGatewayRouteRuleMatch struct {
	// Headers defines HTTP request header matching rules.
	// Based on Gateway API's HTTPHeaderMatch specification.
	// \n\n
	// <b>Matching Types:</b>
	// - Currently only exact header matching is supported
	// - RegularExpression matching is planned for future releases
	// \n\n
	// <b>Common Use Cases:</b>
	// - Route based on API versions
	// - Route based on client identifiers
	// - Route based on model preferences
	// \n\n
	// +listType=map
	// +listMapKey=name
	// +optional
	// +kubebuilder:validation:MaxItems=16
	// +kubebuilder:validation:XValidation:rule="self.all(match, match.type != 'RegularExpression')", message="currently only exact match is supported"
	Headers []gwapiv1.HTTPHeaderMatch `json:"headers,omitempty"`
}

type AIGatewayFilterConfig struct {
	// Type specifies the filter implementation to use. Currently supports "ExtProc" (External Processing) only.\n\n
	// Currently, only ExternalProcess is supported, and default is ExternalProcess.
	//
	// +kubebuilder:default=ExternalProcess
	Type AIGatewayFilterConfigType `json:"type"`

	// ExternalProcess is the configuration for the external process filter.\n\n
	// This is optional, and if not set, the default values of Deployment spec will be used.
	//
	// +optional
	ExternalProcess *AIGatewayFilterConfigExternalProcess `json:"externalProcess,omitempty"`
}

// AIGatewayFilterConfigType defines the available filter implementations.
//
// +kubebuilder:validation:Enum=ExtProc
type AIGatewayFilterConfigType string

const (
	AIGatewayFilterConfigTypeExternalProcess AIGatewayFilterConfigType = "ExternalProcess"
	AIGatewayFilterConfigTypeDynamicModule   AIGatewayFilterConfigType = "DynamicModule" // Reserved for https://github.com/envoyproxy/ai-gateway/issues/90
)

type AIGatewayFilterConfigExternalProcess struct {
	// Replicas is the number of desired pods of the external process deployment.
	//
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
	// Kubernetes resource requirements for the external process container.
	//
	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`
	// TODO: maybe adding the option not to deploy the external process filter and let the user deploy it manually?
	// 	Not sure if it is worth it as we are migrating to dynamic modules.
}

// +kubebuilder:object:root=true

// AIServiceBackend is a resource that represents a single backend for AIGatewayRoute.
//
// A AIServiceBackend is "attached" to a Backend which is either a k8s Service or a Envoy Gateway Backend resource.
//
// When a backend with an attached AIServiceBackend is used as a routing target in the AIGatewayRoute (more precisely, the HTTPRouteSpec defined in the AIGatewayRoute), the ai-gateway will generate the necessary configuration to do the backend specific logic in the final HTTPRoute.
type AIServiceBackend struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// Spec defines the details of AIServiceBackend.
	Spec AIServiceBackendSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// AIServiceBackendList contains a list of AIServiceBackends.
type AIServiceBackendList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AIServiceBackend `json:"items"`
}

// AIServiceBackendSpec details the AIServiceBackend configuration.
type AIServiceBackendSpec struct {
	// APISchema specifies the API schema of the output format of requests from Envoy that this AIServiceBackend can accept as incoming requests.\n\n
	// Based on this schema, the ai-gateway will perform the necessary transformation for the pair of `AIGatewayRouteSpec.APISchema` and `AIServiceBackendSpec.APISchema`.
	//
	// +kubebuilder:validation:Required
	APISchema VersionedAPISchema `json:"schema"`
	// BackendRef is the reference to the Backend resource that this AIServiceBackend corresponds to.\n\n
	// A backend can be of either k8s Service or Backend resource of Envoy Gateway.
	//
	// +kubebuilder:validation:Required
	BackendRef gwapiv1.BackendObjectReference `json:"backendRef"`

	// BackendSecurityPolicyRef is the name of the BackendSecurityPolicy resources this backend is being attached to.
	//
	// +optional
	BackendSecurityPolicyRef *gwapiv1.LocalObjectReference `json:"backendSecurityPolicyRef,omitempty"`

	// TODO: maybe add backend-level LLMRequestCost configuration that overrides the AIGatewayRoute-level LLMRequestCost.
	// 	That may be useful for the backend that has a different cost calculation logic.
}

// VersionedAPISchema defines the API schema of either AIGatewayRoute (the input) or AIServiceBackend (the output).
// This allows the ai-gateway to understand the input and perform the necessary transformation depending on the API schema pair (input, output).
//
// Note that this is vendor specific, and the stability of the API schema is not guaranteed by the ai-gateway, but by the vendor via proper versioning.
type VersionedAPISchema struct {
	// Name is the name of the API schema of the AIGatewayRoute or AIServiceBackend.
	//
	// +kubebuilder:validation:Enum=OpenAI;AWSBedrock
	Name APISchema `json:"name"`

	// Version is the version of the API schema.
	Version string `json:"version,omitempty"`
}

// APISchema defines the API schema.
type APISchema string

const (
	// APISchemaOpenAI is the OpenAI schema.
	//
	// https://github.com/openai/openai-openapi
	APISchemaOpenAI APISchema = "OpenAI"
	// APISchemaAWSBedrock is the AWS Bedrock schema.
	//
	// https://docs.aws.amazon.com/bedrock/latest/APIReference/API_Operations_Amazon_Bedrock_Runtime.html
	APISchemaAWSBedrock APISchema = "AWSBedrock"
)

const (
	// AIModelHeaderKey is the header key whose value is extracted from the request by the ai-gateway.

	// This can be used to describe the routing behavior in HTTPRoute referenced by AIGatewayRoute.
	AIModelHeaderKey = "x-ai-eg-model"
)

// BackendSecurityPolicyType specifies the type of auth mechanism used to access a backend.
type BackendSecurityPolicyType string

const (
	BackendSecurityPolicyTypeAPIKey         BackendSecurityPolicyType = "APIKey"
	BackendSecurityPolicyTypeAWSCredentials BackendSecurityPolicyType = "AWSCredentials"
)

// +kubebuilder:object:root=true

// BackendSecurityPolicy specifies configuration for authentication and authorization rules on the traffic exiting the gateway to the backend.
type BackendSecurityPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              BackendSecurityPolicySpec `json:"spec,omitempty"`
}

// BackendSecurityPolicySpec specifies authentication rules on access the provider from the Gateway.\n\n
// Only one mechanism to access a backend(s) can be specified.
//
// Only one type of `BackendSecurityPolicy` can be defined.
// +kubebuilder:validation:MaxProperties=2
type BackendSecurityPolicySpec struct {
	// Type specifies the auth mechanism used to access the provider. Currently, only `APIKey`, AND `AWSCredentials` are supported.
	//
	// +kubebuilder:validation:Enum=APIKey;AWSCredentials
	Type BackendSecurityPolicyType `json:"type"`

	// APIKey is a mechanism to access a backend(s). The API key will be injected into the Authorization header.
	//
	// +optional
	APIKey *BackendSecurityPolicyAPIKey `json:"apiKey,omitempty"`

	// AWSCredentials is a mechanism to access a backend(s). AWS specific logic will be applied.
	//
	// +optional
	AWSCredentials *BackendSecurityPolicyAWSCredentials `json:"awsCredentials,omitempty"`
}

// +kubebuilder:object:root=true

// BackendSecurityPolicyList contains a list of BackendSecurityPolicy
type BackendSecurityPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackendSecurityPolicy `json:"items"`
}

// BackendSecurityPolicyAPIKey specifies the API key.
type BackendSecurityPolicyAPIKey struct {
	// SecretRef is the reference to the secret containing the API key. ai-gateway must be given the permission to read this secret.\n\n
	// The key of the secret should be "apiKey".
	SecretRef *gwapiv1.SecretObjectReference `json:"secretRef"`
}

// BackendSecurityPolicyAWSCredentials contains the supported authentication mechanisms to access aws
type BackendSecurityPolicyAWSCredentials struct {
	// Region specifies the AWS region associated with the policy.
	//
	// +kubebuilder:validation:MinLength=1
	Region string `json:"region"`

	// CredentialsFile specifies the credentials file to use for the AWS provider.
	//
	// +optional
	CredentialsFile *AWSCredentialsFile `json:"credentialsFile,omitempty"`

	// OIDCExchangeToken specifies the oidc configurations used to obtain an oidc token. The oidc token will be used to obtain temporary credentials to access AWS.
	//
	// +optional
	OIDCExchangeToken *AWSOIDCExchangeToken `json:"oidcExchangeToken,omitempty"`
}

// `AWSCredentialsFile` specifies the credentials file to use for the AWS provider.\n\n
// Envoy reads the secret file, and the profile to use is specified by the `Profile` field.\n\n
type AWSCredentialsFile struct {
	// SecretRef is the reference to the credential file.
	//
	// The secret should contain the AWS credentials file keyed on "credentials".
	SecretRef *gwapiv1.SecretObjectReference `json:"secretRef"`

	// Profile is the profile to use in the credentials file.
	//
	// +kubebuilder:default=default
	Profile string `json:"profile,omitempty"`
}

// `AWSOIDCExchangeToken` specifies credentials to obtain oidc token from a sso server.\n\n
// For AWS, the controller will query STS to obtain AWS AccessKeyId, SecretAccessKey, and SessionToken, and store them in a temporary credentials file.\n\n
type AWSOIDCExchangeToken struct {
	// OIDC is used to obtain oidc tokens via an SSO server which will be used to exchange for temporary AWS credentials.\n\n
	OIDC egv1a1.OIDC `json:"oidc"`

	// GrantType is the method application gets access token.\n\n
	//
	// +optional
	GrantType string `json:"grantType,omitempty"`

	// Aud defines the audience that this ID Token is intended for.
	//
	// +optional
	Aud string `json:"aud,omitempty"`

	// AwsRoleArn is the AWS IAM Role with the permission to use specific resources in AWS account which maps to the temporary AWS security credentials exchanged using the authentication token issued by OIDC provider.
	AwsRoleArn string `json:"awsRoleArn"`
}

// `LLMRequestCost` defines how to capture and track token usage metrics for Large Language Model requests.
type LLMRequestCost struct {
	// MetadataKey is the key under which the token usage metric will be stored
	// in Envoy's dynamic metadata (namespace: io.envoy.ai_gateway).\n\n
	// <b>Example keys:</b>
	// - `llm_input_token`
	// - `llm_output_token`
	// - `llm_total_token`
	//
	// +kubebuilder:validation:Required
	MetadataKey string `json:"metadataKey"`

	// Type specifies which token metric to capture.
	// See `LLMRequestCostType` for available options.
	//
	// +kubebuilder:validation:Enum=OutputToken;InputToken;TotalToken;CEL
	Type LLMRequestCostType `json:"type"`
	// `CELExpression` is the CEL expression to calculate the cost of the request.\n\n
	// The CEL expression must return a signed or unsigned integer. If the return value is negative, it will be error.\n\n
	// The expression can use the following variables:\n\n
	// <b>model</b>: the model name extracted from the request content. Type: string.
	// <b>backend</b>: the backend name in the form of "name.namespace". Type: string.
	// <b>input_tokens</b>: the number of input tokens. Type: unsigned integer.
	// <b>output_tokens</b>: the number of output tokens. Type: unsigned integer.
	// <b>total_tokens</b>: the total number of tokens. Type: unsigned integer.
	// \n\n
	// For example, the following expressions are valid:\n\n
	// ```
	// model == 'llama' ?  input_tokens + output_token * 0.5 : total_tokens
	//
	// backend == 'foo.default' ?  input_tokens + output_tokens : total_tokens
	//
	// input_tokens + output_tokens + total_tokens
	//
	// input_tokens * output_tokens
	// ```
	//
	// +optional
	CELExpression *string `json:"celExpression,omitempty"`
}

// LLMRequestCostType defines the types of token metrics that can be captured.
// +kubebuilder:validation:Enum=InputToken;OutputToken;TotalToken
type LLMRequestCostType string

const (
	// LLMRequestCostTypeInputToken is the cost type of the input token.
	LLMRequestCostTypeInputToken LLMRequestCostType = "InputToken"
	// LLMRequestCostTypeOutputToken is the cost type of the output token.
	LLMRequestCostTypeOutputToken LLMRequestCostType = "OutputToken"
	// LLMRequestCostTypeTotalToken is the cost type of the total token.
	LLMRequestCostTypeTotalToken LLMRequestCostType = "TotalToken"
	// LLMRequestCostTypeCEL is for calculating the cost using the CEL expression.
	LLMRequestCostTypeCEL LLMRequestCostType = "CEL"
)

const (
	// AIGatewayFilterMetadataNamespace is the namespace for the ai-gateway filter metadata.
	AIGatewayFilterMetadataNamespace = "io.envoy.ai_gateway"
)
