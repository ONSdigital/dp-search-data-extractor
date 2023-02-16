package steps

import (
	"context"
	"net/http"
	"time"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	dphttp "github.com/ONSdigital/dp-net/http"
	"github.com/ONSdigital/dp-search-data-extractor/clients"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/dp-search-data-extractor/service"
	"github.com/ONSdigital/dp-search-data-extractor/service/mock"
)

const (
	WaitEventTimeout = 15 * time.Second // maximum time that the component test will wait for a kafka event
)

type Component struct {
	ErrorFeature     componenttest.ErrorFeature
	inputData        models.ZebedeeData
	serviceList      *service.ExternalServiceList
	KafkaConsumer    kafka.IConsumerGroup
	KafkaProducer    kafka.IProducer
	zebedeeClient    clients.ZebedeeClient
	datasetClient    clients.DatasetClient
	errorChan        chan error
	svc              *service.Service
	cfg              *config.Config
	waitEventTimeout time.Duration
}

func NewComponent() *Component {
	c := &Component{
		errorChan:        make(chan error),
		waitEventTimeout: WaitEventTimeout,
	}

	consumer := kafkatest.NewMessageConsumer(false)
	consumer.CheckerFunc = funcCheck
	c.KafkaConsumer = consumer

	producer := kafkatest.NewMessageProducer(false)
	producer.CheckerFunc = funcCheck
	c.KafkaProducer = producer

	cfg, err := config.Get()
	if err != nil {
		return nil
	}

	c.cfg = cfg

	initMock := &mock.InitialiserMock{
		DoGetKafkaConsumerFunc: c.DoGetConsumer,
		DoGetKafkaProducerFunc: c.DoGetProducer,
		DoGetHealthCheckFunc:   c.DoGetHealthCheck,
		DoGetHTTPServerFunc:    c.DoGetHTTPServer,
		DoGetZebedeeClientFunc: c.DoGetZebedeeClient,
		DoGetDatasetClientFunc: c.DoGetDatasetClient,
	}

	c.serviceList = service.NewServiceList(initMock)

	return c
}

func (c *Component) DoGetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (service.HealthChecker, error) {
	return &mock.HealthCheckerMock{
		AddCheckFunc: func(name string, checker healthcheck.Checker) error { return nil },
		StartFunc:    func(ctx context.Context) {},
		StopFunc:     func() {},
	}, nil
}

func (c *Component) DoGetHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
	return dphttp.NewServer(bindAddr, router)
}

func (c *Component) DoGetConsumer(ctx context.Context, cfg *config.Kafka) (kafkaConsumer kafka.IConsumerGroup, err error) {
	return c.KafkaConsumer, nil
}

func (c *Component) DoGetProducer(ctx context.Context, cfg *config.Kafka) (kafkaConsumer kafka.IProducer, err error) {
	return c.KafkaProducer, nil
}

func (c *Component) DoGetZebedeeClient(cfg *config.Config) clients.ZebedeeClient {
	return c.zebedeeClient
}

func (c *Component) DoGetDatasetClient(cfg *config.Config) clients.DatasetClient {
	return c.datasetClient
}

func funcCheck(ctx context.Context, state *healthcheck.CheckState) error {
	return nil
}
