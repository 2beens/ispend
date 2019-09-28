package ispend

import (
	"gopkg.in/yaml.v2"
)

const DBTypePostgres = "postgres"
const DBTypeInMemory = "mem"
const PostgresProduction = "production"
const PostgresDev = "dev"

type YamlConfig struct {
	DBType      string `yaml:"dbtype"`
	PostgresEnv string `yaml:"postgres_env"`
	DBProd      struct {
		Host     string
		Port     int
		Name     string
		User     string
		Password string
		SSLMode  string `yaml:"sslMode"`
	} `yaml:"postgres_production"`
	DBDev struct {
		Host     string
		Port     int
		Name     string
		User     string
		Password string
		SSLMode  string `yaml:"sslMode"`
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

func (c *YamlConfig) GetPostgresDBPassword() string {
	if c.PostgresEnv == PostgresProduction {
		return c.DBProd.Password
	}
	return c.DBDev.Password
}

func (c *YamlConfig) GetPostgresDBSSLMode() string {
	if c.PostgresEnv == PostgresProduction {
		return c.DBProd.SSLMode
	}
	return c.DBDev.SSLMode
}