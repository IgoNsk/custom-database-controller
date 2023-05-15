package customdatabase

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
