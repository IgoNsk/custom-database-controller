package usecases

import (
	"context"
)

func (c *Controller) deleteHandler(ctx context.Context, customDatabaseName string) error {
	logger := loggerFromHandlerContext(ctx)
	logger.Info("Delete CustomDatabase resource")

	customDatabase := c.domainService.CreateCustomDatabaseEntity(customDatabaseName)

	err := c.databaseManager.DropDatabase(ctx, customDatabase.Database.Name)
	if err != nil {
		return err
	}

	err = c.databaseManager.DropUser(ctx, customDatabase.Database.User)
	if err != nil {
		return err
	}

	return nil
}
