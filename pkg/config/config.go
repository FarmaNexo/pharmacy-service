// pkg/config/config.go
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config representa toda la configuración del servicio
type Config struct {
	Environment string         `mapstructure:"environment"`
	Server      ServerConfig   `mapstructure:"server"`
	Database    DatabaseConfig `mapstructure:"database"`
	JWT         JWTConfig      `mapstructure:"jwt"`
	Redis       RedisConfig    `mapstructure:"redis"`
	AWS         AWSConfig      `mapstructure:"aws"`
	SQS         SQSConfig      `mapstructure:"sqs"`
	S3          S3Config       `mapstructure:"s3"`
	Services    ServicesConfig `mapstructure:"services"`
	Log         LogConfig      `mapstructure:"log"`
}

// ServicesConfig — URLs HTTP a otros microservicios.
type ServicesConfig struct {
	CatalogService CatalogServiceConfig `mapstructure:"catalog_service"`
}

// CatalogServiceConfig — URL del catalog-service. Apunta al ALB interno.
// Usado para resolver lookups durante el procesamiento de INVENTORY_DISCOVERED.
type CatalogServiceConfig struct {
	BaseURL string `mapstructure:"base_url"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	Host         string        `mapstructure:"host"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"db_name"`
	Schema          string        `mapstructure:"schema"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type JWTConfig struct {
	Secret              string        `mapstructure:"secret"`
	AccessTokenDuration time.Duration `mapstructure:"access_token_duration"`
	Issuer              string        `mapstructure:"issuer"`
}

type RedisConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	Password   string `mapstructure:"password"`
	DB         int    `mapstructure:"db"`
	MaxRetries int    `mapstructure:"max_retries"`
	PoolSize   int    `mapstructure:"pool_size"`
}

func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

type AWSConfig struct {
	Region   string `mapstructure:"region"`
	Endpoint string `mapstructure:"endpoint"`
}

type SQSConfig struct {
	// PharmacyEventsQueueURL — eventos OUTBOUND publicados por pharmacy-service.
	PharmacyEventsQueueURL string `mapstructure:"pharmacy_events_queue_url"`
	// CatalogEventsQueueURL — eventos publicados por catalog-service que
	// pharmacy podría consumir (no usado actualmente, mantenido por compat).
	CatalogEventsQueueURL string `mapstructure:"catalog_events_queue_url"`
	// ScraperEventsQueueURL — INBOUND: PHARMACY_DISCOVERED + INVENTORY_DISCOVERED
	// del scraper (Tier 5).
	ScraperEventsQueueURL string `mapstructure:"scraper_events_queue_url"`
}

type S3Config struct {
	PharmaciesBucket string `mapstructure:"pharmacies_bucket"`
}

type LogConfig struct {
	Level    string `mapstructure:"level"`
	Encoding string `mapstructure:"encoding"`
}

func LoadConfig(environment string) (*Config, error) {
	v := viper.New()

	v.SetConfigType("yaml")

	// Buscar el archivo config en múltiples rutas
	configFile := findConfigFile(environment)
	if configFile == "" {
		return nil, fmt.Errorf("config file config.%s.yaml not found", environment)
	}

	// Leer el archivo YAML como string
	content, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading config file %s: %w", configFile, err)
	}

	// Expandir variables de entorno ${VAR} en el YAML
	expanded := os.ExpandEnv(string(content))

	// Pasar el YAML expandido a Viper
	if err := v.ReadConfig(strings.NewReader(expanded)); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	config.Environment = environment

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

func findConfigFile(environment string) string {
	filename := fmt.Sprintf("config.%s.yaml", environment)
	paths := []string{"./configs", "../configs", "../../configs"}

	for _, dir := range paths {
		path := fmt.Sprintf("%s/%s", dir, filename)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

func validateConfig(config *Config) error {
	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}
	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if config.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}
	if config.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}
	if len(config.JWT.Secret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters")
	}
	return nil
}

func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode,
	)
}

func (c *Config) IsProduction() bool  { return c.Environment == "production" }
func (c *Config) IsLocal() bool       { return c.Environment == "local" }
func (c *Config) IsDevelopment() bool  { return c.Environment == "development" }
func (c *Config) IsQA() bool          { return c.Environment == "qa" }
func (c *Config) IsUAT() bool         { return c.Environment == "uat" }
