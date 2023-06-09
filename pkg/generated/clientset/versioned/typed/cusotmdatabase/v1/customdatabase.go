/*

 */
// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1 "k8s.io/custom-database/pkg/apis/cusotmdatabase/v1"
	scheme "k8s.io/custom-database/pkg/generated/clientset/versioned/scheme"
)

// CustomDatabasesGetter has a method to return a CustomDatabaseInterface.
// A group's client should implement this interface.
type CustomDatabasesGetter interface {
	CustomDatabases(namespace string) CustomDatabaseInterface
}

// CustomDatabaseInterface has methods to work with CustomDatabase resources.
type CustomDatabaseInterface interface {
	Create(ctx context.Context, customDatabase *v1.CustomDatabase, opts metav1.CreateOptions) (*v1.CustomDatabase, error)
	Update(ctx context.Context, customDatabase *v1.CustomDatabase, opts metav1.UpdateOptions) (*v1.CustomDatabase, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.CustomDatabase, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.CustomDatabaseList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.CustomDatabase, err error)
	CustomDatabaseExpansion
}

// customDatabases implements CustomDatabaseInterface
type customDatabases struct {
	client rest.Interface
	ns     string
}

// newCustomDatabases returns a CustomDatabases
func newCustomDatabases(c *IgorV1Client, namespace string) *customDatabases {
	return &customDatabases{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the customDatabase, and returns the corresponding customDatabase object, and an error if there is any.
func (c *customDatabases) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.CustomDatabase, err error) {
	result = &v1.CustomDatabase{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("customdatabases").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of CustomDatabases that match those selectors.
func (c *customDatabases) List(ctx context.Context, opts metav1.ListOptions) (result *v1.CustomDatabaseList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.CustomDatabaseList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("customdatabases").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested customDatabases.
func (c *customDatabases) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("customdatabases").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a customDatabase and creates it.  Returns the server's representation of the customDatabase, and an error, if there is any.
func (c *customDatabases) Create(ctx context.Context, customDatabase *v1.CustomDatabase, opts metav1.CreateOptions) (result *v1.CustomDatabase, err error) {
	result = &v1.CustomDatabase{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("customdatabases").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(customDatabase).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a customDatabase and updates it. Returns the server's representation of the customDatabase, and an error, if there is any.
func (c *customDatabases) Update(ctx context.Context, customDatabase *v1.CustomDatabase, opts metav1.UpdateOptions) (result *v1.CustomDatabase, err error) {
	result = &v1.CustomDatabase{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("customdatabases").
		Name(customDatabase.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(customDatabase).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the customDatabase and deletes it. Returns an error if one occurs.
func (c *customDatabases) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("customdatabases").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *customDatabases) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("customdatabases").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched customDatabase.
func (c *customDatabases) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.CustomDatabase, err error) {
	result = &v1.CustomDatabase{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("customdatabases").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
