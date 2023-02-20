package steps

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
	"time"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-healthcheck/healthcheck"
	kafka "github.com/ONSdigital/dp-kafka/v3"
	"github.com/ONSdigital/dp-kafka/v3/kafkatest"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/models"
	"github.com/ONSdigital/dp-search-data-extractor/service"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/maxcnunes/httpfake"
)

const (
	WaitEventTimeout = 15 * time.Second // maximum time that the component test will wait for a kafka event
)

var (
	BuildTime = "1625046891"
	GitCommit = "7434fe334d9f51b7239f978094ea29d10ac33b16"
	Version   = ""
)

type Component struct {
	componenttest.ErrorFeature
	DatasetAPI *httpfake.HTTPFake // Dataset API mock at HTTP level
	Zebedee    *httpfake.HTTPFake // Zebedee mock at HTTP level
	// KafkaConsumer    kafka.IConsumerGroup
	// KafkaProducer    kafka.IProducer
	// ZebedeeClient    clients.ZebedeeClient
	// DatasetClient    clients.DatasetClient
	inputData        models.ZebedeeData
	errorChan        chan error
	svc              *service.Service
	signals          chan os.Signal
	cfg              *config.Config
	wg               *sync.WaitGroup
	waitEventTimeout time.Duration
	testETag         string
	ctx              context.Context
}

func NewComponent(t *testing.T) *Component {
	service.GetKafkaConsumer = func(ctx context.Context, cfg *config.Kafka) (kafka.IConsumerGroup, error) {
		consumer := kafkatest.NewMessageConsumer(false)
		consumer.CheckerFunc = funcCheck
		return consumer, nil
	}

	service.GetKafkaProducer = func(ctx context.Context, cfg *config.Kafka) (kafka.IProducer, error) {
		producer := kafkatest.NewMessageProducer(false)
		producer.CheckerFunc = funcCheck
		return producer, nil
	}

	return &Component{
		DatasetAPI:       httpfake.New(),
		Zebedee:          httpfake.New(),
		errorChan:        make(chan error),
		waitEventTimeout: WaitEventTimeout,
		wg:               &sync.WaitGroup{},
		testETag:         "13c7791bafdbaaf5e6660754feb1a58cd6aaa892",
		ctx:              context.Background(),
	}
}

// initService initialises the server, the mocks and waits for the dependencies to be ready
func (c *Component) initService(ctx context.Context) error {
	// register interrupt signals
	c.signals = make(chan os.Signal, 1)
	signal.Notify(c.signals, syscall.SIGINT, syscall.SIGTERM)

	// Read config
	cfg, err := config.Get()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	cfg.DatasetAPIURL = c.DatasetAPI.ResolveURL("")
	cfg.ZebedeeURL = c.Zebedee.ResolveURL("")

	log.Info(ctx, "config used by component tests", log.Data{"cfg": cfg})

	// Create service and initialise it
	c.svc = service.New()
	if err = c.svc.Init(ctx, cfg, BuildTime, GitCommit, Version); err != nil {
		return fmt.Errorf("unexpected service Init error in NewComponent: %w", err)
	}

	c.cfg = cfg

	// wait for producer to be initialised and consumer to be in consuming state
	<-c.svc.Producer.Channels().Initialised
	log.Info(ctx, "service kafka producer initialised")
	c.svc.Consumer.StateWait(kafka.Consuming)
	log.Info(ctx, "service kafka consumer is in consuming state")

	return nil
}

func (c *Component) startService(ctx context.Context) error {
	if err := c.svc.Start(ctx, c.errorChan); err != nil {
		return fmt.Errorf("unexpected error while starting service: %w", err)
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		// blocks until an os interrupt or a fatal error occurs
		select {
		case err := <-c.errorChan:
			if errClose := c.svc.Close(ctx); errClose != nil {
				log.Warn(ctx, "error closing server during error handing", log.Data{"close_error": errClose})
			}
			panic(fmt.Errorf("unexpected error received from errorChan: %w", err))
		case sig := <-c.signals:
			log.Info(ctx, "os signal received", log.Data{"signal": sig})
		}

		if err := c.svc.Close(ctx); err != nil {
			panic(fmt.Errorf("unexpected error during service graceful shutdown: %w", err))
		}
	}()

	return nil
}

// Close kills the application under test.
func (c *Component) Close() {
	// kill application
	c.signals <- os.Interrupt

	// wait for graceful shutdown to finish (or timeout)
	c.wg.Wait()
}

// Reset re-initialises the service under test and the api mocks.
// Note that the service under test should not be started yet
// to prevent race conditions if it tries to call un-initialised dependencies (steps)
func (c *Component) Reset() error {
	if err := c.initService(c.ctx); err != nil {
		return fmt.Errorf("failed to initialise service: %w", err)
	}

	c.DatasetAPI.Reset()
	c.Zebedee.Reset()

	return nil
}

// 	consumer := kafkatest.NewMessageConsumer(false)
// 	consumer.CheckerFunc = funcCheck
// 	c.KafkaConsumer = consumer

// 	producer := kafkatest.NewMessageProducer(false)
// 	producer.CheckerFunc = funcCheck
// 	c.KafkaProducer = producer

// 	cfg, err := config.Get()
// 	if err != nil {
// 		return nil
// 	}

// 	c.cfg = cfg

// 	service.GetHealthCheck = c.GetHealthCheck
// 	service.GetHTTPServer = c.GetHTTPServer
// 	service.GetKafkaConsumer = c.GetConsumer
// 	service.GetKafkaProducer = c.GetProducer
// 	service.GetZebedee = c.GetZebedeeClient
// 	service.GetDatasetClient = c.GetDatasetClient

// 	return c
// }

// func (c *Component) GetHealthCheck(cfg *config.Config, buildTime, gitCommit, version string) (service.HealthChecker, error) {
// 	return &mock.HealthCheckerMock{
// 		AddAndGetCheckFunc: func(name string, checker healthcheck.Checker) (*healthcheck.Check, error) { return nil, nil },
// 		StartFunc:          func(ctx context.Context) {},
// 		StopFunc:           func() {},
// 	}, nil
// }

// func (c *Component) GetHTTPServer(bindAddr string, router http.Handler) service.HTTPServer {
// 	return dphttp.NewServer(bindAddr, router)
// }

// func (c *Component) GetConsumer(ctx context.Context, cfg *config.Kafka) (kafkaConsumer kafka.IConsumerGroup, err error) {
// 	return c.KafkaConsumer, nil
// }

// func (c *Component) GetProducer(ctx context.Context, cfg *config.Kafka) (kafkaConsumer kafka.IProducer, err error) {
// 	return c.KafkaProducer, nil
// }

// func (c *Component) GetZebedeeClient(cfg *config.Config) clients.ZebedeeClient {
// 	return c.zebedeeClient
// }

// func (c *Component) GetDatasetClient(cfg *config.Config) clients.DatasetClient {
// 	return c.datasetClient
// }

func funcCheck(ctx context.Context, state *healthcheck.CheckState) error {
	return nil
}
