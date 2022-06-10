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
	KafkaAddr                  []string      `envconfig:"KAFKA_ADDR"`
	KafkaVersion               string        `envconfig:"KAFKA_VERSION"`
	KafkaOffsetOldest          bool          `envconfig:"KAFKA_OFFSET_OLDEST"`
	KafkaNumWorkers            int           `envconfig:"KAFKA_NUM_WORKERS"`
	KafkaSecProtocol           string        `envconfig:"KAFKA_SEC_PROTO"`
	KafkaSecCACerts            string        `envconfig:"KAFKA_SEC_CA_CERTS"`
	KafkaSecClientCert         string        `envconfig:"KAFKA_SEC_CLIENT_CERT"`
	KafkaSecClientKey          string        `envconfig:"KAFKA_SEC_CLIENT_KEY"          json:"-"`
	KafkaSecSkipVerify         bool          `envconfig:"KAFKA_SEC_SKIP_VERIFY"`
	ContentUpdatedGroup        string        `envconfig:"KAFKA_CONTENT_UPDATED_GROUP"`
	ContentUpdatedTopic        string        `envconfig:"KAFKA_CONTENT_UPDATED_TOPIC"`
	KafkaProducerTopic         string        `envconfig:"KAFKA_PRODUCER_TOPIC"`
	ZebedeeURL                 string        `envconfig:"ZEBEDEE_URL"`
	KeywordsLimit              int           `envconfig:"KEYWORDS_LIMITS"`
	DatasetAPIURL              string        `envconfig:"DATASET_API_URL"`
	ServiceAuthToken           string        `envconfig:"SERVICE_AUTH_TOKEN"`
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
		KafkaAddr:                  []string{"localhost:9092", "localhost:9093", "localhost:9094"},
		KafkaVersion:               "1.0.2",
		KafkaSecProtocol:           "",
		KafkaSecCACerts:            "",
		KafkaSecClientCert:         "",
		KafkaSecClientKey:          "",
		KafkaSecSkipVerify:         false,
		KafkaOffsetOldest:          true,
		KafkaNumWorkers:            1,
		ContentUpdatedGroup:        "dp-search-data-extractor",
		ContentUpdatedTopic:        "content-updated",
		KafkaProducerTopic:         "search-data-import",
		ZebedeeURL:                 "http://localhost:8082",
		KeywordsLimit:              -1,
		DatasetAPIURL:              "http://localhost:22000",
		ServiceAuthToken:           "",
	}

	return cfg, envconfig.Process("", cfg)
}
