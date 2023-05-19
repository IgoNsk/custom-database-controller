package usecases

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/diff"
	kubeinformers "k8s.io/client-go/informers"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	core "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2/ktesting"

	"k8s.io/custom-database/internal/customdatabase"
	fakeadapter "k8s.io/custom-database/internal/customdatabase/adapters/fake"
	customdatabasecontroller "k8s.io/custom-database/pkg/apis/cusotmdatabase/v1"
	"k8s.io/custom-database/pkg/generated/clientset/versioned/fake"
	informers "k8s.io/custom-database/pkg/generated/informers/externalversions"
)

func TestCreateDatabaseAndSecret(t *testing.T) {
	f := newFixture(t)

	customDatabaseItem := newCustomDatabase("test")
	_, ctx := ktesting.NewTestContext(t)

	f.customDatabaseLister = append(f.customDatabaseLister, customDatabaseItem)
	f.objects = append(f.objects, customDatabaseItem)

	expCustomDb := customdatabase.CreatedDatabaseInfo{
		Host:     customdatabase.Host{"localhost", 5432},
		Database: customdatabase.Database{Name: "test", User: "test", Password: "testtesttest"},
	}

	expSecret := newSecret(customDatabaseItem)
	expFinalSecret := newSecretWithDBInfo(expSecret, expCustomDb)

	f.expectCreateSecretAction(expSecret)
	f.expectUpdateSecretAction(expFinalSecret)
	f.expectExistsDatabase(expCustomDb)

	f.run(ctx, getKey(customDatabaseItem, t))
}

func TestDoNothing(t *testing.T) {
	f := newFixture(t)

	customDatabaseItem := newCustomDatabase("test")
	_, ctx := ktesting.NewTestContext(t)

	f.customDatabaseLister = append(f.customDatabaseLister, customDatabaseItem)
	f.objects = append(f.objects, customDatabaseItem)

	expCustomDb := customdatabase.CreatedDatabaseInfo{
		Host:     customdatabase.Host{"localhost", 5432},
		Database: customdatabase.Database{Name: "test", User: "test", Password: "testtesttest"},
	}
	f.databases = append(f.databases, expCustomDb)

	expFinalSecret := makeReadableSecret(newSecretWithDBInfo(newSecret(customDatabaseItem), expCustomDb))
	f.secretLister = append(f.secretLister, expFinalSecret)
	f.kubeobjects = append(f.kubeobjects, expFinalSecret)

	f.expectExistsDatabase(expCustomDb)

	f.run(ctx, getKey(customDatabaseItem, t))
}

func TestUpdateSecret(t *testing.T) {
	f := newFixture(t)

	oldCustomDatabaseItem := newCustomDatabase("test")
	updatedCustomDatabaseItem := newCustomDatabaseWithCustomSecret("test", "test-secret-updated")
	_, ctx := ktesting.NewTestContext(t)

	f.customDatabaseLister = append(f.customDatabaseLister, updatedCustomDatabaseItem)
	f.objects = append(f.objects, updatedCustomDatabaseItem)

	expCustomDb := customdatabase.CreatedDatabaseInfo{
		Host:     customdatabase.Host{"localhost", 5432},
		Database: customdatabase.Database{Name: "test", User: "test", Password: "testtesttest"},
	}
	f.databases = append(f.databases, expCustomDb)

	oldFinalSecret := makeReadableSecret(newSecretWithDBInfo(newSecret(oldCustomDatabaseItem), expCustomDb))
	f.secretLister = append(f.secretLister, oldFinalSecret)
	f.kubeobjects = append(f.kubeobjects, oldFinalSecret)

	expSecret := newSecret(updatedCustomDatabaseItem)
	expFinalSecret := newSecretWithDBInfo(expSecret, expCustomDb)
	f.expectCreateSecretAction(expSecret)
	f.expectUpdateSecretAction(expFinalSecret)
	f.expectExistsDatabase(expCustomDb)

	f.run(ctx, getKey(updatedCustomDatabaseItem, t))
}

