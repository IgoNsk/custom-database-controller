package fake

import (
	"context"
	"fmt"
	"sync"

	"k8s.io/custom-database/internal/customdatabase"
)

// DbManager fake implementation for tests
type DbManager struct {
	Users         map[string]string
	Databases     map[string]struct{}
	User2Database map[string][]string

	mu sync.Mutex
}

func NewDbManager() *DbManager {
	return &DbManager{
		Users:         make(map[string]string),
		Databases:     make(map[string]struct{}),
		User2Database: make(map[string][]string),
		mu:            sync.Mutex{},
	}
}

func (am *DbManager) CreateUser(_ context.Context, userName, password string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, isExists := am.Users[userName]; isExists {
		return customdatabase.ErrUserAlreadyExists
	}

	am.Users[userName] = password

	return nil
}

func (am *DbManager) ChangeUserPassword(_ context.Context, userName, password string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, isExists := am.Users[userName]; !isExists {
		return fmt.Errorf("user doesn't exist")
	}

	am.Users[userName] = password

	return nil
}

func (am *DbManager) DropUser(_ context.Context, userName string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, isExists := am.Users[userName]; !isExists {
		return fmt.Errorf("user doesn't exist")
	}

	delete(am.Users, userName)

	return nil
}

func (am *DbManager) CreateDatabase(_ context.Context, database string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, isExists := am.Databases[database]; isExists {
		return customdatabase.ErrDatabaseAlreadyExists
	}

	am.Databases[database] = struct{}{}

	return nil
}

func (am *DbManager) DropDatabase(_ context.Context, database string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, isExists := am.Databases[database]; !isExists {
		return fmt.Errorf("database doesn't exist")
	}

	delete(am.Databases, database)

	return nil
}

func (am *DbManager) GrantUserToDatabase(_ context.Context, userName, database string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.User2Database[userName] = append(am.User2Database[userName], database)
	return nil
}
