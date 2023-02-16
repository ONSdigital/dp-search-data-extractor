package service

import (
	"context"
	"errors"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/dp-search-data-extractor/clients"
	"github.com/ONSdigital/dp-search-data-extractor/config"
)

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	HealthCheck   bool
	KafkaConsumer bool
	KafkaProducer bool
	Init          Initialiser
	ZebedeeClient bool
	DatasetClient bool
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
		HealthCheck:   false,
		KafkaConsumer: false,
		KafkaProducer: false,
		Init:          initialiser,
		ZebedeeClient: false,
		DatasetClient: false,
	}
}

// Init implements the Initialiser interface to initialise dependencies
type Init struct{}

// GetHTTPServer creates an http server and sets the Server flag to true
func (e *ExternalServiceList) GetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := e.Init.DoGetHTTPServer(bindAddr, router)
	return s
}

// GetHealthCheck creates a healthcheck with versionInfo and sets teh HealthCheck flag to true
func (e *ExternalServiceList) GetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	hc, err := e.Init.DoGetHealthCheck(cfg, buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	e.HealthCheck = true
	return hc, nil
}

// DoGetHTTPServer creates an HTTP Server with the provided bind address and router
func (e *Init) DoGetHTTPServer(bindAddr string, router http.Handler) HTTPServer {
	s := dphttp.NewServer(bindAddr, router)
	s.HandleOSSignals = false
	return s
}

// DoGetHealthCheck creates a healthcheck with versionInfo
func (e *Init) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return nil, err
	}
	hc := healthcheck.New(versionInfo, cfg.HealthCheckCriticalTimeout, cfg.HealthCheckInterval)
	return &hc, nil
}

// GetZebedee return zebedee client
func (e *ExternalServiceList) GetZebedee(cfg *config.Config) clients.ZebedeeClient {
	zebedeeClient := e.Init.DoGetZebedeeClient(cfg)
	e.ZebedeeClient = true
	return zebedeeClient
}

// DoGetZebedeeClient gets and initialises the Zebedee Client
func (e *Init) DoGetZebedeeClient(cfg *config.Config) clients.ZebedeeClient {
	zebedeeClient := zebedee.New(cfg.ZebedeeURL)
	return zebedeeClient
}

// GetDatasetClient return DatasetAPI client
func (e *ExternalServiceList) GetDatasetClient(cfg *config.Config) clients.DatasetClient {
	datasetClient := e.Init.DoGetDatasetClient(cfg)
	e.DatasetClient = true
	return datasetClient
}

// DoGetZebedeeClient gets and initialises the Zebedee Client
func (e *Init) DoGetDatasetClient(cfg *config.Config) clients.DatasetClient {
	datasetClient := dataset.NewAPIClient(cfg.DatasetAPIURL)
	return datasetClient
}

// GetKafkaConsumer creates a Kafka consumer and sets the consumer flag to true
func (e *ExternalServiceList) GetKafkaConsumer(ctx context.Context, cfg *config.Config) (kafka.IConsumerGroup, error) {
	consumer, err := e.Init.DoGetKafkaConsumer(ctx, cfg.Kafka)
	if err != nil {
		return nil, err
	}
	e.KafkaConsumer = true
	return consumer, nil
}

// DoGetKafkaConsumer returns a Kafka Consumer group
func (e *Init) DoGetKafkaConsumer(ctx context.Context, cfg *config.Kafka) (kafka.IConsumerGroup, error) {
	if cfg == nil {
		return nil, errors.New("cannot create a kafka consumer without kafka config")
	}
	kafkaOffset := kafka.OffsetNewest
	if cfg.OffsetOldest {
		kafkaOffset = kafka.OffsetOldest
	}
	cgConfig := &kafka.ConsumerGroupConfig{
		BrokerAddrs:       cfg.Addr,
		Topic:             cfg.ContentUpdatedTopic,
		GroupName:         cfg.ContentUpdatedGroup,
		MinBrokersHealthy: &cfg.ConsumerMinBrokersHealthy,
		KafkaVersion:      &cfg.Version,
		Offset:            &kafkaOffset,
	}
	if cfg.SecProtocol == config.KafkaTLSProtocolFlag {
		cgConfig.SecurityConfig = kafka.GetSecurityConfig(
			cfg.SecCACerts,
			cfg.SecClientCert,
			cfg.SecClientKey,
			cfg.SecSkipVerify,
		)
	}
	return kafka.NewConsumerGroup(ctx, cgConfig)
}

// GetKafkaProducer creates a Kafka producer and sets the producder flag to true
func (e *ExternalServiceList) GetKafkaProducer(ctx context.Context, cfg *config.Config) (kafka.IProducer, error) {
	producer, err := e.Init.DoGetKafkaProducer(ctx, cfg.Kafka)
	if err != nil {
		return nil, err
	}
	e.KafkaProducer = true
	return producer, nil
}

func (e *Init) DoGetKafkaProducer(ctx context.Context, cfg *config.Kafka) (kafka.IProducer, error) {
	if cfg == nil {
		return nil, errors.New("cannot create a kafka producer without kafka config")
	}
	pConfig := &kafka.ProducerConfig{
		BrokerAddrs:       cfg.Addr,
		Topic:             cfg.ProducerTopic,
		MinBrokersHealthy: &cfg.ProducerMinBrokersHealthy,
		KafkaVersion:      &cfg.Version,
		MaxMessageBytes:   &cfg.MaxBytes,
	}
	if cfg.SecProtocol == config.KafkaTLSProtocolFlag {
		pConfig.SecurityConfig = kafka.GetSecurityConfig(
			cfg.SecCACerts,
			cfg.SecClientCert,
			cfg.SecClientKey,
			cfg.SecSkipVerify,
		)
	}
	return kafka.NewProducer(ctx, pConfig)
}
