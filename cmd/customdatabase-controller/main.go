package main

import (
	"flag"
	"fmt"
	"time"

	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	commonDatabase "k8s.io/custom-database/pkg/postgres"
	"k8s.io/custom-database/pkg/signals"
	"k8s.io/klog/v2"

	"k8s.io/custom-database/internal/customdatabase"
	"k8s.io/custom-database/internal/customdatabase/adapters/postgres"
	clientset "k8s.io/custom-database/pkg/generated/clientset/versioned"
	informers "k8s.io/custom-database/pkg/generated/informers/externalversions"
)

var (
	masterURL  string
	kubeconfig string
	workers    int

	pgHost          string
	pgPort          int
	pgAdminUser     string
	pgAdminPassword string
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	// set up signals so we handle the shutdown signal gracefully
	ctx := signals.SetupSignalHandler()
	logger := klog.FromContext(ctx)

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		logger.Error(err, "Error building kubeconfig")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		logger.Error(err, "Error building kubernetes clientset")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	exampleClient, err := clientset.NewForConfig(cfg)
	if err != nil {
		logger.Error(err, "Error building kubernetes clientset")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	exampleInformerFactory := informers.NewSharedInformerFactory(exampleClient, time.Second*30)

	dbPool, err := commonDatabase.NewDBWithPoolSettings(
		ctx,
		makePostgresqlConnectionDSN(pgHost, pgPort, pgAdminUser, pgAdminPassword),
		commonDatabase.DefaultPoolSettings,
		commonDatabase.DatabaseWithLogger(logger),
		commonDatabase.DatabaseWithBinaryParams(),
	)
	if err != nil {
		logger.Error(err, "Error running commonDatabase connection pool")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
	defer dbPool.Close() // todo сделать консистентно с текущим кодом

	pgDbManager := postgres.NewDbManager(dbPool.DB())
	customDatabaseDomainService, err := customdatabase.NewDomainService(pgHost, pgPort)
	if err != nil {
		logger.Error(err, "Error running commonDatabase connection pool")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}

	controller := customdatabase.NewController(
		ctx, kubeClient, exampleClient,
		kubeInformerFactory.Core().V1().Secrets(),
		exampleInformerFactory.Igor().V1().CustomDatabases(),
		pgDbManager,
		customDatabaseDomainService,
	)

	// notice that there is no need to run Start methods in a separate goroutine. (i.e. go kubeInformerFactory.Start(ctx.done())
	// Start method is non-blocking and runs all registered informers in a dedicated goroutine.
	kubeInformerFactory.Start(ctx.Done())
	exampleInformerFactory.Start(ctx.Done())

	if err = controller.Run(ctx, workers); err != nil {
		logger.Error(err, "Error running controller")
		klog.FlushAndExit(klog.ExitFlushTimeout, 1)
	}
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
	flag.IntVar(&workers, "workers", 2, "Controller run with amount of workers")

	flag.StringVar(&pgHost, "pg_host", "localhost", "Postgresql server host name")
	flag.IntVar(&pgPort, "workers", 5432, "Postgresql server port")
	flag.StringVar(&pgAdminUser, "pg_admin_user", "", "Postgresql user with privileges to create databases and roles")
	flag.StringVar(&pgAdminPassword, "pg_admin_password", "", "Postgresql admin user's password")
}

func makePostgresqlConnectionDSN(host string, port int, username, password string) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/postgres?sslmode=disable",
		username, password, host, port,
	)
}
