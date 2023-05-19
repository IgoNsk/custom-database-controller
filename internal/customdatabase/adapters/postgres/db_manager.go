package postgres

import (
	"context"
	"database/sql"
	"github.com/lib/pq"

	"k8s.io/custom-database/internal/customdatabase"
)

// DbManager
//
// We can't use statements here. Generally anything that modifies schemas doesn't support them.
type DbManager struct {
	db DB
}

type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func NewDbManager(db DB) *DbManager {
	return &DbManager{db}
}

func (am *DbManager) CreateUser(ctx context.Context, userName, password string) error {
	// https://www.postgresql.org/docs/current/sql-createrole.html
	_, err := am.db.ExecContext(
		ctx, "CREATE ROLE "+pq.QuoteIdentifier(userName)+" WITH LOGIN ENCRYPTED PASSWORD "+pq.QuoteLiteral(password),
	)
	if err != nil {
		pgError := err.(*pq.Error)
		if pgError.Code == "42710" {
			return customdatabase.ErrUserAlreadyExists
		}
		return err
	}

	return nil
}

func (am *DbManager) ChangeUserPassword(ctx context.Context, userName, password string) error {
	// https://www.postgresql.org/docs/current/sql-createrole.html
	_, err := am.db.ExecContext(
		ctx, "ALTER ROLE "+pq.QuoteIdentifier(userName)+" WITH ENCRYPTED PASSWORD "+pq.QuoteLiteral(password),
	)
	if err != nil {
		return err
	}

	return nil
}

func (am *DbManager) DropUser(ctx context.Context, userName string) error {
	// https://www.postgresql.org/docs/current/sql-createrole.html
	_, err := am.db.ExecContext(
		ctx, "DROP ROLE IF EXISTS "+pq.QuoteIdentifier(userName),
	)
	if err != nil {
		return err
	}

	return nil
}

func (am *DbManager) CreateDatabase(ctx context.Context, database string) error {
	// https://www.postgresql.org/docs/current/sql-createdatabase.html
	_, err := am.db.ExecContext(ctx, "CREATE DATABASE "+pq.QuoteIdentifier(database))
	if err != nil {
		pgError := err.(*pq.Error)
		if pgError.Code == "42P04" {
			return customdatabase.ErrDatabaseAlreadyExists
		}
		return err
	}

	return nil
}

func (am *DbManager) DropDatabase(ctx context.Context, database string) error {
	// https://www.postgresql.org/docs/current/sql-createdatabase.html
	res, err := am.db.ExecContext(ctx, "DROP DATABASE IF EXISTS "+pq.QuoteIdentifier(database))
	if err != nil {
		return err
	}

	_ = res
	return nil
}

func (am *DbManager) GrantUserToDatabase(ctx context.Context, userName, database string) error {
	// https://www.postgresql.org/docs/current/sql-grant.html
	_, err := am.db.ExecContext(
		ctx, "GRANT ALL PRIVILEGES ON DATABASE "+pq.QuoteIdentifier(database)+" TO "+pq.QuoteIdentifier(userName),
	)
	if err != nil {
		return err
	}
	return nil
}
