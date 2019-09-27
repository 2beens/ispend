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
		Port     string
		Name     string
		User     string
		Password string
		SSLMode  string
	} `yaml:"postgres_production"`
	DBDev struct {
		Host     string
		Port     string
		Name     string
		User     string
		Password string
		SSLMode  string
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