func TestDeleteDatabaseAndSecret(t *testing.T) {
	f := newFixture(t)

	customDatabaseItem := newCustomDatabase("test")
	_, ctx := ktesting.NewTestContext(t)

	expCustomDb := customdatabase.CreatedDatabaseInfo{
		Host:     customdatabase.Host{"localhost", 5432},
		Database: customdatabase.Database{Name: "test", User: "test", Password: "testtesttest"},
	}
	f.databases = append(f.databases, expCustomDb)

	expFinalSecret := makeReadableSecret(newSecretWithDBInfo(newSecret(customDatabaseItem), expCustomDb))
	f.secretLister = append(f.secretLister, expFinalSecret)
	f.kubeobjects = append(f.kubeobjects, expFinalSecret)

	// secret will be deleted, because his owner was deleted. It will do k8s, not controller
	f.notExpectExistsDatabase(expCustomDb)

	f.run(ctx, getKey(customDatabaseItem, t))
}

var (
	alwaysReady        = func() bool { return true }
	noResyncPeriodFunc = func() time.Duration { return 0 }
)

type fixture struct {
	t *testing.T

	client     *fake.Clientset
	kubeclient *k8sfake.Clientset

	// Objects to put in the store.
	customDatabaseLister []*customdatabasecontroller.CustomDatabase
	secretLister         []*corev1.Secret
	databases            []customdatabase.CreatedDatabaseInfo

	// Actions expected to happen on the client.
	kubeactions          []core.Action
	actions              []core.Action
	expectedDatabases    []customdatabase.CreatedDatabaseInfo
	notExpectedDatabases []customdatabase.CreatedDatabaseInfo

	// Objects from here preloaded into NewSimpleFake.
	kubeobjects []runtime.Object
	objects     []runtime.Object
}

func newFixture(t *testing.T) *fixture {
	f := &fixture{}
	f.t = t
	f.objects = []runtime.Object{}
	f.kubeobjects = []runtime.Object{}
	return f
}

func newCustomDatabase(name string) *customdatabasecontroller.CustomDatabase {
	return newCustomDatabaseWithCustomSecret(name, fmt.Sprintf("%s-secret", name))
}

func newCustomDatabaseWithCustomSecret(name, secretName string) *customdatabasecontroller.CustomDatabase {
	return &customdatabasecontroller.CustomDatabase{
		TypeMeta: metav1.TypeMeta{APIVersion: customdatabasecontroller.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: metav1.NamespaceDefault,
		},
		Spec: customdatabasecontroller.CustomDatabaseSpec{
			SecretName: secretName,
		},
	}
}

func (f *fixture) newController(ctx context.Context) (
	*Controller, informers.SharedInformerFactory, kubeinformers.SharedInformerFactory, *fakeadapter.DbManager,
) {
	f.client = fake.NewSimpleClientset(f.objects...)
	f.kubeclient = k8sfake.NewSimpleClientset(f.kubeobjects...)

	i := informers.NewSharedInformerFactory(f.client, noResyncPeriodFunc())
	k8sI := kubeinformers.NewSharedInformerFactory(f.kubeclient, noResyncPeriodFunc())

	domainService, _ := customdatabase.NewDomainService("localhost", 5432)
	databaseManager := fakeadapter.NewDbManager()

	c := NewController(ctx, f.kubeclient, f.client,
		k8sI.Core().V1().Secrets(),
		i.Igor().V1().CustomDatabases(),
		databaseManager,
		domainService,
	)

	c.customDatabasesSynced = alwaysReady
	c.secretSynced = alwaysReady
	c.recorder = &record.FakeRecorder{}

	for _, f := range f.customDatabaseLister {
		i.Igor().V1().CustomDatabases().Informer().GetIndexer().Add(f)
	}

	for _, d := range f.secretLister {
		k8sI.Core().V1().Secrets().Informer().GetIndexer().Add(d)
	}

	for _, d := range f.databases {
		databaseManager.CreateDatabase(context.TODO(), d.Database.Name)
		databaseManager.CreateUser(context.TODO(), d.Database.User, d.Database.Password)
		databaseManager.GrantUserToDatabase(context.TODO(), d.Database.User, d.Database.Name)
	}

	return c, i, k8sI, databaseManager
}

