package customdatabase

import "fmt"

var (
	ErrDatabaseAlreadyExists = fmt.Errorf("database already exists")
	ErrUserAlreadyExists     = fmt.Errorf("user already exists")
)

type CreatedDatabaseInfo struct {
	Host
	Database
}

type Host struct {
	DSN  string
	Port int
}

type Database struct {
	Name     string
	User     string
	Password string
}

type DomainService struct {
	dbServerHost string
	dbServerPort int
}

func NewDomainService(host string, port int) (*DomainService, error) {
	if host == "" {
		return nil, fmt.Errorf("host should be not empty")
	}

	if port == 0 {
		return nil, fmt.Errorf("port should be not empty")
	}

	return &DomainService{
		dbServerHost: host,
		dbServerPort: port,
	}, nil
}

func (ds *DomainService) CreateDatabaseCreds(name string) CreatedDatabaseInfo {
	return CreatedDatabaseInfo{
		Host: Host{
			DSN:  ds.dbServerHost,
			Port: ds.dbServerPort,
		},
		Database: Database{
			Name: name,
			User: name,
			// todo create random password
			Password: name + name + name,
		},
	}
}
