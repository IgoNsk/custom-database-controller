/*

 */
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	v1 "k8s.io/custom-database/pkg/apis/cusotmdatabase/v1"
)

// FakeCustomDatabases implements CustomDatabaseInterface
type FakeCustomDatabases struct {
	Fake *FakeCustomdatabaseV1
	ns   string
}

var customdatabasesResource = v1.SchemeGroupVersion.WithResource("customdatabases")

var customdatabasesKind = v1.SchemeGroupVersion.WithKind("CustomDatabase")

// Get takes name of the customDatabase, and returns the corresponding customDatabase object, and an error if there is any.
func (c *FakeCustomDatabases) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.CustomDatabase, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(customdatabasesResource, c.ns, name), &v1.CustomDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.CustomDatabase), err
}

// List takes label and field selectors, and returns the list of CustomDatabases that match those selectors.
func (c *FakeCustomDatabases) List(ctx context.Context, opts metav1.ListOptions) (result *v1.CustomDatabaseList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(customdatabasesResource, customdatabasesKind, c.ns, opts), &v1.CustomDatabaseList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1.CustomDatabaseList{ListMeta: obj.(*v1.CustomDatabaseList).ListMeta}
	for _, item := range obj.(*v1.CustomDatabaseList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested customDatabases.
func (c *FakeCustomDatabases) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(customdatabasesResource, c.ns, opts))

}

// Create takes the representation of a customDatabase and creates it.  Returns the server's representation of the customDatabase, and an error, if there is any.
func (c *FakeCustomDatabases) Create(ctx context.Context, customDatabase *v1.CustomDatabase, opts metav1.CreateOptions) (result *v1.CustomDatabase, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(customdatabasesResource, c.ns, customDatabase), &v1.CustomDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.CustomDatabase), err
}

// Update takes the representation of a customDatabase and updates it. Returns the server's representation of the customDatabase, and an error, if there is any.
func (c *FakeCustomDatabases) Update(ctx context.Context, customDatabase *v1.CustomDatabase, opts metav1.UpdateOptions) (result *v1.CustomDatabase, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(customdatabasesResource, c.ns, customDatabase), &v1.CustomDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.CustomDatabase), err
}

// Delete takes name of the customDatabase and deletes it. Returns an error if one occurs.
func (c *FakeCustomDatabases) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteActionWithOptions(customdatabasesResource, c.ns, name, opts), &v1.CustomDatabase{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeCustomDatabases) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(customdatabasesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1.CustomDatabaseList{})
	return err
}

// Patch applies the patch and returns the patched customDatabase.
func (c *FakeCustomDatabases) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.CustomDatabase, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(customdatabasesResource, c.ns, name, pt, data, subresources...), &v1.CustomDatabase{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1.CustomDatabase), err
}