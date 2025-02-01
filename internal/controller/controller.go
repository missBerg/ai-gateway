package controller

import (
	"context"
	"fmt"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	aigv1a1 "github.com/envoyproxy/ai-gateway/api/v1alpha1"
	"github.com/envoyproxy/ai-gateway/internal/controller/token_rotators"
)

func init() { MustInitializeScheme(scheme) }

// scheme contains the necessary schemas for the AI Gateway.
var scheme = runtime.NewScheme()

// MustInitializeScheme initializes the scheme with the necessary schemas for the AI Gateway.
// This is exported for the testing purposes.
func MustInitializeScheme(scheme *runtime.Scheme) {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(aigv1a1.AddToScheme(scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
	utilruntime.Must(egv1a1.AddToScheme(scheme))
	utilruntime.Must(gwapiv1.Install(scheme))
	utilruntime.Must(gwapiv1b1.Install(scheme))
}

// Options defines the program configurable options that may be passed on the command line.
type Options struct {
	ExtProcLogLevel      string
	ExtProcImage         string
	EnableLeaderElection bool
}

// StartControllers starts the controllers for the AI Gateway.
// This blocks until the manager is stopped.
//
// Note: this is tested with envtest, hence the test exists outside of this package. See /tests/controller_test.go.
func StartControllers(ctx context.Context, config *rest.Config, logger logr.Logger, options Options) error {
	opt := ctrl.Options{
		Scheme:           scheme,
		LeaderElection:   options.EnableLeaderElection,
		LeaderElectionID: "envoy-ai-gateway-controller",
	}

	mgr, err := ctrl.NewManager(config, opt)
	if err != nil {
		return fmt.Errorf("failed to create new controller manager: %w", err)
	}

	c := mgr.GetClient()
	indexer := mgr.GetFieldIndexer()
	if err = applyIndexing(indexer.IndexField); err != nil {
		return fmt.Errorf("failed to apply indexing: %w", err)
	}

	sinkChan := make(chan ConfigSinkEvent, 100)
	routeC := NewAIGatewayRouteController(c, kubernetes.NewForConfigOrDie(config), logger.
		WithName("ai-gateway-route"), sinkChan)
	if err = ctrl.NewControllerManagedBy(mgr).
		For(&aigv1a1.AIGatewayRoute{}).
		Owns(&egv1a1.EnvoyExtensionPolicy{}).
		Owns(&gwapiv1.HTTPRoute{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(routeC); err != nil {
		return fmt.Errorf("failed to create controller for AIGatewayRoute: %w", err)
	}

	backendC := NewAIServiceBackendController(c, kubernetes.NewForConfigOrDie(config), logger.
		WithName("ai-service-backend"), sinkChan)
	if err = ctrl.NewControllerManagedBy(mgr).
		For(&aigv1a1.AIServiceBackend{}).
		Complete(backendC); err != nil {
		return fmt.Errorf("failed to create controller for AIServiceBackend: %w", err)
	}

	// Create and start the Token Manager
	tokenManager := NewTokenManager(logger.WithName("token-manager"), sinkChan, c)
	go tokenManager.Start(ctx)

	// Create AWS credentials rotator
	awsCredsRotator := NewAWSCredentialsRotator(
		c,
		kubernetes.NewForConfigOrDie(config),
		logger.WithName("aws-credentials-rotator"),
	)
	if err := tokenManager.RegisterRotator(awsCredsRotator); err != nil {
		return fmt.Errorf("failed to register AWS credentials rotator: %w", err)
	}

	// Get channels for OIDC rotator
	rotationChan := tokenManager.GetRotationChannel()
	scheduleChan := make(chan token_rotators.RotationEvent, 100)

	// Create AWS OIDC rotator
	awsOIDCRotator := NewAWSOIDCRotator(
		c,
		kubernetes.NewForConfigOrDie(config),
		logger.WithName("aws-oidc-rotator"),
		rotationChan,
		scheduleChan,
	)
	if err := tokenManager.RegisterRotator(awsOIDCRotator); err != nil {
		return fmt.Errorf("failed to register AWS OIDC rotator: %w", err)
	}

	// Start the OIDC rotator
	go func() {
		if err := awsOIDCRotator.Start(ctx); err != nil && err != context.Canceled {
			logger.Error(err, "OIDC rotator failed")
		}
	}()

	// Start goroutine to forward scheduled events to token manager
	go func() {
		for {
			select {
			case event := <-scheduleChan:
				if err := tokenManager.RequestRotation(ctx, event); err != nil {
					logger.Error(err, "failed to schedule rotation",
						"type", event.Type,
						"namespace", event.Namespace,
						"name", event.Name)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Create Backend Security Policy controller with Token Manager
	backendSecurityPolicyC := newBackendSecurityPolicyController(
		c,
		kubernetes.NewForConfigOrDie(config),
		logger.WithName("backend-security-policy"),
		sinkChan,
		tokenManager,
	)

	// Register BackendSecurityPolicy controller
	if err = ctrl.NewControllerManagedBy(mgr).
		For(&aigv1a1.BackendSecurityPolicy{}).
		Complete(backendSecurityPolicyC); err != nil {
		return fmt.Errorf("failed to create controller for BackendSecurityPolicy: %w", err)
	}

	secretC := NewSecretController(c, kubernetes.NewForConfigOrDie(config), logger.
		WithName("secret"), sinkChan)
	if err = ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}).
		Complete(secretC); err != nil {
		return fmt.Errorf("failed to create controller for Secret: %w", err)
	}

	sink := newConfigSink(c, kubernetes.NewForConfigOrDie(config), logger.
		WithName("config-sink"), sinkChan, options.ExtProcImage, options.ExtProcLogLevel)

	// Before starting the manager, initialize the config sink to sync all AIServiceBackend and AIGatewayRoute objects in the cluster.
	if err = sink.init(ctx); err != nil {
		return fmt.Errorf("failed to initialize config sink: %w", err)
	}

	if err = mgr.Start(ctx); err != nil { // This blocks until the manager is stopped.
		return fmt.Errorf("failed to start controller manager: %w", err)
	}
	return nil
}

const (
	// k8sClientIndexSecretToReferencingBackendSecurityPolicy is the index name that maps
	// from a Secret to the BackendSecurityPolicy that references it.
	k8sClientIndexSecretToReferencingBackendSecurityPolicy = "SecretToReferencingBackendSecurityPolicy"
	// k8sClientIndexBackendToReferencingAIGatewayRoute is the index name that maps from a Backend to the
	// AIGatewayRoute that references it.
	k8sClientIndexBackendToReferencingAIGatewayRoute = "BackendToReferencingAIGatewayRoute"
	// k8sClientIndexBackendSecurityPolicyToReferencingAIServiceBackend is the index name that maps from a BackendSecurityPolicy
	// to the AIServiceBackend that references it.
	k8sClientIndexBackendSecurityPolicyToReferencingAIServiceBackend = "BackendSecurityPolicyToReferencingAIServiceBackend"
)

func applyIndexing(indexer func(ctx context.Context, obj client.Object, field string, extractValue client.IndexerFunc) error) error {
	err := indexer(context.Background(), &aigv1a1.AIGatewayRoute{},
		k8sClientIndexBackendToReferencingAIGatewayRoute, aiGatewayRouteIndexFunc)
	if err != nil {
		return fmt.Errorf("failed to index field for AIGatewayRoute: %w", err)
	}
	err = indexer(context.Background(), &aigv1a1.AIServiceBackend{},
		k8sClientIndexBackendSecurityPolicyToReferencingAIServiceBackend, aiServiceBackendIndexFunc)
	if err != nil {
		return fmt.Errorf("failed to index field for AIServiceBackend: %w", err)
	}
	err = indexer(context.Background(), &aigv1a1.BackendSecurityPolicy{},
		k8sClientIndexSecretToReferencingBackendSecurityPolicy, backendSecurityPolicyIndexFunc)
	if err != nil {
		return fmt.Errorf("failed to index field for BackendSecurityPolicy: %w", err)
	}
	return nil
}

func aiGatewayRouteIndexFunc(o client.Object) []string {
	aiGatewayRoute := o.(*aigv1a1.AIGatewayRoute)
	var ret []string
	for _, rule := range aiGatewayRoute.Spec.Rules {
		for _, backend := range rule.BackendRefs {
			key := fmt.Sprintf("%s.%s", backend.Name, aiGatewayRoute.Namespace)
			ret = append(ret, key)
		}
	}
	return ret
}

func aiServiceBackendIndexFunc(o client.Object) []string {
	aiServiceBackend := o.(*aigv1a1.AIServiceBackend)
	var ret []string
	if ref := aiServiceBackend.Spec.BackendSecurityPolicyRef; ref != nil {
		ret = append(ret, fmt.Sprintf("%s.%s", ref.Name, aiServiceBackend.Namespace))
	}
	return ret
}

func backendSecurityPolicyIndexFunc(o client.Object) []string {
	backendSecurityPolicy := o.(*aigv1a1.BackendSecurityPolicy)
	var key string
	switch backendSecurityPolicy.Spec.Type {
	case aigv1a1.BackendSecurityPolicyTypeAPIKey:
		apiKey := backendSecurityPolicy.Spec.APIKey
		key = getSecretNameAndNamespace(apiKey.SecretRef, backendSecurityPolicy.Namespace)
	case aigv1a1.BackendSecurityPolicyTypeAWSCredentials:
		awsCreds := backendSecurityPolicy.Spec.AWSCredentials
		if awsCreds.CredentialsFile != nil {
			key = getSecretNameAndNamespace(awsCreds.CredentialsFile.SecretRef, backendSecurityPolicy.Namespace)
		}
	}
	return []string{key}
}

func getSecretNameAndNamespace(secretRef *gwapiv1.SecretObjectReference, namespace string) string {
	if secretRef.Namespace != nil {
		return fmt.Sprintf("%s.%s", secretRef.Name, *secretRef.Namespace)
	}
	return fmt.Sprintf("%s.%s", secretRef.Name, namespace)
}