func (f *fixture) run(ctx context.Context, customDatabaseName string) {
	f.runController(ctx, customDatabaseName, true, false)
}

func (f *fixture) runExpectError(ctx context.Context, customDatabaseName string) {
	f.runController(ctx, customDatabaseName, true, true)
}

func (f *fixture) runController(ctx context.Context, customDatabaseName string, startInformers bool, expectError bool) {
	c, i, k8sI, databaseManager := f.newController(ctx)
	if startInformers {
		i.Start(ctx.Done())
		k8sI.Start(ctx.Done())
	}

	err := c.syncHandler(ctx, customDatabaseName)
	if !expectError && err != nil {
		f.t.Errorf("error syncing customDatabase: %v", err)
	} else if expectError && err == nil {
		f.t.Error("expected error syncing customDatabase, got nil")
	}

	actions := filterInformerActions(f.client.Actions())
	for i, action := range actions {
		if len(f.actions) < i+1 {
			f.t.Errorf("%d unexpected actions: %+v", len(actions)-len(f.actions), actions[i:])
			break
		}

		expectedAction := f.actions[i]
		checkAction(expectedAction, action, f.t)
	}

	if len(f.actions) > len(actions) {
		f.t.Errorf("%d additional expected actions:%+v", len(f.actions)-len(actions), f.actions[len(actions):])
	}

	k8sActions := filterInformerActions(f.kubeclient.Actions())
	for i, action := range k8sActions {
		if len(f.kubeactions) < i+1 {
			f.t.Errorf("%d unexpected actions: %+v", len(k8sActions)-len(f.kubeactions), k8sActions[i:])
			break
		}

		expectedAction := f.kubeactions[i]
		checkAction(expectedAction, action, f.t)
	}

	if len(f.kubeactions) > len(k8sActions) {
		f.t.Errorf("%d additional expected actions:%+v", len(f.kubeactions)-len(k8sActions), f.kubeactions[len(k8sActions):])
	}

	// todo refactoring
	for _, db := range f.expectedDatabases {
		if _, isExists := databaseManager.Databases[db.Name]; !isExists {
			f.t.Errorf("%s database didn't create", db.Name)
		}
		if userPassword, isExists := databaseManager.Users[db.User]; isExists {
			if userPassword != db.Password {
				f.t.Errorf("%s user's wassword wrong: expected %s, given %s", db.Name, db.Password, userPassword)
			}
		} else {
			f.t.Errorf("%s user didn't create", db.User)
		}
		if grantedDBs, isExists := databaseManager.User2Database[db.User]; isExists {
			isDBGranted := false
			for _, grantedDB := range grantedDBs {
				if grantedDB == db.Name {
					isDBGranted = true
				}
			}
			if !isDBGranted {
				f.t.Errorf("%s user didn't grant to %s", db.User, db.Name)
			}
		} else {
			f.t.Errorf("%s database didn't create", db.Name)
		}
	}

	for _, notExpectedDB := range f.notExpectedDatabases {
		if _, isExists := databaseManager.Databases[notExpectedDB.Name]; isExists {
			f.t.Errorf("%s database shouldn't exist", notExpectedDB.Name)
		}
		if _, isExists := databaseManager.Users[notExpectedDB.User]; isExists {
			f.t.Errorf("%s user shouldn't exist", notExpectedDB.User)
		}
	}
}

