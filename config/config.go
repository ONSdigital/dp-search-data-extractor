package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

// KafkaTLSProtocol informs service to use TLS protocol for kafka
const KafkaTLSProtocol = "TLS"

// Config represents service configuration for dp-search-data-extractor
type Config struct {
	BindAddr                          string        `envconfig:"BIND_ADDR"`
	EnableTopicTagging                bool          `envconfig:"TOPIC_TAGGING_ENABLED"`
	TopicCacheUpdateInterval          time.Duration `envconfig:"TOPIC_CACHE_UPDATE_INTERVAL"`
	TopicAPIURL                       string        `envconfig:"TOPIC_API_URL"`
	GracefulShutdownTimeout           time.Duration `envconfig:"GRACEFUL_SHUTDOWN_TIMEOUT"`
	HealthCheckInterval               time.Duration `envconfig:"HEALTHCHECK_INTERVAL"`
	HealthCheckCriticalTimeout        time.Duration `envconfig:"HEALTHCHECK_CRITICAL_TIMEOUT"`
	EnableZebedeeCallbacks            bool          `envconfig:"ENABLE_ZEBEDEE_CALLBACKS"`
	EnableDatasetAPICallbacks         bool          `envconfig:"ENABLE_DATASET_API_CALLBACKS"`
	EnableSearchContentUpdatedHandler bool          `envconfig:"ENABLE_SEARCH_CONTENT_UPDATED_HANDLER"`
	ZebedeeURL                        string        `envconfig:"ZEBEDEE_URL"`
	KeywordsLimit                     int           `envconfig:"KEYWORDS_LIMITS"`
	DatasetAPIURL                     string        `envconfig:"DATASET_API_URL"`
	ServiceAuthToken                  string        `envconfig:"SERVICE_AUTH_TOKEN"            json:"-"`
	StopConsumingOnUnhealthy          bool          `envconfig:"STOP_CONSUMING_ON_UNHEALTHY"`
	Kafka                             *Kafka
}

// Kafka contains the config required to connect to Kafka
type Kafka struct {
	ContentUpdatedGroup       string   `envconfig:"KAFKA_CONTENT_UPDATED_GROUP"`
	ContentUpdatedTopic       string   `envconfig:"KAFKA_CONTENT_UPDATED_TOPIC"`
	SearchContentTopic        string   `envconfig:"KAFKA_SEARCH_CONTENT_UPDATED_TOPIC"`
	SearchDataImportedTopic   string   `envconfig:"KAFKA_SEARCH_IMPORT_PRODUCER_TOPIC"`
	SearchContentDeletedTopic string   `envconfig:"KAFKA_SEARCH_DELETED_PRODUCER_TOPIC"`
	Addr                      []string `envconfig:"KAFKA_ADDR"`
	Version                   string   `envconfig:"KAFKA_VERSION"`
	OffsetOldest              bool     `envconfig:"KAFKA_OFFSET_OLDEST"`
	NumWorkers                int      `envconfig:"KAFKA_NUM_WORKERS"`
	SecProtocol               string   `envconfig:"KAFKA_SEC_PROTO"`
	SecCACerts                string   `envconfig:"KAFKA_SEC_CA_CERTS"            json:"-"`
	SecClientCert             string   `envconfig:"KAFKA_SEC_CLIENT_CERT"         json:"-"`
	SecClientKey              string   `envconfig:"KAFKA_SEC_CLIENT_KEY"          json:"-"`
	SecSkipVerify             bool     `envconfig:"KAFKA_SEC_SKIP_VERIFY"`
	MaxBytes                  int      `envconfig:"KAFKA_MAX_BYTES"`
	ConsumerMinBrokersHealthy int      `envconfig:"KAFKA_CONSUMER_MIN_BROKERS_HEALTHY"`
	ProducerMinBrokersHealthy int      `envconfig:"KAFKA_PRODUCER_MIN_BROKERS_HEALTHY"`
}

var cfg *Config

// Get returns the default config with any modifications through environment
// variables
func Get() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{
		BindAddr:                          "localhost:25800",
		EnableTopicTagging:                false,
		TopicCacheUpdateInterval:          30 * time.Minute,
		GracefulShutdownTimeout:           5 * time.Second,
		HealthCheckInterval:               30 * time.Second,
		HealthCheckCriticalTimeout:        90 * time.Second,
		EnableZebedeeCallbacks:            false,
		ZebedeeURL:                        "http://localhost:8082",
		KeywordsLimit:                     -1,
		TopicAPIURL:                       "http://localhost:25300",
		EnableDatasetAPICallbacks:         false,
		DatasetAPIURL:                     "http://localhost:22000",
		EnableSearchContentUpdatedHandler: false,
		ServiceAuthToken:                  "",
		StopConsumingOnUnhealthy:          true,
		Kafka: &Kafka{
			ContentUpdatedGroup:       "dp-search-data-extractor",
			ContentUpdatedTopic:       "content-updated",
			SearchContentTopic:        "search-content-updated",
			SearchDataImportedTopic:   "search-data-import",
			SearchContentDeletedTopic: "search-content-deleted",
			Addr:                      []string{"localhost:9092", "localhost:9093", "localhost:9094"},
			Version:                   "1.0.2",
			OffsetOldest:              true,
			NumWorkers:                1,
			SecProtocol:               "",
			SecCACerts:                "",
			SecClientCert:             "",
			SecClientKey:              "",
			SecSkipVerify:             false,
			MaxBytes:                  2000000,
			ConsumerMinBrokersHealthy: 1,
			ProducerMinBrokersHealthy: 1,
		},
	}

	return cfg, envconfig.Process("", cfg)
}
