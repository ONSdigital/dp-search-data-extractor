package service

import (
	"context"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	dpkafka "github.com/ONSdigital/dp-kafka/v2"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/dp-search-data-extractor/clients"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/log.go/v2/log"
)

// KafkaTLSProtocolFlag informs service to use TLS protocol for kafka
const KafkaTLSProtocolFlag = "TLS"

// ExternalServiceList holds the initialiser and initialisation state of external services.
type ExternalServiceList struct {
	HealthCheck   bool
	KafkaConsumer bool
	KafkaProducer bool
	Init          Initialiser
	ZebedeeClient bool
}

// NewServiceList creates a new service list with the provided initialiser
func NewServiceList(initialiser Initialiser) *ExternalServiceList {
	return &ExternalServiceList{
		HealthCheck:   false,
		KafkaConsumer: false,
		KafkaProducer: false,
		Init:          initialiser,
		ZebedeeClient: false,
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

// GetKafkaConsumer creates a Kafka consumer and sets the consumer flag to true
func (e *ExternalServiceList) GetKafkaConsumer(ctx context.Context, cfg *config.Config) (dpkafka.IConsumerGroup, error) {
	consumer, err := e.Init.DoGetKafkaConsumer(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.KafkaConsumer = true
	return consumer, nil
}

// DoGetKafkaConsumer returns a Kafka Consumer group
func (e *Init) DoGetKafkaConsumer(ctx context.Context, cfg *config.Config) (dpkafka.IConsumerGroup, error) {
	cgChannels := dpkafka.CreateConsumerGroupChannels(1)

	kafkaOffset := dpkafka.OffsetNewest
	if cfg.KafkaOffsetOldest {
		kafkaOffset = dpkafka.OffsetOldest
	}
	cConfig := &dpkafka.ConsumerGroupConfig{
		KafkaVersion: &cfg.KafkaVersion,
		Offset:       &kafkaOffset,
	}
	if cfg.KafkaSecProtocol == KafkaTLSProtocolFlag {
		cConfig.SecurityConfig = dpkafka.GetSecurityConfig(
			cfg.KafkaSecCACerts,
			cfg.KafkaSecClientCert,
			cfg.KafkaSecClientKey,
			cfg.KafkaSecSkipVerify,
		)
	}

	kafkaConsumer, err := dpkafka.NewConsumerGroup(
		ctx,
		cfg.KafkaAddr,
		cfg.ContentPublishedTopic,
		cfg.ContentPublishedGroup,
		cgChannels,
		cConfig,
	)
	if err != nil {
		return nil, err
	}

	return kafkaConsumer, nil
}

// GetKafkaProducer creates a Kafka producer and sets the producder flag to true
func (e *ExternalServiceList) GetKafkaProducer(ctx context.Context, cfg *config.Config) (dpkafka.IProducer, error) {
	producer, err := e.Init.DoGetKafkaProducer(ctx, cfg)
	if err != nil {
		return nil, err
	}
	e.KafkaProducer = true
	return producer, nil
}

func (e *Init) DoGetKafkaProducer(ctx context.Context, cfg *config.Config) (dpkafka.IProducer, error) {
	pChannels := dpkafka.CreateProducerChannels()
	pConfig := &dpkafka.ProducerConfig{
		KafkaVersion: &cfg.KafkaVersion,
	}

	if cfg.KafkaSecProtocol == KafkaTLSProtocolFlag {
		pConfig.SecurityConfig = dpkafka.GetSecurityConfig(
			cfg.KafkaSecCACerts,
			cfg.KafkaSecClientCert,
			cfg.KafkaSecClientKey,
			cfg.KafkaSecSkipVerify,
		)
	}
	producer, err := dpkafka.NewProducer(ctx, cfg.KafkaAddr, cfg.KafkaProducerTopic, pChannels, pConfig)
	if err != nil {
		log.Fatal(ctx, "kafka producer returned an error", err, log.Data{"topic": cfg.KafkaProducerTopic})
		return nil, err
	}

	return producer, nil
}