// checkAction verifies that expected and actual actions are equal and both have
// same attached resources
func checkAction(expected, actual core.Action, t *testing.T) {
	if !(expected.Matches(actual.GetVerb(), actual.GetResource().Resource) && actual.GetSubresource() == expected.GetSubresource()) {
		t.Errorf("Expected\n\t%#v\ngot\n\t%#v", expected, actual)
		return
	}

	if reflect.TypeOf(actual) != reflect.TypeOf(expected) {
		t.Errorf("Action has wrong type. Expected: %t. Got: %t", expected, actual)
		return
	}

	switch a := actual.(type) {
	case core.CreateActionImpl:
		e, _ := expected.(core.CreateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()

		if !reflect.DeepEqual(expObject, object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
		}
	case core.UpdateActionImpl:
		e, _ := expected.(core.UpdateActionImpl)
		expObject := e.GetObject()
		object := a.GetObject()

		if !reflect.DeepEqual(expObject, object) {
			t.Errorf("Action %s %s has wrong object\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expObject, object))
		}
	case core.PatchActionImpl:
		e, _ := expected.(core.PatchActionImpl)
		expPatch := e.GetPatch()
		patch := a.GetPatch()

		if !reflect.DeepEqual(expPatch, patch) {
			t.Errorf("Action %s %s has wrong patch\nDiff:\n %s",
				a.GetVerb(), a.GetResource().Resource, diff.ObjectGoPrintSideBySide(expPatch, patch))
		}
	default:
		t.Errorf("Uncaptured Action %s %s, you should explicitly add a case to capture it",
			actual.GetVerb(), actual.GetResource().Resource)
	}
}

// filterInformerActions filters list and watch actions for testing resources.
// Since list and watch don't change resource state we can filter it to lower
// nose level in our tests.
func filterInformerActions(actions []core.Action) []core.Action {
	ret := []core.Action{}
	for _, action := range actions {
		if len(action.GetNamespace()) == 0 &&
			(action.Matches("list", "customdatabases") ||
				action.Matches("watch", "customdatabases") ||
				action.Matches("list", "secrets") ||
				action.Matches("watch", "secrets")) {
			continue
		}
		ret = append(ret, action)
	}

	return ret
}

func (f *fixture) expectCreateSecretAction(s *corev1.Secret) {
	f.kubeactions = append(f.kubeactions, core.NewCreateAction(schema.GroupVersionResource{Resource: "secrets"}, s.Namespace, s))
}

func (f *fixture) expectUpdateSecretAction(s *corev1.Secret) {
	f.kubeactions = append(f.kubeactions, core.NewUpdateAction(schema.GroupVersionResource{Resource: "secrets"}, s.Namespace, s))
}

func (f *fixture) expectDeleteSecretAction(s *corev1.Secret) {
	f.kubeactions = append(f.kubeactions, core.NewDeleteAction(schema.GroupVersionResource{Resource: "secrets"}, s.Namespace, s.Name))
}

func (f *fixture) expectExistsDatabase(cdr customdatabase.CreatedDatabaseInfo) {
	f.expectedDatabases = append(f.expectedDatabases, cdr)
}

func (f fixture) notExpectExistsDatabase(cdr customdatabase.CreatedDatabaseInfo) {
	f.notExpectedDatabases = append(f.notExpectedDatabases, cdr)
}

func getKey(customDatabase *customdatabasecontroller.CustomDatabase, t *testing.T) string {
	key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(customDatabase)
	if err != nil {
		t.Errorf("Unexpected error getting key for customDatabase %v: %v", customDatabase.Name, err)
		return ""
	}
	return key
}

func makeReadableSecret(s *corev1.Secret) *corev1.Secret {
	rs := s.DeepCopy()

	if len(rs.StringData) == 0 {
		return rs
	}

	rs.Data = make(map[string][]byte)
	for k, v := range rs.StringData {
		rs.Data[k] = []byte(v)
	}

	rs.StringData = nil
	return rs
}
