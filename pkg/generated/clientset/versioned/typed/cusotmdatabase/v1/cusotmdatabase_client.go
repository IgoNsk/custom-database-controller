/*

 */
// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"net/http"

	rest "k8s.io/client-go/rest"
	v1 "k8s.io/custom-database/pkg/apis/cusotmdatabase/v1"
	"k8s.io/custom-database/pkg/generated/clientset/versioned/scheme"
)

type IgorV1Interface interface {
	RESTClient() rest.Interface
	CustomDatabasesGetter
}

// IgorV1Client is used to interact with features provided by the igor.yatsevich.ru group.
type IgorV1Client struct {
	restClient rest.Interface
}

func (c *IgorV1Client) CustomDatabases(namespace string) CustomDatabaseInterface {
	return newCustomDatabases(c, namespace)
}

// NewForConfig creates a new IgorV1Client for the given config.
// NewForConfig is equivalent to NewForConfigAndClient(c, httpClient),
// where httpClient was generated with rest.HTTPClientFor(c).
func NewForConfig(c *rest.Config) (*IgorV1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	httpClient, err := rest.HTTPClientFor(&config)
	if err != nil {
		return nil, err
	}
	return NewForConfigAndClient(&config, httpClient)
}

// NewForConfigAndClient creates a new IgorV1Client for the given config and http client.
// Note the http client provided takes precedence over the configured transport values.
func NewForConfigAndClient(c *rest.Config, h *http.Client) (*IgorV1Client, error) {
	config := *c
	if err := setConfigDefaults(&config); err != nil {
		return nil, err
	}
	client, err := rest.RESTClientForConfigAndClient(&config, h)
	if err != nil {
		return nil, err
	}
	return &IgorV1Client{client}, nil
}

// NewForConfigOrDie creates a new IgorV1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *IgorV1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return client
}

// New creates a new IgorV1Client for the given RESTClient.
func New(c rest.Interface) *IgorV1Client {
	return &IgorV1Client{c}
}

func setConfigDefaults(config *rest.Config) error {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = "/apis"
	config.NegotiatedSerializer = scheme.Codecs.WithoutConversion()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	return nil
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *IgorV1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}
	return c.restClient
}
