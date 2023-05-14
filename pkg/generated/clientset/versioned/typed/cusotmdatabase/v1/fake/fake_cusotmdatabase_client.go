/*

 */
// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
	v1 "k8s.io/custom-database/pkg/generated/clientset/versioned/typed/cusotmdatabase/v1"
)

type FakeCustomdatabaseV1 struct {
	*testing.Fake
}

func (c *FakeCustomdatabaseV1) CustomDatabases(namespace string) v1.CustomDatabaseInterface {
	return &FakeCustomDatabases{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeCustomdatabaseV1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}