package steps

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	componenttest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/dp-search-data-extractor/config"
	"github.com/ONSdigital/dp-search-data-extractor/service"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/maxcnunes/httpfake"
)

const (
	WaitEventTimeout = 5 * time.Second // maximum time that the component test consumer will wait for a kafka event
)

var (
	BuildTime = "1625046891"
	GitCommit = "7434fe334d9f51b7239f978094ea29d10ac33b16"
	Version   = ""
)

type Component struct {
	componenttest.ErrorFeature
	DatasetAPI       *httpfake.HTTPFake // Dataset API mock at HTTP level
	Zebedee          *httpfake.HTTPFake // Zebedee mock at HTTP level
	errorChan        chan error
	svc              *service.Service
	cfg              *config.Config
	wg               *sync.WaitGroup
	signals          chan os.Signal
	waitEventTimeout time.Duration
	testETag         string
	ctx              context.Context
	kakfkaScenario   *componenttest.KafkaScenario
}

func NewComponent(kafkaScenario *componenttest.KafkaScenario) *Component {
	c := &Component{
		DatasetAPI:       httpfake.New(),
		Zebedee:          httpfake.New(),
		errorChan:        make(chan error),
		waitEventTimeout: WaitEventTimeout,
		wg:               &sync.WaitGroup{},
		testETag:         "13c7791bafdbaaf5e6660754feb1a58cd6aaa892",
		ctx:              context.Background(),
		kakfkaScenario:   kafkaScenario,
	}
	return c
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

	cfg.HealthCheckInterval = time.Second
	cfg.DatasetAPIURL = c.DatasetAPI.ResolveURL("")
	cfg.EnableDatasetAPICallbacks = true
	cfg.ZebedeeURL = c.Zebedee.ResolveURL("")
	cfg.EnableZebedeeCallbacks = true
	cfg.EnableSearchContentUpdatedHandler = true

	cfg.Kafka.ContentUpdatedTopic = c.kakfkaScenario.GetMappedTopic("content-updated")
	cfg.Kafka.SearchContentTopic = c.kakfkaScenario.GetMappedTopic("search-content-updated")
	cfg.Kafka.SearchDataImportedTopic = c.kakfkaScenario.GetMappedTopic("search-data-import")
	cfg.Kafka.SearchContentDeletedTopic = c.kakfkaScenario.GetMappedTopic("search-content-deleted")
	cfg.Kafka.Addr = c.kakfkaScenario.KafkaFeature.GetBrokers(ctx)

	log.Info(ctx, "config used by component tests", log.Data{"cfg": cfg})

	// Create service and initialise it
	c.svc = service.New()
	if err = c.svc.Init(ctx, cfg, BuildTime, GitCommit, Version); err != nil {
		return fmt.Errorf("unexpected service Init error in NewComponent: %w", err)
	}

	c.cfg = cfg

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

// Close kills the application under test and waits for it to complete the graceful shutdown, or timeout
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
