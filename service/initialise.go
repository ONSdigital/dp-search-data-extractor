package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ONSdigital/dp-api-clients-go/v2/dataset"
	"github.com/ONSdigital/dp-api-clients-go/v2/zebedee"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	dphttp "github.com/ONSdigital/dp-net/v3/http"
	"github.com/ONSdigital/dp-search-data-extractor/clients"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	topicCli "github.com/ONSdigital/dp-topic-api/sdk"
	"github.com/ONSdigital/log.go/v2/log"
)

// GetHTTPServer creates an HTTP Server with the provided bind address and router
var GetHTTPServer = func(bindAddr string, router http.Handler) HTTPServer {
	s := dphttp.NewServer(bindAddr, router)
	s.HandleOSSignals = false
	return s
}

// GetHealthCheck creates a healthcheck with versionInfo
var GetHealthCheck = func(cfg *config.Config, buildTime, gitCommit, version string) (HealthChecker, error) {
	versionInfo, err := healthcheck.NewVersionInfo(buildTime, gitCommit, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get version info: %w", err)
	}
	hc := healthcheck.New(
		versionInfo,
		cfg.HealthCheckCriticalTimeout,
		cfg.HealthCheckInterval,
	)
	return &hc, nil
}

// GetZebedee gets the Zebedee Client
var GetZebedee = func(cfg *config.Config) clients.ZebedeeClient {
	if !cfg.EnableZebedeeCallbacks {
		log.Info(context.Background(), "returning nil zebedee client as callbacks are disabled")
		return nil
	}
	return zebedee.New(cfg.ZebedeeURL)
}

// GetDatasetClient gets the Dataset API client
var GetDatasetClient = func(cfg *config.Config) clients.DatasetClient {
	if !cfg.EnableDatasetAPICallbacks {
		log.Info(context.Background(), "returning nil Dataset client as Dataset API callbacks are disabled")
		return nil
	}
	return dataset.NewAPIClient(cfg.DatasetAPIURL)
}

// GetTopicClient gets the Topic API client
var GetTopicClient = func(cfg *config.Config) topicCli.Clienter {
	return topicCli.New(cfg.TopicAPIURL)
}

// GetKafkaConsumer returns a Kafka Consumer group
var GetKafkaConsumer = func(ctx context.Context, cfg *config.Kafka, topic string) (kafka.IConsumerGroup, error) {
	if cfg == nil {
		return nil, errors.New("cannot create a kafka consumer without kafka config")
	}
	kafkaOffset := kafka.OffsetNewest
	if cfg.OffsetOldest {
		kafkaOffset = kafka.OffsetOldest
	}
	cgConfig := &kafka.ConsumerGroupConfig{
		BrokerAddrs:       cfg.Addr,
		Topic:             topic,
		GroupName:         cfg.ContentUpdatedGroup,
		MinBrokersHealthy: &cfg.ConsumerMinBrokersHealthy,
		KafkaVersion:      &cfg.Version,
		Offset:            &kafkaOffset,
	}
	if cfg.SecProtocol == config.KafkaTLSProtocol {
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
var GetKafkaProducer = func(ctx context.Context, cfg *config.Kafka) (kafka.IProducer, error) {
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
	if cfg.SecProtocol == config.KafkaTLSProtocol {
		pConfig.SecurityConfig = kafka.GetSecurityConfig(
			cfg.SecCACerts,
			cfg.SecClientCert,
			cfg.SecClientKey,
			cfg.SecSkipVerify,
		)
	}
	return kafka.NewProducer(ctx, pConfig)
}
