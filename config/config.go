package config

// Config : Struct holding configuration for the daemon
type Config struct {
	DeploymentType string `default:"production"`

	FileDir string `required:"true"`

	PostgresHost       string
	PostgresUser       string
	PostgresPassword   string
	PostgresPort       int64
	PostgresPortEnv    string
	PostgresDb         string
	PostgresRequireSSL string `default:"require"`
}
