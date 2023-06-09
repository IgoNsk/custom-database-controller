package usecases

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	informerscorev1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	listerscorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"k8s.io/custom-database/internal/customdatabase"
	clientset "k8s.io/custom-database/pkg/generated/clientset/versioned"
	samplescheme "k8s.io/custom-database/pkg/generated/clientset/versioned/scheme"
	informers "k8s.io/custom-database/pkg/generated/informers/externalversions/cusotmdatabase/v1"
	listers "k8s.io/custom-database/pkg/generated/listers/cusotmdatabase/v1"
)

const controllerAgentName = "custom-database-controller"

const (
	// SuccessSynced is used as part of the Event 'reason' when a CustomDatabase is synced
	SuccessSynced = "Synced"
	// MessageResourceSynced is the message used for an Event fired when a CustomDatabase
	// is synced successfully
	MessageResourceSynced = "CustomDatabase synced successfully"
)

// Controller is the controller implementation for CustomDatabase resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	sampleclientset clientset.Interface

	secretLister listerscorev1.SecretLister
	secretSynced cache.InformerSynced

	customDatabasesLister listers.CustomDatabaseLister
	customDatabasesSynced cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface
	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	databaseManager DatabaseManager

	// we use here concrete DomainService instead of interface, because this component - is a business logic, that can't
	// be different or changed. Also this component - pure, without any side effects.
	domainService *customdatabase.DomainService
}

// DatabaseManager interface of component, that encapsulated Postgresql service for management databases and roles
type DatabaseManager interface {
	CreateDatabase(ctx context.Context, database string) error
	DropDatabase(ctx context.Context, database string) error

	CreateUser(ctx context.Context, userName, password string) error
	ChangeUserPassword(ctx context.Context, userName, password string) error
	DropUser(ctx context.Context, userName string) error

	GrantUserToDatabase(ctx context.Context, userName, database string) error
}

// NewController returns a new sample controller
func NewController(
	ctx context.Context,
	kubeclientset kubernetes.Interface,
	sampleclientset clientset.Interface,
	secretInformer informerscorev1.SecretInformer,
	customDatabaseInformer informers.CustomDatabaseInformer,
	databaseManager DatabaseManager,
	domainService *customdatabase.DomainService,
) *Controller {
	logger := klog.FromContext(ctx)

	// Create event broadcaster
	// Add custom-database-controller types to the default Kubernetes Scheme so Events can be
	// logged for custom-database-controller types.
	utilruntime.Must(samplescheme.AddToScheme(scheme.Scheme))
	logger.V(4).Info("Creating event broadcaster")

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:         kubeclientset,
		sampleclientset:       sampleclientset,
		customDatabasesLister: customDatabaseInformer.Lister(),
		customDatabasesSynced: customDatabaseInformer.Informer().HasSynced,
		secretLister:          secretInformer.Lister(),
		secretSynced:          secretInformer.Informer().HasSynced,
		workqueue:             workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "CustomDatabases"),
		recorder:              recorder,
		databaseManager:       databaseManager,
		domainService:         domainService,
	}

	logger.Info("Setting up event handlers")
	// Set up an event handler for when CustomDatabase resources change
	customDatabaseInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueCustomDatabase,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueCustomDatabase(new)
		},
		DeleteFunc: controller.enqueueCustomDatabase,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(ctx context.Context, workers int) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()
	logger := klog.FromContext(ctx)

	// Start the informer factories to begin populating the informer caches
	logger.Info("Starting CustomDatabase controller")

	// Wait for the caches to be synced before starting workers
	logger.Info("Waiting for informer caches to sync")

	if ok := cache.WaitForCacheSync(ctx.Done(), c.customDatabasesSynced, c.secretSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	logger.Info("Starting workers", "count", workers)
	// Launch two workers to process CustomDatabase resources
	for i := 0; i < workers; i++ {
		go wait.UntilWithContext(ctx, c.runWorker, time.Second)
	}

	logger.Info("Started workers")
	<-ctx.Done()
	logger.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem(ctx context.Context) bool {
	obj, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	logger := klog.FromContext(ctx)

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// CustomDatabase resource to be synced.
		if err := c.syncHandler(ctx, key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		logger.Info("Successfully synced", "resourceName", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two.
func (c *Controller) syncHandler(ctx context.Context, key string) error {
	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	ctx = contextWithResourceNameLogger(ctx, name)

	// Get the CustomDatabase resource with this namespace/name
	customDatabase, err := c.customDatabasesLister.CustomDatabases(namespace).Get(name)
	if err != nil {
		// The CustomDatabase resource may no longer exist, in this case we will delete database
		if errors.IsNotFound(err) {
			return c.deleteHandler(ctx, name)
		}
		return err
	}

	return c.addOrUpdateHandler(ctx, customDatabase)
}

// enqueueCustomDatabase takes a CustomDatabase resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than CustomDatabase.
func (c *Controller) enqueueCustomDatabase(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

func contextWithResourceNameLogger(ctx context.Context, name string) context.Context {
	logger := klog.LoggerWithValues(klog.FromContext(ctx), "resourceName", name)

	ctx = context.WithValue(ctx, "syncHandlerLogger", logger)

	return ctx
}

func loggerFromHandlerContext(ctx context.Context) klog.Logger {
	if logger, ok := ctx.Value("syncHandlerLogger").(klog.Logger); ok {
		return logger
	}

	return klog.FromContext(ctx)
}
