package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type StorageConfig struct {
	Host          string `yaml:"host" env-required:"true"`
	Port          int    `yaml:"port" env-required:"true"`
	User          string `yaml:"user" env-required:"true"`
	Password      string `yaml:"pass" env-required:"true"`
	DB            string `yaml:"db" env-required:"true"`
	ShouldMigrate bool   `yaml:"should-migrate" env-default:"false"`
	Secure        bool   `yaml:"tls" env-default:"false"`
}

type CRUDConfig struct {
	StorageConfig    `yaml:"storage" env-required:"true"`
	HTTPServerConfig `yaml:"http"`
}

type HTTPServerConfig struct {
	Address string `yaml:"address" env-default:"localhost:8080"`
	Swagger SwaggerConfig
}

type SwaggerConfig struct {
	UIPath   string `yaml:"address" env-default:"swagger/index.html"`
	SpecPath string `yaml:"address" env-default:"swagger/swagger.yml"`
}

func MustLoadCRUDConfig() *CRUDConfig {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg CRUDConfig
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
