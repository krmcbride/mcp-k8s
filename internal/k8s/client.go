// Package k8s provides a Kubernetes client factory with context switching support
// and utilities for dynamic resource operations across multiple clusters.
package k8s

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

// k8sClients bundles together Kubernetes clients needed for dynamic operations.
// This includes both the dynamic client (for CRUD operations on any resource type)
// and the REST mapper (for converting between Kinds and Resources).
type k8sClients struct {
	dynamic    dynamic.Interface
	restMapper meta.RESTMapper
}

// GetDynamicClientForContext creates a Kubernetes dynamic client for the specified context.
// A dynamic client can work with any Kubernetes resource type without needing generated Go types.
//
// Parameters:
//   - k8sContext: The name of the kubeconfig context to use. If empty, uses the current context.
//
// Returns:
//   - A dynamic client interface for performing CRUD operations on any Kubernetes resource
//   - An error if the client creation fails (e.g., invalid context, connection issues)
//
// Example usage:
//
//	client, err := GetDynamicClientForContext("production")
//	pods, err := client.Resource(podGVR).Namespace("default").List(ctx, metav1.ListOptions{})
func GetDynamicClientForContext(k8sContext string) (dynamic.Interface, error) {
	clients, err := getClientsForContext(k8sContext)
	if err != nil {
		return nil, err
	}
	return clients.dynamic, nil
}

// Helper that creates both a dynamic client and REST mapper for a specific Kubernetes context.
//
// The function creates:
// - A dynamic client: Can work with any Kubernetes resource type (built-in or CRD)
// - A REST mapper: Converts between GVK (Group/Version/Kind) and GVR (Group/Version/Resource)
//
// This bundling is useful because operations that need dynamic clients often also need
// REST mapping capabilities (e.g., converting "Pod" to "pods").
func getClientsForContext(k8sContext string) (*k8sClients, error) {
	kubeConfig := getKubeConfigForContext(k8sContext)

	config, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// Create discovery client for REST mapper
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}

	// Create REST mapper
	groupResources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return nil, err
	}
	restMapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	return &k8sClients{
		dynamic:    dynamicClient,
		restMapper: restMapper,
	}, nil
}

// Helper that creates a ClientConfig for a specific context.
// This handles the kubeconfig loading and context switching logic.
//
// The function:
// - Uses the standard kubeconfig loading rules (checks KUBECONFIG env, then ~/.kube/config)
// - Allows overriding the context (empty string means use current context)
// - Returns a deferred loading config (config is only loaded when actually needed)
//
// This separation allows us to centralize kubeconfig handling and makes testing easier.
func getKubeConfigForContext(k8sContext string) clientcmd.ClientConfig {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	if k8sContext == "" {
		configOverrides = nil
	} else {
		configOverrides.CurrentContext = k8sContext
	}

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)
}
