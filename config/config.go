package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config represents service configuration for dp-search-data-extractor
type Config struct {
	BindAddr                   string        `envconfig:"BIND_ADDR"`
	GracefulShutdownTimeout    time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval        time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	KafkaAddr                  []string      `envconfig:"KAFKA_ADDR"                     json:"-"`
	KafkaVersion               string        `envconfig:"KAFKA_VERSION"`
	KafkaOffsetOldest          bool          `envconfig:"KAFKA_OFFSET_OLDEST"`
	KafkaNumWorkers            int           `envconfig:"KAFKA_NUM_WORKERS"`
	KafkaSecProtocol           string        `envconfig:"KAFKA_SEC_PROTO"`
	KafkaSecCACerts            string        `envconfig:"KAFKA_SEC_CA_CERTS"`
	KafkaSecClientCert         string        `envconfig:"KAFKA_SEC_CLIENT_CERT"`
	KafkaSecClientKey          string        `envconfig:"KAFKA_SEC_CLIENT_KEY"          json:"-"`
	KafkaSecSkipVerify         bool          `envconfig:"KAFKA_SEC_SKIP_VERIFY"`	
	ContentPublishedGroup      string        `envconfig:"KAFKA_CONTENT_PUBLISHED_GROUP"`
	ContentPublishedTopic      string        `envconfig:"KAFKA_CONTENT_PUBLISHED_TOPIC"`
	OutputFilePath             string        `envconfig:"OUTPUT_FILE_PATH"`
	KafkaProducerTopic         string        `envconfig:"KAFKA_PRODUCER_TOPIC"`
	ZebedeeURL                 string        `envconfig:"ZEBEDEE_URL"`
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                   "localhost:25800",
		GracefulShutdownTimeout:    5 * time.Second,
		HealthCheckInterval:        30 * time.Second,
		HealthCheckCriticalTimeout: 90 * time.Second,
		KafkaAddr:                  []string{"localhost:9092"},
		KafkaVersion:               "1.0.2",
		KafkaOffsetOldest:          true,
		KafkaNumWorkers:            1,
		ContentPublishedGroup:      "dp-search-data-extractor",
		ContentPublishedTopic:      "content-published",
		OutputFilePath:             "/tmp/dpSearchDataExtractor.txt",
		KafkaProducerTopic:         "search-data-import",
		ZebedeeURL:                 "http://localhost:8082",
	}

	return cfg, envconfig.Process("", cfg)
}
