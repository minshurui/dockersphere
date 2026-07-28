package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	Docker DockerConfig `mapstructure:"docker"`
	App    AppConfig    `mapstructure:"app"`
	Task   TaskConfig   `mapstructure:"task"`
	Audit  AuditConfig  `mapstructure:"audit"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DockerConfig struct {
	Host string `mapstructure:"host"`
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Version string `mapstructure:"version"`
}

type TaskConfig struct {
	WorkerPoolSize int    `mapstructure:"worker_pool_size"`
	Repository     string `mapstructure:"repository"`
}

type AuditConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	DBPath  string `mapstructure:"db_path"`
}

func Load(cfgFile string) (*Config, error) {
	v := viper.New()

	v.SetConfigType("yaml")
	if cfgFile != "" {
		v.SetConfigFile(cfgFile)
	} else {
		v.SetConfigName("config")
		v.AddConfigPath("./configs")
		v.AddConfigPath(".")
	}

	// Environment variable overrides
	v.SetEnvPrefix("DOCKERSPHERE")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("docker.host", "unix:///var/run/docker.sock")
	v.SetDefault("app.name", "DockerSphere")
	v.SetDefault("app.version", "0.3.0")
	v.SetDefault("task.worker_pool_size", 4)
	v.SetDefault("task.repository", "memory")
	v.SetDefault("audit.enabled", true)
	v.SetDefault("audit.db_path", "data/audit.db")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
