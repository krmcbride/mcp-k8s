package k8s

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sClients struct {
	Dynamic    dynamic.Interface
	RESTMapper meta.RESTMapper
}

func getClientsForContext(k8sContext string) (*K8sClients, error) {
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

	return &K8sClients{
		Dynamic:    dynamicClient,
		RESTMapper: restMapper,
	}, nil
}

func GetDynamicClientForContext(k8sContext string) (dynamic.Interface, error) {
	clients, err := getClientsForContext(k8sContext)
	if err != nil {
		return nil, err
	}
	return clients.Dynamic, nil
}

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
