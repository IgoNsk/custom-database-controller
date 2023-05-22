package usecases

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	"k8s.io/custom-database/internal/customdatabase"
	v1 "k8s.io/custom-database/pkg/apis/cusotmdatabase/v1"
)

func (c *Controller) addOrUpdateHandler(
	ctx context.Context, customDatabaseReq *v1.CustomDatabase,
) error {
	var err error

	logger := loggerFromHandlerContext(ctx)
	logger.Info("Add or update CustomDatabase resource")

	// validate input data
	if customDatabaseReq.Spec.SecretName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: secretName name must be specified", customDatabaseReq.Name))
		return nil
	}

	// Get the secret with the name specified in CustomDatabase.spec
	storedSecret, err := c.secretLister.Secrets(customDatabaseReq.Namespace).Get(customDatabaseReq.Spec.SecretName)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	}

	isSecretNotExists := storedSecret == nil
	customDatabase := c.domainService.CreateCustomDatabaseEntity(customDatabaseReq.Name)

	// actualize information about Database objects
	err = c.actualizeDatabaseInStorage(ctx, customDatabase, isSecretNotExists)
	if err != nil {
		return err
	}

	// After all object in Database was created successful, store information about created database is Secret
	err = c.actualizeSecretInStorage(
		ctx, isSecretNotExists, secretWithDBInfo(newEmptySecret(customDatabaseReq), customDatabase),
	)
	if err != nil {
		return err
	}

	c.recorder.Event(customDatabaseReq, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return nil
}

func (c *Controller) actualizeDatabaseInStorage(
	ctx context.Context, customDatabase customdatabase.Entity, isSecretNotExists bool,
) error {
	var err error
	logger := loggerFromHandlerContext(ctx)

	// Create database in Postgresql
	err = c.databaseManager.CreateDatabase(ctx, customDatabase.Database.Name)
	if err == customdatabase.ErrDatabaseAlreadyExists {
		logger.Info("database already exists", "db_name", customDatabase.Database.Name)
	} else if err != nil {
		return err
	}

	// Create Postgresql user for given CustomDatabase
	err = c.databaseManager.CreateUser(ctx, customDatabase.Database.User, customDatabase.Database.Password)
	if err == customdatabase.ErrUserAlreadyExists {
		logger.Info("user already exists", "user_name", customDatabase.Database.User)
		if isSecretNotExists {
			logger.Info("secretName was changed, we have to update user password and store it in new secret",
				"user_name", customDatabase.Database.User,
			)

			err = c.databaseManager.ChangeUserPassword(ctx, customDatabase.Database.User, customDatabase.Database.Password)
			if err != nil {
				return err
			}
		} else {
			logger.Info("user already stored in actual secret",
				"user_name", customDatabase.Database.User,
			)
		}
	} else if err != nil {
		return err
	}

	// Connect user with database - grant all privileges to database for user
	err = c.databaseManager.GrantUserToDatabase(ctx, customDatabase.Database.User, customDatabase.Database.Name)
	if err != nil {
		return err
	}

	return nil
}

// We only create new Secrets. If Secret exists - we expect that it contains actual CustomDatabase variables.
// Reason - in CustomDatabase can change only SecretName value.
// It means, that we have or not actual secret - without any third condition
func (c *Controller) actualizeSecretInStorage(
	ctx context.Context, isSecretNotExists bool, secretNewState *corev1.Secret,
) error {
	var err error
	logger := loggerFromHandlerContext(ctx)

	if isSecretNotExists {
		logger.Info("Create secret resource for CustomDatabase", "secretName", secretNewState.Name)
		_, err = c.kubeclientset.CoreV1().Secrets(secretNewState.Namespace).Create(ctx, secretNewState, metav1.CreateOptions{})

		if err != nil {
			return err
		}
	}

	return nil
}
