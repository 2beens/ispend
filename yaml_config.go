package ispend

import (
	"gopkg.in/yaml.v2"
)

const DBTypePostgres = "postgres"
const DBTypeInMemory = "mem"
const PostgresProduction = "production"
const PostgresDev = "dev"

type YamlConfig struct {
	PingTimeout int    `yaml:"db_ping_timeout"`
	DBType      string `yaml:"dbtype"`
	// TODO: should have one general env variable
	PostgresEnv string `yaml:"postgres_env"`
	LogsPath    string `yaml:"logs_path"`

	Graphite struct {
		Enabled bool
		Host    string
		Port    int
	}

	DBProd struct {
		Host    string
		Port    int
		Name    string
		User    string
		SSLMode string `yaml:"sslMode"`
	} `yaml:"postgres_production"`
	DBDev struct {
		Host    string
		Port    int
		Name    string
		User    string
		SSLMode string `yaml:"sslMode"`
	} `yaml:"postgres_dev"`
}

func NewYamlConfig(yamlContent []byte) (*YamlConfig, error) {
	yc := &YamlConfig{}
	err := yaml.Unmarshal(yamlContent, yc)
	if err != nil {
		return nil, err
	}
	return yc, nil
}

func (c *YamlConfig) GetPostgresHost() string {
	if c.PostgresEnv == PostgresProduction {
		return c.DBProd.Host
	}
	return c.DBDev.Host
}

func (c *YamlConfig) GetPostgresPort() int {
	if c.PostgresEnv == PostgresProduction {
		return c.DBProd.Port
	}
	return c.DBDev.Port
}

func (c *YamlConfig) GetPostgresDBName() string {
	if c.PostgresEnv == PostgresProduction {
		return c.DBProd.Name
	}
	return c.DBDev.Name
}

func (c *YamlConfig) GetPostgresDBUsername() string {
	if c.PostgresEnv == PostgresProduction {
		return c.DBProd.User
	}
	return c.DBDev.User
}

func (c *YamlConfig) GetPostgresDBSSLMode() string {
	if c.PostgresEnv == PostgresProduction {
		return c.DBProd.SSLMode
	}
	return c.DBDev.SSLMode
}
